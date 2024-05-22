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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere/kubekey/v4/builtin"
	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
)

type CommonOptions struct {
	// Playbook which to execute.
	Playbook string
	// HostFile is the path of host file
	InventoryFile string
	// ConfigFile is the path of config file
	ConfigFile string
	// Set value in config
	Set []string
	// WorkDir is the baseDir which command find any resource (project etc.)
	WorkDir string
	// Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.
	Debug bool
}

func newCommonOptions() CommonOptions {
	o := CommonOptions{}
	wd, err := os.Getwd()
	if err != nil {
		klog.ErrorS(err, "get current dir error")
		o.WorkDir = "/tmp/kk"
	} else {
		o.WorkDir = wd
	}
	return o
}

func (o *CommonOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.StringVar(&o.WorkDir, "work-dir", o.WorkDir, "the base Dir for kubekey. Default current dir. ")
	gfs.StringVarP(&o.ConfigFile, "config", "c", o.ConfigFile, "the config file path. support *.yaml ")
	gfs.StringSliceVar(&o.Set, "set", o.Set, "set value in config. format --set key=val")
	gfs.StringVarP(&o.InventoryFile, "inventory", "i", o.InventoryFile, "the host list file path. support *.ini")
	gfs.BoolVarP(&o.Debug, "debug", "d", o.Debug, "Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.")
	return fss
}

func (o *CommonOptions) completeRef(pipeline *kubekeyv1.Pipeline) (*kubekeyv1.Config, *kubekeyv1.Inventory, error) {
	if !filepath.IsAbs(o.WorkDir) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, nil, fmt.Errorf("get current dir error: %v", err)
		}
		o.WorkDir = filepath.Join(wd, o.WorkDir)
	}

	config, err := genConfig(o.ConfigFile)
	if err != nil {
		return nil, nil, fmt.Errorf("generate config error: %v", err)
	}
	if wd, err := config.GetValue("work_dir"); err == nil && wd != nil {
		// if work_dir is defined in config, use it. otherwise use current dir.
		o.WorkDir = wd.(string)
	} else if err := config.SetValue("work_dir", o.WorkDir); err != nil {
		return nil, nil, fmt.Errorf("work_dir to config error: %v", err)
	}
	for _, s := range o.Set {
		ss := strings.Split(s, "=")
		if len(ss) != 2 {
			return nil, nil, fmt.Errorf("--set value should be k=v")
		}
		if err := setValue(config, ss[0], ss[1]); err != nil {
			return nil, nil, fmt.Errorf("--set value to config error: %v", err)
		}
	}

	pipeline.Spec.ConfigRef = &corev1.ObjectReference{
		Kind:            config.Kind,
		Namespace:       config.Namespace,
		Name:            config.Name,
		UID:             config.UID,
		APIVersion:      config.APIVersion,
		ResourceVersion: config.ResourceVersion,
	}

	inventory, err := genInventory(o.InventoryFile)
	if err != nil {
		return nil, nil, fmt.Errorf("generate inventory error: %v", err)
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
		cdata = builtin.DefaultConfig
	}
	if err != nil {
		return nil, fmt.Errorf("read config file error: %v", err)
	}
	if err := yaml.Unmarshal(cdata, config); err != nil {
		return nil, fmt.Errorf("unmarshal config file error: %v", err)
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
		cdata = builtin.DefaultInventory
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

func setValue(config *kubekeyv1.Config, key, val string) error {
	switch {
	case strings.HasPrefix(val, "{") && strings.HasSuffix(val, "{"):
		var value map[string]any
		err := json.Unmarshal([]byte(val), &value)
		if err != nil {
			return err
		}
		return config.SetValue(key, value)
	case strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]"):
		var value []any
		err := json.Unmarshal([]byte(val), &value)
		if err != nil {
			return err
		}
		return config.SetValue(key, value)
	default:
		return config.SetValue(key, val)
	}
}
