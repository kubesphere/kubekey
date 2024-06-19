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

package variable

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

type GetFunc func(Variable) (any, error)

type MergeFunc func(Variable) error

type Variable interface {
	Key() string
	Get(GetFunc) (any, error)
	Merge(MergeFunc) error
}

// New variable. generate value from config args. and render to source.
func New(client ctrlclient.Client, pipeline kubekeyv1.Pipeline) (Variable, error) {
	// new source
	s, err := source.New(filepath.Join(_const.RuntimeDirFromPipeline(pipeline), _const.RuntimePipelineVariableDir))
	if err != nil {
		klog.V(4).ErrorS(err, "create file source failed", "path", filepath.Join(_const.RuntimeDirFromPipeline(pipeline), _const.RuntimePipelineVariableDir), "pipeline", ctrlclient.ObjectKeyFromObject(&pipeline))
		return nil, err
	}
	// get config
	var config = &kubekeyv1.Config{}
	if err := client.Get(context.Background(), types.NamespacedName{Namespace: pipeline.Spec.ConfigRef.Namespace, Name: pipeline.Spec.ConfigRef.Name}, config); err != nil {
		klog.V(4).ErrorS(err, "get config from pipeline error", "config", pipeline.Spec.ConfigRef, "pipeline", ctrlclient.ObjectKeyFromObject(&pipeline))
		return nil, err
	}
	// get inventory
	var inventory = &kubekeyv1.Inventory{}
	if err := client.Get(context.Background(), types.NamespacedName{Namespace: pipeline.Spec.InventoryRef.Namespace, Name: pipeline.Spec.InventoryRef.Name}, inventory); err != nil {
		klog.V(4).ErrorS(err, "get inventory from pipeline error", "inventory", pipeline.Spec.InventoryRef, "pipeline", ctrlclient.ObjectKeyFromObject(&pipeline))
		return nil, err
	}
	v := &variable{
		key:    string(pipeline.UID),
		source: s,
		value: &value{
			Config:    *config,
			Inventory: *inventory,
			Hosts:     make(map[string]host),
		},
	}
	for _, hostname := range convertGroup(*inventory)["all"].([]string) {
		v.value.Hosts[hostname] = host{
			RemoteVars:  make(map[string]any),
			RuntimeVars: make(map[string]any),
		}
	}

	// read data from source
	data, err := v.source.Read()
	if err != nil {
		klog.V(4).ErrorS(err, "read data from source error", "pipeline", ctrlclient.ObjectKeyFromObject(&pipeline))
		return nil, err
	}
	for k, d := range data {
		// set hosts
		h := host{}
		if err := json.Unmarshal(d, &h); err != nil {
			klog.V(4).ErrorS(err, "unmarshal host error", "pipeline", ctrlclient.ObjectKeyFromObject(&pipeline))
			return nil, err
		}
		v.value.Hosts[strings.TrimSuffix(k, ".json")] = h
	}

	return v, nil
}
