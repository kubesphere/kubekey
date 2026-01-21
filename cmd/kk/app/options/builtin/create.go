//go:build builtin
// +build builtin

/*
Copyright 2024 The KubeSphere Authors.

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
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

// ======================================================================================
//                                  create cluster
// ======================================================================================

// NewCreateClusterOptions for newCreateClusterCommand
func NewCreateClusterOptions() *CreateClusterOptions {
	// set default value
	o := &CreateClusterOptions{
		CommonOptions: options.NewCommonOptions(),
		Kubernetes:    defaultKubeVersion,
	}
	o.GetInventoryFunc = getInventory

	return o
}

// CreateClusterOptions for NewCreateClusterOptions
type CreateClusterOptions struct {
	options.CommonOptions
	// kubernetes version which the cluster will install.
	Kubernetes string
}

// Flags add to newCreateClusterCommand
func (o *CreateClusterOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))

	return fss
}

// Complete options. create Playbook, Config and Inventory
func (o *CreateClusterOptions) Complete(cmd *cobra.Command, args []string) (*kkcorev1.Playbook, error) {
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "create-cluster-",
			Namespace:    o.Namespace,
			Annotations: map[string]string{
				kkcorev1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	if len(args) != 1 {
		return nil, errors.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}
	o.Playbook = args[0]

	playbook.Spec = kkcorev1.PlaybookSpec{
		Playbook: o.Playbook,
	}
	if err := o.CommonOptions.Complete(playbook); err != nil {
		return nil, err
	}

	return playbook, o.completeConfig()
}

func (o *CreateClusterOptions) completeConfig() error {
	if _, ok, _ := unstructured.NestedFieldNoCopy(o.Config.Value(), "kubernetes", "kube_version"); !ok {
		if err := unstructured.SetNestedField(o.Config.Value(), o.Kubernetes, "kubernetes", "kube_version"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "kube_version")
		}
	}

	return nil
}

// ======================================================================================
//                                  create config
// ======================================================================================

// NewCreateConfigOptions for newCreateConfigCommand
func NewCreateConfigOptions() *CreateConfigOptions {
	// set default value
	return &CreateConfigOptions{
		Kubernetes: defaultKubeVersion,
	}
}

// CreateConfigOptions for NewCreateConfigOptions
type CreateConfigOptions struct {
	// kubernetes version which the config will install.
	Kubernetes string
	// OutputPath for config file (can be file or directory path)
	OutputPath string
}

// Flags add to newCreateConfigCommand
func (o *CreateConfigOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	kfs := fss.FlagSet("config")
	kfs.StringVar(&o.Kubernetes, "with-kubernetes", o.Kubernetes, fmt.Sprintf("Specify a supported version of kubernetes. default is %s", o.Kubernetes))
	kfs.StringVarP(&o.OutputPath, "output", "o", o.OutputPath, "Output path for config file (file or directory). if not set will output to stdout")

	return fss
}

// Run executes the create config operation. It reads the default config file for the specified
// Kubernetes version and either writes it to the specified output directory or prints it to stdout.
// If an output directory is specified, it creates a config file named "config-<kubernetes-version>.yaml".
func (o *CreateConfigOptions) Run() error {
	// Read the default config file for the specified Kubernetes version
	data, err := getConfig(o.Kubernetes)
	if err != nil {
		return err
	}
	if o.OutputPath != "" {
		// Resolve output path (file or directory)
		filename, err := resolveOutputPath(o.OutputPath, fmt.Sprintf("config-%s.yaml", o.Kubernetes))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filename, data, 0644); err != nil {
			return errors.Wrapf(err, "failed to write config file to %s", filename)
		}
		fmt.Printf("write config file to %s success.\n", filename)
	} else {
		// Print config to stdout if no output directory specified
		fmt.Println(string(data))
	}

	return nil
}

// ======================================================================================
//                                  create inventory
// ======================================================================================

// NewCreateInventoryOptions for newCreateInventoryCommand
func NewCreateInventoryOptions() *CreateInventoryOptions {
	// set default value
	return &CreateInventoryOptions{}
}

// CreateInventoryOptions for NewCreateInventoryOptions
type CreateInventoryOptions struct {
	// OutputPath for inventory file (can be file or directory path)
	OutputPath string
}

// Flags add to newCreateInventoryCommand
func (o *CreateInventoryOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	kfs := fss.FlagSet("inventory")
	kfs.StringVarP(&o.OutputPath, "output", "o", o.OutputPath, "Output path for inventory file (file or directory). if not set will output to stdout")

	return fss
}

func (o *CreateInventoryOptions) Run() error {

	data, err := getInventoryData()
	if err != nil {
		return err
	}

	if o.OutputPath != "" {
		// Resolve output path (file or directory)
		filename, err := resolveOutputPath(o.OutputPath, "inventory.yaml")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filename, data, 0644); err != nil {
			return errors.Wrapf(err, "failed to write inventory file to %s", filename)
		}
		fmt.Printf("write inventory file to %s success.\n", filename)
	} else {
		// Print inventory to stdout if no output directory specified
		fmt.Println(string(data))
	}

	return nil
}

// resolveOutputPath determines if the output path is a file or directory.
// If it's a directory (or looks like one), it returns directory/defaultFilename.
// If it's a file (or looks like one), it returns the path directly.
func resolveOutputPath(output, defaultFilename string) (string, error) {
	// If output is empty, return empty (will use stdout)
	if output == "" {
		return "", nil
	}

	// Check if path exists
	info, err := os.Stat(output)
	if err == nil {
		// Path exists
		if info.IsDir() {
			// It's a directory, append default filename
			return filepath.Join(output, defaultFilename), nil
		}
		// It's a file, use it directly
		return output, nil
	}

	// Explicitly check if error is "not exist"
	if !os.IsNotExist(err) {
		// An error other than "not exist" occurred, e.g. permission denied
		return "", errors.Wrapf(err, "failed to stat output path %q", output)
	}

	// Path doesn't exist, use heuristics
	// If it has a file extension, treat as file
	ext := filepath.Ext(output)
	if ext != "" {
		return output, nil
	}

	// If it ends with separator, treat as directory
	if strings.HasSuffix(output, string(filepath.Separator)) {
		return filepath.Join(output, defaultFilename), nil
	}

	// Default: treat as directory (backward compatible)
	return filepath.Join(output, defaultFilename), nil
}
