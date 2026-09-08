//go:build builtin
// +build builtin

/*
Copyright 2026 The KubeSphere Authors.

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

package builtin

import (
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	cliflag "k8s.io/component-base/cli/flag"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/convert"
)

// NewConvertOptions for newConvertCommand
func NewConvertOptions() *ConvertOptions {
	return &ConvertOptions{}
}

// ConvertOptions for NewConvertOptions
type ConvertOptions struct {
	// Input is the path of the KubeKey v3 (v1alpha2) cluster configuration file.
	Input string
	// OutputPath is the output directory for the generated files. When empty,
	// both documents are printed to stdout as a multi-document YAML stream.
	OutputPath string
}

// Flags add to newConvertCommand
func (o *ConvertOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	cfs := fss.FlagSet("convert")
	cfs.StringVar(&o.Input, "input", o.Input, "Path of the KubeKey v3 (v1alpha2) cluster configuration file (required)")
	cfs.StringVarP(&o.OutputPath, "output", "o", o.OutputPath, "Output directory for inventory.yaml and config.yaml. if not set will output to stdout")

	return fss
}

// Run executes the conversion: it reads the v3 cluster configuration, converts
// it and either writes inventory.yaml/config.yaml to the output directory or
// prints both documents to stdout.
func (o *ConvertOptions) Run() error {
	if o.Input == "" {
		return errors.New("--input is required, please set it to the path of the v3 cluster configuration file")
	}
	data, err := os.ReadFile(o.Input)
	if err != nil {
		return errors.Wrapf(err, "failed to read v3 cluster configuration file %q", o.Input)
	}

	inventoryYAML, configYAML, warnings, err := convert.ConvertInventoryAndConfig(data)
	if err != nil {
		return err
	}

	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", w)
	}
	if len(warnings) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d field(s) could not be converted automatically, please review the output files.\n", len(warnings))
	}

	if o.OutputPath == "" {
		fmt.Printf("---\n%s\n---\n%s\n", inventoryYAML, configYAML)
		return nil
	}

	info, err := os.Stat(o.OutputPath)
	if err != nil {
		return errors.Wrapf(err, "failed to access output path %q, it must be an existing directory", o.OutputPath)
	}
	if !info.IsDir() {
		return errors.Errorf("output path %q is not a directory, please set -o to the directory for inventory.yaml and config.yaml", o.OutputPath)
	}

	for name, content := range map[string][]byte{
		"inventory.yaml": inventoryYAML,
		"config.yaml":    configYAML,
	} {
		filename := fmt.Sprintf("%s/%s", o.OutputPath, name)
		if err := os.WriteFile(filename, content, _const.PermFilePublic); err != nil {
			return errors.Wrapf(err, "failed to write %s", filename)
		}
		fmt.Printf("write %s to %s success.\n", name, filename)
	}

	return nil
}
