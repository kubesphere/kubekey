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
	"github.com/kubesphere/kubekey/pkg/scale"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:   "scale",
	Short: "Scale a cluster according to the new nodes information from the specified configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := util.InitLogger(opt.Verbose)
		return scale.ScaleCluster(opt.ClusterCfgFile, "", "", logger, false, opt.Verbose, opt.SkipCheck)
	},
}

func init() {
	rootCmd.AddCommand(scaleCmd)
	scaleCmd.Flags().StringVarP(&opt.ClusterCfgFile, "file", "f", "", "Path to a configuration file")
	scaleCmd.Flags().BoolVarP(&opt.SkipCheck, "yes", "y", false, "Skip pre-check of the installation")
}
