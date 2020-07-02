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
	"github.com/kubesphere/kubekey/pkg/config"
	"github.com/spf13/cobra"
)

var addons, name, clusterCfgPath string

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Create cluster configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		var ksVersion string
		if kubesphere && len(args) > 0 {
			ksVersion = args[0]
		} else {
			ksVersion = ""
		}
		err := config.GenerateClusterObj(addons, kubernetes, ksVersion, name, clusterCfgPath, kubesphere)
		if err != nil {
			return err
		}
		return err
	},
}

func init() {
	createCmd.AddCommand(configCmd)
	configCmd.Flags().StringVarP(&addons, "with-storage", "", "", "Add storage plugins")
	configCmd.Flags().StringVarP(&name, "name", "", "config-sample", "Specify a name of cluster object")
	configCmd.Flags().StringVarP(&clusterCfgPath, "file", "f", "", "Specify a configuration file path")
	configCmd.Flags().StringVarP(&kubernetes, "with-kubernetes", "", "v1.17.8", "Specify a supported version of kubernetes")
	configCmd.Flags().BoolVarP(&kubesphere, "with-kubesphere", "", false, "Deploy a specific version of kubesphere (default v3.0.0)")
}
