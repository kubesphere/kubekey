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
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
)

// CombineMappingNode combines two yaml.Node objects representing mapping nodes.
// If b is nil or zero, returns a.
// If both a and b are mapping nodes, appends a's content to b.
// Returns b in all other cases.
//
// Parameters:
//   - a: First yaml.Node to combine
//   - b: Second yaml.Node to combine
//
// Returns:
//   - Combined yaml.Node, with b taking precedence
func CombineMappingNode(a, b *yaml.Node) *yaml.Node {
	if b == nil || b.IsZero() {
		return a
	}

	if a.Kind == yaml.MappingNode && b.Kind == yaml.MappingNode {
		b.Content = append(b.Content, a.Content...)
	}

	return b
}

// CombineVariables merge multiple variables into one variable.
// It recursively combines two maps, where values from m2 override values from m1 if keys overlap.
// For nested maps, it will recursively merge their contents.
// For non-map values or when either input is nil, m2's value takes precedence.
//
// Parameters:
//   - m1: The first map to merge (base map)
//   - m2: The second map to merge (override map)
//
// Returns:
//   - A new map containing the merged key-value pairs from both input maps
func CombineVariables(m1, m2 map[string]any) map[string]any {
	var f func(val1, val2 any) any
	f = func(val1, val2 any) any {
		// If both values are non-nil maps, merge them recursively
		if val1 != nil && val2 != nil &&
			reflect.TypeOf(val1).Kind() == reflect.Map && reflect.TypeOf(val2).Kind() == reflect.Map {
			mergedVars := make(map[string]any)
			// Copy all values from val1 first
			for _, k := range reflect.ValueOf(val1).MapKeys() {
				mergedVars[k.String()] = reflect.ValueOf(val1).MapIndex(k).Interface()
			}

			// Merge in values from val2, recursively handling nested maps
			for _, k := range reflect.ValueOf(val2).MapKeys() {
				mergedVars[k.String()] = f(mergedVars[k.String()], reflect.ValueOf(val2).MapIndex(k).Interface())
			}

			return mergedVars
		}

		// For non-map values or nil inputs, return val2
		return val2
	}

	// Initialize result map
	mv := make(map[string]any)

	// Copy all key-value pairs from m1
	for k, v := range m1 {
		mv[k] = v
	}

	// Merge in values from m2
	for k, v := range m2 {
		mv[k] = f(mv[k], v)
	}

	return mv
}

