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
	"context"
	"io/fs"
	"os"

	"github.com/google/gops/agent"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
)

func NewControllerManagerCommand() *cobra.Command {
	o := options.NewControllerManagerServerOptions()

	cmd := &cobra.Command{
		Use:   "controller-manager",
		Short: "kubekey controller manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.GOPSEnabled {
				// Add agent to report additional information such as the current stack trace, Go version, memory stats, etc.
				// Bind to a random port on address 127.0.0.1
				if err := agent.Listen(agent.Options{}); err != nil {
					return err
				}
			}

			o.Complete(cmd, args)
			// create workdir directory,if not exists
			_const.SetWorkDir(o.WorkDir)
			if _, err := os.Stat(o.WorkDir); os.IsNotExist(err) {
				if err := os.MkdirAll(o.WorkDir, fs.ModePerm); err != nil {
					return err
				}
			}
			return run(signals.SetupSignalHandler(), o)
		},
	}

	fs := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		fs.AddFlagSet(f)
	}
	return cmd
}

func run(ctx context.Context, o *options.ControllerManagerServerOptions) error {
	return manager.NewControllerManager(manager.ControllerManagerOptions{
		ControllerGates:         o.ControllerGates,
		MaxConcurrentReconciles: o.MaxConcurrentReconciles,
		LeaderElection:          o.LeaderElection,
	}).Run(ctx)
}
