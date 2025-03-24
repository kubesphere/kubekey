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
	"path/filepath"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

var internalCommand = make([]*cobra.Command, 0)

// NewRootCommand console command.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "kk",
		Long: "kubekey is a daemon that execute command in a node",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := options.InitGOPS(); err != nil {
				return errors.WithStack(err)
			}

			return options.InitProfiling(cmd.Context())
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			return options.FlushProfiling()
		},
	}
	cmd.SetContext(signals.SetupSignalHandler())
	// add common flag
	flags := cmd.PersistentFlags()
	options.AddProfilingFlags(flags)
	options.AddKlogFlags(flags)
	options.AddGOPSFlags(flags)
	// add children command
	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newPlaybookCommand())
	cmd.AddCommand(newVersionCommand())
	// internal command
	cmd.AddCommand(internalCommand...)

	return cmd
}

// CommandRunE executes the main command logic for the application.
// It sets up the necessary configurations, creates the inventory and playbook
// resources, and then runs the command manager.
//
// Parameters:
//   - ctx: The context for controlling the execution flow.
//   - workdir: The working directory path.
//   - playbook: The playbook resource to be created and managed.
//   - config: The configuration resource.
//   - inventory: The inventory resource to be created.
//
// Returns:
//   - error: An error if any step in the process fails, otherwise nil.
func CommandRunE(ctx context.Context, workdir string, playbook *kkcorev1.Playbook, config *kkcorev1.Config, inventory *kkcorev1.Inventory) error {
	restconfig := &rest.Config{}
	if err := proxy.RestConfig(filepath.Join(workdir, _const.RuntimeDir), restconfig); err != nil {
		return errors.Wrap(err, "failed to get restconfig")
	}
	client, err := ctrlclient.New(restconfig, ctrlclient.Options{
		Scheme: _const.Scheme,
	})
	if err != nil {
		return errors.Wrap(err, "failed to get runtime-client")
	}
	// create inventory
	if err := client.Create(ctx, inventory); err != nil {
		return errors.Wrap(err, "failed to create inventory")
	}
	// create playbook
	if err := client.Create(ctx, playbook); err != nil {
		return errors.Wrap(err, "failed to create playbook")
	}

	return manager.NewCommandManager(manager.CommandManagerOptions{
		Workdir:   workdir,
		Playbook:  playbook,
		Config:    config,
		Inventory: inventory,
		Client:    client,
	}).Run(ctx)
}
