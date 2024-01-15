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

package _const

import (
	"path/filepath"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
)

var workDirOnce = &sync.Once{}

// SetWorkDir sets the workdir once.
func SetWorkDir(wd string) {
	workDirOnce.Do(func() {
		workDir = wd
	})
}

// GetWorkDir returns the workdir.
func GetWorkDir() string {
	return workDir
}

func ResourceFromObject(obj runtime.Object) string {
	switch obj.(type) {
	case *kubekeyv1.Pipeline, *kubekeyv1.PipelineList:
		return RuntimePipelineDir
	case *kubekeyv1.Config, *kubekeyv1.ConfigList:
		return RuntimeConfigDir
	case *kubekeyv1.Inventory, *kubekeyv1.InventoryList:
		return RuntimeInventoryDir
	case *kubekeyv1alpha1.Task, *kubekeyv1alpha1.TaskList:
		return RuntimePipelineTaskDir
	default:
		return ""
	}
}

func RuntimeDirFromObject(obj runtime.Object) string {
	resource := ResourceFromObject(obj)
	if resource == "" {
		klog.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
		return ""
	}
	mo, ok := obj.(metav1.Object)

	if !ok {
		klog.Errorf("Failed convert to metav1.Object: %s", obj.GetObjectKind().GroupVersionKind().String())
		return ""
	}
	return filepath.Join(workDir, RuntimeDir, mo.GetNamespace(), resource, mo.GetName())
}
