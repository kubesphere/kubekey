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

	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
)

// Manager shared dependencies such as Addr and , and provides them to Runnable.
type Manager interface {
	// Run the driver
	Run(ctx context.Context) error
}

// CommandManagerOptions for NewCommandManager
type CommandManagerOptions struct {
	*kkcorev1.Pipeline
	*kkcorev1.Config
	*kkcorev1.Inventory

	ctrlclient.Client
}

// NewCommandManager return a new commandManager
func NewCommandManager(o CommandManagerOptions) Manager {
	return &commandManager{
		Pipeline:  o.Pipeline,
		Config:    o.Config,
		Inventory: o.Inventory,
		Client:    o.Client,
		logOutput: os.Stdout,
	}
}

// ControllerManagerOptions for NewControllerManager
type ControllerManagerOptions struct {
	MaxConcurrentReconciles int
	LeaderElection          bool
}

// NewControllerManager return a new controllerManager
func NewControllerManager(o ControllerManagerOptions) Manager {
	return &controllerManager{
		MaxConcurrentReconciles: o.MaxConcurrentReconciles,
		LeaderElection:          o.LeaderElection,
	}
}
