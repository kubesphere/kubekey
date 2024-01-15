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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
)

func newRunCommand() *cobra.Command {
	o := options.NewKubeKeyRunOptions()

	cmd := &cobra.Command{
		Use:   "run [playbook]",
		Short: "run a playbook",
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.GOPSEnabled {
				// Add agent to report additional information such as the current stack trace, Go version, memory stats, etc.
				// Bind to a random port on address 127.0.0.1
				if err := agent.Listen(agent.Options{}); err != nil {
					return err
				}
			}
			kk, err := o.Complete(cmd, args)
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
			// convert option to kubekeyv1.Pipeline
			return run(signals.SetupSignalHandler(), kk, o.ConfigFile, o.InventoryFile)
		},
	}

	fs := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		fs.AddFlagSet(f)
	}
	return cmd
}

func run(ctx context.Context, kk *kubekeyv1.Pipeline, configFile string, inventoryFile string) error {
	// convert configFile
	config := &kubekeyv1.Config{}
	cdata, err := os.ReadFile(configFile)
	if err != nil {
		klog.Errorf("read config file error %v", err)
		return err
	}
	if err := yaml.Unmarshal(cdata, config); err != nil {
		klog.Errorf("unmarshal config file error %v", err)
		return err
	}
	if config.Namespace == "" {
		config.Namespace = corev1.NamespaceDefault
	}
	kk.Spec.ConfigRef = &corev1.ObjectReference{
		Kind:            config.Kind,
		Namespace:       config.Namespace,
		Name:            config.Name,
		UID:             config.UID,
		APIVersion:      config.APIVersion,
		ResourceVersion: config.ResourceVersion,
	}

	// convert inventoryFile
	inventory := &kubekeyv1.Inventory{}
	idata, err := os.ReadFile(inventoryFile)
	if err := yaml.Unmarshal(idata, inventory); err != nil {
		klog.Errorf("unmarshal inventory file error %v", err)
		return err
	}
	if inventory.Namespace == "" {
		inventory.Namespace = corev1.NamespaceDefault
	}
	kk.Spec.InventoryRef = &corev1.ObjectReference{
		Kind:            inventory.Kind,
		Namespace:       inventory.Namespace,
		Name:            inventory.Name,
		UID:             inventory.UID,
		APIVersion:      inventory.APIVersion,
		ResourceVersion: inventory.ResourceVersion,
	}
	return manager.NewCommandManager(manager.CommandManagerOptions{
		Pipeline:  kk,
		Config:    config,
		Inventory: inventory,
	}).Run(ctx)
}
