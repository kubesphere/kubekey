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
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

// // CTX cancel by shutdown signal
// var CTX = signals.SetupSignalHandler()

// CommonOptions holds the configuration options for executing a playbook.
// It includes paths to various configuration files, runtime settings, and
// debug options.
type CommonOptions struct {
	// Playbook specifies the playbook to execute.
	Playbook string
	// InventoryFile is the path to the host file.
	InventoryFile string
	// ConfigFile is the path to the configuration file.
	ConfigFile string
	// Set contains values to set in the configuration.
	Set []string
	// Workdir is the base directory where the command finds any resources (e.g., project files).
	Workdir string
	// Artifact is the path to the offline package for kubekey.
	Artifact string
	// Debug indicates whether to retain runtime data after a successful execution of the playbook.
	// This includes task execution status and parameters.
	Debug bool
	// Namespace specifies the namespace for all resources.
	Namespace string

	// Config is the kubekey core configuration.
	Config *kkcorev1.Config
	// Inventory is the kubekey core inventory.
	Inventory *kkcorev1.Inventory
}

// NewCommonOptions creates a new CommonOptions object with default values.
// It sets the default namespace, working directory, and initializes the Config and Inventory objects.
func NewCommonOptions() CommonOptions {
	o := CommonOptions{
		Namespace: metav1.NamespaceDefault,
	}

	// Set the working directory to the current directory joined with "kubekey".
	wd, err := os.Getwd()
	if err != nil {
		klog.ErrorS(err, "get current dir error")
		o.Workdir = "/root/kubekey"
	} else {
		o.Workdir = filepath.Join(wd, "kubekey")
	}

	// Initialize the Config object with default API version and kind.
	o.Config = &kkcorev1.Config{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kkcorev1.SchemeGroupVersion.String(),
			Kind:       "Config",
		},
	}

	// Initialize the Inventory object with default API version, kind, and name.
	o.Inventory = &kkcorev1.Inventory{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kkcorev1.SchemeGroupVersion.String(),
			Kind:       "Inventory",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
	}

	return o
}

// Run executes the main command logic for the application.
// It sets up the necessary configurations, creates the inventory and playbook
// resources, and then runs the command manager.
func (o *CommonOptions) Run(ctx context.Context, playbook *kkcorev1.Playbook) error {
	// create workdir directory,if not exists
	if _, err := os.Stat(o.Workdir); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to stat local dir %q for playbook %q", o.Workdir, ctrlclient.ObjectKeyFromObject(playbook))
		}
		// if dir is not exist, create it.
		if err := os.MkdirAll(o.Workdir, os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to create local dir %q for playbook %q", o.Workdir, ctrlclient.ObjectKeyFromObject(playbook))
		}
	}
	restconfig := &rest.Config{}
	if err := proxy.RestConfig(filepath.Join(o.Workdir, _const.RuntimeDir), restconfig); err != nil {
		return err
	}
	client, err := ctrlclient.New(restconfig, ctrlclient.Options{
		Scheme: _const.Scheme,
	})
	if err != nil {
		return errors.Wrap(err, "failed to runtime-client")
	}
	// create inventory
	if err := client.Create(ctx, o.Inventory); err != nil {
		return errors.Wrap(err, "failed to create inventory")
	}
	// create playbook
	if err := client.Create(ctx, playbook); err != nil {
		return errors.Wrap(err, "failed to create playbook")
	}

	return manager.NewCommandManager(manager.CommandManagerOptions{
		Workdir:   o.Workdir,
		Playbook:  playbook,
		Config:    o.Config,
		Inventory: o.Inventory,
		Client:    client,
	}).Run(ctx)
}

// Flags returns a NamedFlagSets object that contains the command-line flags
// for the CommonOptions. These flags can be used to configure the CommonOptions
// from the command line.
func (o *CommonOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.StringVar(&o.Workdir, "workdir", o.Workdir, "the base Dir for kubekey. Default current dir. ")
	gfs.StringVarP(&o.Artifact, "artifact", "a", "", "Path to a KubeKey artifact")
	gfs.StringVarP(&o.ConfigFile, "config", "c", o.ConfigFile, "the config file path. support *.yaml ")
	gfs.StringArrayVar(&o.Set, "set", o.Set, "set value in config. format --set key=val or --set k1=v1,k2=v2")
	gfs.StringVarP(&o.InventoryFile, "inventory", "i", o.InventoryFile, "the host list file path. support *.yaml")
	gfs.BoolVarP(&o.Debug, "debug", "d", o.Debug, "Debug mode, after a successful execution of Playbook, will retain runtime data, which includes task execution status and parameters.")
	gfs.StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "the namespace which playbook will be executed, all reference resources(playbook, config, inventory, task) should in the same namespace")

	return fss
}

