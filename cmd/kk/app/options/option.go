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
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

// InventoryFunc defines a function type that returns a pointer to a kkcorev1.Inventory and an error.
// It is used to provide a custom way to retrieve or generate an Inventory object.
type InventoryFunc func() (*kkcorev1.Inventory, error)

// ConfigFunc defines a function type that returns a pointer to a kkcorev1.Config and an error.
// It is used to provide a custom way to retrieve or generate a Config object.
type ConfigFunc func() (*kkcorev1.Config, error)

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
	// Namespace specifies the namespace for all resources.
	Namespace string

	// Config is the kubekey core configuration.
	Config *kkcorev1.Config
	// Inventory is the kubekey core inventory.
	Inventory        *kkcorev1.Inventory
	GetInventoryFunc InventoryFunc
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
		ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault, Name: "default"},
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
	// create or update inventory
	if err := client.Get(ctx, ctrlclient.ObjectKeyFromObject(o.Inventory), o.Inventory); err != nil {
		if apierrors.IsNotFound(err) {
			if err := client.Create(ctx, o.Inventory); err != nil {
				return errors.Wrap(err, "failed to create inventory")
			}
		} else {
			return errors.Wrap(err, "failed to get inventory")
		}
	} else {
		if err := client.Update(ctx, o.Inventory); err != nil {
			return errors.Wrap(err, "failed to update inventory")
		}
	}
	// create playbook
	if err := client.Create(ctx, playbook); err != nil {
		return errors.Wrap(err, "failed to create playbook")
	}

	return manager.NewCommandManager(manager.CommandManagerOptions{
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

	if o.ConfigFile != "" {
		data, err := os.ReadFile(o.ConfigFile)
		if err != nil {
			return errors.Wrapf(err, "failed to get config from file %q", o.ConfigFile)
		}
		if err := yaml.Unmarshal(data, o.Config); err != nil {
			return errors.Wrapf(err, "failed to unmarshal config from file %q", o.ConfigFile)
		}
	}
	if o.InventoryFile != "" {
		data, err := os.ReadFile(o.InventoryFile)
		if err != nil {
			return errors.Wrapf(err, "failed to get inventory from file %q", o.InventoryFile)
		}
		if err := yaml.Unmarshal(data, o.Inventory); err != nil {
			return errors.Wrapf(err, "failed to unmarshal inventory from file %q", o.InventoryFile)
		}
	} else if o.GetInventoryFunc != nil {
		inventory, err := o.GetInventoryFunc()
		if err != nil {
			return err
		}
		o.Inventory = inventory
	}

	// Complete the configuration.
	if err := o.completeConfig(); err != nil {
		return err
	}
	playbook.Spec.Config = ptr.Deref(o.Config, kkcorev1.Config{})
	// Complete the inventory reference.
	if err := o.completeInventory(o.Inventory); err != nil {
		return err
	}
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
		for _, setVal := range strings.Split(s, ",") {
			if err := parseAndSetValue(o.Config, setVal); err != nil {
				return err
			}
		}
	}

	return nil
}

// genConfig generate config by ConfigFile and set value by command args.
func (o *CommonOptions) completeInventory(inventory *kkcorev1.Inventory) error {
	// set value by command args
	if o.Namespace != "" {
		inventory.Namespace = o.Namespace
	}

	return nil
}

// parseAndSetValue parses a key=value pair and sets the value in the config.
// It supports Helm-style --set syntax:
// - Simple values: key=value
// - Nested objects: outer.inner=value
// - Array indexing: array[0]=value
// - Array element nesting: array[0].field=value
// - Multiple values: key1=value1,key2=value2
func parseAndSetValue(config *kkcorev1.Config, setVal string) error {
	i := strings.Index(setVal, "=")
	if i == 0 || i == -1 {
		return errors.New("--set value should be k=v")
	}

	key := setVal[:i]
	val := unescapeString(setVal[i+1:])

	// Parse the key to extract field path with array indices
	fieldPath := parseKey(key)

	return setNestedValue(config, fieldPath, val)
}

// parseKey parses a key string into a field path, handling array indices.
// Examples:
// - "a.b.c" -> ["a", "b", "c"]
// - "a[0].b" -> ["a", "0", "b"]
// - "a[0][1].b" -> ["a", "0", "1", "b"]
func parseKey(key string) []string {
	var result []string
	var current strings.Builder

	for i := 0; i < len(key); i++ {
		switch key[i] {
		case '[':
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		case ']':
			// Skip
		case '.':
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(key[i])
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// setNestedValue sets a value in the config based on a field path and string value.
// It supports different value types:
// - JSON objects (starting with '{' and ending with '}')
// - JSON arrays (starting with '[' and ending with ']')
// - Boolean values (true/false, yes/no, y/n - case insensitive)
// - Numeric values (integers and floats)
// - String values (default case)
// It also supports array indexing in the field path (e.g., items[0].field)
func setNestedValue(config *kkcorev1.Config, fieldPath []string, val string) error {
	var value any
	var err error

	switch {
	case strings.HasPrefix(val, "{") && strings.HasSuffix(val, "}"):
		err = json.Unmarshal([]byte(val), &value)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal json object value for \"--set %s\"", strings.Join(fieldPath, "."))
		}
	case strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]"):
		err = json.Unmarshal([]byte(val), &value)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal json array value for \"--set %s\"", strings.Join(fieldPath, "."))
		}
	case strings.EqualFold(val, "TRUE") || strings.EqualFold(val, "YES") || strings.EqualFold(val, "Y"):
		value = true
	case strings.EqualFold(val, "FALSE") || strings.EqualFold(val, "NO") || strings.EqualFold(val, "N"):
		value = false
	case isNumeric(val):
		// Try to parse as integer first, then float
		if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
			value = intVal
		} else if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			value = floatVal
		} else {
			value = val
		}
	default:
		value = val
	}

	// Navigate through the field path and set the value
	// Handle array indices properly
	if err := setValueByPath(config.Value(), fieldPath, value); err != nil {
		return errors.Wrapf(err, "failed to set \"--set %s=%s\" to config", strings.Join(fieldPath, "."), val)
	}

	return nil
}

