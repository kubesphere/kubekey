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

package manager

import (
	"context"
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
)

// Manager defines the interface for different types of managers that can run operations
type Manager interface {
	// Run executes the manager's main functionality with the given context
	Run(ctx context.Context) error
}

// CommandManagerOptions contains the configuration options for creating a new command manager
type CommandManagerOptions struct {
	*kkcorev1.Playbook
	*kkcorev1.Config
	*kkcorev1.Inventory

	ctrlclient.Client
}

// NewCommandManager creates and returns a new command manager instance with the provided options
func NewCommandManager(o CommandManagerOptions) Manager {
	return &commandManager{
		Playbook:  o.Playbook,
		Inventory: o.Inventory,
		Client:    o.Client,
		logOutput: os.Stdout,
	}
}

// NewControllerManager creates and returns a new controller manager instance with the provided options
func NewControllerManager(o *options.ControllerManagerServerOptions) Manager {
	return &controllerManager{
		ControllerManagerServerOptions: o,
	}
}

// WebManagerOptions contains the configuration options for creating a new web manager
type WebManagerOptions struct {
	Workdir    string
	Port       int
	SchemaPath string
	UIPath     string
	ctrlclient.Client
	*rest.Config
}

// NewWebManager creates and returns a new web manager instance with the provided options
func NewWebManager(o WebManagerOptions) Manager {
	return &webManager{
		workdir:    o.Workdir,
		port:       o.Port,
		schemaPath: o.SchemaPath,
		uiPath:     o.UIPath,
		Client:     o.Client,
		Config:     o.Config,
	}
}
