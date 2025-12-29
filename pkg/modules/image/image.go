/*
Copyright 2024 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package image

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/cockroachdb/errors"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The Image module provides functionality for managing container image operations in KubeKey.
It supports pulling images from remote registries, pushing images to private registries,
and copying images between different sources (local directories or remote registries).

This module uses the OCI (Open Container Initiative) distribution specification for
image transfer operations, providing compatibility with Docker registries and OCI registries.

New Configuration Format (recommended):
The module accepts configuration through the following parameters:

image:
  platform: []string           # optional: list of target platforms (e.g., ["linux/amd64", "linux/arm64"])
  manifests: []string          # required: list of image manifests to operate on
  pattern: string             # optional: regex pattern to match images
  auths:                      # optional: image registry authentication information
    - repo: string            # optional: target image registry
      username: string        # optional: username for authentication
      password: string        # optional: password for authentication
      skipTLSVerify: bool    # optional: skip TLS verification
      ca_cert: string        # optional: CA certificate path
  src: string                 # optional: source image reference
                               #   - "oci://{{ .module.image.reference }}/{{ .module.image.reference.repository }}:{{ .module.image.reference.reference }}" means pull from remote registry
                               #   - "local://{{ .module.image.localPath }}/{{ .module.image.reference.repository }}:{{ .module.image.reference.reference }}" means pull from local directory
  dest: string                # optional: destination image reference
                               #   - "local://{{ .module.image.localPath }}/{{ .module.image.reference.repository }}:{{ .module.image.reference.reference }}" means save to local directory
                               #   - "oci://{{ .image_registry.auth.registry }}{{ .module.image.reference.repository }}:{{ .module.image.reference.reference }}" means push to remote registry

Operation Types (determined by src and dest):
- src=oci://, dest=local://  -> pull image from remote registry to local directory
- src=local://, dest=oci:// -> push image from local directory to remote registry
- src=local://, dest=local://-> copy image from one local directory to another

Usage Examples in Playbook Tasks:
1. Pull images from registry:
   ```yaml
   - name: Pull container images
     image:
       manifests:
         - nginx:latest
         - prometheus:v2.45.0
       src: "oci://{{ .module.image.reference }}/{{ .module.image.reference.repository }}:{{ .module.image.reference.reference }}"
       dest: "local:///var/lib/kubekey/images"
       platform:
         - linux/amd64
         - linux/arm64
       auths:
         - repo: docker.io
           username: MyDockerAccount
           password: my_password
     register: pull_result
   ```

2. Push images to private registry:
   ```yaml
   - name: Push images to private registry
     image:
       src: "local:///var/lib/kubekey/images"
       dest: "oci://registry.example.com/{{ .module.image.reference.repository }}:{{ .module.image.reference.reference }}"
       pattern: ".*"
       auths:
         - repo: registry.example.com
           username: admin
           password: secret
     register: push_result
   ```

3. Copy image from local to local
   ```yaml
   - name: local to local copy
     image:
       manifests:
         - nginx:latest
         - prometheus:v2.45.0
       src: "local:///path/to/source/images"
       dest: "local:///path/to/dest/images"
   ```

Return Values:
- On success: Returns "Success" in stdout
- On failure: Returns error message in stderr
*/

const defaultRegistry = "docker.io"

const (
	// PolicyStrict: all requested platforms must exist in the image, otherwise return error.
	// This is the default policy.
	PolicyStrict = "strict"
	// PolicyWarn: log warning if some requested platforms are missing, but continue copying.
	PolicyWarn = "warn"
)

