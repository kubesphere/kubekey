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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

var (
	emptyGetFunc GetFunc = func(Variable) (any, error) {
		return nil, errors.New("nil value returned")
	}
	emptyMergeFunc MergeFunc = func(Variable) error {
		return nil
	}
)

// GetFunc get data from variable
type GetFunc func(Variable) (any, error)

// MergeFunc merge data to variable
type MergeFunc func(Variable) error

// Variable store all vars which pipeline used.
type Variable interface {
	Get(getFunc GetFunc) (any, error)
	Merge(mergeFunc MergeFunc) error
}

// New variable. generate value from config args. and render to source.
func New(ctx context.Context, client ctrlclient.Client, pipeline kkcorev1.Pipeline, st source.SourceType) (Variable, error) {
	var err error
	// new source
	var s source.Source

	switch st {
	case source.MemorySource:
		s = source.NewMemorySource()
	case source.FileSource:
		s, err = source.NewFileSource(filepath.Join(_const.RuntimeDirFromPipeline(pipeline), _const.RuntimePipelineVariableDir))
		if err != nil {
			klog.V(4).ErrorS(err, "create file source failed", "path", filepath.Join(_const.RuntimeDirFromPipeline(pipeline), _const.RuntimePipelineVariableDir), "pipeline", ctrlclient.ObjectKeyFromObject(&pipeline))

			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported source type: %v", st)
	}

	// get config
	var config = &kkcorev1.Config{}
	if pipeline.Spec.ConfigRef != nil {
		if err := client.Get(ctx, types.NamespacedName{Namespace: pipeline.Spec.ConfigRef.Namespace, Name: pipeline.Spec.ConfigRef.Name}, config); err != nil {
			klog.V(4).ErrorS(err, "get config from pipeline error", "config", pipeline.Spec.ConfigRef, "pipeline", ctrlclient.ObjectKeyFromObject(&pipeline))

			return nil, err
		}
	}

	// get inventory
	var inventory = &kkcorev1.Inventory{}
	if pipeline.Spec.InventoryRef != nil {
		if err := client.Get(ctx, types.NamespacedName{Namespace: pipeline.Spec.InventoryRef.Namespace, Name: pipeline.Spec.InventoryRef.Name}, inventory); err != nil {
			klog.V(4).ErrorS(err, "get inventory from pipeline error", "inventory", pipeline.Spec.InventoryRef, "pipeline", ctrlclient.ObjectKeyFromObject(&pipeline))

			return nil, err
		}
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

	if gd, ok := convertGroup(*inventory)["all"].([]string); ok {
		for _, hostname := range gd {
			v.value.Hosts[hostname] = host{
				RemoteVars:  make(map[string]any),
				RuntimeVars: make(map[string]any),
			}
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
