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
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v2/cmd/ctl/options"
)

type ArtifactImagesOptions struct {
	CommonOptions *options.CommonOptions
}

func NewArtifactImagesOptions() *ArtifactImagesOptions {
	return &ArtifactImagesOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdArtifactImages creates a new `kubekey artifact image` command
func NewCmdArtifactImages() *cobra.Command {
	o := NewArtifactImagesOptions()
	cmd := &cobra.Command{
		Use:     "images",
		Aliases: []string{"image", "i"},
		Short:   "manage KubeKey artifact image",
	}

	o.CommonOptions.AddCommonFlag(cmd)
	cmd.AddCommand(NewCmdArtifactImagesPush())

	return cmd
}
