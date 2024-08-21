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
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

// ctx cancel by shutdown signal
var ctx = signals.SetupSignalHandler()

var internalCommand = make([]*cobra.Command, 0)

func registerInternalCommand(command *cobra.Command) {
	for _, c := range internalCommand {
		if c.Name() == command.Name() {
			// command has register. skip
			return
		}
	}
	internalCommand = append(internalCommand, command)
}

// NewRootCommand console command.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "kk",
		Long: "kubekey is a daemon that execute command in a node",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if err := options.InitGOPS(); err != nil {
				return err
			}

			return options.InitProfiling(ctx)
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return options.FlushProfiling()
		},
	}
	cmd.SetContext(ctx)

	// add common flag
	flags := cmd.PersistentFlags()
	options.AddProfilingFlags(flags)
	options.AddKlogFlags(flags)
	options.AddGOPSFlags(flags)

	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newPipelineCommand())
	cmd.AddCommand(newVersionCommand())
	// internal command
	cmd.AddCommand(internalCommand...)

	return cmd
}
