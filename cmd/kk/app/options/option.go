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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
)

var defaultConfig = &kubekeyv1.Config{
	TypeMeta: metav1.TypeMeta{
		APIVersion: kubekeyv1.SchemeGroupVersion.String(),
		Kind:       "Config",
	},
	ObjectMeta: metav1.ObjectMeta{Name: "default"}}
var defaultInventory = &kubekeyv1.Inventory{
	TypeMeta: metav1.TypeMeta{
		APIVersion: kubekeyv1.SchemeGroupVersion.String(),
		Kind:       "Inventory",
	},
	ObjectMeta: metav1.ObjectMeta{Name: "default"}}

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
	// Artifact is the path of offline package for kubekey.
	Artifact string
	// Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.
	Debug bool
	// Namespace for all resources.
	Namespace string
}

func newCommonOptions() CommonOptions {
	o := CommonOptions{
		Namespace: metav1.NamespaceDefault,
	}
	wd, err := os.Getwd()
	if err != nil {
		klog.ErrorS(err, "get current dir error")
		o.WorkDir = "/tmp/kubekey"
	} else {
		o.WorkDir = filepath.Join(wd, "kubekey")
	}
	return o
}

func (o *CommonOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.StringVar(&o.WorkDir, "work-dir", o.WorkDir, "the base Dir for kubekey. Default current dir. ")
	gfs.StringVarP(&o.Artifact, "artifact", "a", "", "Path to a KubeKey artifact")
	gfs.StringVarP(&o.ConfigFile, "config", "c", o.ConfigFile, "the config file path. support *.yaml ")
	gfs.StringArrayVar(&o.Set, "set", o.Set, "set value in config. format --set key=val or --set k1=v1,k2=v2")
	gfs.StringVarP(&o.InventoryFile, "inventory", "i", o.InventoryFile, "the host list file path. support *.ini")
	gfs.BoolVarP(&o.Debug, "debug", "d", o.Debug, "Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.")
	gfs.StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "the namespace which pipeline will be executed, all reference resources(pipeline, config, inventory, task) should in the same namespace")
	return fss
}

func (o *CommonOptions) completeRef(pipeline *kubekeyv1.Pipeline) (*kubekeyv1.Config, *kubekeyv1.Inventory, error) {
	if !filepath.IsAbs(o.WorkDir) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, nil, fmt.Errorf("get current dir error: %w", err)
		}
		o.WorkDir = filepath.Join(wd, o.WorkDir)
	}

	config, err := o.genConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("generate config error: %w", err)
	}
	pipeline.Spec.ConfigRef = &corev1.ObjectReference{
		Kind:            config.Kind,
		Namespace:       config.Namespace,
		Name:            config.Name,
		UID:             config.UID,
		APIVersion:      config.APIVersion,
		ResourceVersion: config.ResourceVersion,
	}

	inventory, err := o.genInventory()
	if err != nil {
		return nil, nil, fmt.Errorf("generate inventory error: %w", err)
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

// genConfig generate config by ConfigFile and set value by command args.
func (o *CommonOptions) genConfig() (*kubekeyv1.Config, error) {
	config := defaultConfig.DeepCopy()
	if o.ConfigFile != "" {
		cdata, err := os.ReadFile(o.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("read config file error: %w", err)
		}
		config = &kubekeyv1.Config{}
		if err := yaml.Unmarshal(cdata, config); err != nil {
			return nil, fmt.Errorf("unmarshal config file error: %w", err)
		}
	}
	// set by command args
	if o.Namespace != "" {
		config.Namespace = o.Namespace
	}
	if wd, err := config.GetValue("work_dir"); err == nil && wd != nil {
		// if work_dir is defined in config, use it. otherwise use current dir.
		o.WorkDir = wd.(string)
	} else if err := config.SetValue("work_dir", o.WorkDir); err != nil {
		return nil, fmt.Errorf("work_dir to config error: %w", err)
	}
	if o.Artifact != "" {
		// override artifact_file in config
		if err := config.SetValue("artifact_file", o.Artifact); err != nil {
			return nil, fmt.Errorf("artifact file to config error: %w", err)
		}
	}
	for _, s := range o.Set {
		for _, setVal := range strings.Split(unescapeString(s), ",") {
			i := strings.Index(setVal, "=")
			if i == 0 || i == -1 {
				return nil, fmt.Errorf("--set value should be k=v")
			}
			if err := setValue(config, setVal[:i], setVal[i+1:]); err != nil {
				return nil, fmt.Errorf("--set value to config error: %w", err)
			}
		}
	}

	return config, nil
}

// genConfig generate config by ConfigFile and set value by command args.
func (o *CommonOptions) genInventory() (*kubekeyv1.Inventory, error) {
	inventory := defaultInventory.DeepCopy()
	if o.InventoryFile != "" {
		cdata, err := os.ReadFile(o.InventoryFile)
		if err != nil {
			klog.V(4).ErrorS(err, "read config file error")
			return nil, err
		}
		inventory = &kubekeyv1.Inventory{}
		if err := yaml.Unmarshal(cdata, inventory); err != nil {
			klog.V(4).ErrorS(err, "unmarshal config file error")
			return nil, err
		}
	}
	// set by command args
	if o.Namespace != "" {
		inventory.Namespace = o.Namespace
	}

	return inventory, nil
}

// setValue set key: val in config.
// if val is json string. convert to map or slice
// if val is TRUE,YES,Y. convert to bool type true.
// if val is FALSE,NO,N. convert to bool type false.
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
	case strings.EqualFold(val, "TRUE") || strings.EqualFold(val, "YES") || strings.EqualFold(val, "Y"):
		return config.SetValue(key, true)
	case strings.EqualFold(val, "FALSE") || strings.EqualFold(val, "NO") || strings.EqualFold(val, "N"):
		return config.SetValue(key, false)
	default:
		return config.SetValue(key, val)
	}
}

// unescapeString handles common escape sequences
func unescapeString(s string) string {
	replacements := map[string]string{
		`\\`: `\`,
		`\"`: `"`,
		`\'`: `'`,
		`\n`: "\n",
		`\r`: "\r",
		`\t`: "\t",
		`\b`: "\b",
		`\f`: "\f",
	}

	// Iterate over the replacements map and replace escape sequences in the string
	for o, n := range replacements {
		s = strings.ReplaceAll(s, o, n)
	}

	return s
}
