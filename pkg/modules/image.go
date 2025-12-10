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

package modules

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/kubesphere/kubekey/v4/pkg/utils"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/containerd/containerd/images"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	imagev1 "github.com/opencontainers/image-spec/specs-go/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The Image module handles container image operations including pulling images from registries and pushing images to private registries.

Configuration:
Users can specify image operations through the following parameters:

image:
  pull:                    # optional: pull configuration
    manifests: []string    # required: list of image manifests to pull
    images_dir: string     # required: directory to store pulled images
    skipTLSVerify: bool    # optional: skip TLS verification
    autus:                 # optional: target image repo access information, slice type
      - repo: string       # optional: target image repo
        username: string   # optional: target image repo access username
        password: string   # optional: target image repo access password
  push:                    # optional: push configuration
    username: string       # optional: registry username
    password: string       # optional: registry password
    images_dir: string     # required: directory containing images to push
    skipTLSVerify: bool    # optional: skip TLS verification
    src_pattern: string            # optional: source image pattern to push (regex supported). If not specified, all images in images_dir will be pushed
    dest: string           # required: destination registry and image name. Supports template syntax for dynamic values
  copy:
    from:
      type: string           # required: image source type, file or hub
      path: string           # optional: when image source type is file, then required, means image file path
      manifests: []string    # required: list of image manifests to pull
      skipTLSVerify: bool    # optional: skip TLS verification
      autus:                 # optional: target image repo access information, slice type
        - repo: string       # optional: target image repo
          username: string   # optional: target image repo access username
          password: string   # optional: target image repo access password
    to:
      type: string           # required: image target type, file or hub
      path: string           # required: image target path
      skipTLSVerify: bool    # optional: skip TLS verification
      pattern: string        # optional: source image pattern to push (regex supported). If not specified, all images in images_dir will be pushed
      autus:                 # optional: target image repo access information, slice type
        - repo: string       # optional: target image repo
          username: string   # optional: target image repo access username
          password: string   # optional: target image repo access password

Usage Examples in Playbook Tasks:
1. Pull images from registry:
   ```yaml
   - name: Pull container images
     image:
       pull:
         manifests:
           - nginx:latest
           - prometheus:v2.45.0
         images_dir: /path/to/images
         auths:
           - repo: docker.io
             username: MyDockerAccount
             password: my_password
           - repo: my.dockerhub.local
             username: MyHubAccount
             password: my_password
     register: pull_result
   ```

2. Push images to private registry:
   ```yaml
   - name: Push images to private registry
     image:
       push:
         username: admin
         password: secret
         namespace_override: custom-ns
         images_dir: /path/to/images
		 dest: registry.example.com/{{ . }}
     register: push_result
   ```

3. Copy image from file to file
   ```yaml
   - name: file to file
     image:
       copy:
         from:
           path: "/path/from/images"
           manifests:
            - nginx:latest
            - prometheus:v2.45.0
         to:
           path: /path/to/images
   ```

Return Values:
- On success: Returns "Success" in stdout
- On failure: Returns error message in stderr
*/

const defaultRegistry = "docker.io"

// imageArgs holds the configuration for image operations
type imageArgs struct {
	pull *imagePullArgs
	push *imagePushArgs
	copy *imageCopyArgs
}

// imagePullArgs contains parameters for pulling images
type imagePullArgs struct {
	imagesDir     string
	manifests     []string
	skipTLSVerify *bool
	platform      []string
	auths         []imageAuth
}

