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

package cri

import (
	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/spf13/cobra"
)

type MigrateCriOptions struct {
	CommonOptions    *options.CommonOptions
	ClusterCfgFile   string
	Kubernetes       string
	EnableKubeSphere bool
	KubeSphere       string
	DownloadCmd      string
	Artifact         string
	Type             string
	Role             string
}

func NewMigrateCriOptions() *MigrateCriOptions {
	return &MigrateCriOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdDeleteCluster creates a new delete cluster command
func NewCmdMigrateCri() *cobra.Command {
	o := NewMigrateCriOptions()
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate a container",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *MigrateCriOptions) Run() error {
	arg := common.Argument{
		FilePath:          o.ClusterCfgFile,
		Debug:             o.CommonOptions.Verbose,
		KubernetesVersion: o.Kubernetes,
		Type:              o.Type,
		Role:              o.Role,
	}
	return pipelines.MigrateCri(arg, o.DownloadCmd)
}

func (o *MigrateCriOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Role, "role", "", "", "Role groups for migrating. Support: master, worker, all.")
	cmd.Flags().StringVarP(&o.Type, "type", "", "", "Type of target CRI. Support: docker, containerd.")
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
	cmd.Flags().StringVarP(&o.DownloadCmd, "download-cmd", "", "curl -L -o %s %s",
		`The user defined command to download the necessary binary files. The first param '%s' is output path, the second param '%s', is the URL`)
	cmd.Flags().StringVarP(&o.Artifact, "artifact", "a", "", "Path to a KubeKey artifact")
}

func (o *MigrateCriOptions) Validate() error {
	if o.Role == "" {
		return errors.New("node Role can not be empty")
	}
	if o.Role != common.Worker && o.Role != common.Master && o.Role != "all" {

		return errors.Errorf("node Role is invalid: %s", o.Role)
	}
	if o.Type == "" {
		return errors.New("cri Type can not be empty")
	}
	if o.Type != common.Docker && o.Type != common.Conatinerd {
		return errors.Errorf("cri Type is invalid: %s", o.Type)
	}
	if o.ClusterCfgFile == "" {
		return errors.New("configuration file can not be empty")
	}
	return nil
}
