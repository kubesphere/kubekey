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

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Create a Kubernetes or KubeSphere cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		var ksVersion string
		if opt.Kubesphere && len(args) > 0 {
			ksVersion = args[0]
		} else {
			ksVersion = ""
		}
		logger := util.InitLogger(opt.Verbose)
		return install.CreateCluster(opt.ClusterCfgFile, opt.Kubernetes, ksVersion, logger, opt.Kubesphere, opt.Verbose, opt.SkipCheck, opt.SkipPullImages)
	},
}

func init() {
	createCmd.AddCommand(clusterCmd)

	clusterCmd.Flags().StringVarP(&opt.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	clusterCmd.Flags().StringVarP(&opt.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
	clusterCmd.Flags().BoolVarP(&opt.Kubesphere, "with-kubesphere", "", false, "Deploy a specific version of kubesphere (default v3.0.0)")
	clusterCmd.Flags().BoolVarP(&opt.SkipCheck, "yes", "y", false, "Skip pre-check of the installation")
	clusterCmd.Flags().BoolVarP(&opt.SkipPullImages, "skip-pull-images", "", false, "Skip pre pull images")
}
