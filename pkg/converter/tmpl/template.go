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
	"bytes"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/converter/internal"
)

// ParseBool parse template string to bool
func ParseBool(ctx map[string]any, inputs []string) (bool, error) {
	for _, input := range inputs {
		if !IsTmplSyntax(input) {
			input = "{{ " + input + " }}"
		}

		tl, err := internal.Template.Parse(input)
		if err != nil {
			return false, fmt.Errorf("failed to parse template '%s': %w", input, err)
		}

		result := bytes.NewBuffer(nil)
		if err := tl.Execute(result, ctx); err != nil {
			return false, fmt.Errorf("failed to execute template '%s': %w", input, err)
		}
		klog.V(6).InfoS(" parse template succeed", "result", result.String())
		if result.String() != "true" {
			return false, nil
		}
	}

	return true, nil
}

// ParseString parse template string to actual string
func ParseString(ctx map[string]any, input string) (string, error) {
	if !IsTmplSyntax(input) {
		return input, nil
	}

	tl, err := internal.Template.Parse(input)
	if err != nil {
		return "", fmt.Errorf("failed to parse template '%s': %w", input, err)
	}

	result := bytes.NewBuffer(nil)
	if err := tl.Execute(result, ctx); err != nil {
		return "", fmt.Errorf("failed to execute template '%s': %w", input, err)
	}
	klog.V(6).InfoS(" parse template succeed", "result", result.String())

	return strings.TrimSpace(result.String()), nil
}

// IsTmplSyntax Check if the string conforms to the template syntax.
func IsTmplSyntax(s string) bool {
	return strings.Contains(s, "{{") && strings.Contains(s, "}}")
}
