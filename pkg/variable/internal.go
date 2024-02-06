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
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"k8s.io/klog/v2"
	"k8s.io/utils/strings/slices"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

type variable struct {
	// key is the unique Identifier of the variable. usually the UID of the pipeline.
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
	kubekeyv1.Config    `json:"-"`
	kubekeyv1.Inventory `json:"-"`
	// Hosts store the variable for running tasks on specific hosts
	Hosts map[string]host `json:"hosts"`
	// Location is the complete location index.
	// This index can help us determine the specific location of the task,
	// enabling us to retrieve the task's parameters and establish the execution order.
	Location []location `json:"location"`
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

// getGlobalVars get defined variable from inventory and config
func (v *value) getGlobalVars(hostname string) VariableData {
	// get host vars
	hostVars := Extension2Variables(v.Inventory.Spec.Hosts[hostname])
	// set inventory_hostname to hostVars
	// inventory_hostname" is the hostname configured in the inventory file.
	hostVars = mergeVariables(hostVars, VariableData{
		"inventory_hostname": hostname,
	})
	// merge group vars to host vars
	for _, gv := range v.Inventory.Spec.Groups {
		if slices.Contains(gv.Hosts, hostname) {
			hostVars = mergeVariables(hostVars, Extension2Variables(gv.Vars))
		}
	}
	// merge inventory vars to host vars
	hostVars = mergeVariables(hostVars, Extension2Variables(v.Inventory.Spec.Vars))
	// merge config vars to host vars
	hostVars = mergeVariables(hostVars, Extension2Variables(v.Config.Spec))

	// external vars
	hostVars = mergeVariables(hostVars, VariableData{
		"groups": convertGroup(v.Inventory),
	})

	return hostVars
}

type location struct {
	// UID is current location uid
	UID string `json:"uid"`
	// PUID is the parent uid for current location
	PUID string `json:"puid"`
	// Name is the name of current location
	Name string `json:"name"`
	// Vars is the variable of current location
	Vars VariableData `json:"vars,omitempty"`

	Block  []location `json:"block,omitempty"`
	Always []location `json:"always,omitempty"`
	Rescue []location `json:"rescue,omitempty"`
}

// VariableData is the variable data
type VariableData map[string]any

func (v VariableData) String() string {
	data, err := json.Marshal(v)
	if err != nil {
		klog.V(4).ErrorS(err, "marshal in error", "data", v)
		return ""
	}
	return string(data)
}

type host struct {
	Vars        VariableData            `json:"vars"`
	RuntimeVars map[string]VariableData `json:"runtime"`
}

func (v *variable) Key() string {
	return v.key
}

func (v *variable) Get(option GetOption) (any, error) {
	return option.filter(*v.value)
}

func (v *variable) Merge(mo ...MergeOption) error {
	v.Lock()
	defer v.Unlock()

	old := v.value.deepCopy()
	for _, o := range mo {
		if err := o.mergeTo(v.value); err != nil {
			return err
		}
	}

	if !reflect.DeepEqual(old.Location, v.value.Location) {
		if err := v.syncLocation(); err != nil {
			klog.ErrorS(err, "sync location error")
		}
	}

	for hn, hv := range v.value.Hosts {
		if !reflect.DeepEqual(old.Hosts[hn], hv) {
			if err := v.syncHosts(hn); err != nil {
				klog.ErrorS(err, "sync host error", "hostname", hn)
			}
		}
	}

	return nil
}

func (v *variable) syncLocation() error {
	data, err := json.MarshalIndent(v.value.Location, "", "  ")
	if err != nil {
		klog.ErrorS(err, "marshal location data error")
		return err
	}
	if err := v.source.Write(data, _const.RuntimePipelineVariableLocationFile); err != nil {
		klog.V(4).ErrorS(err, "write location data to local file error", "filename", _const.RuntimePipelineVariableLocationFile)
		return err
	}
	return nil
}

// syncHosts sync hosts data to local file. If hostname is empty, sync all hosts
func (v *variable) syncHosts(hostname ...string) error {
	for _, hn := range hostname {
		if hv, ok := v.value.Hosts[hn]; ok {
			data, err := json.MarshalIndent(hv, "", "  ")
			if err != nil {
				klog.ErrorS(err, "marshal host data error", "hostname", hn)
				return err
			}
			if err := v.source.Write(data, fmt.Sprintf("%s.json", hn)); err != nil {
				klog.ErrorS(err, "write host data to local file error", "hostname", hn, "filename", fmt.Sprintf("%s.json", hn))
			}
		}
	}
	return nil
}

// mergeSlice with skip repeat value
func mergeSlice(g1, g2 []string) []string {
	uniqueValues := make(map[string]bool)
	mg := []string{}

	// Add values from the first slice
	for _, v := range g1 {
		if !uniqueValues[v] {
			uniqueValues[v] = true
			mg = append(mg, v)
		}
	}

	// Add values from the second slice
	for _, v := range g2 {
		if !uniqueValues[v] {
			uniqueValues[v] = true
			mg = append(mg, v)
		}
	}

	return mg
}
