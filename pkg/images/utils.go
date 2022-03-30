/*
 Copyright 2022 The KubeSphere Authors.

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

package images

import (
	"encoding/json"
	"fmt"
	"github.com/containerd/containerd/platforms"
	"github.com/containers/image/v5/types"
	manifesttypes "github.com/estesp/manifest-tool/v2/pkg/types"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"strings"
)

var defaultUserAgent = "kubekey"

type dockerImageOptions struct {
	arch           string
	os             string
	variant        string
	username       string
	password       string
	dockerCertPath string
	tlsVerify      bool
}

func (d *dockerImageOptions) systemContext() *types.SystemContext {
	ctx := &types.SystemContext{
		ArchitectureChoice:          d.arch,
		OSChoice:                    d.os,
		VariantChoice:               d.variant,
		DockerRegistryUserAgent:     defaultUserAgent,
		DockerInsecureSkipTLSVerify: types.NewOptionalBool(d.tlsVerify),
	}
	return ctx
}

type srcImageOptions struct {
	dockerImage   dockerImageOptions
	imageName     string
	sharedBlobDir string
}

func (s *srcImageOptions) systemContext() *types.SystemContext {
	ctx := s.dockerImage.systemContext()
	ctx.DockerCertPath = s.dockerImage.dockerCertPath
	ctx.OCISharedBlobDirPath = s.sharedBlobDir
	ctx.DockerAuthConfig = &types.DockerAuthConfig{
		Username: s.dockerImage.username,
		Password: s.dockerImage.password,
	}

	return ctx
}

type destImageOptions struct {
	dockerImage dockerImageOptions
	imageName   string
}

func (d *destImageOptions) systemContext() *types.SystemContext {
	ctx := d.dockerImage.systemContext()
	ctx.DockerCertPath = d.dockerImage.dockerCertPath
	ctx.DockerAuthConfig = &types.DockerAuthConfig{
		Username: d.dockerImage.username,
		Password: d.dockerImage.password,
	}

	return ctx
}

type registryAuth struct {
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	PlainHTTP bool   `json:"plainHTTP,omitempty"`
}

func Auths(manifest *common.ArtifactManifest) (auths map[string]registryAuth) {
	if len(manifest.Spec.ManifestRegistry.Auths.Raw) == 0 {
		return
	}

	err := json.Unmarshal(manifest.Spec.ManifestRegistry.Auths.Raw, &auths)
	if err != nil {
		logger.Log.Fatalf("Failed to Parse Registry Auths configuration: %v", manifest.Spec.ManifestRegistry.Auths.Raw)
		return
	}

	return
}

// ParseArchVariant
// Ex:
// amd64 returns amd64, ""
// arm/v8 returns arm, v8
func ParseArchVariant(platform string) (string, string) {
	osArchArr := strings.Split(platform, "/")

	variant := ""
	arch := osArchArr[0]
	if len(osArchArr) > 1 {
		variant = osArchArr[1]
	}
	return arch, variant
}

func ParseImageWithArchTag(ref string) (string, ocispec.Platform) {
	n := strings.LastIndex(ref, "-")
	if n < 0 {
		logger.Log.Fatalf("get arch or variant index failed: %s", ref)
	}
	archOrVariant := ref[n+1:]

	// try to parse the arch-only case
	specifier := fmt.Sprintf("linux/%s", archOrVariant)
	if p, err := platforms.Parse(specifier); err == nil && isKnownArch(p.Architecture) {
		return ref[:n], p
	}

	archStr := ref[:n]
	a := strings.LastIndex(archStr, "-")
	if a < 0 {
		logger.Log.Fatalf("get arch index failed: %s", ref)
	}
	arch := archStr[a+1:]

	// parse the case where both arch and variant exist
	specifier = fmt.Sprintf("linux/%s/%s", arch, archOrVariant)
	p, err := platforms.Parse(specifier)
	if err != nil {
		logger.Log.Fatalf("parse image %s failed: %s", ref, err.Error())
	}

	return ref[:a], p
}

func isKnownArch(arch string) bool {
	switch arch {
	case "386", "amd64", "amd64p32", "arm", "armbe", "arm64", "arm64be", "ppc64", "ppc64le", "loong64", "mips", "mipsle", "mips64", "mips64le", "mips64p32", "mips64p32le", "ppc", "riscv", "riscv64", "s390", "s390x", "sparc", "sparc64", "wasm":
		return true
	}
	return false
}

// ParseImageTag
// Get a repos name and returns the right reposName + tag
// The tag can be confusing because of a port in a repository name.
//     Ex: localhost.localdomain:5000/samalba/hipache:latest
func ParseImageTag(repos string) (string, string) {
	n := strings.LastIndex(repos, ":")
	if n < 0 {
		return repos, ""
	}
	if tag := repos[n+1:]; !strings.Contains(tag, "/") {
		return repos[:n], tag
	}
	return repos, ""
}

func NewManifestSpec(image string, entries []manifesttypes.ManifestEntry) manifesttypes.YAMLInput {
	var srcImages []manifesttypes.ManifestEntry

	var tags []string
	for _, e := range entries {
		_, tag := ParseImageTag(e.Image)
		tags = append(tags, tag)
		newTag := strings.ReplaceAll(tag, "amd64", "arm64")
		tags = append(tags, newTag)
		srcImages = append(srcImages, manifesttypes.ManifestEntry{
			Image:    e.Image,
			Platform: e.Platform,
		})
	}

	return manifesttypes.YAMLInput{
		Image:     image,
		Tags:      tags,
		Manifests: srcImages,
	}
}