// CombineSlice combines two string slices while skipping duplicate values.
// It maintains the order of elements from g1 followed by unique elements from g2.
//
// Parameters:
//   - g1: The first slice of strings
//   - g2: The second slice of strings
//
// Returns:
//   - A new slice containing unique strings from both input slices,
//     preserving order with g1 elements appearing before unique g2 elements
func CombineSlice(g1, g2 []string) []string {
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
//	map[string][]string: A map where keys are group names and values are lists of hostnames.
func ConvertGroup(inv kkcorev1.Inventory) map[string][]string {
	groups := make(map[string][]string)
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
		groups[gn] = hostsInGroup(inv, gn)
		if hosts, ok := groups[gn]; ok {
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

// hostsInGroup get a host_name slice in a given group
// if the given group contains other group. convert other group to host_name slice.
func hostsInGroup(inv kkcorev1.Inventory, groupName string) []string {
	if v, ok := inv.Spec.Groups[groupName]; ok {
		var hosts []string
		for _, cg := range v.Groups {
			hosts = CombineSlice(hostsInGroup(inv, cg), hosts)
		}

		return CombineSlice(hosts, v.Hosts)
	}

	return make([]string, 0)
}

// PrintVar get variable by key
func PrintVar(ctx map[string]any, keys ...string) (any, error) {
	sv, found, err := unstructured.NestedFieldNoCopy(ctx, keys...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !found {
		return nil, errors.Errorf("cannot find variable %q", strings.Join(keys, "."))
	}
	// try to marshal by json
	return sv, nil
}

// StringVar get string value by key
func StringVar(ctx map[string]any, args map[string]any, keys ...string) (string, error) {
	// convert to string
	sv, found, err := unstructured.NestedString(args, keys...)
	if err != nil {
		return "", errors.WithStack(err)
	}
	if !found {
		return "", errors.Errorf("cannot find variable %q", strings.Join(keys, "."))
	}

	return tmpl.ParseFunc(ctx, sv, func(b []byte) string { return string(b) })
}

// StringSliceVar get string slice value by key
func StringSliceVar(ctx map[string]any, args map[string]any, keys ...string) ([]string, error) {
	val, found, err := unstructured.NestedFieldNoCopy(args, keys...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !found {
		return nil, errors.Errorf("cannot find variable %q", strings.Join(keys, "."))
	}

	switch valv := val.(type) {
	case []string:
		var ss []string
		for _, a := range valv {
			as, err := tmpl.ParseFunc(ctx, a, func(b []byte) string { return string(b) })
			if err != nil {
				return nil, err
			}
			ss = append(ss, as)
		}

		return ss, nil
	case []any:
		var ss []string

		for _, a := range valv {
			av, ok := a.(string)
			if !ok {
				klog.V(6).InfoS("variable is not string", "key", keys)

				return nil, nil
			}

			as, err := tmpl.ParseFunc(ctx, av, func(b []byte) string { return string(b) })
			if err != nil {
				return nil, err
			}

			ss = append(ss, as)
		}

		return ss, nil
	case string:
		as, err := tmpl.Parse(ctx, valv)
		if err != nil {
			return nil, err
		}

		var ss []string
		if err := json.Unmarshal(as, &ss); err == nil {
			return ss, nil
		}

		return []string{string(as)}, nil
	default:
		return nil, errors.Errorf("unsupported variable %q type", strings.Join(keys, "."))
	}
}

// IntVar get int value by key
func IntVar(ctx map[string]any, args map[string]any, keys ...string) (*int, error) {
	val, found, err := unstructured.NestedFieldNoCopy(args, keys...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !found {
		return nil, errors.Errorf("cannot find variable %q", strings.Join(keys, "."))
	}

	// default convert to int
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ptr.To(int(v.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u := v.Uint()
		if u > uint64(^uint(0)>>1) {
			return nil, errors.Errorf("variable %q value %d overflows int", strings.Join(keys, "."), u)
		}

		return ptr.To(int(u)), nil
	case reflect.Float32, reflect.Float64:
		return ptr.To(int(v.Float())), nil
	case reflect.String:
		vs, err := tmpl.ParseFunc(ctx, v.String(), func(b []byte) string { return string(b) })
		if err != nil {
			return nil, err
		}

		atoi, err := strconv.Atoi(vs)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert string %q to int of key %q", vs, strings.Join(keys, "."))
		}

		return ptr.To(atoi), nil
	default:
		return nil, errors.Errorf("unsupported variable %q type", strings.Join(keys, "."))
	}
}

// BoolVar get bool value by key
func BoolVar(ctx map[string]any, args map[string]any, keys ...string) (*bool, error) {
	val, found, err := unstructured.NestedFieldNoCopy(args, keys...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !found {
		return nil, errors.Errorf("cannot find variable %q", strings.Join(keys, "."))
	}
	// default convert to int
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Bool:
		return ptr.To(v.Bool()), nil
	case reflect.String:
		vs, err := tmpl.ParseBool(ctx, v.String())
		if err != nil {
			return nil, err
		}

		return ptr.To(vs), nil
	}

	return nil, errors.Errorf("unsupported variable %q type", strings.Join(keys, "."))
}

// DurationVar get time.Duration value by key
func DurationVar(ctx map[string]any, args map[string]any, key string) (time.Duration, error) {
	stringVar, err := StringVar(ctx, args, key)
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
func Extension2Slice(ctx map[string]any, ext runtime.RawExtension) []any {
	if len(ext.Raw) == 0 {
		return nil
	}

	var data []any
	// try parse yaml string which defined  by single value or multi value
	if err := json.Unmarshal(ext.Raw, &data); err == nil {
		return data
	}
	// try converter template string
	val, err := Extension2String(ctx, ext)
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
func Extension2String(ctx map[string]any, ext runtime.RawExtension) (string, error) {
	if len(ext.Raw) == 0 {
		return "", nil
	}

	var input = string(ext.Raw)
	// try to escape string
	if ns, err := strconv.Unquote(string(ext.Raw)); err == nil {
		input = ns
	}

	result, err := tmpl.Parse(ctx, input)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
