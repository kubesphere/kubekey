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
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/container"
	coreutil "github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/spf13/cobra"
)

type ArtifactExportOptions struct {
	CommonOptions *options.CommonOptions

	ManifestFile string
	Output       string
	CriSocket    string
	DownloadCmd  string
}

func NewArtifactExportOptions() *ArtifactExportOptions {
	return &ArtifactExportOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCreateCluster creates a new create cluster command
func NewCmdCreateCluster() *cobra.Command {
	o := NewArtifactExportOptions()
	cmd := &cobra.Command{
		Use:   "export",
		Short: "export a KubeKey artifact",
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

func (o *ArtifactExportOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error
	if o.Output == "" {
		o.Output = "kubekey-artifact.tar.gz"
	}
	if o.CriSocket == "" {
		o.CriSocket, err = container.DetectCRISocket()
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *ArtifactExportOptions) Validate(args []string) error {
	if o.ManifestFile == "" {
		return fmt.Errorf("--manifest can not be an empty string")
	}
	if !coreutil.IsExist(o.CriSocket) {
		return fmt.Errorf("can not found the socket file %s", o.CriSocket)
	}
	return nil
}

func (o *ArtifactExportOptions) Run() error {
	arg := common.ArtifactArgument{
		ManifestFile: o.ManifestFile,
		Output:       o.Output,
		CriSocket:    o.CriSocket,
		Debug:        o.CommonOptions.Verbose,
		IgnoreErr:    o.CommonOptions.IgnoreErr,
	}

	return pipelines.ArtifactExport(arg, o.DownloadCmd)
}

func (o *ArtifactExportOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ManifestFile, "manifest", "m", "", "Path to a manifest file")
	cmd.Flags().StringVarP(&o.Output, "output", "o", "", "Path to a output path")
	cmd.Flags().StringVar(&o.CriSocket, "cri-socket", "", "Path to the CRI socket to connect. If empty KubeKey will try to auto-detect this value")
	cmd.Flags().StringVarP(&o.DownloadCmd, "download-cmd", "", "curl -L -o %s %s",
		`The user defined command to download the necessary binary files. The first param '%s' is output path, the second param '%s', is the URL`)
}
