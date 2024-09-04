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
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
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
	kkcorev1.Config    `json:"-"`
	kkcorev1.Inventory `json:"-"`
	// Hosts store the variable for running tasks on specific hosts
	Hosts map[string]host `json:"hosts"`
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

// getParameterVariable get defined variable from inventory and config
func (v value) getParameterVariable() map[string]any {
	globalHosts := make(map[string]any)

	for hostname := range v.Hosts {
		// get host vars
		hostVars := Extension2Variables(v.Inventory.Spec.Hosts[hostname])
		// set inventory_name to hostVars
		// "inventory_name" is the hostname configured in the inventory file.
		hostVars[_const.VariableInventoryName] = hostname
		if _, ok := hostVars[_const.VariableHostName]; !ok {
			hostVars[_const.VariableHostName] = hostname
		}
		// merge group vars to host vars
		for _, gv := range v.Inventory.Spec.Groups {
			if slices.Contains(gv.Hosts, hostname) {
				hostVars = combineVariables(hostVars, Extension2Variables(gv.Vars))
			}
		}
		// set default localhost
		setlocalhostVarialbe(hostname, v, hostVars)
		// merge inventory vars to host vars
		hostVars = combineVariables(hostVars, Extension2Variables(v.Inventory.Spec.Vars))
		// merge config vars to host vars
		hostVars = combineVariables(hostVars, Extension2Variables(v.Config.Spec))
		globalHosts[hostname] = hostVars
	}

	var externalVal = make(map[string]any)
	// external vars
	for hostname := range globalHosts {
		var val = make(map[string]any)
		val = combineVariables(val, map[string]any{
			_const.VariableGlobalHosts: globalHosts,
		})
		val = combineVariables(val, map[string]any{
			_const.VariableGroups: convertGroup(v.Inventory),
		})
		externalVal[hostname] = val
	}

	return combineVariables(globalHosts, externalVal)
}

type host struct {
	// RemoteVars sources from remote node config. as gather_fact.scope all tasks. it should not be changed.
	RemoteVars map[string]any `json:"remote"`
	// RuntimeVars sources from runtime. store which defined in each appeared block vars.
	RuntimeVars map[string]any `json:"runtime"`
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
		// write to source
		data, err := json.MarshalIndent(hv, "", "  ")
		if err != nil {
			klog.ErrorS(err, "marshal host data error", "hostname", hn)

			return err
		}

		if err := v.source.Write(data, hn+".json"); err != nil {
			klog.ErrorS(err, "write host data to local file error", "hostname", hn, "filename", hn+".json")
		}
	}

	return nil
}

// ***************************** GetFunc ***************************** //

// GetHostnames get all hostnames from a group or host
var GetHostnames = func(name []string) GetFunc {
	if len(name) == 0 {
		return emptyGetFunc
	}

	return func(v Variable) (any, error) {
		vv, ok := v.(*variable)
		if !ok {
			return nil, errors.New("variable type error")
		}
		var hs []string
		for _, n := range name {
			// add host to hs
			if _, ok := vv.value.Hosts[n]; ok {
				hs = append(hs, n)
			}
			// add group's host to gs
			for gn, gv := range convertGroup(vv.value.Inventory) {
				if gn == n {
					if gvd, ok := gv.([]string); ok {
						hs = mergeSlice(hs, gvd)
					}

					break
				}
			}

			// Add the specified host in the specified group to the hs.
			regexForIndex := regexp.MustCompile(`^(.*)\[\d\]$`)
			if match := regexForIndex.FindStringSubmatch(strings.TrimSpace(n)); match != nil {
				index, err := strconv.Atoi(match[2])
				if err != nil {
					klog.V(4).ErrorS(err, "convert index to int error", "index", match[2])

					return nil, err
				}
				if group, ok := convertGroup(vv.value.Inventory)[match[1]].([]string); ok {
					if index >= len(group) {
						return nil, fmt.Errorf("index %v out of range for group %s", index, group)
					}
					hs = append(hs, group[index])
				}
			}

			// add random host in group
			regexForRandom := regexp.MustCompile(`^(.+?)\s*\|\s*random$`)
			if match := regexForRandom.FindStringSubmatch(strings.TrimSpace(n)); match != nil {
				if group, ok := convertGroup(vv.value.Inventory)[match[1]].([]string); ok {
					hs = append(hs, group[rand.Intn(len(group))])
				}
			}
		}

		return hs, nil
	}
}

