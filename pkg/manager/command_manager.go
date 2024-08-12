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
	"io"
	"os"
	"time"

	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/executor"
)

type commandManager struct {
	*kkcorev1.Pipeline
	*kkcorev1.Config
	*kkcorev1.Inventory

	ctrlclient.Client

	logOutput io.Writer
}

func (m *commandManager) Run(ctx context.Context) error {
	fmt.Fprint(m.logOutput, `

 _   __      _          _   __           
| | / /     | |        | | / /           
| |/ / _   _| |__   ___| |/ /  ___ _   _ 
|    \| | | | '_ \ / _ \    \ / _ \ | | |
| |\  \ |_| | |_) |  __/ |\  \  __/ |_| |
\_| \_/\__,_|_.__/ \___\_| \_/\___|\__, |
                                    __/ |
                                   |___/

`)
	fmt.Fprintf(m.logOutput, "%s [Pipeline %s] start\n", time.Now().Format(time.TimeOnly+" MST"), ctrlclient.ObjectKeyFromObject(m.Pipeline))
	cp := m.Pipeline.DeepCopy()
	defer func() {
		fmt.Fprintf(m.logOutput, "%s [Pipeline %s] finish. total: %v,success: %v,ignored: %v,failed: %v\n", time.Now().Format(time.TimeOnly+" MST"), ctrlclient.ObjectKeyFromObject(m.Pipeline),
			m.Pipeline.Status.TaskResult.Total, m.Pipeline.Status.TaskResult.Success, m.Pipeline.Status.TaskResult.Ignored, m.Pipeline.Status.TaskResult.Failed)
		if !m.Pipeline.Spec.Debug && m.Pipeline.Status.Phase == kkcorev1.PipelinePhaseSucceed {
			fmt.Fprintf(m.logOutput, "%s [Pipeline %s] clean runtime directory\n", time.Now().Format(time.TimeOnly+" MST"), ctrlclient.ObjectKeyFromObject(m.Pipeline))
			// clean runtime directory
			if err := os.RemoveAll(_const.GetRuntimeDir()); err != nil {
				klog.ErrorS(err, "clean runtime directory error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline), "runtime_dir", _const.GetRuntimeDir())
			}
		}
		if m.Pipeline.Spec.JobSpec.Schedule != "" { // if pipeline is cornJob. it's always running.
			m.Pipeline.Status.Phase = kkcorev1.PipelinePhaseRunning
		}
		// update pipeline status
		if err := m.Client.Status().Patch(ctx, m.Pipeline, ctrlclient.MergeFrom(cp)); err != nil {
			klog.ErrorS(err, "update pipeline error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		}
	}()

	m.Pipeline.Status.Phase = kkcorev1.PipelinePhaseSucceed
	if err := executor.NewTaskExecutor(m.Client, m.Pipeline, m.logOutput).Exec(ctx); err != nil {
		klog.ErrorS(err, "executor tasks error", "pipeline", ctrlclient.ObjectKeyFromObject(m.Pipeline))
		m.Pipeline.Status.Phase = kkcorev1.PipelinePhaseFailed
		m.Pipeline.Status.Reason = err.Error()
		return err
	}

	return nil
}
