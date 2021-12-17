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

package create

import (
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/spf13/cobra"
)

type CreateOptions struct {
	CommonOptions *options.CommonOptions
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCreate creates a new create command
func NewCmdCreate() *cobra.Command {
	o := NewCreateOptions()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a cluster or a cluster configuration file",
	}

	o.CommonOptions.AddCommonFlag(cmd)

	cmd.AddCommand(NewCmdCreateCluster())
	cmd.AddCommand(NewCmdCreateConfig())
	cmd.AddCommand(NewCmdCreateManifest())
	return cmd
}
