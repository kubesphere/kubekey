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

package internal

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	apistorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
	"k8s.io/klog/v2"
)

const (
	// when delete resource, add suffix deleteTagSuffix to the file name.
	// after delete event is handled, the file will be deleted from disk.
	deleteTagSuffix = "-deleted"
	// the file type of resource will store local.
	// NOTE: the variable will store in playbook dir, add file-suffix to distinct it.
	yamlSuffix = ".yaml"
)

func newFileStorage(prefix string, resource schema.GroupResource, codec runtime.Codec, newFunc func() runtime.Object) (apistorage.Interface, factory.DestroyFunc) {
	return &fileStorage{
			prefix:    prefix,
			versioner: apistorage.APIObjectVersioner{},
			resource:  resource,
			codec:     codec,
			newFunc:   newFunc,
		}, func() {
			// do nothing
		}
}

type fileStorage struct {
	prefix    string
	versioner apistorage.Versioner
	codec     runtime.Codec
	resource  schema.GroupResource

	newFunc func() runtime.Object
}

// ReadinessCheck implements storage.Interface.
func (s *fileStorage) ReadinessCheck() error {
	// only need filesystem is ok. nothing to check.
	return nil
}

var _ apistorage.Interface = &fileStorage{}

// Versioner of local resource files.
func (s fileStorage) Versioner() apistorage.Versioner {
	return s.versioner
}

// Create local resource files.
func (s fileStorage) Create(_ context.Context, key string, obj, out runtime.Object, _ uint64) error {
	// set resourceVersion to obj
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return errors.Wrapf(err, "failed to get meta from object %q", key)
	}
	metaObj.SetResourceVersion("1")
	// create file to local disk
	if _, err := os.Stat(filepath.Dir(key)); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "failed to check dir %q", filepath.Dir(key))
		}
		if err := os.MkdirAll(filepath.Dir(key), os.ModePerm); err != nil {
			return errors.Wrapf(err, "failed to create dir %q", filepath.Dir(key))
		}
	}

	data, err := runtime.Encode(s.codec, obj)
	if err != nil {
		return errors.Wrapf(err, "failed to encode object %q", key)
	}
	if err := decode(s.codec, data, out); err != nil {
		return err
	}
	// render to file
	if err := os.WriteFile(key+yamlSuffix, data, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create object %q", key)
	}

	return nil
}

// Delete local resource files.
func (s fileStorage) Delete(ctx context.Context, key string, out runtime.Object, preconditions *apistorage.Preconditions, validateDeletion apistorage.ValidateObjectFunc, cachedExistingObject runtime.Object) error {
	if cachedExistingObject != nil {
		out = cachedExistingObject
	} else {
		if err := s.Get(ctx, key, apistorage.GetOptions{}, out); err != nil {
			return err
		}
	}

	if err := preconditions.Check(key, out); err != nil {
		return errors.Wrapf(err, "failed to check preconditions for object %q", key)
	}

	if err := validateDeletion(ctx, out); err != nil {
		return err
	}

	// delete object: rename file to trigger watcher, it will actual delete by watcher.
	if err := os.Rename(key+yamlSuffix, key+yamlSuffix+deleteTagSuffix); err != nil {
		return errors.Wrapf(err, "failed to rename object %q to del-object %q", key+yamlSuffix, key+yamlSuffix+deleteTagSuffix)
	}

	return nil
}

// Watch local resource files.
func (s fileStorage) Watch(_ context.Context, key string, _ apistorage.ListOptions) (watch.Interface, error) {
	return newFileWatcher(s.prefix, key, s.codec, s.newFunc)
}

// Get local resource files.
func (s fileStorage) Get(_ context.Context, key string, _ apistorage.GetOptions, out runtime.Object) error {
	data, err := os.ReadFile(key + yamlSuffix)
	if err != nil {
		return errors.Wrapf(err, "failed to read object file %q", key)
	}

	return decode(s.codec, data, out)
}

