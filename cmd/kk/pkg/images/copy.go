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
	"encoding/json"
	"fmt"
	"github.com/containerd/containerd/platforms"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"helm.sh/helm/v3/pkg/registry"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
	"os"
)

const (
	MediaTypeConfig       = "application/vnd.docker.container.image.v1+json"
	MediaTypeManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"
	MediaTypeManifest     = "application/vnd.docker.distribution.manifest.v2+json"
	MediaTypeForeignLayer = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip"
)

type CopyImageOptions struct {
	srcImage           *srcImageOptions
	destImage          *destImageOptions
	imageListSelection copy.ImageListSelection
}

func (c *CopyImageOptions) Copy() error {
	policyContext, err := getPolicyContext()
	if err != nil {
		return err
	}
	defer policyContext.Destroy()

	srcRef, err := alltransports.ParseImageName(c.srcImage.imageName)
	if err != nil {
		return err
	}
	destRef, err := alltransports.ParseImageName(c.destImage.imageName)
	if err != nil {
		return err
	}

	srcContext := c.srcImage.systemContext()
	destContext := c.destImage.systemContext()

	_, err = copy.Image(context.Background(), policyContext, destRef, srcRef, &copy.Options{
		ReportWriter:       os.Stdout,
		SourceCtx:          srcContext,
		DestinationCtx:     destContext,
		ImageListSelection: c.imageListSelection,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *CopyImageOptions) OrasCopy() error {
	srcTagger, err := c.srcImage.NewSrcTagger()
	if err != nil {
		return err
	}
	destTagger, err := c.destImage.NewDestTagger()
	if err != nil {
		return err
	}
	options := oras.DefaultCopyOptions

	options.MapRoot = func(ctx context.Context, src content.ReadOnlyStorage, root ocispec.Descriptor) (ocispec.Descriptor, error) {
		return mapRoot(ctx, src, root, ocispec.Platform{
			Architecture: c.srcImage.dockerImage.arch,
			OS:           c.srcImage.dockerImage.os,
			Variant:      c.srcImage.dockerImage.variant,
		})
	}

	_, err = oras.Copy(context.TODO(), srcTagger, c.srcImage.imageName, destTagger, c.destImage.imageName, options)
	if err != nil {
		return err
	}

	return nil
}

func getPolicyContext() (*signature.PolicyContext, error) {
	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	return signature.NewPolicyContext(policy)
}

type Index struct {
	Manifests []Manifest
}

type Manifest struct {
	Annotations annotations
	Platform    ocispec.Platform `json:"platform,omitempty"`
}

type annotations struct {
	RefName string `json:"org.opencontainers.image.ref.name"`
}

func NewIndex() *Index {
	return &Index{
		Manifests: []Manifest{},
	}
}

func mapRoot(ctx context.Context, src content.ReadOnlyStorage, root ocispec.Descriptor, p ocispec.Platform) (ocispec.Descriptor, error) {

	all, err := content.FetchAll(ctx, src, root)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	switch root.MediaType {
	case MediaTypeManifestList, ocispec.MediaTypeImageIndex:
		var index ocispec.Index
		if err := json.Unmarshal(all, &index); err != nil {
			return ocispec.Descriptor{}, err
		}
		for _, manifest := range index.Manifests {
			matcher := platforms.NewMatcher(*manifest.Platform)
			if matcher.Match(p) {
				return manifest, nil
			}
		}
		return ocispec.Descriptor{}, fmt.Errorf("%s: %w: no matching manifest was found in the manifest list", root.Digest, errdef.ErrNotFound)
	default:
		var configDesc ocispec.Manifest
		err = json.Unmarshal(all, &configDesc)
		if err != nil {
			return ocispec.Descriptor{}, err
		}

		if configDesc.Config.MediaType == registry.ConfigMediaType {
			return root, nil
		}

		fetch, err := src.Fetch(ctx, configDesc.Config)
		if err != nil {
			return ocispec.Descriptor{}, err
		}
		var configPlat ocispec.Platform
		err = json.NewDecoder(fetch).Decode(&configPlat)
		if err != nil {
			return ocispec.Descriptor{}, err
		}
		matcher := platforms.NewMatcher(configPlat)
		if matcher.Match(p) {
			return root, nil
		}

		return ocispec.Descriptor{}, fmt.Errorf("fail to recognize platform from unknown config %s", root.MediaType)
	}

}
