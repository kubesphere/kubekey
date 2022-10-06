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

package init

import (
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
)

type InitOptions struct {
	CommonOptions *options.CommonOptions
}

func NewInitOptions() *InitOptions {
	return &InitOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdInit create a new init command
func NewCmdInit() *cobra.Command {
	o := NewInitOptions()
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initializes the installation environment",
	}
	o.CommonOptions.AddCommonFlag(cmd)
	cmd.AddCommand(NewCmdInitOs())
	cmd.AddCommand(NewCmdInitRegistry())
	return cmd
}
