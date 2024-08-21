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

package converter

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"

	kkcorev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1alpha1"
	kkprojectv1 "github.com/kubesphere/kubekey/v4/pkg/apis/project/v1"
)

// MarshalBlock marshal block to task
func MarshalBlock(role string, hosts []string, when []string, block kkprojectv1.Block) *kkcorev1alpha1.Task {
	task := &kkcorev1alpha1.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: kkcorev1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
			Annotations: map[string]string{
				kkcorev1alpha1.TaskAnnotationRole: role,
			},
		},
		Spec: kkcorev1alpha1.TaskSpec{
			Name:        block.Name,
			Hosts:       hosts,
			IgnoreError: block.IgnoreErrors,
			Retries:     block.Retries,
			When:        when,
			FailedWhen:  block.FailedWhen.Data,
			Register:    block.Register,
		},
	}

	if block.Loop != nil {
		data, err := json.Marshal(block.Loop)
		if err != nil {
			klog.V(4).ErrorS(err, "Marshal loop failed", "task", task.Name, "block", block.Name)
		}
		task.Spec.Loop = runtime.RawExtension{Raw: data}
	}

	return task
}

// GroupHostBySerial group hosts by serial
func GroupHostBySerial(hosts []string, serial []any) ([][]string, error) {
	if len(serial) == 0 {
		return [][]string{hosts}, nil
	}

	// convertSerial to []int
	var sis = make([]int, len(serial))
	// the count for sis
	var count int
	for i, a := range serial {
		switch val := a.(type) {
		case int:
			sis[i] = val
		case string:
			if strings.HasSuffix(val, "%") {
				b, err := strconv.ParseFloat(val[:len(val)-1], 64)
				if err != nil {
					return nil, fmt.Errorf("convert serial %v to float error: %w", a, err)
				}
				sis[i] = int(math.Ceil(float64(len(hosts)) * b / 100.0))
			} else {
				b, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("convert serial %v to int error: %w", a, err)
				}
				sis[i] = b
			}
		default:
			return nil, errors.New("unknown serial type. only support int or percent")
		}
		if sis[i] == 0 {
			return nil, fmt.Errorf("serial %v should not be zero", a)
		}
		count += sis[i]
	}

	if len(hosts) > count {
		for i := 0.0; i < float64(len(hosts)-count)/float64(sis[len(sis)-1]); i++ {
			sis = append(sis, sis[len(sis)-1])
		}
	}

	// total result
	result := make([][]string, len(sis))
	var begin, end int
	for i, si := range sis {
		end += si
		if end > len(hosts) {
			end = len(hosts)
		}
		result[i] = hosts[begin:end]
		begin += si
	}

	return result, nil
}