// setValueByPath sets a value at the specified field path, handling array indices.
// fieldPath can contain numeric strings like "0", "1", etc. which are treated as array indices.
func setValueByPath(root map[string]interface{}, fieldPath []string, value any) error {
	if len(fieldPath) == 0 {
		return nil
	}

	return setValueAtPath(root, fieldPath, 0, value)
}

// setValueAtPath recursively navigates to the correct position and sets the value.
func setValueAtPath(current map[string]interface{}, fieldPath []string, index int, value any) error {
	if index >= len(fieldPath) {
		return nil
	}

	field := fieldPath[index]

	// Check if this field is an array index
	arrayIndex, isArrayIndex := parseArrayIndex(field)

	if isArrayIndex {
		// This is an array index, but we need to look at the previous field to get the array
		// This case should be handled by setArrayElement
		return errors.Errorf("invalid field path: array index %d found without parent field", arrayIndex)
	}

	// This is a regular field name
	nextIndex := index + 1
	if nextIndex >= len(fieldPath) {
		// We're setting this field directly
		current[field] = value
		return nil
	}

	// Check if the next part is an array index
	nextField := fieldPath[nextIndex]
	nextArrayIndex, isNextArrayIndex := parseArrayIndex(nextField)

	if isNextArrayIndex {
		// The next field is an array index, navigate to the array element
		return setArrayElementWithNestedValue(current, field, nextArrayIndex, fieldPath[nextIndex+1:], value)
	}

	// The next field is a regular field, navigate to nested object
	if current[field] == nil {
		current[field] = make(map[string]interface{})
	}

	nested, ok := current[field].(map[string]interface{})
	if !ok {
		return errors.Errorf("field %q is not a nested object", field)
	}

	return setValueAtPath(nested, fieldPath, nextIndex, value)
}

// parseArrayIndex checks if a string is a valid array index (non-negative integer).
// Returns the integer index and true if it's an array index, otherwise 0 and false.
func parseArrayIndex(s string) (int, bool) {
	// Empty string is not a valid index
	if s == "" {
		return 0, false
	}

	idx, err := strconv.Atoi(s)
	if err != nil || idx < 0 {
		return 0, false
	}

	return idx, true
}

// setArrayElementWithNestedValue sets a nested value inside an array element.
func setArrayElementWithNestedValue(current map[string]interface{}, field string, index int, remainingPath []string, value any) error {
	arr, exists := current[field]
	if !exists {
		// Create the array with the appropriate size
		arr = make([]any, index+1)
		current[field] = arr
	}

	arrSlice, ok := arr.([]any)
	if !ok {
		return errors.Errorf("field %q is not an array", field)
	}

	// Expand the array if necessary
	for len(arrSlice) <= index {
		arrSlice = append(arrSlice, nil)
	}

	// If there are no remaining path elements, set the value directly
	if len(remainingPath) == 0 {
		arrSlice[index] = value
		current[field] = arrSlice
		return nil
	}

	// Ensure the element at index is a map for nested access
	if arrSlice[index] == nil {
		arrSlice[index] = make(map[string]interface{})
	}

	elem, ok := arrSlice[index].(map[string]interface{})
	if !ok {
		return errors.Errorf("array element at index %d is not a map", index)
	}

	// Navigate to the nested field and set the value
	return setNestedField(elem, remainingPath[0], remainingPath[1:], value)
}

// setNestedField sets a nested value in a map.
func setNestedField(current map[string]interface{}, field string, remainingPath []string, value any) error {
	if len(remainingPath) == 0 {
		current[field] = value
		return nil
	}

	// Check if the next field is an array index
	nextField := remainingPath[0]
	nextIndex, isArrayIndex := parseArrayIndex(nextField)

	if isArrayIndex {
		// The next field is an array index
		arr, exists := current[field]
		if !exists {
			// Create the array with empty objects
			arr = make([]any, nextIndex+1)
			current[field] = arr
		}

		arrSlice, ok := arr.([]any)
		if !ok {
			return errors.Errorf("field %q is not an array", field)
		}

		// Expand the array if necessary
		for len(arrSlice) <= nextIndex {
			arrSlice = append(arrSlice, nil)
		}

		// Ensure the element at index is a map
		if arrSlice[nextIndex] == nil {
			arrSlice[nextIndex] = make(map[string]interface{})
		}

		elem, ok := arrSlice[nextIndex].(map[string]interface{})
		if !ok {
			return errors.Errorf("array element at index %d is not a map", nextIndex)
		}

		// Continue navigating
		return setNestedField(elem, remainingPath[1], remainingPath[2:], value)
	} else {
		// The next field is a regular field
		if current[field] == nil {
			current[field] = make(map[string]interface{})
		}

		nested, ok := current[field].(map[string]interface{})
		if !ok {
			return errors.Errorf("field %q is not a nested object", field)
		}

		return setNestedField(nested, nextField, remainingPath[1:], value)
	}
}

// isNumeric checks if a string represents a numeric value
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// unescapeString handles common escape sequences
func unescapeString(s string) string {
	replacements := map[string]string{
		`\\`: `\`,
		`\,`: `,`,
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
