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

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
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

// GetRuntimeDir returns the absolute path of the runtime directory.
func GetRuntimeDir() string {
	return filepath.Join(workDir, RuntimeDir)
}

func RuntimeDirFromPipeline(obj kkcorev1.Pipeline) string {
	return filepath.Join(GetRuntimeDir(), kkcorev1.SchemeGroupVersion.String(),
		RuntimePipelineDir, obj.Namespace, obj.Name)
}
