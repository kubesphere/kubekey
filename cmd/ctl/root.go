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

package ctl

import (
	"github.com/kubesphere/kubekey/cmd/ctl/add"
	"github.com/kubesphere/kubekey/cmd/ctl/cert"
	"github.com/kubesphere/kubekey/cmd/ctl/completion"
	"github.com/kubesphere/kubekey/cmd/ctl/create"
	"github.com/kubesphere/kubekey/cmd/ctl/delete"
	initOs "github.com/kubesphere/kubekey/cmd/ctl/init"
	"github.com/kubesphere/kubekey/cmd/ctl/upgrade"
	"github.com/kubesphere/kubekey/cmd/ctl/version"
	"github.com/spf13/cobra"
)

func NewDefaultKubeKeyCommand() *cobra.Command {
	return NewDefaultKubeKeyCommandWithArgs()
}

func NewDefaultKubeKeyCommandWithArgs() *cobra.Command {
	cmd := NewKubeKeyCommand()

	return cmd
}

// NewKubeKeyCommand creates a new kubekey root command
func NewKubeKeyCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "kk",
		Short: "Kubernetes/KubeSphere Deploy Tool",
		Long: `Deploy a Kubernetes or KubeSphere cluster efficiently, flexibly and easily. There are three scenarios to use KubeKey.
1. Install Kubernetes only
2. Install Kubernetes and KubeSphere together in one command
3. Install Kubernetes first, then deploy KubeSphere on it using https://github.com/kubesphere/ks-installer`,
	}

	cmds.AddCommand(initOs.NewCmdInit())

	cmds.AddCommand(create.NewCmdCreate())
	cmds.AddCommand(delete.NewCmdDelete())
	cmds.AddCommand(add.NewCmdAdd())
	cmds.AddCommand(upgrade.NewCmdUpgrade())
	cmds.AddCommand(cert.NewCmdCerts())

	cmds.AddCommand(completion.NewCmdCompletion())
	cmds.AddCommand(version.NewCmdVersion())
	return cmds
}