type imageAuth struct {
	Repo      string `json:"repo"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Insecure  *bool  `json:"insecure"`
	PlainHTTP *bool  `json:"plain_http"`
}

type fetchResult struct {
	IsIndex   bool
	IndexDesc *imagev1.Descriptor
	Index     *imagev1.Index
	Manifests []*manifestInfo
}

type manifestInfo struct {
	Desc       imagev1.Descriptor
	Content    []byte
	Platform   *imagev1.Platform
	SourceRepo *remote.Repository
}

// pull retrieves images from a remote registry and stores them locally
func (i imagePullArgs) pull(ctx context.Context, platform []string) error {

	manifests := i.manifests
	if len(manifests) == 0 {
		return nil
	}

	maxWorkers := 10

	// 创建任务队列
	tasks := make(chan string, len(manifests))

	worker := utils.Worker[string]{
		MaxWorkerCount: maxWorkers,
		TaskChan:       tasks,
		ExecFunc: func(img string) error {
			return i.downloadSingleImage(ctx, img, platform)
		},
	}

	worker.Do(ctx)

	// 发送任务
	for _, img := range manifests {
		tasks <- img
	}
	close(tasks)

	// 等待所有 worker 完成
	go func() {
		worker.Wait()
	}()

	// 收集结果
	var collectedErrors = worker.CollectedErrors()

	if len(collectedErrors) > 0 {
		return fmt.Errorf("download errors: %v", strings.Join(collectedErrors, "; "))
	}
	return nil
}

func (i imagePullArgs) downloadSingleImage(ctx context.Context, img string, platform []string) error {
	img = normalizeImageNameSimple(img)
	src, err := remote.NewRepository(img)
	if err != nil {
		return errors.Wrapf(err, "failed to get remote image %s", img)
	}
	src.Client = &auth.Client{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipTLSVerifyFunc(img, i.auths, *i.skipTLSVerify),
				},
			},
		},
		Cache:      auth.NewCache(),
		Credential: authFunc(i.auths),
	}
	dst, err := newLocalRepository(filepath.Join(src.Reference.Registry, src.Reference.Repository)+":"+src.Reference.Reference, i.imagesDir)
	if err != nil {
		return err
	}
	src.PlainHTTP = plainHTTPFunc(img, i.auths, false)

	return imageSrcToDst(ctx, src, dst, img, platform)
}

func imageSrcToDst(ctx context.Context, src, dst *remote.Repository, img string, platform []string) error {
	var err error
	if len(platform) == 0 || (len(platform) == 1 && strings.TrimSpace(platform[0]) == "*") {
		_, err = oras.Copy(ctx, src, src.Reference.Reference, dst, "", oras.DefaultCopyOptions)
		if err != nil {
			err = errors.Wrapf(err, "failed to pull image %q to local dir", img)
		}
		return err
	}
	fetchResult, defaultMediaType, err := fetchManifestsFromMultiArch(ctx, src, src.Reference.Reference)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch manifests")
	}
	if !fetchResult.IsIndex {
		_, err = oras.Copy(ctx, src, src.Reference.Reference, dst, "", oras.DefaultCopyOptions)
		if err != nil {
			err = errors.Wrapf(err, "failed to pull image %q to local dir", img)
		}
		return err
	}
	// filter target platform
	var filteredManifests []*manifestInfo
	for _, manifest := range fetchResult.Manifests {
		// some arm architecture is arm64/v7 or arm68/v8 , support all of then
		for _, arch := range platform {
			if strings.Contains(manifest.Platform.Architecture, arch) {
				manifest.SourceRepo = src
				filteredManifests = append(filteredManifests, manifest)
				break
			}
		}
	}
	if len(filteredManifests) == 0 {
		klog.Warningf("Image %s has no manifests matched for platform: %s", img, platform)
		return nil
	}
	// push all filtered manifests and layers
	for _, manifest := range filteredManifests {
		if err = pushManifestWithLayers(ctx, src, dst, manifest); err != nil {
			return errors.Wrapf(err, "failed to push manifest for %s/%s",
				manifest.Platform.OS, manifest.Platform.Architecture)
		}
	}
	err = createAndPushIndex(ctx, dst, filteredManifests, dst.Reference.Reference, defaultMediaType)
	if err != nil {
		return errors.Wrapf(err, "failed to pull image %q to local dir", img)
	}
	return nil
}

func fetchManifestsFromMultiArch(ctx context.Context, repo *remote.Repository, ref string) (*fetchResult, string, error) {
	desc, rc, err := repo.FetchReference(ctx, ref)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch reference %s: %w", ref, err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read content: %w", err)
	}

	result := &fetchResult{}

	if desc.MediaType == imagev1.MediaTypeImageIndex ||
		desc.MediaType == "application/vnd.docker.distribution.manifest.list.v2+json" {
		// multi arch image
		result.IsIndex = true
		result.IndexDesc = &desc

		var index imagev1.Index
		if err := json.Unmarshal(content, &index); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal index: %w", err)
		}
		result.Index = &index

		for _, manifestDesc := range index.Manifests {
			if manifestDesc.MediaType != imagev1.MediaTypeImageManifest &&
				manifestDesc.MediaType != "application/vnd.docker.distribution.manifest.v2+json" {
				continue
			}

			manifestInfo, err := fetchSingleManifest(ctx, repo, manifestDesc)
			if err != nil {
				return nil, "", fmt.Errorf("failed to fetch manifest %s: %w", manifestDesc.Digest, err)
			}

			result.Manifests = append(result.Manifests, manifestInfo)
		}
	} else if desc.MediaType == imagev1.MediaTypeImageManifest ||
		desc.MediaType == "application/vnd.docker.distribution.manifest.v2+json" {
		// single arch image
		result.IsIndex = false
		info, err := fetchSingleManifestFromContent(content, &desc)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse manifest: %w", err)
		}
		result.Manifests = []*manifestInfo{info}
	} else {
		return nil, "", fmt.Errorf("unsupported media type: %s", desc.MediaType)
	}

	return result, desc.MediaType, nil
}

func fetchSingleManifest(ctx context.Context, repo *remote.Repository, desc imagev1.Descriptor) (*manifestInfo, error) {
	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest content: %w", err)
	}

	return fetchSingleManifestFromContent(content, &desc)
}

func fetchSingleManifestFromContent(content []byte, desc *imagev1.Descriptor) (*manifestInfo, error) {
	var manifest imagev1.Manifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	platform := desc.Platform
	if platform == nil {
		// read platform from config
		// but if config has no platform info ,then use default unknown
		if manifest.Config.Platform != nil {
			platform = manifest.Config.Platform
		} else {
			platform = &imagev1.Platform{
				Architecture: "unknown",
				OS:           "unknown",
			}
		}
	}

	return &manifestInfo{
		Desc:     *desc,
		Content:  content,
		Platform: platform,
	}, nil
}

func pushManifestWithLayers(ctx context.Context, srcRepo, dstRepo *remote.Repository, manifestInfo *manifestInfo) error {
	var manifest imagev1.Manifest
	if err := json.Unmarshal(manifestInfo.Content, &manifest); err != nil {
		return fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	// push config layer
	if err := copyBlob(ctx, srcRepo, dstRepo, manifest.Config); err != nil {
		return fmt.Errorf("failed to copy config: %w", err)
	}

	// push all layers
	for _, layer := range manifest.Layers {
		if err := copyBlob(ctx, srcRepo, dstRepo, layer); err != nil {
			return fmt.Errorf("failed to copy layer %s: %w", layer.Digest, err)
		}
	}

	// push manifests
	manifestDesc := imagev1.Descriptor{
		MediaType: manifest.MediaType,
		Digest:    digest.FromBytes(manifestInfo.Content),
		Size:      int64(len(manifestInfo.Content)),
		Platform:  manifestInfo.Platform,
	}

	exists, err := dstRepo.Exists(ctx, manifestDesc)
	if err == nil && exists {
		return nil
	}

	err = dstRepo.Push(ctx, manifestDesc, bytes.NewReader(manifestInfo.Content))
	if err != nil {
		return fmt.Errorf("failed to push manifest: %w", err)
	}

	return nil
}

func copyBlob(ctx context.Context, srcRepo, dstRepo *remote.Repository, desc imagev1.Descriptor) error {
	exists, err := dstRepo.Exists(ctx, desc)
	if err == nil && exists {
		return nil
	}

	rc, err := srcRepo.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("failed to fetch blob %s: %w", desc.Digest, err)
	}
	defer rc.Close()

	err = dstRepo.Push(ctx, desc, rc)
	if err != nil {
		return fmt.Errorf("failed to push blob %s: %w", desc.Digest, err)
	}

	return nil
}

func createAndPushIndex(ctx context.Context, dstRepo *remote.Repository, manifests []*manifestInfo, targetTag, defaultMediaType string) error {
	var descList = make([]imagev1.Descriptor, 0)
	for _, info := range manifests {
		var manifest imagev1.Manifest
		if err := json.Unmarshal(info.Content, &manifest); err != nil {
			return fmt.Errorf("failed to unmarshal manifest: %w", err)
		}
		desc := imagev1.Descriptor{
			MediaType: manifest.MediaType,
			Digest:    digest.FromBytes(info.Content),
			Size:      int64(len(info.Content)),
			Platform:  info.Platform,
		}
		descList = append(descList, desc)
	}

	index := imagev1.Index{
		Versioned: specs.Versioned{SchemaVersion: 2},
		MediaType: defaultMediaType,
		Manifests: descList,
		Annotations: map[string]string{
			"org.opencontainers.image.created":  time.Now().UTC().Format(time.RFC3339),
			"org.opencontainers.image.ref.name": targetTag,
		},
	}

	indexJSON, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	indexDesc := imagev1.Descriptor{
		MediaType: defaultMediaType,
		Digest:    digest.FromBytes(indexJSON),
		Size:      int64(len(indexJSON)),
	}

	err = dstRepo.PushReference(ctx, indexDesc, bytes.NewReader(indexJSON), targetTag)
	if err != nil {
		return fmt.Errorf("failed to push index: %w", err)
	}

	return nil
}

func dockerHostParser(img string) string {
	// if image is like docker.io/xxx/xxx:tag, then download by pull func will store it to registry-1.docker.io
	// so we should change host from docker.io to registry-1.docker.io
	splitedImg := strings.Split(img, "/")
	if len(splitedImg) == 1 {
		return img
	}
	if splitedImg[0] != "docker.io" {
		return img
	}
	splitedImg[0] = "registry-1.docker.io"
	return strings.Join(splitedImg, "/")
}

func authFunc(auths []imageAuth) func(ctx context.Context, hostport string) (auth.Credential, error) {
	var creds = make(map[string]auth.Credential)
	for _, inputAuth := range auths {
		var rp = inputAuth.Repo
		if rp == "docker.io" || rp == "" {
			rp = "registry-1.docker.io"
		}
		creds[rp] = auth.Credential{
			Username: inputAuth.Username,
			Password: inputAuth.Password,
		}
	}
	return func(_ context.Context, hostport string) (auth.Credential, error) {
		cred, ok := creds[hostport]
		if !ok {
			cred = auth.EmptyCredential
		}
		return cred, nil
	}
}

func skipTLSVerifyFunc(img string, auths []imageAuth, defaults bool) bool {
	imgHost := strings.Split(img, "/")[0]
	for _, a := range auths {
		if imgHost == a.Repo {
			if a.Insecure != nil {
				return *a.Insecure
			}
			return defaults
		}
	}
	return defaults
}

func plainHTTPFunc(img string, auths []imageAuth, defaults bool) bool {
	imgHost := strings.Split(img, "/")[0]
	for _, a := range auths {
		if imgHost == a.Repo {
			if a.PlainHTTP != nil {
				return *a.PlainHTTP
			}
			return defaults
		}
	}
	return defaults
}

// imagePushArgs contains parameters for pushing images
type imagePushArgs struct {
	imagesDir     string
	skipTLSVerify *bool
	srcPattern    *regexp.Regexp
	destTmpl      string
	auths         []imageAuth
}

// push uploads local images to a remote registry
func (i imagePushArgs) push(ctx context.Context, hostVars map[string]any) error {
	manifests, err := findLocalImageManifests(i.imagesDir)
	if err != nil {
		return err
	}
	klog.V(5).Info("manifests found", "manifests", manifests)

	for _, img := range manifests {
		// match regex by src
		if i.srcPattern != nil && !i.srcPattern.MatchString(img) {
			// skip
			continue
		}
		src, err := newLocalRepository(img, i.imagesDir)
		if err != nil {
			return err
		}
		dest := i.destTmpl
		if kkprojectv1.IsTmplSyntax(dest) {
			// add temporary variable
			_ = unstructured.SetNestedField(hostVars, src.Reference.Registry, "module", "image", "src", "reference", "registry")
			_ = unstructured.SetNestedField(hostVars, src.Reference.Repository, "module", "image", "src", "reference", "repository")
			_ = unstructured.SetNestedField(hostVars, src.Reference.Reference, "module", "image", "src", "reference", "reference")
			dest, err = tmpl.ParseFunc(hostVars, dest, func(b []byte) string { return string(b) })
			if err != nil {
				return err
			}
		}
		dst, err := remote.NewRepository(dest)
		if err != nil {
			return errors.Wrapf(err, "failed to get remote repository %q", dest)
		}
		dst.Client = &auth.Client{
			Client: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: skipTLSVerifyFunc(dest, i.auths, *i.skipTLSVerify),
					},
				},
			},
			Cache:      auth.NewCache(),
			Credential: authFunc(i.auths),
		}

		dst.PlainHTTP = plainHTTPFunc(dest, i.auths, false)

		if _, err = oras.Copy(ctx, src, src.Reference.Reference, dst, dst.Reference.Reference, oras.DefaultCopyOptions); err != nil {
			return errors.Wrapf(err, "failed to push image %q to remote", img)
		}
	}

	return nil
}

type imageCopyArgs struct {
	Platform []string            `json:"platform"`
	From     imageCopyTargetArgs `json:"from"`
	To       imageCopyTargetArgs `json:"to"`
}

type imageCopyTargetArgs struct {
	Path      string `json:"path"`
	manifests []string
	Pattern   *regexp.Regexp
}

func (i *imageCopyArgs) parseFromVars(vars, cp map[string]any) error {
	i.Platform, _ = variable.StringSliceVar(vars, cp, "platform")

	i.From.manifests, _ = variable.StringSliceVar(vars, cp, "from", "manifests")

	i.From.Path, _ = variable.StringVar(vars, cp, "from", "path")

	toPath, _ := variable.PrintVar(cp, "to", "path")
	if destStr, ok := toPath.(string); !ok {
		return errors.New("\"copy.to.path\" must be a string")
	} else if destStr == "" {
		return errors.New("\"copy.to.path\" should not be empty")
	} else {
		i.To.Path = destStr
	}
	srcPattern, _ := variable.StringVar(vars, cp, "to", "src_pattern")
	if srcPattern != "" {
		pattern, err := regexp.Compile(srcPattern)
		if err != nil {
			return errors.Wrap(err, "\"to.pattern\" should be a valid regular expression. ")
		}
		i.To.Pattern = pattern
	}
	return nil
}

func (i *imageCopyArgs) copy(ctx context.Context, hostVars map[string]any) error {
	if sts, err := os.Stat(i.From.Path); err != nil || !sts.IsDir() {
		return errors.New("\"copy.from.path\" must be a exist directory")
	}
	for _, img := range i.From.manifests {
		img = normalizeImageNameSimple(img)
		if i.To.Pattern != nil && !i.To.Pattern.MatchString(img) {
			// skip
			continue
		}
		src, err := newLocalRepository(dockerHostParser(img), i.From.Path)
		if err != nil {
			return err
		}
		dest := i.To.Path
		if kkprojectv1.IsTmplSyntax(dest) {
			// add temporary variable
			_ = unstructured.SetNestedField(hostVars, src.Reference.Registry, "module", "image", "src", "reference", "registry")
			_ = unstructured.SetNestedField(hostVars, src.Reference.Repository, "module", "image", "src", "reference", "repository")
			_ = unstructured.SetNestedField(hostVars, src.Reference.Reference, "module", "image", "src", "reference", "reference")
			dest, err = tmpl.ParseFunc(hostVars, dest, func(b []byte) string { return string(b) })
			if err != nil {
				return err
			}
		}
		dst, err := newLocalRepository(filepath.Join(src.Reference.Registry, src.Reference.Repository)+":"+src.Reference.Reference, dest)
		if err != nil {
			return err
		}

		err = imageSrcToDst(ctx, src, dst, img, i.Platform)
		if err != nil {
			return err
		}
	}

	return nil
}

// newImageArgs creates a new imageArgs instance from raw configuration
func newImageArgs(_ context.Context, raw runtime.RawExtension, vars map[string]any) (*imageArgs, error) {
	ia := &imageArgs{}
	// check args
	args := variable.Extension2Variables(raw)
	binaryDir, _ := variable.StringVar(vars, args, _const.BinaryDir)

	if pullArgs, ok := args["pull"]; ok {
		pull, ok := pullArgs.(map[string]any)
		if !ok {
			return nil, errors.New("\"pull\" should be map")
		}
		ipl := &imagePullArgs{}
		ipl.manifests, _ = variable.StringSliceVar(vars, pull, "manifests")
		ipl.auths = make([]imageAuth, 0)
		pullAuths := make([]imageAuth, 0)
		_ = variable.AnyVar(vars, pull, &pullAuths, "auths")
		for _, a := range pullAuths {
			a.Repo, _ = tmpl.ParseFunc(vars, a.Repo, func(b []byte) string { return string(b) })
			a.Username, _ = tmpl.ParseFunc(vars, a.Username, func(b []byte) string { return string(b) })
			a.Password, _ = tmpl.ParseFunc(vars, a.Password, func(b []byte) string { return string(b) })
			ipl.auths = append(ipl.auths, a)
		}
		ipl.imagesDir, _ = variable.StringVar(vars, pull, "images_dir")
		ipl.skipTLSVerify, _ = variable.BoolVar(vars, pull, "skip_tls_verify")
		if ipl.skipTLSVerify == nil {
			ipl.skipTLSVerify = ptr.To(false)
		}
		ipl.platform, _ = variable.StringSliceVar(vars, pull, "platform")
		// check args
		if len(ipl.manifests) == 0 {
			return nil, errors.New("\"pull.manifests\" is required")
		}
		if ipl.imagesDir == "" {
			if binaryDir == "" {
				return nil, errors.New("\"pull.images_dir\" is required")
			}
			ipl.imagesDir = filepath.Join(binaryDir, _const.BinaryImagesDir)
		}
		ia.pull = ipl
	}
	// if namespace_override is not empty, it will override the image manifests namespace_override. (namespace maybe multi sub path)
	// push to private registry
	if pushArgs, ok := args["push"]; ok {
		push, ok := pushArgs.(map[string]any)
		if !ok {
			return nil, errors.New("\"push\" should be map")
		}

		ips := &imagePushArgs{}
		ips.auths = make([]imageAuth, 0)
		pullAuths := make([]imageAuth, 0)
		_ = variable.AnyVar(vars, push, &pullAuths, "auths")
		for _, a := range pullAuths {
			a.Repo, _ = tmpl.ParseFunc(vars, a.Repo, func(b []byte) string { return string(b) })
			a.Username, _ = tmpl.ParseFunc(vars, a.Username, func(b []byte) string { return string(b) })
			a.Password, _ = tmpl.ParseFunc(vars, a.Password, func(b []byte) string { return string(b) })
			ips.auths = append(ips.auths, a)
		}
		ips.imagesDir, _ = variable.StringVar(vars, push, "images_dir")
		srcPattern, _ := variable.StringVar(vars, push, "src_pattern")
		destTmpl, _ := variable.PrintVar(push, "dest")
		ips.skipTLSVerify, _ = variable.BoolVar(vars, push, "skip_tls_verify")
		if ips.skipTLSVerify == nil {
			ips.skipTLSVerify = ptr.To(false)
		}
		// check args
		if ips.imagesDir == "" {
			if binaryDir == "" {
				return nil, errors.New("\"push.images_dir\" is required")
			}
			ips.imagesDir = filepath.Join(binaryDir, _const.BinaryImagesDir)
		}
		if srcPattern != "" {
			pattern, err := regexp.Compile(srcPattern)
			if err != nil {
				return nil, errors.Wrap(err, "\"push.src\" should be a valid regular expression. ")
			}
			ips.srcPattern = pattern
		}
		if destStr, ok := destTmpl.(string); !ok {
			return nil, errors.New("\"push.dest\" must be a string")
		} else if destStr == "" {
			return nil, errors.New("\"push.dest\" should not be empty")
		} else {
			ips.destTmpl = destStr
		}
		ia.push = ips
	}

	if cpArgs, ok := args["copy"]; ok {
		cp, ok := cpArgs.(map[string]any)
		if !ok {
			return nil, errors.New("\"copy\" should be map")
		}

		cps := &imageCopyArgs{}

		err := cps.parseFromVars(vars, cp)
		if err != nil {
			return nil, err
		}

		ia.copy = cps
	}

	return ia, nil
}

// ModuleImage handles the "image" module, managing container image operations including pulling and pushing images
func ModuleImage(ctx context.Context, options ExecOptions) (string, string, error) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return StdoutFailed, StderrGetHostVariable, err
	}

	ia, err := newImageArgs(ctx, options.Args, ha)
	if err != nil {
		return StdoutFailed, StderrParseArgument, err
	}

	// pull image manifests to local dir
	if ia.pull != nil {
		if err := ia.pull.pull(ctx, ia.pull.platform); err != nil {
			return StdoutFailed, "failed to pull image", err
		}
	}
	// push image to private registry
	if ia.push != nil {
		if err := ia.push.push(ctx, ha); err != nil {
			return StdoutFailed, "failed to push image", err
		}
	}

	if ia.copy != nil {
		if err := ia.copy.copy(ctx, ha); err != nil {
			return StdoutFailed, "failed to push image", err
		}
	}

	return StdoutSuccess, "", nil
}

// findLocalImageManifests get image manifests with whole image's name.
func findLocalImageManifests(localDir string) ([]string, error) {
	if _, err := os.Stat(localDir); err != nil {
		klog.V(4).ErrorS(err, "failed to stat local directory", "image_dir", localDir)
		// images is not exist, skip
		return make([]string, 0), nil
	}

	var manifests []string
	if err := filepath.WalkDir(localDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "failed to walkdir %s", path)
		}

		if path == filepath.Join(localDir, "blobs") {
			return filepath.SkipDir
		}

		if d.IsDir() || filepath.Base(path) == "manifests" {
			return nil
		}

		file, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %q", path)
		}

		var data map[string]any
		if err := json.Unmarshal(file, &data); err != nil {
			klog.V(4).ErrorS(err, "unmarshal manifests file error", "file", path)
			// skip un-except file (empty)
			return nil
		}

		mediaType, ok := data["mediaType"].(string)
		if !ok {
			return errors.New("invalid mediaType")
		}
		if mediaType == imagev1.MediaTypeImageIndex || mediaType == imagev1.MediaTypeImageManifest || // oci multi or single schema
			mediaType == images.MediaTypeDockerSchema2ManifestList || mediaType == images.MediaTypeDockerSchema2Manifest { // docker multi or single schema
			subpath, err := filepath.Rel(localDir, path)
			if err != nil {
				return errors.Wrap(err, "failed to get relative filepath")
			}
			if strings.HasPrefix(filepath.Base(subpath), "sha256:") {
				// only found image tag.
				return nil
			}
			// the parent dir of subpath is "manifests". should delete it
			manifests = append(manifests, filepath.Dir(filepath.Dir(subpath))+":"+filepath.Base(subpath))
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return manifests, nil
}

// newLocalRepository local dir images repository
func newLocalRepository(reference, localDir string) (*remote.Repository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse reference %q", reference)
	}
	// store in each domain

	return &remote.Repository{
		Reference: ref,
		Client:    &http.Client{Transport: &imageTransport{baseDir: localDir}},
	}, nil
}

var responseNotFound = &http.Response{Proto: "Local", StatusCode: http.StatusNotFound}
var responseNotAllowed = &http.Response{Proto: "Local", StatusCode: http.StatusMethodNotAllowed}
var responseServerError = &http.Response{Proto: "Local", StatusCode: http.StatusInternalServerError}
var responseCreated = &http.Response{Proto: "Local", StatusCode: http.StatusCreated}
var responseOK = &http.Response{Proto: "Local", StatusCode: http.StatusOK}

// const domain = "internal"
const apiPrefix = "/v2/"

type imageTransport struct {
	baseDir string
}

// RoundTrip deal http.Request in local dir images.
func (i imageTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	var resp *http.Response

	switch request.Method {
	case http.MethodHead:
		resp = i.head(request)
	case http.MethodPost:
		resp = i.post(request)
	case http.MethodPut:
		resp = i.put(request)
	case http.MethodGet:
		resp = i.get(request)
	default:
		resp = responseNotAllowed
	}

	if resp != nil {
		resp.Request = request
	}

	return resp, nil
}

// head method for http.MethodHead. check if file is exist in blobs dir or manifests dir
func (i imageTransport) head(request *http.Request) *http.Response {
	if strings.HasSuffix(filepath.Dir(request.URL.Path), "blobs") { // blobs
		filename := filepath.Join(i.baseDir, "blobs", filepath.Base(request.URL.Path))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound
		}

		return responseOK
	} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "manifests") { // manifests
		filename := filepath.Join(i.baseDir, request.Host, strings.TrimPrefix(request.URL.Path, apiPrefix))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound
		}

		file, err := os.ReadFile(filename)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError
		}

		var data map[string]any
		if err := json.Unmarshal(file, &data); err != nil {
			klog.V(4).ErrorS(err, "failed to unmarshal file", "filename", filename)

			return responseServerError
		}

		mediaType, ok := data["mediaType"].(string)
		if !ok {
			klog.V(4).ErrorS(nil, "unknown mediaType", "filename", filename)

			return responseServerError
		}

		return &http.Response{
			Proto:      "Local",
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{mediaType},
			},
			ContentLength: int64(len(file)),
		}
	}

	return responseNotAllowed
}

// post method for http.MethodPost, accept request.
func (i imageTransport) post(request *http.Request) *http.Response {
	if strings.HasSuffix(request.URL.Path, "/uploads/") {
		return &http.Response{
			Proto:      "Local",
			StatusCode: http.StatusAccepted,
			Header: http.Header{
				"Location": []string{filepath.Dir(request.URL.Path)},
			},
			Request: request,
		}
	}

	return responseNotAllowed
}

// put method for http.MethodPut, create file in blobs dir or manifests dir
func (i imageTransport) put(request *http.Request) *http.Response {
	if strings.HasSuffix(request.URL.Path, "/uploads") { // blobs
		defer request.Body.Close()

		filename := filepath.Join(i.baseDir, "blobs", request.URL.Query().Get("digest"))
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			fmt.Println(err, "failed to create dir", "dir", filepath.Dir(filename))

			return responseServerError
		}

		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			fmt.Println(err, "failed to create file", "filename", filename)
			return responseServerError
		}

		defer func() {
			if err = file.Sync(); err != nil {
				fmt.Println(err, "failed to sync file", "filename", filename)
			}
			if err = file.Close(); err != nil {
				fmt.Println(err, "failed to close file", "filename", filename)
			}
		}()

		if _, err = io.Copy(file, request.Body); err != nil {
			fmt.Println(err, "failed to write file", "filename", filename)

			return responseServerError
		}

		return responseCreated
	} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "/manifests") { // manifests
		body, err := io.ReadAll(request.Body)
		if err != nil {
			fmt.Println(err, "failed to read request")

			return responseServerError
		}
		defer request.Body.Close()

		filename := filepath.Join(i.baseDir, request.Host, strings.TrimPrefix(request.URL.Path, apiPrefix))
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			fmt.Println(err, "failed to create dir", "dir", filepath.Dir(filename))

			return responseServerError
		}

		if err := os.WriteFile(filename, body, os.ModePerm); err != nil {
			fmt.Println(err, "failed to write file", "filename", filename)

			return responseServerError
		}

		return responseCreated
	}

	return responseNotAllowed
}

// get method for http.MethodGet, get file in blobs dir or manifest dir
func (i imageTransport) get(request *http.Request) *http.Response {
	if strings.HasSuffix(filepath.Dir(request.URL.Path), "blobs") { // blobs
		filename := filepath.Join(i.baseDir, "blobs", filepath.Base(request.URL.Path))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound
		}

		file, err := os.Open(filename)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError
		}

		fStat, err := file.Stat()
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError
		}

		return &http.Response{
			Proto:         "Local",
			StatusCode:    http.StatusOK,
			ContentLength: fStat.Size(),
			Body:          file,
		}
	} else if strings.HasSuffix(filepath.Dir(request.URL.Path), "manifests") { // manifests
		filename := filepath.Join(i.baseDir, request.Host, strings.TrimPrefix(request.URL.Path, apiPrefix))
		if _, err := os.Stat(filename); err != nil {
			klog.V(4).ErrorS(err, "failed to stat blobs", "filename", filename)

			return responseNotFound
		}

		file, err := os.ReadFile(filename)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to read file", "filename", filename)

			return responseServerError
		}

		var data map[string]any
		if err := json.Unmarshal(file, &data); err != nil {
			klog.V(4).ErrorS(err, "failed to unmarshal file data", "filename", filename)

			return responseServerError
		}

		mediaType, ok := data["mediaType"].(string)
		if !ok {
			klog.V(4).ErrorS(nil, "unknown mediaType", "filename", filename)

			return responseServerError
		}

		return &http.Response{
			Proto:      "Local",
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{mediaType},
			},
			ContentLength: int64(len(file)),
			Body:          io.NopCloser(bytes.NewReader(file)),
		}
	}

	return responseNotAllowed
}

func normalizeImageNameSimple(image string) string {
	parts := strings.Split(image, "/")

	switch len(parts) {
	case 1:
		// image like: ubuntu -> docker.io/library/ubuntu
		return fmt.Sprintf("%s/library/%s", defaultRegistry, image)
	case 2:
		// image like: project/xx or registry/project
		firstPart := parts[0]
		if firstPart == "localhost" || (strings.Contains(firstPart, ".") || strings.Contains(firstPart, ":")) {
			return image
		}
		return fmt.Sprintf("%s/%s", defaultRegistry, image)
	default:
		// image like: registry/project/xx/sub
		firstPart := parts[0]
		if firstPart == "localhost" || (strings.Contains(firstPart, ".") || strings.Contains(firstPart, ":")) {
			return image
		}
		return fmt.Sprintf("%s/%s", defaultRegistry, image)
	}
}

func init() {
	utilruntime.Must(RegisterModule("image", ModuleImage))
}