// Complete finalizes the CommonOptions by setting up the working directory,
// generating the configuration, and completing the inventory reference for the playbook.
func (o *CommonOptions) Complete(playbook *kkcorev1.Playbook) error {
	// Ensure the working directory is an absolute path.
	if !filepath.IsAbs(o.Workdir) {
		wd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "get current dir error")
		}
		o.Workdir = filepath.Join(wd, o.Workdir)
	}
	// Generate and complete the configuration.
	if err := o.completeConfig(); err != nil {
		return err
	}
	playbook.Spec.Config = ptr.Deref(o.Config, kkcorev1.Config{})
	// Complete the inventory reference.
	o.completeInventory(o.Inventory)
	playbook.Spec.InventoryRef = &corev1.ObjectReference{
		Kind:            o.Inventory.Kind,
		Namespace:       o.Inventory.Namespace,
		Name:            o.Inventory.Name,
		UID:             o.Inventory.UID,
		APIVersion:      o.Inventory.APIVersion,
		ResourceVersion: o.Inventory.ResourceVersion,
	}

	return nil
}

// genConfig generate config by ConfigFile and set value by command args.
func (o *CommonOptions) completeConfig() error {
	// set value by command args
	if o.Workdir != "" {
		if err := unstructured.SetNestedField(o.Config.Value(), o.Workdir, _const.Workdir); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", _const.Workdir)
		}
	}
	if o.Artifact != "" {
		// override artifact_file in config
		if err := unstructured.SetNestedField(o.Config.Value(), o.Artifact, "artifact_file"); err != nil {
			return errors.Wrapf(err, "failed to set %q to config", "artifact_file")
		}
	}
	for _, s := range o.Set {
		for _, setVal := range strings.Split(unescapeString(s), ",") {
			i := strings.Index(setVal, "=")
			if i == 0 || i == -1 {
				return errors.New("--set value should be k=v")
			}
			if err := setValue(o.Config, setVal[:i], setVal[i+1:]); err != nil {
				return err
			}
		}
	}

	return nil
}

// genConfig generate config by ConfigFile and set value by command args.
func (o *CommonOptions) completeInventory(inventory *kkcorev1.Inventory) {
	// set value by command args
	if o.Namespace != "" {
		inventory.Namespace = o.Namespace
	}
}

// setValue set key: val in config.
// If val is json string. convert to map or slice
// If val is TRUE,YES,Y. convert to bool type true.
// If val is FALSE,NO,N. convert to bool type false.
func setValue(config *kkcorev1.Config, key, val string) error {
	switch {
	case strings.HasPrefix(val, "{") && strings.HasSuffix(val, "{"):
		var value map[string]any
		err := json.Unmarshal([]byte(val), &value)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal json object value for \"--set %s\"", key)
		}

		return errors.Wrapf(unstructured.SetNestedMap(config.Value(), value, key),
			"failed to set \"--set %s\" to config", key)
	case strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]"):
		var value []any
		err := json.Unmarshal([]byte(val), &value)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal json array value for \"--set %s\"", key)
		}

		return errors.Wrapf(unstructured.SetNestedSlice(config.Value(), value, key),
			"failed to set \"--set %s\" to config", key)
	case strings.EqualFold(val, "TRUE") || strings.EqualFold(val, "YES") || strings.EqualFold(val, "Y"):
		return errors.Wrapf(unstructured.SetNestedField(config.Value(), true, key),
			"failed to set \"--set %s\" to config", key)
	case strings.EqualFold(val, "FALSE") || strings.EqualFold(val, "NO") || strings.EqualFold(val, "N"):
		return errors.Wrapf(unstructured.SetNestedField(config.Value(), false, key),
			"failed to set \"--set %s\" to config", key)
	default:
		return errors.Wrapf(unstructured.SetNestedField(config.Value(), val, key),
			"failed to set \"--set %s\" to config", key)
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
