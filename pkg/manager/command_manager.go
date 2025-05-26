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
	"io"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/executor"
)

type commandManager struct {
	*kkcorev1.Playbook
	*kkcorev1.Inventory

	ctrlclient.Client

	logOutput io.Writer
}

// Run command Manager. print log and run playbook executor.
func (m *commandManager) Run(ctx context.Context) error {
	return executor.NewPlaybookExecutor(ctx, m.Client, m.Playbook, m.logOutput).Exec(ctx)
}
