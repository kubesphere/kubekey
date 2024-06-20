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
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

func newRunCommand() *cobra.Command {
	o := options.NewKubeKeyRunOptions()

	cmd := &cobra.Command{
		Use:   "run [playbook]",
		Short: "run a playbook by playbook file. the file source can be git or local",
		RunE: func(cmd *cobra.Command, args []string) error {
			kk, config, inventory, err := o.Complete(cmd, args)
			if err != nil {
				return err
			}
			// set workdir
			_const.SetWorkDir(o.WorkDir)
			// create workdir directory,if not exists
			if _, err := os.Stat(o.WorkDir); os.IsNotExist(err) {
				if err := os.MkdirAll(o.WorkDir, fs.ModePerm); err != nil {
					return err
				}
			}
			return run(signals.SetupSignalHandler(), kk, config, inventory)
		},
	}

	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}
	return cmd
}

func run(ctx context.Context, pipeline *kubekeyv1.Pipeline, config *kubekeyv1.Config, inventory *kubekeyv1.Inventory) error {
	restconfig, err := proxy.NewConfig(&rest.Config{})
	if err != nil {
		return fmt.Errorf("could not get rest config: %w", err)
	}
	client, err := ctrlclient.New(restconfig, ctrlclient.Options{
		Scheme: _const.Scheme,
	})
	if err != nil {
		return fmt.Errorf("could not get runtime-client: %w", err)
	}

	// create config, inventory and pipeline
	if err := client.Create(ctx, config); err != nil {
		klog.ErrorS(err, "Create config error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		return err
	}
	if err := client.Create(ctx, inventory); err != nil {
		klog.ErrorS(err, "Create inventory error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		return err
	}
	pipeline.Status.Phase = kubekeyv1.PipelinePhaseRunning
	if err := client.Create(ctx, pipeline); err != nil {
		klog.ErrorS(err, "Create pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))
		return err
	}

	return manager.NewCommandManager(manager.CommandManagerOptions{
		Pipeline:  pipeline,
		Config:    config,
		Inventory: inventory,
		Client:    client,
	}).Run(ctx)
}
