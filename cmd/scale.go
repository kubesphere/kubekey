/*
Copyright 2020 The KubeSphere Authors.

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

package cmd

import (
	"github.com/kubesphere/kubekey/pkg/install"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

func NewCmdScaleCluster() *cobra.Command {
	var (
		clusterCfgFile string
		//pkgDir         string
		verbose bool
	)
	var clusterCmd = &cobra.Command{
		Use:   "scale",
		Short: "Scale a cluster according to the new nodes information from the specified configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := util.InitLogger(verbose)
			//return scale.ScaleCluster(clusterCfgFile, logger, pkgDir, Verbose)
			return install.CreateCluster(clusterCfgFile, logger, false, verbose)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "file", "f", "", "configuration file name")
	//clusterCmd.Flags().StringVarP(&pkgDir, "pkg", "", "", "release package (offline)")
	clusterCmd.Flags().BoolVarP(&verbose, "debug", "", true, "")
	return clusterCmd
}
