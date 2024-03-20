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
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/registry"
	"github.com/kubesphere/kubekey/v3/version"
	"net/http"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"os"
	"strings"

	"github.com/containerd/containerd/platforms"
	"github.com/containers/image/v5/types"
	manifesttypes "github.com/estesp/manifest-tool/v2/pkg/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
)

var defaultUserAgent = "kubekey"

type dockerImageOptions struct {
	arch           string
	os             string
	variant        string
	username       string
	password       string
	dockerCertPath string
	SkipTLSVerify  bool
}

func (d *dockerImageOptions) systemContext() *types.SystemContext {
	ctx := &types.SystemContext{
		ArchitectureChoice:          d.arch,
		OSChoice:                    d.os,
		VariantChoice:               d.variant,
		DockerRegistryUserAgent:     defaultUserAgent,
		DockerInsecureSkipTLSVerify: types.NewOptionalBool(d.SkipTLSVerify),
	}
	return ctx
}

func (d *dockerImageOptions) AuthClient() (*auth.Client, error) {
	config := &tls.Config{
		InsecureSkipVerify: d.SkipTLSVerify,
	}

	if d.dockerCertPath != "" {
		ca, cert, key, err := registry.LookupCertsFile(d.dockerCertPath)
		if err != nil {
			return nil, err
		}
		keyPair, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		config.Certificates = append(config.Certificates, keyPair)
		pool := x509.NewCertPool()
		file, err := os.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		pool.AppendCertsFromPEM(file)
		config.RootCAs = pool
	}

	client := &auth.Client{
		Client: &http.Client{Transport: &http.Transport{TLSClientConfig: config}},
		Cache:  auth.NewCache(),
		Credential: func(ctx context.Context, hostport string) (auth.Credential, error) {
			if d.username == "" && d.password == "" {
				return auth.EmptyCredential, nil
			}
			if d.username == "" {
				return auth.Credential{RefreshToken: d.password}, nil
			}
			return auth.Credential{
				Username: d.username,
				Password: d.password,
			}, nil
		},
	}
	client.SetUserAgent(fmt.Sprintf("%s/%s", defaultUserAgent, version.Get().String()))
	return client, nil
}

type srcImageOptions struct {
	dockerImage   dockerImageOptions
	imageName     string
	inputPath     string
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

func (s *srcImageOptions) NewSrcTagger() (oras.ReadOnlyGraphTarget, error) {
	if s.inputPath == "" {
		return s.dockerImage.NewRepository(s.imageName)
	}
	if s.inputPath != "" {
		_, err := os.Stat(s.inputPath)
		if err != nil {
			return nil, err
		}

		return oci.New(s.inputPath)

	}

	return nil, fmt.Errorf("unknown path type %s", s.imageName)
}

func (d *destImageOptions) NewDestTagger() (oras.GraphTarget, error) {
	if d.outputPath == "" {
		return d.dockerImage.NewRepository(d.imageName)
	}
	return oci.New(d.outputPath)

}

func (d *dockerImageOptions) NewRepository(path string) (*remote.Repository, error) {

	repository, err := remote.NewRepository(path)
	if err != nil {
		return nil, err
	}
	client, err := d.AuthClient()
	if err != nil {
		return nil, err
	}
	repository.Client = client

	return repository, nil

}

type destImageOptions struct {
	dockerImage dockerImageOptions
	imageName   string
	outputPath  string
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
//
//	Ex: localhost.localdomain:5000/samalba/hipache:latest
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

	for _, e := range entries {
		srcImages = append(srcImages, manifesttypes.ManifestEntry{
			Image:    e.Image,
			Platform: e.Platform,
		})
	}

	return manifesttypes.YAMLInput{
		Image:     image,
		Manifests: srcImages,
	}
}

func validateImageName(imageFullName string) error {
	image := strings.Split(imageFullName, "/")
	if len(image) < 3 {
		return errors.Errorf("image %s is invalid, image PATH need contain at least two slash-separated", imageFullName)
	}
	if len(strings.Split(image[len(image)-1], ":")) != 2 {
		return errors.Errorf(`image %s is invalid, image PATH need contain ":"`, imageFullName)
	}
	return nil
}

func suffixImageName(imageFullName []string) string {
	if len(imageFullName) >= 2 {
		return strings.Join(imageFullName, "/")
	}
	return imageFullName[0]
}
