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
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
)

// combineVariables merge multiple variables into one variable
// v2 will override v1 if variable is repeated
func combineVariables(v1, v2 map[string]any) map[string]any {
	var f func(val1, val2 any) any
	f = func(val1, val2 any) any {
		if val1 != nil && reflect.TypeOf(val1).Kind() == reflect.Map &&
			val2 != nil && reflect.TypeOf(val2).Kind() == reflect.Map {
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
	for k, v := range v1 {
		mv[k] = v
	}
	for k, v := range v2 {
		mv[k] = f(mv[k], v)
	}
	return mv

}

func convertGroup(inv kubekeyv1.Inventory) map[string]any {
	groups := make(map[string]any)
	all := []string{"localhost"} // set default host
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

// StringVar get string value by key
func StringVar(d map[string]any, args map[string]any, key string) (string, error) {
	val, ok := args[key]
	if !ok {
		return "", fmt.Errorf("cannot find variable \"%s\"", key)
	}

	sv, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("variable \"%s\" is not string", key)
	}
	return tmpl.ParseString(d, sv)
}

// StringSliceVar get string slice value by key
func StringSliceVar(d map[string]any, vars map[string]any, key string) ([]string, error) {
	val, ok := vars[key]
	if !ok {
		return nil, fmt.Errorf("cannot find variable \"%s\"", key)
	}
	switch val.(type) {
	case []any:
		var ss []string
		for _, a := range val.([]any) {
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
		as, err := tmpl.ParseString(d, val.(string))
		if err != nil {
			return nil, err
		}
		var ss []string
		switch {
		case regexp.MustCompile(`^<\[\](.*?) Value>$`).MatchString(as):
			// in pongo2 cannot get slice value. add extension filter value.
			var input = val.(string)
			// try to escape string
			if ns, err := strconv.Unquote(val.(string)); err == nil {
				input = ns
			}
			vv := GetValue(d, input)
			if _, ok := vv.([]any); ok {
				ss = make([]string, len(vv.([]any)))
				for i, a := range vv.([]any) {
					ss[i] = a.(string)
				}
			}
		default:
			// value is simple string
			if err := json.Unmarshal([]byte(as), &ss); err != nil {
				// if is not json format. only return a value contains this
				return []string{as}, nil
			}
		}
		return ss, nil
	default:
		return nil, fmt.Errorf("unsupport variable \"%s\" type", key)
	}
}

// IntVar get int value by key
func IntVar(d map[string]any, vars map[string]any, key string) (int, error) {
	val, ok := vars[key]
	if !ok {
		return 0, fmt.Errorf("cannot find variable \"%s\"", key)
	}
	// default convert to float64
	switch val.(type) {
	case float64:
		return int(val.(float64)), nil
	case string:
		vs, err := tmpl.ParseString(d, val.(string))
		if err != nil {
			return 0, err
		}
		return strconv.Atoi(vs)
	default:
		return 0, fmt.Errorf("unsupport variable \"%s\" type", key)
	}
}

// Extension2Variables convert extension to variables
func Extension2Variables(ext runtime.RawExtension) map[string]any {
	if len(ext.Raw) == 0 {
		return nil
	}

	var data map[string]any
	if err := yaml.Unmarshal(ext.Raw, &data); err != nil {
		klog.V(4).ErrorS(err, "failed to unmarshal extension to variables")
	}
	return data
}

// Extension2Slice convert extension to slice
func Extension2Slice(d map[string]any, ext runtime.RawExtension) []any {
	if len(ext.Raw) == 0 {
		return nil
	}

	var data []any
	if err := yaml.Unmarshal(ext.Raw, &data); err == nil {
		return data
	}

	val, err := Extension2String(d, ext)
	if err != nil {
		klog.ErrorS(err, "extension2string error", "input", string(ext.Raw))
	}
	// parse value by pongo2. if
	switch {
	case regexp.MustCompile(`^<\[\](.*?) Value>$`).MatchString(val):
		// in pongo2 cannot get slice value. add extension filter value.
		var input = string(ext.Raw)
		// try to escape string
		if ns, err := strconv.Unquote(string(ext.Raw)); err == nil {
			input = ns
		}
		vv := GetValue(d, input)
		if _, ok := vv.([]any); ok {
			return vv.([]any)
		}
	default:
		// value is simple string
		return []any{val}
	}

	return data
}

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

// GetValue from VariableData by key path
func GetValue(value map[string]any, keys string) any {
	switch {
	case strings.HasPrefix(keys, "{{") && strings.HasSuffix(keys, "}}"):
		// the keys like {{ a.b.c }}. return value[a][b][c]
		var result any = value
		for _, k := range strings.Split(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(keys, "{{"), "}}")), ".") {
			result = result.(map[string]any)[k]
		}
		return result
	default:
		return nil
	}
}

// parseVariable parse all string values to the actual value.
func parseVariable(v any, parseTmplFunc func(string) (string, error)) error {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Map:
		for _, kv := range reflect.ValueOf(v).MapKeys() {
			val := reflect.ValueOf(v).MapIndex(kv)
			if vv, ok := val.Interface().(string); ok {
				if tmpl.IsTmplSyntax(vv) {
					newValue, err := parseTmplFunc(vv)
					if err != nil {
						return err
					}
					switch {
					case strings.ToUpper(newValue) == "TRUE":
						reflect.ValueOf(v).SetMapIndex(kv, reflect.ValueOf(true))
					case strings.ToUpper(newValue) == "FALSE":
						reflect.ValueOf(v).SetMapIndex(kv, reflect.ValueOf(false))
					default:
						reflect.ValueOf(v).SetMapIndex(kv, reflect.ValueOf(newValue))
					}
				}
			} else {
				if err := parseVariable(val.Interface(), parseTmplFunc); err != nil {
					return err
				}
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < reflect.ValueOf(v).Len(); i++ {
			val := reflect.ValueOf(v).Index(i)
			if vv, ok := val.Interface().(string); ok {
				if tmpl.IsTmplSyntax(vv) {
					newValue, err := parseTmplFunc(vv)
					if err != nil {
						return err
					}
					switch {
					case strings.ToUpper(newValue) == "TRUE":

						val.Set(reflect.ValueOf(true))
					case strings.ToUpper(newValue) == "FALSE":
						val.Set(reflect.ValueOf(false))
					default:
						val.Set(reflect.ValueOf(newValue))
					}
				}
			} else {
				if err := parseVariable(val.Interface(), parseTmplFunc); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