// GetList local resource files.
func (s fileStorage) GetList(_ context.Context, key string, opts apistorage.ListOptions, listObj runtime.Object) error {
	listPtr, err := meta.GetItemsPtr(listObj)
	if err != nil {
		return errors.Wrapf(err, "failed to get object list items of %q", key)
	}
	v, err := conversion.EnforcePtr(listPtr)
	if err != nil || v.Kind() != reflect.Slice {
		return errors.Wrapf(err, "need ptr to slice of %q", key)
	}

	// Build matching rules for resource version and continue key.
	resourceVersionMatchRule, continueKeyMatchRule, err := s.buildMatchRules(key, opts, &sync.Once{})
	if err != nil {
		return err
	}

	// Get the root entries in the directory corresponding to 'key'.
	rootEntries, isAllNamespace, err := s.getRootEntries(key)
	if err != nil {
		return err
	}

	var lastKey string
	var hasMore bool
	// Iterate over root entries, processing either directories or files.
	for i, entry := range rootEntries {
		if isAllNamespace {
			// Process namespace directory.
			err = s.processNamespaceDirectory(key, entry, v, continueKeyMatchRule, resourceVersionMatchRule, &lastKey, &hasMore, opts, listObj)
		} else {
			// Process individual resource file.
			err = s.processResourceFile(key, entry, v, continueKeyMatchRule, resourceVersionMatchRule, &lastKey, opts, listObj)
		}
		if err != nil {
			return err
		}
		// Check if we have reached the limit of results requested by the client.
		if opts.Predicate.Limit != 0 && int64(v.Len()) >= opts.Predicate.Limit {
			hasMore = i != len(rootEntries)-1

			break
		}
	}
	// Handle the final result after all entries have been processed.
	return s.handleResult(listObj, v, lastKey, hasMore)
}

// buildMatchRules creates the match rules for resource version and continue key based on the given options.
func (s fileStorage) buildMatchRules(key string, opts apistorage.ListOptions, startReadOnce *sync.Once) (func(uint64) bool, func(string) bool, error) {
	resourceVersionMatchRule := func(uint64) bool { return true }
	continueKeyMatchRule := func(key string) bool { return strings.HasSuffix(key, yamlSuffix) }

	switch {
	case opts.Recursive && opts.Predicate.Continue != "":
		// If continue token is present, set up a rule to start reading after the continueKey.
		continueKey, _, err := apistorage.DecodeContinue(opts.Predicate.Continue, key)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "invalid continue token of %q", key)
		}

		continueKeyMatchRule = func(key string) bool {
			startRead := false
			if key == continueKey {
				startReadOnce.Do(func() { startRead = true })
			}

			return startRead && key != continueKey
		}
	case opts.ResourceVersion != "":
		// Handle resource version matching based on the provided match rule.
		parsedRV, err := s.versioner.ParseResourceVersion(opts.ResourceVersion)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "invalid resource version of %q", key)
		}
		switch opts.ResourceVersionMatch {
		case metav1.ResourceVersionMatchNotOlderThan:
			resourceVersionMatchRule = func(u uint64) bool { return u >= parsedRV }
		case metav1.ResourceVersionMatchExact:
			resourceVersionMatchRule = func(u uint64) bool { return u == parsedRV }
		case "":
			// Legacy case: match all resource versions.
		default:
			return nil, nil, errors.Errorf("unknown ResourceVersionMatch value %q of %q", opts.ResourceVersionMatch, key)
		}
	}

	return resourceVersionMatchRule, continueKeyMatchRule, nil
}

// getRootEntries reads the directory entries at the given key path.
func (s fileStorage) getRootEntries(key string) ([]os.DirEntry, bool, error) {
	var allNamespace bool
	switch len(filepath.SplitList(strings.TrimPrefix(key, s.prefix))) {
	case 0: // read all namespace's resources
		// Traverse the resource storage directory. startRead after continueKey.
		// get all resources from key. key is runtimeDir
		allNamespace = true
	case 1: // read a namespace's resources
		// Traverse the resource storage directory. startRead after continueKey.
		// get all resources from key. key is runtimeDir
		allNamespace = false
	default:
		return nil, false, errors.Errorf("key is invalid: %s", key)
	}
	rootEntries, err := os.ReadDir(key)

	return rootEntries, allNamespace, err
}