// GetParamVariable get param variable which is combination of inventory, config.
// if hostname is empty, return all host's param variable.
var GetParamVariable = func(hostname string) GetFunc {
	return func(v Variable) (any, error) {
		vv, ok := v.(*variable)
		if !ok {
			return nil, errors.New("variable type error")
		}
		if hostname == "" {
			return vv.value.getParameterVariable(), nil
		}

		return vv.value.getParameterVariable()[hostname], nil
	}
}

// GetAllVariable get all variable for a given host
var GetAllVariable = func(hostName string) GetFunc {
	return func(v Variable) (any, error) {
		vv, ok := v.(*variable)
		if !ok {
			return nil, errors.New("variable type error")
		}
		result := make(map[string]any)
		// find from runtime
		result = combineVariables(result, vv.value.Hosts[hostName].RuntimeVars)
		// find from remote
		result = combineVariables(result, vv.value.Hosts[hostName].RemoteVars)
		// find from global.
		if vv, ok := vv.value.getParameterVariable()[hostName]; ok {
			if vvd, ok := vv.(map[string]any); ok {
				result = combineVariables(result, vvd)
			}
		}

		return result, nil
	}
}

// GetHostMaxLength get the max length for all hosts
var GetHostMaxLength = func() GetFunc {
	return func(v Variable) (any, error) {
		vv, ok := v.(*variable)
		if !ok {
			return nil, errors.New("variable type error")
		}
		var hostNameMaxLen int
		for k := range vv.value.Hosts {
			hostNameMaxLen = max(len(k), hostNameMaxLen)
		}

		return hostNameMaxLen, nil
	}
}

// ***************************** MergeFunc ***************************** //

// MergeRemoteVariable merge variable to remote.
var MergeRemoteVariable = func(data map[string]any, hostname string) MergeFunc {
	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}

		if hostname == "" {
			return errors.New("when merge source is remote. HostName cannot be empty")
		}
		if _, ok := vv.value.Hosts[hostname]; !ok {
			return fmt.Errorf("when merge source is remote. HostName %s not exist", hostname)
		}

		// it should not be changed
		if hv := vv.value.Hosts[hostname]; len(hv.RemoteVars) == 0 {
			hv.RemoteVars = data
			vv.value.Hosts[hostname] = hv
		}

		return nil
	}
}

// MergeRuntimeVariable parse variable by specific host and merge to the host.
var MergeRuntimeVariable = func(data map[string]any, hosts ...string) MergeFunc {
	if len(data) == 0 || len(hosts) == 0 {
		// skip
		return emptyMergeFunc
	}

	return func(v Variable) error {
		for _, hostName := range hosts {
			vv, ok := v.(*variable)
			if !ok {
				return errors.New("variable type error")
			}
			// merge to specify host
			curVariable, err := v.Get(GetAllVariable(hostName))
			if err != nil {
				return err
			}
			// parse variable
			if err := parseVariable(data, func(s string) (string, error) {
				// parse use total variable. the task variable should not contain template syntax.
				cv, ok := curVariable.(map[string]any)
				if !ok {
					return "", errors.New("variable type error")
				}

				return tmpl.ParseString(combineVariables(data, cv), s)
			}); err != nil {
				return err
			}

			if _, ok := v.(*variable); !ok {
				return errors.New("variable type error")
			}
			hv := vv.value.Hosts[hostName]
			hv.RuntimeVars = combineVariables(hv.RuntimeVars, data)
			vv.value.Hosts[hostName] = hv
		}

		return nil
	}
}

// MergeAllRuntimeVariable parse variable by specific host and merge to all hosts.
var MergeAllRuntimeVariable = func(data map[string]any, hostName string) MergeFunc {
	return func(v Variable) error {
		vv, ok := v.(*variable)
		if !ok {
			return errors.New("variable type error")
		}
		// merge to specify host
		curVariable, err := v.Get(GetAllVariable(hostName))
		if err != nil {
			return err
		}
		// parse variable
		if err := parseVariable(data, func(s string) (string, error) {
			// parse use total variable. the task variable should not contain template syntax.
			cv, ok := curVariable.(map[string]any)
			if !ok {
				return "", errors.New("variable type error")
			}

			return tmpl.ParseString(combineVariables(data, cv), s)
		}); err != nil {
			return err
		}

		for h := range vv.value.Hosts {
			if _, ok := v.(*variable); !ok {
				return errors.New("variable type error")
			}
			hv := vv.value.Hosts[h]
			hv.RuntimeVars = combineVariables(hv.RuntimeVars, data)
			vv.value.Hosts[h] = hv
		}

		return nil
	}
}
