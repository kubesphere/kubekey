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

var (
	clusterCfgFile string
	kubernetes     string
	kubesphere     bool
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Create a Kubernetes or KubeSphere cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		var ksVersion string
		if kubesphere && len(args) > 0 {
			ksVersion = args[0]
		} else {
			ksVersion = ""
		}
		logger := util.InitLogger(verbose)
		return install.CreateCluster(clusterCfgFile, kubernetes, ksVersion, logger, kubesphere, verbose)
	},
}

func init() {
	createCmd.AddCommand(clusterCmd)

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "file", "f", "", "Path to a configuration file")
	clusterCmd.Flags().StringVarP(&kubernetes, "with-kubernetes", "", "v1.17.6", "Specify a supported version of kubernetes")
	clusterCmd.Flags().BoolVarP(&kubesphere, "with-kubesphere", "", false, "Deploy a specific version of kubesphere (default v3.0.0)")
}
