/*
 Copyright 2021 The KubeSphere Authors.

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

package plugin

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/options"
)

var (
	ValidPluginFilenamePrefixes = []string{"kk"}
)

func NewCmdPlugin(streams options.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "plugin [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "Provides utilities for interacting with plugins",
	}

	cmd.AddCommand(NewCmdPluginList(streams))
	return cmd
}
