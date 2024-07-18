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

package tmpl

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v6"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/version"
)

func init() {
	utilruntime.Must(pongo2.RegisterFilter("defined", filterDefined))
	utilruntime.Must(pongo2.RegisterFilter("version", filterVersion))
	utilruntime.Must(pongo2.RegisterFilter("pow", filterPow))
	utilruntime.Must(pongo2.RegisterFilter("match", filterMatch))
	utilruntime.Must(pongo2.RegisterFilter("to_json", filterToJson))
	utilruntime.Must(pongo2.RegisterFilter("to_yaml", filterToYaml))
	utilruntime.Must(pongo2.RegisterFilter("ip_range", filterIpRange))
	utilruntime.Must(pongo2.RegisterFilter("get", filterGet))
	utilruntime.Must(pongo2.RegisterFilter("rand", filterRand))
}

func filterDefined(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.IsNil() {
		return pongo2.AsValue(false), nil
	}
	return pongo2.AsValue(true), nil
}

func filterVersion(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	inVersion, err := version.ParseGeneric(in.String())
	if err != nil {
		return pongo2.AsValue(nil), &pongo2.Error{
			Sender:    "filter:version in",
			OrigError: err,
		}
	}
	paramString := strings.TrimSpace(param.String())
	switch {
	case strings.HasPrefix(paramString, ">="):
		compareVersion := strings.TrimSpace(paramString[2:])
		ci, err := inVersion.Compare(compareVersion)
		if err != nil {
			return pongo2.AsValue(nil), &pongo2.Error{
				Sender:    "filter:version",
				OrigError: fmt.Errorf("converter second param error: %w", err),
			}
		}
		return pongo2.AsValue(ci >= 0), nil
	case strings.HasPrefix(paramString, "<="):
		compareVersion := strings.TrimSpace(paramString[2:])
		ci, err := inVersion.Compare(compareVersion)
		if err != nil {
			return pongo2.AsValue(nil), &pongo2.Error{
				Sender:    "filter:version",
				OrigError: fmt.Errorf("converter second param error: %w", err),
			}
		}
		return pongo2.AsValue(ci <= 0), nil
	case strings.HasPrefix(paramString, "=="):
		compareVersion := strings.TrimSpace(paramString[2:])
		ci, err := inVersion.Compare(compareVersion)
		if err != nil {
			return pongo2.AsValue(nil), &pongo2.Error{
				Sender:    "filter:version",
				OrigError: fmt.Errorf("converter second param error: %w", err),
			}
		}
		return pongo2.AsValue(ci == 0), nil
	case strings.HasPrefix(paramString, ">"):
		compareVersion := strings.TrimSpace(paramString[1:])
		ci, err := inVersion.Compare(compareVersion)
		if err != nil {
			return pongo2.AsValue(nil), &pongo2.Error{
				Sender:    "filter:version",
				OrigError: fmt.Errorf("converter second param error: %w", err),
			}
		}
		return pongo2.AsValue(ci == 1), nil
	case strings.HasPrefix(paramString, "<"):
		compareVersion := strings.TrimSpace(paramString[1:])
		ci, err := inVersion.Compare(compareVersion)
		if err != nil {
			return pongo2.AsValue(nil), &pongo2.Error{
				Sender:    "filter:version",
				OrigError: fmt.Errorf("converter second param error: %w", err),
			}
		}
		return pongo2.AsValue(ci == -1), nil
	default:
		return pongo2.AsValue(nil), &pongo2.Error{
			Sender:    "filter:version",
			OrigError: fmt.Errorf("converter first param error: %w", err),
		}
	}
}

func filterPow(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(math.Pow(in.Float(), param.Float())), nil
}

func filterMatch(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	match, err := regexp.Match(param.String(), []byte(in.String()))
	if err != nil {
		return pongo2.AsValue(nil), &pongo2.Error{Sender: "filter:match", OrigError: err}
	}
	return pongo2.AsValue(match), nil
}

func filterToJson(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	data, err := json.Marshal(in.Interface())
	if err != nil {
		return pongo2.AsValue(nil), &pongo2.Error{
			Sender:    "to_json",
			OrigError: fmt.Errorf("parse in to json: %w", err),
		}
	}
	result := string(data)
	if param.IsInteger() {
		result = Indent(param.Integer(), result)
	}
	return pongo2.AsValue(result), nil
}

func filterToYaml(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.IsNil() {
		return pongo2.AsValue(nil), nil
	}
	data, err := yaml.Marshal(in.Interface())
	if err != nil {
		return pongo2.AsValue(nil), &pongo2.Error{
			Sender:    "to_yaml",
			OrigError: fmt.Errorf("parse in to json: %w", err),
		}
	}
	result := string(data)
	if result == "{}\n" || result == "{}" {
		return pongo2.AsValue(nil), nil
	}
	if !param.IsNil() && param.IsInteger() {
		result = Indent(param.Integer(), result)
	}
	return pongo2.AsValue(result), nil
}

func filterIpRange(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.IsNil() || !in.IsString() {
		return pongo2.AsValue(nil), &pongo2.Error{
			Sender:    "ip_range",
			OrigError: fmt.Errorf("input is not format string"),
		}
	}
	var ipRange = make([]string, 0)
	for _, s := range strings.Split(in.String(), ",") {
		ipRange = append(ipRange, ParseIp(s)...)
	}
	// if param is integer. return a single value
	if param.IsInteger() {
		index := param.Integer()
		// handle negative number
		if index < 0 {
			index = max(len(ipRange)+index, 0)
		}
		index = max(index, 0)
		index = min(index, len(ipRange)-1)
		return pongo2.AsValue(ipRange[index]), nil
	}
	if param.IsString() {
		comp := strings.Split(param.String(), ":")
		switch len(comp) {
		case 1: //  return a single value
			index := pongo2.AsValue(comp[0]).Integer()
			// handle negative number
			if index < 0 {
				index = max(len(ipRange)+index, 0)
			}
			index = max(index, 0)
			index = min(index, len(ipRange)-1)
			return pongo2.AsValue(ipRange[index]), nil
		case 2: // return a slice
			// start with [x:len]
			from := pongo2.AsValue(comp[0]).Integer()
			from = max(from, 0)
			from = min(from, len(ipRange)-1)

			to := pongo2.AsValue(comp[1]).Integer()
			// handle missing y
			if strings.TrimSpace(comp[1]) == "" {
				to = len(ipRange) - 1
			}
			to = max(to, from)
			to = min(to, len(ipRange)-1)

			return pongo2.AsValue(ipRange[from:to]), nil
		default:
			return nil, &pongo2.Error{
				Sender:    "filter:ip_range",
				OrigError: fmt.Errorf("ip_range string must have the format 'from:to' or a single number format 'index'"),
			}
		}
	}

	return pongo2.AsValue(ipRange), nil
}

// filterGet get value from map or array
func filterGet(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	var result *pongo2.Value
	in.Iterate(func(idx, count int, key, value *pongo2.Value) bool {
		if param.IsInteger() && idx == param.Integer() {
			result = in.Index(idx)
			return false
		}
		if param.IsString() && key.String() == param.String() {
			result = pongo2.AsValue(value.Interface())
			return false
		}
		return true
	}, func() {
		result = pongo2.AsValue(nil)
	})
	return result, nil
}

func filterRand(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	if !param.IsInteger() {
		return pongo2.AsValue(nil), &pongo2.Error{
			Sender:    "rand",
			OrigError: fmt.Errorf("param is not format int"),
		}
	}
	return pongo2.AsValue(rand.String(param.Integer())), nil
}
