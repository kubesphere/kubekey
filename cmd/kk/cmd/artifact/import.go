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

package artifact

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/phase/artifact"
)

type ArtifactImportOptions struct {
	CommonOptions *options.CommonOptions
	Artifact      string
}

func NewArtifactImportOptions() *ArtifactImportOptions {
	return &ArtifactImportOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdArtifactImport creates a new artifact import command
func NewCmdArtifactImport() *cobra.Command {
	o := NewArtifactImportOptions()
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import a KubeKey offline installation package",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate(args))
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *ArtifactImportOptions) Run() error {
	arg := common.Argument{
		Debug:    o.CommonOptions.Verbose,
		Artifact: o.Artifact,
	}
	return artifact.ArtifactImport(arg)
}

func (o *ArtifactImportOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Artifact, "artifact", "a", "", "Path to a artifact gzip")
}

func (o *ArtifactImportOptions) Validate(_ []string) error {
	if o.Artifact == "" {
		return errors.New("artifact path can not be empty")
	}
	return nil
}