// processNamespaceDirectory handles the traversal and processing of a namespace directory.
func (s fileStorage) processNamespaceDirectory(key string, ns os.DirEntry, v reflect.Value, continueKeyMatchRule func(string) bool, resourceVersionMatchRule func(uint64) bool, lastKey *string, hasMore *bool, opts apistorage.ListOptions, listObj runtime.Object) error {
	if !ns.IsDir() {
		// only need dir. skip
		return nil
	}
	nsDir := filepath.Join(key, ns.Name())
	entries, err := os.ReadDir(nsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return errors.Wrapf(err, "failed to read dir %q", nsDir)
	}

	for _, entry := range entries {
		err := s.processResourceFile(nsDir, entry, v, continueKeyMatchRule, resourceVersionMatchRule, lastKey, opts, listObj)
		if err != nil {
			return err
		}
		// Check if we have reached the limit of results requested by the client.
		if opts.Predicate.Limit != 0 && int64(v.Len()) >= opts.Predicate.Limit {
			*hasMore = true

			return nil
		}
	}

	return nil
}

// processResourceFile handles reading, decoding, and processing a single resource file.
func (s fileStorage) processResourceFile(parentDir string, entry os.DirEntry, v reflect.Value, continueKeyMatchRule func(string) bool, resourceVersionMatchRule func(uint64) bool, lastKey *string, opts apistorage.ListOptions, listObj runtime.Object) error {
	if entry.IsDir() {
		// only need file. skip
		return nil
	}
	currentKey := filepath.Join(parentDir, entry.Name())
	if !continueKeyMatchRule(currentKey) {
		return nil
	}

	data, err := os.ReadFile(currentKey)
	if err != nil {
		return errors.Wrapf(err, "failed to read object %q", currentKey)
	}

	obj, _, err := s.codec.Decode(data, nil, getNewItem(listObj, v))
	if err != nil {
		return errors.Wrapf(err, "failed to decode object %q", currentKey)
	}

	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return errors.Wrapf(err, "failed to get object %q meta", currentKey)
	}

	rv, err := s.versioner.ParseResourceVersion(metaObj.GetResourceVersion())
	if err != nil {
		return errors.Wrapf(err, "failed to parse resource version %q", metaObj.GetResourceVersion())
	}

	// Apply the resource version match rule.
	if !resourceVersionMatchRule(rv) {
		return nil
	}

	// Check if the object matches the given predicate.
	if matched, err := opts.Predicate.Matches(obj); err == nil && matched {
		v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
		*lastKey = currentKey
	}

	return nil
}

