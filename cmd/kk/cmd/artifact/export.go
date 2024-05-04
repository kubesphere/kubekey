/*
 Copyright 2021 The KubeSphere Authors.

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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/pipelines"
)

type ArtifactExportOptions struct {
	CommonOptions *options.CommonOptions

	ManifestFile       string
	Output             string
	CriSocket          string
	DownloadCmd        string
	ImageTransport     string
	SkipRemoveArtifact bool
}

func NewArtifactExportOptions() *ArtifactExportOptions {
	return &ArtifactExportOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdArtifactExport creates a new `kubekey artifact export` command
func NewCmdArtifactExport() *cobra.Command {
	o := NewArtifactExportOptions()
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a KubeKey offline installation package",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(cmd, args))
			util.CheckErr(o.Validate(args))
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *ArtifactExportOptions) Complete(_ *cobra.Command, _ []string) error {
	if o.Output == "" {
		o.Output = "kubekey-artifact.tar.gz"
	}
	return nil
}

func (o *ArtifactExportOptions) Validate(_ []string) error {
	if o.ManifestFile == "" {
		return fmt.Errorf("--manifest can not be an empty string")
	}
	return nil
}

func (o *ArtifactExportOptions) Run() error {
	arg := common.ArtifactArgument{
		ManifestFile:       o.ManifestFile,
		Output:             o.Output,
		CriSocket:          o.CriSocket,
		ImageTransport:     o.ImageTransport,
		Debug:              o.CommonOptions.Verbose,
		IgnoreErr:          o.CommonOptions.IgnoreErr,
		SkipRemoveArtifact: o.SkipRemoveArtifact,
	}

	return pipelines.ArtifactExport(arg, o.DownloadCmd)
}

func (o *ArtifactExportOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ManifestFile, "manifest", "m", "", "Path to a manifest file")
	cmd.Flags().StringVarP(&o.Output, "output", "o", "", "Path to a output path")
	cmd.Flags().StringVarP(&o.DownloadCmd, "download-cmd", "", "curl -L -o %s %s",
		`The user defined command to download the necessary binary files. The first param '%s' is output path, the second param '%s', is the URL`)
	cmd.Flags().StringVarP(&o.ImageTransport, "image-transport", "", "", "Image transport to pull from, take values from [docker, docker-daemon]")
	cmd.Flags().BoolVarP(&o.SkipRemoveArtifact, "skip-remove-artifact", "", false, "Skip remove artifact")

}
