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
	"fmt"
	"net"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
)

// CombineVariables merge multiple variables into one variable
// v2 will override v1 if variable is repeated
func CombineVariables(m1, m2 map[string]any) map[string]any {
	var f func(val1, val2 any) any
	f = func(val1, val2 any) any {
		if val1 != nil && val2 != nil &&
			reflect.TypeOf(val1).Kind() == reflect.Map && reflect.TypeOf(val2).Kind() == reflect.Map {
			mergedVars := make(map[string]any)
			for _, k := range reflect.ValueOf(val1).MapKeys() {
				mergedVars[k.String()] = reflect.ValueOf(val1).MapIndex(k).Interface()
			}

			for _, k := range reflect.ValueOf(val2).MapKeys() {
				mergedVars[k.String()] = f(mergedVars[k.String()], reflect.ValueOf(val2).MapIndex(k).Interface())
			}

			return mergedVars
		}

		return val2
	}
	mv := make(map[string]any)

	for k, v := range m1 {
		mv[k] = v
	}

	for k, v := range m2 {
		mv[k] = f(mv[k], v)
	}

	return mv
}

// ConvertGroup converts the inventory into a map of groups with their respective hosts.
// It ensures that all hosts are included in the "all" group and adds a default localhost if not present.
// It also creates an "ungrouped" group for hosts that are not part of any specific group.
//
// Parameters:
//
//	inv (kkcorev1.Inventory): The inventory containing hosts and groups specifications.
//
// Returns:
//
//	map[string]any: A map where keys are group names and values are lists of hostnames.
func ConvertGroup(inv kkcorev1.Inventory) map[string]any {
	groups := make(map[string]any)
	all := make([]string, 0)

	for hn := range inv.Spec.Hosts {
		all = append(all, hn)
	}

	ungrouped := make([]string, len(all))
	copy(ungrouped, all)

	if !slices.Contains(all, _const.VariableLocalHost) { // set default localhost
		all = append(all, _const.VariableLocalHost)
	}

	groups[_const.VariableGroupsAll] = all

	for gn := range inv.Spec.Groups {
		groups[gn] = HostsInGroup(inv, gn)
		if hosts, ok := groups[gn].([]string); ok {
			for _, v := range hosts {
				if slices.Contains(ungrouped, v) {
					ungrouped = slices.Delete(ungrouped, slices.Index(ungrouped, v), slices.Index(ungrouped, v)+1)
				}
			}
		}
	}

	groups[_const.VariableUnGrouped] = ungrouped

	return groups
}

// HostsInGroup get a host_name slice in a given group
// if the given group contains other group. convert other group to host_name slice.
func HostsInGroup(inv kkcorev1.Inventory, groupName string) []string {
	if v, ok := inv.Spec.Groups[groupName]; ok {
		var hosts []string
		for _, cg := range v.Groups {
			hosts = mergeSlice(HostsInGroup(inv, cg), hosts)
		}

		return mergeSlice(hosts, v.Hosts)
	}

	return nil
}

// mergeSlice with skip repeat value
func mergeSlice(g1, g2 []string) []string {
	uniqueValues := make(map[string]bool)
	mg := make([]string, 0)

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

// parseVariable parse all string values to the actual value.
func parseVariable(v any, parseTmplFunc func(string) (string, error)) error {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Map:
		if err := parseVariableFromMap(v, parseTmplFunc); err != nil {
			return err
		}
	case reflect.Slice, reflect.Array:
		if err := parseVariableFromArray(v, parseTmplFunc); err != nil {
			return err
		}
	}

	return nil
}

