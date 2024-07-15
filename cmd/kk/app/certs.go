//go:build builtin
// +build builtin

/*
Copyright 2024 The KubeSphere Authors.

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
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

func newCertsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certs",
		Short: "cluster certs",
	}

	cmd.AddCommand(newCertsRenewCommand())
	return cmd
}

func newCertsRenewCommand() *cobra.Command {
	o := options.NewCertsRenewOptions()
	cmd := &cobra.Command{
		Use:   "renew",
		Short: "renew a cluster certs",
		RunE: func(cmd *cobra.Command, args []string) error {
			pipeline, config, inventory, err := o.Complete(cmd, []string{"playbooks/certs_renew.yaml"})
			if err != nil {
				return err
			}
			// set workdir
			_const.SetWorkDir(o.WorkDir)
			// create workdir directory,if not exists
			if _, err := os.Stat(o.WorkDir); os.IsNotExist(err) {
				if err := os.MkdirAll(o.WorkDir, os.ModePerm); err != nil {
					return err
				}
			}
			return run(signals.SetupSignalHandler(), pipeline, config, inventory)
		},
	}

	flags := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		flags.AddFlagSet(f)
	}
	return cmd
}

func init() {
	registerInternalCommand(newCertsCommand())
}
