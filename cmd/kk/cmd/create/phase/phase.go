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

package phase

import (
	"github.com/spf13/cobra"
)

func NewPhaseCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "phase",
		Short: "KubeKey create phase",
		Long:  `This is the create phase run cmd`,
	}
	cmds.AddCommand(NewCmdCreateBinary())
	cmds.AddCommand(NewCmdConfigOS())
	cmds.AddCommand(NewCmdCreateImages())
	cmds.AddCommand(NewCmdCreateEtcd())
	cmds.AddCommand(NewCmdCreateInitCluster())
	cmds.AddCommand(NewCmdCreateJoinNodes())
	cmds.AddCommand(NewCmdCreateConfigureKubernetes())
	cmds.AddCommand(NewCmdCreateKubeSphere())
	cmds.AddCommand(NewCmdApplyAddons())

	return cmds
}
