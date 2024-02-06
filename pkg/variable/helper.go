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
	"path/filepath"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// mergeVariables merge multiple variables into one variable
// v2 will override v1 if variable is repeated
func mergeVariables(v1, v2 VariableData) VariableData {
	mergedVars := make(VariableData)
	for k, v := range v1 {
		mergedVars[k] = v
	}
	for k, v := range v2 {
		mergedVars[k] = v
	}
	return mergedVars
}

func findLocation(loc []location, uid string) *location {
	for i := range loc {
		if uid == loc[i].UID {
			return &loc[i]
		}
		// find in block,always,rescue
		if l := findLocation(append(append(loc[i].Block, loc[i].Always...), loc[i].Rescue...), uid); l != nil {
			return l
		}
	}
	return nil
}

func convertGroup(inv kubekeyv1.Inventory) VariableData {
	groups := make(VariableData)
	all := make([]string, 0)
	for hn := range inv.Spec.Hosts {
		all = append(all, hn)
	}
	groups["all"] = all
	for gn := range inv.Spec.Groups {
		groups[gn] = hostsInGroup(inv, gn)
	}
	return groups
}

func hostsInGroup(inv kubekeyv1.Inventory, groupName string) []string {
	if v, ok := inv.Spec.Groups[groupName]; ok {
		var hosts []string
		for _, cg := range v.Groups {
			hosts = mergeSlice(hostsInGroup(inv, cg), hosts)
		}
		return mergeSlice(hosts, v.Hosts)
	}
	return nil
}

// StringVar get string value by key
func StringVar(vars VariableData, key string) *string {
	value, ok := vars[key]
	if !ok {
		klog.V(4).InfoS("cannot find variable", "key", key)
		return nil
	}
	sv, ok := value.(string)
	if !ok {
		klog.V(4).InfoS("variable is not string", "key", key)
		return nil
	}
	return &sv
}

// IntVar get int value by key
func IntVar(vars VariableData, key string) *int {
	value, ok := vars[key]
	if !ok {
		klog.V(4).InfoS("cannot find variable", "key", key)
		return nil
	}
	// default convert to float64
	number, ok := value.(float64)
	if !ok {
		klog.V(4).InfoS("variable is not number", "key", key)
		return nil
	}
	vi := int(number)
	return &vi
}

// StringSliceVar get string slice value by key
func StringSliceVar(vars VariableData, key string) []string {
	value, ok := vars[key]
	if !ok {
		klog.V(4).InfoS("cannot find variable", "key", key)
		return nil
	}
	sv, ok := value.([]any)
	if !ok {
		klog.V(4).InfoS("variable is not string slice", "key", key)
		return nil
	}
	var ss []string
	for _, a := range sv {
		av, ok := a.(string)
		if !ok {
			klog.V(4).InfoS("variable is not string", "key", key)
			return nil
		}
		ss = append(ss, av)
	}
	return ss
}

func Extension2Variables(ext runtime.RawExtension) VariableData {
	if len(ext.Raw) == 0 {
		return nil
	}

	var data VariableData
	if err := yaml.Unmarshal(ext.Raw, &data); err != nil {
		klog.V(4).ErrorS(err, "failed to unmarshal extension to variables")
	}
	return data
}

func Extension2Slice(ext runtime.RawExtension) []any {
	if len(ext.Raw) == 0 {
		return nil
	}

	var data []any
	if err := yaml.Unmarshal(ext.Raw, &data); err != nil {
		klog.V(4).ErrorS(err, "failed to unmarshal extension to slice")
	}
	return data
}

func Extension2String(ext runtime.RawExtension) string {
	if len(ext.Raw) == 0 {
		return ""
	}
	// try to escape string
	if ns, err := strconv.Unquote(string(ext.Raw)); err == nil {
		return ns
	}
	return string(ext.Raw)
}

func RuntimeDirFromPipeline(obj kubekeyv1.Pipeline) string {
	return filepath.Join(_const.GetRuntimeDir(), kubekeyv1.SchemeGroupVersion.String(),
		_const.RuntimePipelineDir, obj.Namespace, obj.Name, _const.RuntimePipelineVariableDir)
}
