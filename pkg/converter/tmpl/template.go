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

	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/converter/internal"
)

// ParseFunc parses a template string using the provided context and parse function.
// It takes a context map C, an input string that may contain template syntax,
// and a parse function that converts the template result to the desired Output type.
// If the input is not a template, it directly applies the parse function.
// For template inputs, it parses and executes the template with the context,
// then applies the parse function to the result.
// Returns the parsed output and any error encountered during template processing.
func ParseFunc[C ~map[string]any, Output any](ctx C, input string, f func([]byte) Output) (Output, error) {
	// If input doesn't contain template syntax, return directly
	if !kkprojectv1.IsTmplSyntax(input) {
		return f(bytes.Trim([]byte(input), "\r\n")), nil
	}
	// Parse the template string
	tl, err := internal.Template.Parse(input)
	if err != nil {
		return f(nil), fmt.Errorf("failed to parse template '%s': %w", input, err)
	}
	// Execute template with provided context
	result := bytes.NewBuffer(nil)
	if err := tl.Execute(result, ctx); err != nil {
		return f(nil), fmt.Errorf("failed to execute template '%s': %w", input, err)
	}
	// Log successful parsing
	klog.V(6).InfoS(" parse template succeed", "result", result.String())

	// Apply parse function to result and return
	return f(bytes.Trim(result.Bytes(), "\r\n")), nil
}

// Parse is a helper function that wraps ParseFunc to directly return bytes.
// It takes a context map C and input string, and returns the parsed bytes.
func Parse[C ~map[string]any](ctx C, input string) ([]byte, error) {
	return ParseFunc(ctx, input, func(o []byte) []byte {
		return o
	})
}

// ParseBool parse template string to bool
func ParseBool(ctx map[string]any, inputs ...string) (bool, error) {
	for _, input := range inputs {
		output, err := ParseFunc(ctx, input, func(o []byte) bool {
			return bytes.EqualFold(o, []byte("true"))
		})
		if err != nil {
			return false, err
		}
		if !output {
			return output, nil
		}
	}

	return true, nil
}
