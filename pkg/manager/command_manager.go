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
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	"github.com/kubesphere/kubekey/v4/pkg/cache"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/controllers"
	"github.com/kubesphere/kubekey/v4/pkg/task"
)

type commandManager struct {
	*kubekeyv1.Pipeline
	*kubekeyv1.Config
	*kubekeyv1.Inventory

	ctrlclient.Client
}

func (m *commandManager) Run(ctx context.Context) error {
	// create config, inventory and pipeline
	klog.Infof("[Pipeline %s] start", ctrlclient.ObjectKeyFromObject(m.Pipeline))
	if err := m.Client.Create(ctx, m.Config); err != nil {
		klog.Errorf("[Pipeline %s] create config error: %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), err)
		return err
	}
	if err := m.Client.Create(ctx, m.Inventory); err != nil {
		klog.Errorf("[Pipeline %s] create inventory error: %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), err)
		return err
	}
	if err := m.Client.Create(ctx, m.Pipeline); err != nil {
		klog.Errorf("[Pipeline %s] create pipeline error: %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), err)
		return err
	}

	defer func() {
		// update pipeline status
		if err := m.Client.Update(ctx, m.Pipeline); err != nil {
			klog.Errorf("[Pipeline %s] update pipeline error: %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), err)
		}

		klog.Infof("[Pipeline %s] finish", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		if !m.Pipeline.Spec.Debug && m.Pipeline.Status.Phase == kubekeyv1.PipelinePhaseSucceed {
			klog.Infof("[Pipeline %s] clean runtime directory", ctrlclient.ObjectKeyFromObject(m.Pipeline))
			// clean runtime directory
			if err := os.RemoveAll(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir)); err != nil {
				klog.Errorf("clean runtime directory %s error: %v", filepath.Join(_const.GetWorkDir(), _const.RuntimeDir), err)
			}
		}
	}()

	klog.Infof("[Pipeline %s] start task controller", ctrlclient.ObjectKeyFromObject(m.Pipeline))
	kd, err := task.NewController(task.ControllerOptions{
		Client: m.Client,
		TaskReconciler: &controllers.TaskReconciler{
			Client:        m.Client,
			VariableCache: cache.LocalVariable,
		},
	})
	if err != nil {
		klog.Errorf("[Pipeline %s] create task controller error: %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), err)
		m.Pipeline.Status.Phase = kubekeyv1.PipelinePhaseFailed
		m.Pipeline.Status.Reason = fmt.Sprintf("create task controller failed: %v", err)
		return err
	}
	// init pipeline status
	m.Pipeline.Status.Phase = kubekeyv1.PipelinePhaseRunning
	if err := kd.AddTasks(ctx, task.AddTaskOptions{
		Pipeline: m.Pipeline,
	}); err != nil {
		klog.Errorf("[Pipeline %s] add task error: %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), err)
		m.Pipeline.Status.Phase = kubekeyv1.PipelinePhaseFailed
		m.Pipeline.Status.Reason = fmt.Sprintf("add task to controller failed: %v", err)
		return err
	}
	// update pipeline status
	if err := m.Client.Update(ctx, m.Pipeline); err != nil {
		klog.Errorf("[Pipeline %s] update pipeline error: %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), err)
	}
	go kd.Start(ctx)

	_ = wait.PollUntilContextCancel(ctx, time.Millisecond*100, false, func(ctx context.Context) (done bool, err error) {
		if err := m.Client.Get(ctx, ctrlclient.ObjectKeyFromObject(m.Pipeline), m.Pipeline); err != nil {
			klog.Errorf("[Pipeline %s] get pipeline error: %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), err)
		}
		if m.Pipeline.Status.Phase == kubekeyv1.PipelinePhaseFailed || m.Pipeline.Status.Phase == kubekeyv1.PipelinePhaseSucceed {
			return true, nil
		}
		return false, nil
	})
	// kill by signal
	if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
		klog.Errorf("[Pipeline %s] manager terminated error: %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), err)
		return err
	}
	klog.Infof("[Pipeline %s] task finish", ctrlclient.ObjectKeyFromObject(m.Pipeline))

	return nil
}
