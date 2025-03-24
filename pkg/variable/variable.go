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
	"reflect"
	"sync"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

// Variable store all vars which playbook used.
type Variable interface {
	Get(getFunc GetFunc) (any, error)
	Merge(mergeFunc MergeFunc) error
}

// New variable. generate value from config args. and render to source.
func New(ctx context.Context, client ctrlclient.Client, playbook kkcorev1.Playbook, st source.SourceType) (Variable, error) {
	var err error
	// new source
	var s source.Source

	switch st {
	case source.MemorySource:
		s = source.NewMemorySource()
	case source.FileSource:
		path := filepath.Join(_const.GetWorkdirFromConfig(playbook.Spec.Config), _const.RuntimeDir, kkcorev1.SchemeGroupVersion.String(), _const.RuntimePlaybookDir, playbook.Namespace, playbook.Name, _const.RuntimePlaybookVariableDir)
		s, err = source.NewFileSource(path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create file source %q", path)
		}
	default:
		return nil, errors.Errorf("unsupported source type: %v", st)
	}

	// get inventory
	var inventory = &kkcorev1.Inventory{}
	if playbook.Spec.InventoryRef != nil {
		if err := client.Get(ctx, types.NamespacedName{Namespace: playbook.Spec.InventoryRef.Namespace, Name: playbook.Spec.InventoryRef.Name}, inventory); err != nil {
			return nil, errors.Wrapf(err, "failed to get inventory %q from playbook %q", playbook.Spec.InventoryRef, ctrlclient.ObjectKeyFromObject(&playbook))
		}
	}

	v := &variable{
		key:    string(playbook.UID),
		source: s,
		value: &value{
			Config:    playbook.Spec.Config,
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
		return nil, errors.Wrapf(err, "failed to read data from source. playbook: %q", ctrlclient.ObjectKeyFromObject(&playbook))
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

type variable struct {
	// key is the unique Identifier of the variable. usually the UID of the playbook.
	key string
	// source is where the variable is stored
	source source.Source
	// value is the data of the variable, which store in memory
	value *value
	// lock is the lock for value
	sync.Mutex
}

// value is the specific data contained in the variable
type value struct {
	Config    kkcorev1.Config
	Inventory kkcorev1.Inventory
	// Hosts store the variable for running tasks on specific hosts
	Hosts map[string]host
}

type host struct {
	// RemoteVars sources from remote node config. as gather_fact.scope all tasks. it should not be changed.
	RemoteVars map[string]any `json:"remote"`
	// RuntimeVars sources from runtime. store which defined in each appeared block vars.
	RuntimeVars map[string]any `json:"runtime"`
}

func (v value) deepCopy() value {
	nv := value{}

	data, err := json.Marshal(v)
	if err != nil {
		return value{}
	}

	if err := json.Unmarshal(data, &nv); err != nil {
		return value{}
	}

	return nv
}

// Get vars
func (v *variable) Get(f GetFunc) (any, error) {
	return f(v)
}

// Merge hosts vars to variable and sync to resource
func (v *variable) Merge(f MergeFunc) error {
	v.Lock()
	defer v.Unlock()

	old := v.value.deepCopy()

	if err := f(v); err != nil {
		return err
	}

	return v.syncSource(old)
}

// syncSource sync hosts vars to source.
func (v *variable) syncSource(old value) error {
	for hn, hv := range v.value.Hosts {
		if reflect.DeepEqual(old.Hosts[hn], hv) {
			// nothing change skip.
			continue
		}
		if err := v.source.Write(map[string]any{
			"remote":  hv.RemoteVars,
			"runtime": hv.RuntimeVars,
		}, hn); err != nil {
			return errors.Wrapf(err, "failed to write host %s variable to source", hn)
		}
	}

	return nil
}
