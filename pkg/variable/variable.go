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
	"errors"
	"fmt"
	"path/filepath"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

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
		path := filepath.Join(_const.GetWorkdirFromConfig(pipeline.Spec.Config), _const.RuntimeDir, kkcorev1.SchemeGroupVersion.String(), _const.RuntimePipelineDir, pipeline.Namespace, pipeline.Name, _const.RuntimePipelineVariableDir)
		s, err = source.NewFileSource(path)
		if err != nil {
			klog.V(4).ErrorS(err, "create file source failed", "path", path)

			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported source type: %v", st)
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
			Config:    pipeline.Spec.Config,
			Inventory: *inventory,
			Hosts:     make(map[string]host),
		},
	}

	if gd, ok := ConvertGroup(*inventory)["all"].([]string); ok {
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
		return nil, fmt.Errorf("failed to read data from source. pipeline: %q. error: %w", ctrlclient.ObjectKeyFromObject(&pipeline), err)
	}

	for k, d := range data {
		// set hosts
		h := host{}
		if val, ok := d["remote"]; ok {
			remoteVars, ok := val.(map[string]any)
			if !ok {
				return nil, errors.New("type assertion failed. expected map[string]any for remote vars")
			}
			h.RemoteVars = remoteVars
		}
		if val, ok := d["runtime"]; ok {
			runtimeVars, ok := val.(map[string]any)
			if !ok {
				return nil, errors.New("type assertion failed. expected map[string]any for runtime vars")
			}
			h.RuntimeVars = runtimeVars
		}
		v.value.Hosts[k] = h
	}

	return v, nil
}
