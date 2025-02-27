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
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/klog/v2"
)

// GetWorkdirFromConfig retrieves the working directory from the provided configuration.
// If the 'workdir' value is set in the configuration and is a string, it returns that value.
// If the 'workdir' value is not set or is not a string, it logs an informational message
// and attempts to get the current working directory of the process.
// If it fails to get the current working directory, it logs another informational message
// and returns a default directory path "/opt/kubekey".
func GetWorkdirFromConfig(config kkcorev1.Config) string {
	workdir, err := config.GetValue(Workdir)
	if err == nil {
		wd, ok := workdir.(string)
		if ok {
			return wd
		}
	}
	klog.Info("work_dir is not set use current dir.")
	wd, err := os.Getwd()
	if err != nil {
		klog.Info("failed to get current dir. use default: /root/kubekey")

		return "/opt/kubekey"
	}

	return wd
}

// // GetRuntimeDir returns the absolute path of the runtime directory.
// func GetRuntimeDir(config kkcorev1.Config) string {
// 	return filepath.Join(GetWorkDir(config), RuntimeDir)
// }

// // RuntimeDirFromPipeline returns the absolute path of the runtime directory for specify Pipeline
// func RuntimeDirFromPipeline(obj kkcorev1.Pipeline) string {
// 	return filepath.Join(GetRuntimeDir(obj.Spec.Config), kkcorev1.SchemeGroupVersion.String(),
// 		RuntimePipelineDir, obj.Namespace, obj.Name)
// }
