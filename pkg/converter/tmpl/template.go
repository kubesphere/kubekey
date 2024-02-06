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
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v6"
	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// ParseBool by pongo2 with not contain "{{ }}". It will add  "{{ }}" to input string.
func ParseBool(v variable.VariableData, inputs []string) (bool, error) {
	for _, input := range inputs {
		// first convert.
		intql, err := pongo2.FromString(input)
		if err != nil {
			klog.V(4).ErrorS(err, "Failed to get string")
			return false, err
		}
		inres, err := intql.Execute(pongo2.Context(v))
		if err != nil {
			klog.V(4).ErrorS(err, "Failed to execute string")
			return false, err
		}

		// trim line break.
		inres = strings.TrimSuffix(inres, "\n")
		inres = fmt.Sprintf("{{ %s }}", inres)

		// second convert.
		tql, err := pongo2.FromString(inres)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to get string")
			return false, err
		}
		result, err := tql.Execute(pongo2.Context(v))
		if err != nil {
			klog.V(4).ErrorS(err, "failed to execute string")
			return false, err
		}
		klog.V(4).InfoS(" parse template succeed", "result", result)
		if result != "True" {
			return false, nil
		}
	}
	return true, nil
}

// ParseString with contain "{{  }}"
func ParseString(v variable.VariableData, input string) (string, error) {
	tql, err := pongo2.FromString(input)
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to get string")
		return input, err
	}
	result, err := tql.Execute(pongo2.Context(v))
	if err != nil {
		klog.V(4).ErrorS(err, "Failed to execute string")
		return input, err
	}
	klog.V(4).InfoS(" parse template succeed", "result", result)
	return result, nil
}

func ParseFile(v variable.VariableData, file []byte) (string, error) {
	tql, err := pongo2.FromBytes(file)
	if err != nil {
		klog.V(4).ErrorS(err, "Transfer file to template error")
		return "", err
	}
	result, err := tql.Execute(pongo2.Context(v))
	if err != nil {
		klog.V(4).ErrorS(err, "exec template error")
		return "", err
	}
	klog.V(4).InfoS(" parse template succeed", "result", result)
	return result, nil
}
