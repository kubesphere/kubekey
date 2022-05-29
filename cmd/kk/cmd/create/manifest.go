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

package create

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/cmd/kk/internal/artifact"
	"github.com/kubesphere/kubekey/cmd/kk/internal/common"
)

type CreateManifestOptions struct {
	CommonOptions *options.CommonOptions

	Name       string
	KubeConfig string
	FileName   string
}

func NewCreateManifestOptions() *CreateManifestOptions {
	return &CreateManifestOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCreateManifest creates a create manifest command
func NewCmdCreateManifest() *cobra.Command {
	o := NewCreateManifestOptions()
	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "Create an offline installation package configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(cmd, args))
			util.CheckErr(o.Run())

		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *CreateManifestOptions) Complete(cmd *cobra.Command, args []string) error {
	if o.Name != "" {
		o.Name = strings.Split(o.Name, ".")[0]
	}
	if o.KubeConfig == "" {
		o.KubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	if o.FileName == "" {
		currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return errors.Wrap(err, "Failed to get current dir")
		}
		o.FileName = filepath.Join(currentDir, fmt.Sprintf("manifest-%s.yaml", o.Name))
	}
	return nil
}

func (o *CreateManifestOptions) Run() error {
	arg := common.Argument{
		FilePath:   o.FileName,
		KubeConfig: o.KubeConfig,
	}
	return artifact.CreateManifest(arg, o.Name)
}

func (o *CreateManifestOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Name, "name", "", "sample", "Specify a name of manifest object")
	cmd.Flags().StringVarP(&o.FileName, "filename", "f", "", "Specify a manifest file path")
	cmd.Flags().StringVar(&o.KubeConfig, "kubeconfig", "", "Specify a kubeconfig file")
}
