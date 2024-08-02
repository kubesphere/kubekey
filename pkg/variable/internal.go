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
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
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
	kubekeyv1.Config    `json:"-"`
	kubekeyv1.Inventory `json:"-"`
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
		hostVars[_const.VariableHostName] = hostname
		// merge group vars to host vars
		for _, gv := range v.Inventory.Spec.Groups {
			if slices.Contains(gv.Hosts, hostname) {
				hostVars = combineVariables(hostVars, Extension2Variables(gv.Vars))
			}
		}
		// set default localhost
		if hostname == _const.VariableLocalHost {
			if _, ok := hostVars[_const.VariableIPv4]; !ok {
				hostVars[_const.VariableIPv4] = getLocalIP(_const.VariableIPv4)
			}
			if _, ok := hostVars[_const.VariableIPv6]; !ok {
				hostVars[_const.VariableIPv6] = getLocalIP(_const.VariableIPv6)
			}
		}

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

func (v *variable) Key() string {
	return v.key
}

func (v *variable) Get(f GetFunc) (any, error) {
	return f(v)
}

func (v *variable) Merge(f MergeFunc) error {
	v.Lock()
	defer v.Unlock()

	old := v.value.deepCopy()
	if err := f(v); err != nil {
		return err
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

// GetHostnames get all hostnames from a group or host
var GetHostnames = func(name []string) GetFunc {
	return func(v Variable) (any, error) {
		if _, ok := v.(*variable); !ok {
			return nil, fmt.Errorf("variable type error")
		}
		data := v.(*variable).value

		var hs []string
		for _, n := range name {
			// add host to hs
			if _, ok := data.Hosts[n]; ok {
				hs = append(hs, n)
			}
			// add group's host to gs
			for gn, gv := range convertGroup(data.Inventory) {
				if gn == n {
					hs = mergeSlice(hs, gv.([]string))
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
				if group, ok := convertGroup(data.Inventory)[match[1]].([]string); ok {
					if index >= len(group) {
						return nil, fmt.Errorf("index %v out of range for group %s", index, group)
					}
					hs = append(hs, group[index])
				}
			}

			// add random host in group
			regexForRandom := regexp.MustCompile(`^(.+?)\s*\|\s*random$`)
			if match := regexForRandom.FindStringSubmatch(strings.TrimSpace(n)); match != nil {
				if group, ok := convertGroup(data.Inventory)[match[1]].([]string); ok {
					hs = append(hs, group[rand.Intn(len(group))])
				}
			}
		}

		return hs, nil
	}
}

// GetParamVariable get param variable which is combination of inventory, config.
var GetParamVariable = func(hostname string) GetFunc {
	return func(v Variable) (any, error) {
		if _, ok := v.(*variable); !ok {
			return nil, fmt.Errorf("variable type error")
		}
		data := v.(*variable).value
		if hostname == "" {
			return data.getParameterVariable(), nil
		}
		return data.getParameterVariable()[hostname], nil
	}
}

// MergeRemoteVariable merge variable to remote.
var MergeRemoteVariable = func(hostname string, data map[string]any) MergeFunc {
	return func(v Variable) error {
		if _, ok := v.(*variable); !ok {
			return fmt.Errorf("variable type error")
		}
		vv := v.(*variable).value

		if hostname == "" {
			return fmt.Errorf("when merge source is remote. HostName cannot be empty")
		}
		if _, ok := vv.Hosts[hostname]; !ok {
			return fmt.Errorf("when merge source is remote. HostName %s not exist", hostname)
		}

		// it should not be changed
		if hv := vv.Hosts[hostname]; len(hv.RemoteVars) == 0 {
			hv.RemoteVars = data
			vv.Hosts[hostname] = hv
		}

		return nil
	}
}

// MergeRuntimeVariable parse variable by specific host and merge to the host.
var MergeRuntimeVariable = func(hostName string, vd map[string]any) MergeFunc {
	return func(v Variable) error {
		vv := v.(*variable).value
		// merge to specify host
		curVariable, err := v.Get(GetAllVariable(hostName))
		if err != nil {
			return err
		}
		// parse variable
		if err := parseVariable(vd, func(s string) (string, error) {
			// parse use total variable. the task variable should not contain template syntax.
			return tmpl.ParseString(combineVariables(vd, curVariable.(map[string]any)), s)
		}); err != nil {
			return err
		}

		if _, ok := v.(*variable); !ok {
			return fmt.Errorf("variable type error")
		}
		hv := vv.Hosts[hostName]
		hv.RuntimeVars = combineVariables(hv.RuntimeVars, vd)
		vv.Hosts[hostName] = hv

		return nil
	}
}

// MergeAllRuntimeVariable parse variable by specific host and merge to all hosts.
var MergeAllRuntimeVariable = func(hostName string, vd map[string]any) MergeFunc {
	return func(v Variable) error {
		vv := v.(*variable).value
		// merge to specify host
		curVariable, err := v.Get(GetAllVariable(hostName))
		if err != nil {
			return err
		}
		// parse variable
		if err := parseVariable(vd, func(s string) (string, error) {
			// parse use total variable. the task variable should not contain template syntax.
			return tmpl.ParseString(combineVariables(vd, curVariable.(map[string]any)), s)
		}); err != nil {
			return err
		}

		for h := range vv.Hosts {
			if _, ok := v.(*variable); !ok {
				return fmt.Errorf("variable type error")
			}
			hv := vv.Hosts[h]
			hv.RuntimeVars = combineVariables(hv.RuntimeVars, vd)
			vv.Hosts[h] = hv
		}

		return nil
	}
}

// GetAllVariable get all variable for a given host
var GetAllVariable = func(hostName string) GetFunc {
	return func(v Variable) (any, error) {
		if _, ok := v.(*variable); !ok {
			return nil, fmt.Errorf("variable type error")
		}
		data := v.(*variable).value
		result := make(map[string]any)
		// find from runtime
		result = combineVariables(result, data.Hosts[hostName].RuntimeVars)
		// find from remote
		result = combineVariables(result, data.Hosts[hostName].RemoteVars)
		// find from global.
		if vv, ok := data.getParameterVariable()[hostName]; ok {
			result = combineVariables(result, vv.(map[string]any))
		}

		return result, nil
	}
}

// GetHostMaxLength get the max length for all hosts
var GetHostMaxLength = func() GetFunc {
	return func(v Variable) (any, error) {
		if _, ok := v.(*variable); !ok {
			return nil, fmt.Errorf("variable type error")
		}
		data := v.(*variable).value
		var hostNameMaxLen int
		for k := range data.Hosts {
			hostNameMaxLen = max(len(k), hostNameMaxLen)
		}
		return hostNameMaxLen, nil
	}
}
