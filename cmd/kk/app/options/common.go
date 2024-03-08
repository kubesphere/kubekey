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

package options

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/project"
)

type CommonOptions struct {
	// Playbook which to execute.
	Playbook string
	// HostFile is the path of host file
	InventoryFile string
	// ConfigFile is the path of config file
	ConfigFile string
	// WorkDir is the baseDir which command find any resource (project etc.)
	WorkDir string
	// Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.
	Debug bool
}

func (o *CommonOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.StringVar(&o.WorkDir, "work-dir", o.WorkDir, "the base Dir for kubekey. Default current dir. ")
	gfs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "the config file path. support *.yaml ")
	gfs.StringVar(&o.InventoryFile, "inventory", o.InventoryFile, "the host list file path. support *.ini")
	gfs.BoolVar(&o.Debug, "debug", o.Debug, "Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.")
	return fss
}

func completeRef(pipeline *kubekeyv1.Pipeline, configFile string, inventoryFile string) (*kubekeyv1.Config, *kubekeyv1.Inventory, error) {
	config, err := genConfig(configFile)
	if err != nil {
		klog.V(4).ErrorS(err, "generate config error", "file", configFile)
		return nil, nil, err
	}
	pipeline.Spec.ConfigRef = &corev1.ObjectReference{
		Kind:            config.Kind,
		Namespace:       config.Namespace,
		Name:            config.Name,
		UID:             config.UID,
		APIVersion:      config.APIVersion,
		ResourceVersion: config.ResourceVersion,
	}

	inventory, err := genInventory(inventoryFile)
	if err != nil {
		klog.V(4).ErrorS(err, "generate inventory error", "file", inventoryFile)
		return nil, nil, err
	}
	pipeline.Spec.InventoryRef = &corev1.ObjectReference{
		Kind:            inventory.Kind,
		Namespace:       inventory.Namespace,
		Name:            inventory.Name,
		UID:             inventory.UID,
		APIVersion:      inventory.APIVersion,
		ResourceVersion: inventory.ResourceVersion,
	}

	return config, inventory, nil
}

func genConfig(configFile string) (*kubekeyv1.Config, error) {
	var (
		config = &kubekeyv1.Config{}
		cdata  []byte
		err    error
	)
	if configFile != "" {
		cdata, err = os.ReadFile(configFile)
	} else {
		cdata, err = project.InternalPipeline.ReadFile(_const.BuiltinConfigFile)
	}
	if err != nil {
		klog.V(4).ErrorS(err, "read config file error")
		return nil, err
	}
	if err := yaml.Unmarshal(cdata, config); err != nil {
		klog.V(4).ErrorS(err, "unmarshal config file error")
		return nil, err
	}
	if config.Namespace == "" {
		config.Namespace = corev1.NamespaceDefault
	}
	return config, nil
}

func genInventory(inventoryFile string) (*kubekeyv1.Inventory, error) {
	var (
		inventory = &kubekeyv1.Inventory{}
		cdata     []byte
		err       error
	)
	if inventoryFile != "" {
		cdata, err = os.ReadFile(inventoryFile)
	} else {
		cdata, err = project.InternalPipeline.ReadFile(_const.BuiltinInventoryFile)
	}
	if err != nil {
		klog.V(4).ErrorS(err, "read config file error")
		return nil, err
	}
	if err := yaml.Unmarshal(cdata, inventory); err != nil {
		klog.V(4).ErrorS(err, "unmarshal config file error")
		return nil, err
	}
	if inventory.Namespace == "" {
		inventory.Namespace = corev1.NamespaceDefault
	}
	return inventory, nil
}
