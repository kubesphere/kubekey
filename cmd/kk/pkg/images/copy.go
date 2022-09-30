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
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"os"
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

func getPolicyContext() (*signature.PolicyContext, error) {
	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	return signature.NewPolicyContext(policy)
}

type Index struct {
	Manifests []Manifest
}

type Manifest struct {
	Annotations annotations
}

type annotations struct {
	RefName string `json:"org.opencontainers.image.ref.name"`
}

func NewIndex() *Index {
	return &Index{
		Manifests: []Manifest{},
	}
}
