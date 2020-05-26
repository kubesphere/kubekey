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
	"github.com/kubesphere/kubekey/cmd/create"
	"github.com/spf13/cobra"
)

var Verbose bool

func NewKubekeyCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "kk",
		Short: "Kubernetes/KubeSphere Deploy Tool",
		Long: "Deploy a Kubernetes or KubeSphere cluster efficiently, flexibly and easily. There are three scenarios to use KubeKey. \n" +
			"1. Install Kubernetes only \n" +
			"2. Install Kubernetes and KubeSphere together in one command \n" +
			"3. Install Kubernetes first, then deploy KubeSphere on it using https://github.com/kubesphere/ks-installer",
	}

	rootCmd.AddCommand(create.NewCmdCreate())
	rootCmd.AddCommand(NewCmdScaleCluster())
	rootCmd.AddCommand(NewCmdVersion())
	rootCmd.AddCommand(NewCmdResetCluster())
	return rootCmd
}