// handleResult processes and finalizes the result before returning it.
func (s fileStorage) handleResult(listObj runtime.Object, v reflect.Value, lastKey string, hasMore bool) error {
	if v.IsNil() {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	if hasMore {
		// If there are more results, set the continuation token for the next query.
		next, err := apistorage.EncodeContinue(lastKey+"\x00", "", 0)
		if err != nil {
			return errors.Wrapf(err, "failed to encode continue %q", lastKey)
		}

		return s.versioner.UpdateList(listObj, 1, next, nil)
	}

	// If no more results, return the final list without continuation.
	return s.versioner.UpdateList(listObj, 1, "", nil)
}

// GuaranteedUpdate local resource file.
func (s fileStorage) GuaranteedUpdate(ctx context.Context, key string, destination runtime.Object, ignoreNotFound bool, preconditions *apistorage.Preconditions, tryUpdate apistorage.UpdateFunc, cachedExistingObject runtime.Object) error {
	var oldObj runtime.Object
	if cachedExistingObject != nil {
		oldObj = cachedExistingObject
	} else {
		oldObj = s.newFunc()
		if err := s.Get(ctx, key, apistorage.GetOptions{IgnoreNotFound: ignoreNotFound}, oldObj); err != nil {
			return err
		}
	}
	if err := preconditions.Check(key, oldObj); err != nil {
		return errors.Wrapf(err, "failed to check preconditions %q", key)
	}
	// set resourceVersion to obj
	metaObj, err := meta.Accessor(oldObj)
	if err != nil {
		return errors.Wrapf(err, "failed to get object %q meta", filepath.Dir(key))
	}
	oldVersion, err := s.versioner.ParseResourceVersion(metaObj.GetResourceVersion())
	if err != nil {
		return errors.Wrapf(err, "failed to parse resource version %q", metaObj.GetResourceVersion())
	}
	out, _, err := tryUpdate(oldObj, apistorage.ResponseMeta{ResourceVersion: oldVersion + 1})
	if err != nil {
		return err
	}

	data, err := runtime.Encode(s.codec, out)
	if err != nil {
		return errors.Wrapf(err, "failed to encode resource file %q", key)
	}
	// render to destination
	if destination != nil {
		err = decode(s.codec, data, destination)
		if err != nil {
			return err
		}
	}
	// render to file
	if err := os.WriteFile(key+yamlSuffix, data, os.ModePerm); err != nil {
		return errors.Wrapf(err, "failed to create resource file %q", key)
	}

	return nil
}

// Count local resource file
func (s fileStorage) Count(key string) (int64, error) {
	// countByNSDir count the crd files by namespace dir.
	countByNSDir := func(dir string) (int64, error) {
		var count int64
		entries, err := os.ReadDir(dir)
		if err != nil { // cannot read namespace dir
			return 0, errors.Wrapf(err, "failed to read namespaces dir %q", dir)
		}
		// count the file
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), yamlSuffix) {
				count++
			}
		}

		return count, nil
	}

	switch len(filepath.SplitList(strings.TrimPrefix(key, s.prefix))) {
	case 0: // count all namespace's resources
		var count int64
		rootEntries, err := os.ReadDir(key)
		if err != nil && !os.IsNotExist(err) {
			return 0, errors.Wrapf(err, "failed to read runtime dir %q", key)
		}
		for _, ns := range rootEntries {
			if !ns.IsDir() {
				continue
			}
			// the next dir is namespace.
			c, err := countByNSDir(filepath.Join(key, ns.Name()))
			if err != nil {
				return 0, err
			}
			count += c
		}

		return count, nil
	case 1: // count a namespace's resources
		return countByNSDir(key)
	default:
		// not support key
		return 0, errors.Errorf("key is invalid: %s", key)
	}
}

// RequestWatchProgress do nothing.
func (s fileStorage) RequestWatchProgress(context.Context) error {
	return nil
}

// decode decodes value of bytes into object. It will also set the object resource version to rev.
// On success, objPtr would be set to the object.
func decode(codec runtime.Codec, value []byte, objPtr runtime.Object) error {
	if _, err := conversion.EnforcePtr(objPtr); err != nil {
		return errors.Wrap(err, "failed to convert output object to pointer")
	}
	_, _, err := codec.Decode(value, nil, objPtr)
	if err != nil {
		return errors.Wrap(err, "failed to decode output object")
	}

	return nil
}

func getNewItem(listObj runtime.Object, v reflect.Value) runtime.Object {
	// For unstructured lists with a target group/version, preserve the group/version in the instantiated list items
	if unstructuredList, isUnstructured := listObj.(*unstructured.UnstructuredList); isUnstructured {
		if apiVersion := unstructuredList.GetAPIVersion(); apiVersion != "" {
			return &unstructured.Unstructured{Object: map[string]any{"apiVersion": apiVersion}}
		}
	}
	// Otherwise just instantiate an empty item
	elem := v.Type().Elem()
	if obj, ok := reflect.New(elem).Interface().(runtime.Object); ok {
		return obj
	}
	klog.V(6).Info("elem is not runtime.Object")

	return nil
}
