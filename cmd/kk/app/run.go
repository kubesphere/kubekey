/*
Copyright 2023 The KubeSphere Authors.

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

package app

import (
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

func newRunCommand() *cobra.Command {
	o := options.NewKubeKeyRunOptions()

	cmd := &cobra.Command{
		Use:   "run [playbook]",
		Short: "run a playbook by playbook file. the file source can be git or local",
		RunE: func(cmd *cobra.Command, args []string) error {
			playbook, err := o.Complete(cmd, args)
			if err != nil {
				return errors.WithStack(err)
			}

			return o.CommonOptions.Run(cmd.Context(), playbook)
		},
	}
	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}

	return cmd
}
