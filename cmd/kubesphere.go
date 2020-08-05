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
	"github.com/kubesphere/kubekey/pkg/deploy"
	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var ksCmd = &cobra.Command{
	Use:   "kubesphere",
	Short: "Deploy KubeSphere on the existing K8s",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deploy.DeployKubeSphere(opt.KsVersion, opt.Registry, opt.Kubeconfig)
	},
}

func init() {
	deployCmd.AddCommand(ksCmd)

	ksCmd.Flags().StringVarP(&opt.Kubeconfig, "kubeconfig", "", "", "Specify a kubeconfig file")
	ksCmd.Flags().StringVarP(&opt.KsVersion, "version", "v", "v3.0.0", "Specify a supported version of kubesphere")
	ksCmd.Flags().StringVarP(&opt.Registry, "registry", "", "", "Specify a image registry address")
}
