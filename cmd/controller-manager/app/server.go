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
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if err := options.InitGOPS(); err != nil {
				return err
			}
			return options.InitProfiling()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return options.FlushProfiling()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
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

	// add common flag
	flags := cmd.PersistentFlags()
	options.AddProfilingFlags(flags)
	options.AddKlogFlags(flags)
	options.AddGOPSFlags(flags)

	fs := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		fs.AddFlagSet(f)
	}

	cmd.AddCommand(newVersionCommand())
	return cmd
}

func run(ctx context.Context, o *options.ControllerManagerServerOptions) error {
	return manager.NewControllerManager(manager.ControllerManagerOptions{
		MaxConcurrentReconciles: o.MaxConcurrentReconciles,
		LeaderElection:          o.LeaderElection,
	}).Run(ctx)
}