// parseVariableFromMap parse to variable when the v is map.
func parseVariableFromMap(v any, parseTmplFunc func(string) (string, error)) error {
	for _, kv := range reflect.ValueOf(v).MapKeys() {
		val := reflect.ValueOf(v).MapIndex(kv)
		if vv, ok := val.Interface().(string); ok {
			if !tmpl.IsTmplSyntax(vv) {
				continue
			}

			newValue, err := parseTmplFunc(vv)
			if err != nil {
				return err
			}

			switch {
			case strings.EqualFold(newValue, "TRUE"):
				reflect.ValueOf(v).SetMapIndex(kv, reflect.ValueOf(true))
			case strings.EqualFold(newValue, "FALSE"):
				reflect.ValueOf(v).SetMapIndex(kv, reflect.ValueOf(false))
			default:
				reflect.ValueOf(v).SetMapIndex(kv, reflect.ValueOf(newValue))
			}
		} else {
			if err := parseVariable(val.Interface(), parseTmplFunc); err != nil {
				return err
			}
		}
	}

	return nil
}

// parseVariableFromArray parse to variable when the v is slice.
func parseVariableFromArray(v any, parseTmplFunc func(string) (string, error)) error {
	for i := range reflect.ValueOf(v).Len() {
		val := reflect.ValueOf(v).Index(i)
		if vv, ok := val.Interface().(string); ok {
			if !tmpl.IsTmplSyntax(vv) {
				continue
			}

			newValue, err := parseTmplFunc(vv)
			if err != nil {
				return err
			}

			switch {
			case strings.EqualFold(newValue, "TRUE"):
				val.Set(reflect.ValueOf(true))
			case strings.EqualFold(newValue, "FALSE"):
				val.Set(reflect.ValueOf(false))
			default:
				val.Set(reflect.ValueOf(newValue))
			}
		} else {
			if err := parseVariable(val.Interface(), parseTmplFunc); err != nil {
				return err
			}
		}
	}

	return nil
}

// getLocalIP get the ipv4 or ipv6 for localhost machine
func getLocalIP(ipType string) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		klog.ErrorS(err, "get network address error")
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipType == _const.VariableIPv4 && ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}

			if ipType == _const.VariableIPv6 && ipNet.IP.To16() != nil && ipNet.IP.To4() == nil {
				return ipNet.IP.String()
			}
		}
	}

	klog.V(4).Infof("connot get local %s address", ipType)

	return ""
}

// StringVar get string value by key
func StringVar(d map[string]any, args map[string]any, key string) (string, error) {
	val, ok := args[key]
	if !ok {
		klog.V(4).InfoS("cannot find variable", "key", key)

		return "", fmt.Errorf("cannot find variable \"%s\"", key)
	}
	// convert to string
	sv, ok := val.(string)
	if !ok {
		klog.V(4).ErrorS(nil, "variable is not string", "key", key)

		return "", fmt.Errorf("variable \"%s\" is not string", key)
	}

	return tmpl.ParseString(d, sv)
}

// StringSliceVar get string slice value by key
func StringSliceVar(d map[string]any, vars map[string]any, key string) ([]string, error) {
	val, ok := vars[key]
	if !ok {
		klog.V(4).InfoS("cannot find variable", "key", key)

		return nil, fmt.Errorf("cannot find variable \"%s\"", key)
	}

	switch valv := val.(type) {
	case []any:
		var ss []string

		for _, a := range valv {
			av, ok := a.(string)
			if !ok {
				klog.V(6).InfoS("variable is not string", "key", key)

				return nil, nil
			}

			as, err := tmpl.ParseString(d, av)
			if err != nil {
				return nil, err
			}

			ss = append(ss, as)
		}

		return ss, nil
	case string:
		as, err := tmpl.ParseString(d, valv)
		if err != nil {
			klog.V(4).ErrorS(err, "parse variable error", "key", key)

			return nil, err
		}

		var ss []string
		if err := json.Unmarshal([]byte(as), &ss); err == nil {
			return ss, nil
		}

		return []string{as}, nil
	default:
		klog.V(4).ErrorS(nil, "unsupported variable type", "key", key)

		return nil, fmt.Errorf("unsupported variable \"%s\" type", key)
	}
}

