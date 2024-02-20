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
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/controllers"
	"github.com/kubesphere/kubekey/v4/pkg/task"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type commandManager struct {
	*kubekeyv1.Pipeline
	*kubekeyv1.Config
	*kubekeyv1.Inventory

	ctrlclient.Client
	*runtime.Scheme
}

func (m *commandManager) Run(ctx context.Context) error {
	// create config, inventory and pipeline
	if err := m.Client.Create(ctx, m.Config); err != nil {
		klog.ErrorS(err, "Create config error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		return err
	}
	if err := m.Client.Create(ctx, m.Inventory); err != nil {
		klog.ErrorS(err, "Create inventory error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		return err
	}
	if err := m.Client.Create(ctx, m.Pipeline); err != nil {
		klog.ErrorS(err, "Create pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		return err
	}
	klog.Infof("[Pipeline %s] start", ctrlclient.ObjectKeyFromObject(m.Pipeline))
	defer func() {
		klog.Infof("[Pipeline %s] finish", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		// update pipeline status
		if err := m.Client.Status().Update(ctx, m.Pipeline); err != nil {
			klog.ErrorS(err, "Update pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		}

		if !m.Pipeline.Spec.Debug && m.Pipeline.Status.Phase == kubekeyv1.PipelinePhaseSucceed {
			klog.Infof("[Pipeline %s] clean runtime directory", ctrlclient.ObjectKeyFromObject(m.Pipeline))
			// clean runtime directory
			if err := os.RemoveAll(_const.GetRuntimeDir()); err != nil {
				klog.ErrorS(err, "Clean runtime directory error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline), "runtime_dir", _const.GetRuntimeDir())
			}
		}
		// kill by signal
		if err := syscall.Kill(os.Getpid(), syscall.SIGTERM); err != nil {
			klog.ErrorS(err, "Kill process error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		}
	}()

	klog.Infof("[Pipeline %s] start task controller", ctrlclient.ObjectKeyFromObject(m.Pipeline))
	kd, err := task.NewController(task.ControllerOptions{
		Scheme:        m.Scheme,
		VariableCache: variable.Cache,
		Client:        m.Client,
		TaskReconciler: &controllers.TaskReconciler{
			Client: m.Client,
		},
	})
	if err != nil {
		klog.ErrorS(err, "Create task controller error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		m.Pipeline.Status.Phase = kubekeyv1.PipelinePhaseFailed
		m.Pipeline.Status.Reason = fmt.Sprintf("create task controller failed: %v", err)
		return err
	}
	// init pipeline status
	m.Pipeline.Status.Phase = kubekeyv1.PipelinePhaseRunning
	if err := kd.AddTasks(ctx, m.Pipeline); err != nil {
		klog.ErrorS(err, "Add task error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		m.Pipeline.Status.Phase = kubekeyv1.PipelinePhaseFailed
		m.Pipeline.Status.Reason = fmt.Sprintf("add task to controller failed: %v", err)
		return err
	}
	// update pipeline status
	if err := m.Client.Status().Update(ctx, m.Pipeline); err != nil {
		klog.ErrorS(err, "Update pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		return err
	}
	klog.Infof("[Pipeline %s] start deal task total %v", ctrlclient.ObjectKeyFromObject(m.Pipeline), m.Pipeline.Status.TaskResult.Total)
	go kd.Start(ctx)

	_ = wait.PollUntilContextCancel(ctx, time.Millisecond*100, false, func(ctx context.Context) (done bool, err error) {
		if err := m.Client.Get(ctx, ctrlclient.ObjectKeyFromObject(m.Pipeline), m.Pipeline); err != nil {
			klog.ErrorS(err, "Get pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
			return false, nil
		}
		if m.Pipeline.Status.Phase == kubekeyv1.PipelinePhaseFailed || m.Pipeline.Status.Phase == kubekeyv1.PipelinePhaseSucceed {
			return true, nil
		}
		return false, nil
	})

	return nil
}
