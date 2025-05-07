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
	"math"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"

	"github.com/cockroachdb/errors"
	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// MarshalBlock marshal block to task
func MarshalBlock(hosts []string, when []string, block kkprojectv1.Block) *kkcorev1alpha1.Task {
	task := &kkcorev1alpha1.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: kkcorev1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
		},
		Spec: kkcorev1alpha1.TaskSpec{
			Name:         block.Name,
			Hosts:        hosts,
			IgnoreError:  block.IgnoreErrors,
			Retries:      block.Retries,
			When:         when,
			FailedWhen:   block.FailedWhen.Data,
			Register:     block.Register,
			RegisterType: block.RegisterType,
		},
	}
	if annotation, ok := block.UnknownField["annotations"].(map[string]string); ok {
		task.ObjectMeta.Annotations = annotation
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
					return nil, errors.Wrapf(err, "convert serial %q to float", val)
				}
				sis[i] = int(math.Ceil(float64(len(hosts)) * b / 100.0))
			} else {
				b, err := strconv.Atoi(val)
				if err != nil {
					return nil, errors.Wrapf(err, "convert serial %q to int", val)
				}
				sis[i] = b
			}
		default:
			return nil, errors.New("unknown serial type. only support int or percent")
		}
		if sis[i] == 0 {
			return nil, errors.Errorf("serial %v should not be zero", a)
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

// ConvertKKClusterToInventoryHost convert inventoryHost which defined in kkclusters.infrastructure.cluster.x-k8s.io to inventoryHost which defined in inventories.kubekey.kubesphere.io .
func ConvertKKClusterToInventoryHost(kkcluster *capkkinfrav1beta1.KKCluster) (kkcorev1.InventoryHost, error) {
	inventoryHosts := make(kkcorev1.InventoryHost)
	for _, ih := range kkcluster.Spec.InventoryHosts {
		vars := make(map[string]any)
		if ih.Vars.Raw != nil {
			if err := json.Unmarshal(ih.Vars.Raw, &vars); err != nil {
				return nil, errors.Wrapf(err, "failed to unmarshal kkcluster.spec.InventoryHost %s to inventoryHost", ih.Name)
			}
		}
		vars[_const.VariableConnector] = ih.Connector
		data, err := json.Marshal(vars)
		if err != nil {
			return nil, errors.Wrapf(err, "marshal kkclusters %s to inventory", ih.Name)
		}
		inventoryHosts[ih.Name] = runtime.RawExtension{Raw: data}
	}

	return inventoryHosts, nil
}

// ConvertMap2Node converts a map[string]any to a yaml.Node by first marshaling to YAML bytes
// then unmarshaling into a Node. This allows working with the YAML node structure directly.
func ConvertMap2Node(m map[string]any) (yaml.Node, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return yaml.Node{}, errors.Wrap(err, "failed to marshal map to yaml")
	}
	var node yaml.Node
	err = yaml.Unmarshal(data, &node)
	if err != nil {
		return yaml.Node{}, errors.Wrap(err, "failed to unmarshal yaml to node")
	}

	return node, nil
}