// IntVar get int value by key
func IntVar(d map[string]any, vars map[string]any, key string) (*int, error) {
	val, ok := vars[key]
	if !ok {
		klog.V(4).InfoS("cannot find variable", "key", key)

		return nil, fmt.Errorf("cannot find variable \"%s\"", key)
	}
	// default convert to int
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ptr.To(int(v.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u := v.Uint()
		if u > uint64(^uint(0)>>1) {
			return nil, fmt.Errorf("variable \"%s\" value %d overflows int", key, u)
		}

		return ptr.To(int(u)), nil
	case reflect.Float32, reflect.Float64:
		return ptr.To(int(v.Float())), nil
	case reflect.String:
		vs, err := tmpl.ParseString(d, v.String())
		if err != nil {
			klog.V(4).ErrorS(err, "parse string variable error", "key", key)

			return nil, err
		}

		atoi, err := strconv.Atoi(vs)
		if err != nil {
			klog.V(4).ErrorS(err, "parse convert string to int error", "key", key)

			return nil, err
		}

		return ptr.To(atoi), nil
	default:
		klog.V(4).ErrorS(nil, "unsupported variable type", "key", key)

		return nil, fmt.Errorf("unsupported variable \"%s\" type", key)
	}
}

// BoolVar get bool value by key
func BoolVar(d map[string]any, args map[string]any, key string) (*bool, error) {
	val, ok := args[key]
	if !ok {
		klog.V(4).InfoS("cannot find variable", "key", key)

		return nil, fmt.Errorf("cannot find variable \"%s\"", key)
	}
	// default convert to int
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Bool:
		return ptr.To(v.Bool()), nil
	case reflect.String:
		vs, err := tmpl.ParseString(d, v.String())
		if err != nil {
			klog.V(4).ErrorS(err, "parse string variable error", "key", key)

			return nil, err
		}

		if strings.EqualFold(vs, "TRUE") {
			return ptr.To(true), nil
		}

		if strings.EqualFold(vs, "FALSE") {
			return ptr.To(false), nil
		}
	}

	return nil, fmt.Errorf("unsupported variable \"%s\" type", key)
}

// DurationVar get time.Duration value by key
func DurationVar(d map[string]any, args map[string]any, key string) (time.Duration, error) {
	stringVar, err := StringVar(d, args, key)
	if err != nil {
		return 0, err
	}

	return time.ParseDuration(stringVar)
}

// Extension2Variables convert runtime.RawExtension to variables
func Extension2Variables(ext runtime.RawExtension) map[string]any {
	if len(ext.Raw) == 0 {
		return make(map[string]any)
	}

	var data map[string]any
	if err := json.Unmarshal(ext.Raw, &data); err != nil {
		klog.V(4).ErrorS(err, "failed to unmarshal extension to variables")
	}

	return data
}

// Extension2Slice convert runtime.RawExtension to slice
// if runtime.RawExtension contains tmpl syntax, parse it.
func Extension2Slice(d map[string]any, ext runtime.RawExtension) []any {
	if len(ext.Raw) == 0 {
		return nil
	}

	var data []any
	// try parse yaml string which defined  by single value or multi value
	if err := json.Unmarshal(ext.Raw, &data); err == nil {
		return data
	}
	// try converter template string
	val, err := Extension2String(d, ext)
	if err != nil {
		klog.ErrorS(err, "extension2string error", "input", string(ext.Raw))
	}

	if err := json.Unmarshal([]byte(val), &data); err == nil {
		return data
	}

	return []any{val}
}

// Extension2String convert runtime.RawExtension to string.
// if runtime.RawExtension contains tmpl syntax, parse it.
func Extension2String(d map[string]any, ext runtime.RawExtension) (string, error) {
	if len(ext.Raw) == 0 {
		return "", nil
	}

	var input = string(ext.Raw)
	// try to escape string
	if ns, err := strconv.Unquote(string(ext.Raw)); err == nil {
		input = ns
	}

	result, err := tmpl.ParseString(d, input)
	if err != nil {
		return "", err
	}

	return result, nil
}