// imageAuth contains authentication information for connecting to image registries.
// It includes credentials, TLS settings, and registry endpoint details.
type imageAuth struct {
	Registry      string `json:"registry"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	SkipTLSVerify *bool  `json:"skip_tls_verify"`
	PlainHTTP     *bool  `json:"plain_http"`
}

// imageArgs holds the configuration for image operations using the new format (src/dest/manifests).
// This struct encapsulates all parameters needed for pulling, pushing, or copying container images.
type imageArgs struct {
	platform  []string       // optional: target platforms (e.g., ["linux/amd64", "linux/arm64"])
	manifests []string       // required: list of image manifests to operate on
	pattern   *regexp.Regexp // optional: regex pattern to match images
	auths     []imageAuth    // optional: image registry authentication information
	src       string         // optional: source image reference (remote or local)
	dest      string         // optional: destination image reference (local or remote)
	policy    string         // optional: policy for image copy, default is strict
}

// newImageArgs creates a new imageArgs instance from raw configuration.
// It supports both the new format (src/dest/manifests) and the deprecated format (pull/push/copy),
// but they cannot be used together. The operation type is determined by src and dest:
// - pull: src=remote registry, dest=local directory
// - push: src=local directory, dest=remote registry
// - copy: src=local directory, dest=local directory
func newImageArgs(_ context.Context, raw runtime.RawExtension, vars map[string]any) (*imageArgs, error) {
	args := variable.Extension2Variables(raw)

	if pull, ok := args["pull"]; ok {
		return transferPull(pull, vars)
	}
	if push, ok := args["push"]; ok {
		return transferPush(push, vars)
	}
	if copy, ok := args["copy"]; ok {
		return transferCopy(copy, vars)
	}

	// New format: parse directly
	ia := &imageArgs{
		manifests: make([]string, 0),
		auths:     make([]imageAuth, 0),
		policy:    PolicyStrict,
	}

	// Parse manifests
	ia.manifests, _ = variable.StringSliceVar(vars, args, "manifests")

	// Parse platform
	ia.platform, _ = variable.StringSliceVar(vars, args, "platform")

	// Parse pattern
	patternStr, _ := variable.StringVar(vars, args, "pattern")
	if patternStr != "" {
		pattern, err := regexp.Compile(patternStr)
		if err != nil {
			return nil, errors.Wrap(err, "\"pattern\" should be a valid regular expression")
		}
		ia.pattern = pattern
	}

	// Parse auths
	auths := make([]imageAuth, 0)
	_ = variable.AnyVar(vars, args, &auths, "auths")
	ia.auths = append(ia.auths, auths...)

	// Parse src
	src, _ := variable.PrintVar(args, "src")
	if src, ok := src.(string); !ok {
		return nil, errors.New("\"src\" must be a string")
	} else if src == "" {
		return nil, errors.New("\"src\" should not be empty")
	} else {
		ia.src = src
	}

	// Parse dest
	dest, _ := variable.PrintVar(args, "dest")
	if dest, ok := dest.(string); !ok {
		return nil, errors.New("\"dest\" must be a string")
	} else if dest == "" {
		return nil, errors.New("\"dest\" should not be empty")
	} else {
		ia.dest = dest
	}

	return ia, nil
}

func (i *imageArgs) copy(ctx context.Context, hostVars map[string]any) error {
	// Determine the manifests to operate on
	images := i.manifests
	// Apply pattern filtering if pattern is provided
	if i.pattern != nil && len(images) > 0 {
		var filtered []string
		for _, img := range images {
			if i.pattern.MatchString(img) {
				filtered = append(filtered, img)
			}
		}
		images = filtered
	}

	if len(images) == 0 {
		// Skip if no manifests specified
		return filepath.SkipDir
	}

	for _, img := range images {
		// Normalize image name (handle docker.io and docker.io/library)
		img = normalizeImageName(img)
		reference, err := registry.ParseReference(img)
		if err != nil {
			return errors.Wrapf(err, "failed to parse reference %q", img)
		}
		_ = unstructured.SetNestedField(hostVars, reference.Registry, "module", "image", "reference", "registry")
		_ = unstructured.SetNestedField(hostVars, reference.Repository, "module", "image", "reference", "repository")
		_ = unstructured.SetNestedField(hostVars, reference.Reference, "module", "image", "reference", "reference")

		src, err := tmpl.ParseFunc(hostVars, i.src, tmpl.StringFunc)
		if err != nil {
			return errors.Wrapf(err, "failed to parse src %q", i.src)
		}
		dest, err := tmpl.ParseFunc(hostVars, i.dest, tmpl.StringFunc)
		if err != nil {
			return errors.Wrapf(err, "failed to parse dest %q", i.dest)
		}
		// Create source repository
		srcRepo, err := newRepository(src, img, i.auths)
		if err != nil {
			return err
		}
		dstRepo, err := newRepository(dest, img, i.auths)
		if err != nil {
			return err
		}
		klog.V(4).InfoS("copy image", "src", src, "dst", dest, "source", srcRepo.Reference.String(), "destination", dstRepo.Reference.String())

		// Handle multi-platform copy with filtering
		if len(i.platform) > 0 && !slices.Contains(i.platform, "all") {
			if err := i.copyWithPlatformFilter(ctx, srcRepo, dstRepo, img, i.platform); err != nil {
				return err
			}
		} else {
			// Original copy (all platforms)
			if _, err := oras.Copy(ctx, srcRepo, srcRepo.Reference.Reference, dstRepo, "", oras.DefaultCopyOptions); err != nil {
				return errors.Wrapf(err, "failed to copy image %q", img)
			}
		}
	}

	return nil
}

func (i *imageArgs) copyWithPlatformFilter(ctx context.Context, src, dst *remote.Repository, ref string, platform []string) error {
	// Build a set of requested platforms
	requestedPlatforms := make(map[string]bool)
	for _, p := range platform {
		requestedPlatforms[p] = true
	}

	// Step 1: Get source manifest
	manifests := src.Manifests()

	// Resolve the reference to get the descriptor
	desc, err := manifests.Resolve(ctx, ref)
	if err != nil {
		return errors.Wrapf(err, "failed to resolve manifest for %s", ref)
	}

	// Fetch the manifest content
	manifestRC, err := manifests.Fetch(ctx, desc)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch manifest for %s", ref)
	}
	defer manifestRC.Close()

	// Parse the manifest
	var manifest map[string]interface{}
	if err := json.NewDecoder(manifestRC).Decode(&manifest); err != nil {
		return errors.Wrapf(err, "failed to parse manifest for %s", ref)
	}

	// Check if it's a manifest list (multi-arch)
	mediaType, _ := manifest["mediaType"].(string)
	switch mediaType {
	case "application/vnd.docker.distribution.manifest.list.v2+json", ocispec.MediaTypeImageIndex:
		// It's a manifest list, filter platforms
		manifestsList, ok := manifest["manifests"].([]interface{})
		if !ok {
			return errors.Errorf("failed to parse manifest list for %s", ref)
		}

		// Build a map of available platforms in the image
		availablePlatforms := make(map[string]interface{})
		for _, m := range manifestsList {
			m, ok := m.(map[string]interface{})
			if !ok {
				continue
			}
			platformInfo, ok := m["platform"].(map[string]interface{})
			if !ok {
				continue
			}
			os, _ := platformInfo["os"].(string)
			arch, _ := platformInfo["architecture"].(string)
			platformKey := os + "/" + arch
			availablePlatforms[platformKey] = m
		}

		// Step 2: Iterate through requested platforms and find matches
		var filteredManifests []interface{}
		var missingPlatforms []string
		for _, requestedPlatform := range platform {
			if m, exists := availablePlatforms[requestedPlatform]; exists {
				filteredManifests = append(filteredManifests, m)
			} else {
				// Platform not found in image
				missingPlatforms = append(missingPlatforms, requestedPlatform)
			}
		}

		// Handle based on policy
		switch i.policy {
		case PolicyStrict:
			if len(filteredManifests) == 0 {
				return errors.Errorf("no matching platforms found for %s in %v", ref, platform)
			}
			if len(missingPlatforms) > 0 {
				return errors.Errorf("image %s missing requested platforms: %v", ref, missingPlatforms)
			}
		case PolicyWarn:
			if len(filteredManifests) == 0 {
				klog.Warningf("no matching platforms found for image, skipping: image=%s, requestedPlatforms=%v", ref, platform)
				return nil
			}
			if len(missingPlatforms) > 0 {
				klog.Warningf("image %s missing some requested platforms, proceeding with available ones: requestedPlatforms=%v, missingPlatforms=%v", ref, platform, missingPlatforms)
			}
		default:
			// Default to strict behavior
			if len(filteredManifests) == 0 {
				return errors.Errorf("no matching platforms found for %s in %v", ref, platform)
			}
			if len(missingPlatforms) > 0 {
				return errors.Errorf("image %s missing requested platforms: %v", ref, missingPlatforms)
			}
		}

		// Step 3: Create new manifest list (index)
		filteredIndex := map[string]interface{}{
			"schemaVersion": 2,
			"mediaType":     mediaType,
			"manifests":     filteredManifests,
		}

		// Serialize the filtered index
		indexBytes, err := json.Marshal(filteredIndex)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal filtered manifest index")
		}

		// Create descriptor for the filtered index
		indexDesc := ocispec.Descriptor{
			MediaType: mediaType,
			Size:      int64(len(indexBytes)),
			Digest:    digest.Digest(computeDigest(indexBytes)),
		}

		// Step 4: Push the filtered manifest list
		err = dst.Manifests().Push(ctx, indexDesc, strings.NewReader(string(indexBytes)))
		if err != nil {
			return errors.Wrapf(err, "failed to push filtered manifest index")
		}

		// Step 5: Copy the actual image content for each filtered platform
		for _, m := range filteredManifests {
			m, ok := m.(map[string]interface{})
			if !ok {
				continue
			}

			digestStr, ok := m["digest"].(string)
			if !ok {
				continue
			}

			// Copy this platform's image content
			_, err = oras.Copy(ctx, src, digestStr, dst, "", oras.DefaultCopyOptions)
			if err != nil {
				return errors.Wrapf(err, "failed to copy platform image")
			}
		}
	case "application/vnd.docker.distribution.manifest.v2+json", "application/vnd.oci.image.manifest.v1+json":
		// This is a single-platform image, check if it matches the requested platform
		// Get config from manifest
		config, ok := manifest["config"].(map[string]interface{})
		if !ok {
			return errors.Errorf("failed to get config from manifest for %s", ref)
		}

		// Get config digest and mediaType
		configDigest, ok := config["digest"].(string)
		if !ok {
			return errors.Errorf("failed to get config digest from manifest for %s", ref)
		}
		configMediaType, _ := config["mediaType"].(string)

		// Fetch config to get platform info
		configDesc := ocispec.Descriptor{
			MediaType: configMediaType,
			Digest:    digest.Digest(configDigest),
		}
		configRC, err := src.Blobs().Fetch(ctx, configDesc)
		if err != nil {
			return errors.Wrapf(err, "failed to fetch config for %s", ref)
		}
		defer configRC.Close()

		var configContent map[string]interface{}
		if err := json.NewDecoder(configRC).Decode(&configContent); err != nil {
			return errors.Wrapf(err, "failed to parse config for %s", ref)
		}

		// Get platform from config
		// For Docker, it's under "architecture" and "os"
		// For OCI, it's under "os" and "architecture" in the root
		os, _ := configContent["os"].(string)
		arch, _ := configContent["architecture"].(string)
		if os == "" || arch == "" {
			return errors.Errorf("failed to get platform info from config for %s", ref)
		}

		platformKey := os + "/" + arch

		// Check if requested platforms contain this platform
		matched := false
		for _, p := range platform {
			if p == platformKey {
				matched = true
			}
		}

		if !matched {
			// No matching platform found, handle based on policy
			switch i.policy {
			case PolicyStrict:
				return errors.Errorf("image %s platform %q does not match any of the requested platforms %v", ref, platformKey, platform)
			case PolicyWarn:
				klog.V(2).InfoS("image platform does not match any of the requested platforms, proceeding",
					"image", ref,
					"imagePlatform", platformKey,
					"requestedPlatforms", platform)
			default:
				// Default to strict behavior
				return errors.Errorf("image %s platform %q does not match any of the requested platforms %v", ref, platformKey, platform)
			}
		}

		// Copy the single platform image
		_, err = oras.Copy(ctx, src, ref, dst, "", oras.DefaultCopyOptions)
		if err != nil {
			return errors.Wrapf(err, "failed to copy image %q", ref)
		}

	default:
		return errors.Errorf("unsupported media type: %s", mediaType)
	}

	return nil
}

// ModuleImage handles the "image" module, managing container image operations including pulling,
// pushing, and copying images between registries and local directories.
func ModuleImage(ctx context.Context, opts internal.ExecOptions) (string, string, error) {
	// get host variable
	ha, err := opts.GetAllVariables()
	if err != nil {
		return internal.StdoutFailed, internal.StderrGetHostVariable, err
	}

	ia, err := newImageArgs(ctx, opts.Args, ha)
	if err != nil {
		return internal.StdoutFailed, internal.StderrParseArgument, err
	}

	if err := ia.copy(ctx, ha); err != nil {
		if errors.Is(err, filepath.SkipDir) {
			return internal.StdoutSkip, "image manifest is empty", nil
		}
		return internal.StdoutFailed, "failed to transfer image", err
	}

	return internal.StdoutSuccess, "", nil
}

// computeDigest computes the SHA256 digest of the given content and returns it in the format "sha256:..."
func computeDigest(content []byte) string {
	hash := sha256.Sum256(content)
	return "sha256:" + hex.EncodeToString(hash[:])
}

// normalizeImageName adds the default registry (docker.io) to image names that don't include a registry.
// Examples: "ubuntu" -> "docker.io/library/ubuntu", "project/xx" -> "docker.io/project/xx"
// Images that already include a registry (e.g., "registry.example.com/image") are returned unchanged.
func normalizeImageName(image string) string {
	parts := strings.Split(image, "/")

	switch len(parts) {
	case 1:
		// Single part (e.g., "ubuntu"): official image, add library and default registry
		return fmt.Sprintf("%s/library/%s", defaultRegistry, image)
	default:
		// Two parts (e.g., "project/xx" or "registry.example.com/project")
		// Check if first part is a registry host
		firstPart := parts[0]
		if govalidator.IsHost(firstPart) {
			// registry/project format: keep as is
			return image
		}
		// project/xx format: add default registry
		return fmt.Sprintf("%s/%s", defaultRegistry, image)
	}
}
