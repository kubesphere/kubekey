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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

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
	yamlSuffix = ".yaml"
)

func newFileStorage(prefix string, resource schema.GroupResource, codec runtime.Codec, newFunc func() runtime.Object) (apistorage.Interface, factory.DestroyFunc, error) {
	return &fileStorage{
			prefix:    prefix,
			versioner: apistorage.APIObjectVersioner{},
			resource:  resource,
			codec:     codec,
			newFunc:   newFunc,
		}, func() {
			// do nothing
		}, nil
}

type fileStorage struct {
	prefix    string
	versioner apistorage.Versioner
	codec     runtime.Codec
	resource  schema.GroupResource

	newFunc func() runtime.Object
}

var _ apistorage.Interface = &fileStorage{}

func (s fileStorage) Versioner() apistorage.Versioner {
	return s.versioner
}

func (s fileStorage) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	// set resourceVersion to obj
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		klog.V(4).ErrorS(err, "failed to get meta object", "path", filepath.Dir(key))
		return err
	}
	metaObj.SetResourceVersion("1")
	// create file to local disk
	if _, err := os.Stat(filepath.Dir(key)); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(key), os.ModePerm); err != nil {
				klog.V(4).ErrorS(err, "failed to create dir", "path", filepath.Dir(key))
				return err
			}
		} else {
			klog.V(4).ErrorS(err, "failed to check dir", "path", filepath.Dir(key))
			return err
		}
	}

	data, err := runtime.Encode(s.codec, obj)
	if err != nil {
		klog.V(4).ErrorS(err, "failed to encode resource file", "path", key)
		return err
	}
	// render to out
	if out != nil {
		err = decode(s.codec, data, out)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to decode resource file", "path", key)
			return err
		}
	}
	// render to file
	if err := os.WriteFile(key+yamlSuffix, data, os.ModePerm); err != nil {
		klog.V(4).ErrorS(err, "failed to create resource file", "path", key)
		return err
	}
	return nil
}

func (s fileStorage) Delete(ctx context.Context, key string, out runtime.Object, preconditions *apistorage.Preconditions, validateDeletion apistorage.ValidateObjectFunc, cachedExistingObject runtime.Object) error {
	if cachedExistingObject != nil {
		out = cachedExistingObject
	} else {
		if err := s.Get(ctx, key, apistorage.GetOptions{}, out); err != nil {
			klog.V(4).ErrorS(err, "failed to get resource", "path", key)
			return err
		}
	}

	if err := preconditions.Check(key, out); err != nil {
		klog.V(4).ErrorS(err, "failed to check preconditions", "path", key)
		return err
	}

	if err := validateDeletion(ctx, out); err != nil {
		klog.V(4).ErrorS(err, "failed to validate deletion", "path", key)
		return err
	}

	// delete object
	// rename file to trigger watcher
	if err := os.Rename(key+yamlSuffix, key+yamlSuffix+deleteTagSuffix); err != nil {
		klog.V(4).ErrorS(err, "failed to rename resource file", "path", key)
		return err
	}
	return nil
}

func (s fileStorage) Watch(ctx context.Context, key string, opts apistorage.ListOptions) (watch.Interface, error) {
	return newFileWatcher(s.prefix, key, s.codec, s.newFunc)
}

func (s fileStorage) Get(ctx context.Context, key string, opts apistorage.GetOptions, out runtime.Object) error {
	data, err := os.ReadFile(key + yamlSuffix)
	if err != nil {
		klog.V(4).ErrorS(err, "failed to read resource file", "path", key)
		return err
	}
	if err := decode(s.codec, data, out); err != nil {
		klog.V(4).ErrorS(err, "failed to decode resource file", "path", key)
		return err
	}
	return nil
}

func (s fileStorage) GetList(ctx context.Context, key string, opts apistorage.ListOptions, listObj runtime.Object) error {
	listPtr, err := meta.GetItemsPtr(listObj)
	if err != nil {
		return err
	}
	v, err := conversion.EnforcePtr(listPtr)
	if err != nil || v.Kind() != reflect.Slice {
		return fmt.Errorf("need ptr to slice: %w", err)
	}

	// lastKey in result.
	var lastKey string
	var hasMore bool
	// resourceVersionMatchRule is a function that returns true if the resource version matches the rule.
	var resourceVersionMatchRule = func(uint64) bool {
		// default rule is to match all resource versions
		return true
	}
	var continueKeyMatchRule = func(key string) bool {
		// default rule
		return strings.HasSuffix(key, yamlSuffix)
	}

	switch {
	case opts.Recursive && opts.Predicate.Continue != "":
		// The format of continueKey is: namespace/resourceName/name.yaml
		// continueKey is localPath which resources store.
		continueKey, _, err := apistorage.DecodeContinue(opts.Predicate.Continue, key)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to parse continueKey", "continueKey", opts.Predicate.Continue)
			return fmt.Errorf("invalid continue token: %w", err)
		}
		startReadOnce := sync.Once{}
		continueKeyMatchRule = func(key string) bool {
			var startRead bool
			if key == continueKey {
				startReadOnce.Do(func() {
					startRead = true
				})
			}
			// start read after continueKey (not contain). Because it has read in last result.
			return startRead && key != continueKey
		}
	case opts.ResourceVersion != "":
		parsedRV, err := s.versioner.ParseResourceVersion(opts.ResourceVersion)
		if err != nil {
			return fmt.Errorf("invalid resource version: %w", err)
		}
		switch opts.ResourceVersionMatch {
		case metav1.ResourceVersionMatchNotOlderThan:
			resourceVersionMatchRule = func(u uint64) bool {
				return u >= parsedRV
			}
		case metav1.ResourceVersionMatchExact:
			resourceVersionMatchRule = func(u uint64) bool {
				return u == parsedRV
			}
		case "": // legacy case
			// use default rule. match all resource versions.
		default:
			return fmt.Errorf("unknown ResourceVersionMatch value: %v", opts.ResourceVersionMatch)
		}
	}

	switch len(filepath.SplitList(strings.TrimPrefix(key, s.prefix))) {
	case 0: // read all namespace's resources
		// Traverse the resource storage directory. startRead after continueKey.
		// Traverse the resource storage directory. startRead after continueKey.
		// get all resources from key. key is runtimeDir
		rootEntries, err := os.ReadDir(key)
		if err != nil && !os.IsNotExist(err) {
			klog.V(4).ErrorS(err, "failed to read runtime dir", "path", key)
			return err
		}
		for _, ns := range rootEntries {
			if !ns.IsDir() {
				continue
			}
			// the next dir is namespace.
			nsDir := filepath.Join(key, ns.Name())
			entries, err := os.ReadDir(nsDir)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				klog.V(4).ErrorS(err, "failed to read namespaces dir", "path", nsDir)
				return err
			}

			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				// the next file is resource name.
				currentKey := filepath.Join(nsDir, e.Name())
				if !continueKeyMatchRule(currentKey) {
					continue
				}
				data, err := os.ReadFile(currentKey)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					klog.V(4).ErrorS(err, "failed to read resource file", "path", currentKey)
					return err
				}

				obj, _, err := s.codec.Decode(data, nil, getNewItem(listObj, v))
				if err != nil {
					klog.V(4).ErrorS(err, "failed to decode resource file", "path", currentKey)
					return err
				}
				metaObj, err := meta.Accessor(obj)
				if err != nil {
					klog.V(4).ErrorS(err, "failed to get meta object", "path", currentKey)
					return err
				}
				rv, err := s.versioner.ParseResourceVersion(metaObj.GetResourceVersion())
				if err != nil {
					klog.V(4).ErrorS(err, "failed to parse resource version", "resourceVersion", obj.(metav1.Object).GetResourceVersion())
					return err
				}
				if !resourceVersionMatchRule(rv) {
					continue
				}
				if matched, err := opts.Predicate.Matches(obj); err == nil && matched {
					v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
					lastKey = currentKey
				}

				if opts.Predicate.Limit != 0 && int64(v.Len()) >= opts.Predicate.Limit {
					// got enough results. Stop the loop.
					goto RESULT
				}
			}
		}
		hasMore = false
	case 1: // read a namespace's resources
		// Traverse the resource storage directory. startRead after continueKey.
		// get all resources from key. key is runtimeDir
		rootEntries, err := os.ReadDir(key)
		if err != nil && !os.IsNotExist(err) {
			klog.V(4).ErrorS(err, "failed to read runtime dir", "path", key)
			return err
		}
		for _, rf := range rootEntries {
			if rf.IsDir() {
				continue
			}
			// the next file is resource name.
			currentKey := filepath.Join(key, rf.Name())
			if !continueKeyMatchRule(currentKey) {
				continue
			}
			data, err := os.ReadFile(currentKey)
			if err != nil {
				klog.V(4).ErrorS(err, "failed to read resource file", "path", currentKey)
				return err
			}

			obj, _, err := s.codec.Decode(data, nil, getNewItem(listObj, v))
			if err != nil {
				klog.V(4).ErrorS(err, "failed to decode resource file", "path", currentKey)
				return err
			}
			metaObj, err := meta.Accessor(obj)
			if err != nil {
				klog.V(4).ErrorS(err, "failed to get meta object", "path", currentKey)
				return err
			}
			rv, err := s.versioner.ParseResourceVersion(metaObj.GetResourceVersion())
			if err != nil {
				klog.V(4).ErrorS(err, "failed to parse resource version", "resourceVersion", obj.(metav1.Object).GetResourceVersion())
				return err
			}
			if !resourceVersionMatchRule(rv) {
				continue
			}
			if matched, err := opts.Predicate.Matches(obj); err == nil && matched {
				v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
				lastKey = currentKey
			}

			if opts.Predicate.Limit != 0 && int64(v.Len()) >= opts.Predicate.Limit {
				// got enough results. Stop the loop.
				goto RESULT
			}
		}
		hasMore = false
	default:
		klog.V(4).ErrorS(nil, "key is invalid", "key", key)
		return fmt.Errorf("key is invalid: %s", key)
	}

RESULT:
	if v.IsNil() {
		// Ensure that we never return a nil Items pointer in the result for consistency.
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	// instruct the client to begin querying from immediately after the last key we returned
	// we never return a key that the client wouldn't be allowed to see
	if hasMore {
		// we want to start immediately after the last key
		next, err := apistorage.EncodeContinue(lastKey+"\x00", key, 0)
		if err != nil {
			return err
		}
		// Unable to calculate remainingItemCount currently.
		// todo Store the resourceVersion in the file data. No resourceVersion strategy for List Object currently.
		// resourceVersion default set 1
		return s.versioner.UpdateList(listObj, 1, next, nil)
	}

	// no continuation
	// resourceVersion default set 1
	return s.versioner.UpdateList(listObj, 1, "", nil)
}

func (s fileStorage) GuaranteedUpdate(ctx context.Context, key string, destination runtime.Object, ignoreNotFound bool, preconditions *apistorage.Preconditions, tryUpdate apistorage.UpdateFunc, cachedExistingObject runtime.Object) error {
	var oldObj runtime.Object
	if cachedExistingObject != nil {
		oldObj = cachedExistingObject
	} else {
		oldObj = s.newFunc()
		if err := s.Get(ctx, key, apistorage.GetOptions{IgnoreNotFound: ignoreNotFound}, oldObj); err != nil {
			klog.V(4).ErrorS(err, "failed to get resource", "path", key)
			return err
		}
	}
	if err := preconditions.Check(key, oldObj); err != nil {
		klog.V(4).ErrorS(err, "failed to check preconditions", "path", key)
		return err
	}
	// set resourceVersion to obj
	metaObj, err := meta.Accessor(oldObj)
	if err != nil {
		klog.V(4).ErrorS(err, "failed to get meta object", "path", filepath.Dir(key))
		return err
	}
	oldVersion, err := s.versioner.ParseResourceVersion(metaObj.GetResourceVersion())
	if err != nil {
		klog.V(4).ErrorS(err, "failed to parse resource version", "resourceVersion", metaObj.GetResourceVersion())
		return err
	}
	out, _, err := tryUpdate(oldObj, apistorage.ResponseMeta{ResourceVersion: oldVersion + 1})
	if err != nil {
		klog.V(4).ErrorS(err, "failed to try update", "path", key)
		return err
	}

	data, err := runtime.Encode(s.codec, out)
	if err != nil {
		klog.V(4).ErrorS(err, "failed to encode resource file", "path", key)
		return err
	}
	// render to destination
	if destination != nil {
		err = decode(s.codec, data, destination)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to decode resource file", "path", key)
			return err
		}
	}
	// render to file
	if err := os.WriteFile(key+yamlSuffix, data, os.ModePerm); err != nil {
		klog.V(4).ErrorS(err, "failed to create resource file", "path", key)
		return err
	}
	return nil
}

func (s fileStorage) Count(key string) (int64, error) {
	switch len(filepath.SplitList(strings.TrimPrefix(key, s.prefix))) {
	case 0: // count all namespace's resources
		var count int64
		rootEntries, err := os.ReadDir(key)
		if err != nil && !os.IsNotExist(err) {
			klog.V(4).ErrorS(err, "failed to read runtime dir", "path", key)
			return 0, err
		}
		for _, ns := range rootEntries {
			if !ns.IsDir() {
				continue
			}
			// the next dir is namespace.
			nsDir := filepath.Join(key, ns.Name())
			entries, err := os.ReadDir(nsDir)
			if err != nil {
				klog.V(4).ErrorS(err, "failed to read namespaces dir", "path", nsDir)
				return 0, err
			}
			// count the file
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), yamlSuffix) {
					count++
				}
			}
		}
		return count, nil
	case 1: // count a namespace's resources
		var count int64
		rootEntries, err := os.ReadDir(key)
		if err != nil && !os.IsNotExist(err) {
			klog.V(4).ErrorS(err, "failed to read runtime dir", "path", key)
			return 0, err
		}
		for _, entry := range rootEntries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), yamlSuffix) {
				count++
			}
		}
		return count, nil
	default:
		klog.V(4).ErrorS(nil, "key is invalid", "key", key)
		return 0, fmt.Errorf("key is invalid: %s", key)
	}
}

func (s fileStorage) RequestWatchProgress(ctx context.Context) error {
	return nil
}

// decode decodes value of bytes into object. It will also set the object resource version to rev.
// On success, objPtr would be set to the object.
func decode(codec runtime.Codec, value []byte, objPtr runtime.Object) error {
	if _, err := conversion.EnforcePtr(objPtr); err != nil {
		return fmt.Errorf("unable to convert output object to pointer: %w", err)
	}
	_, _, err := codec.Decode(value, nil, objPtr)
	if err != nil {
		return err
	}
	return nil
}

func getNewItem(listObj runtime.Object, v reflect.Value) runtime.Object {
	// For unstructured lists with a target group/version, preserve the group/version in the instantiated list items
	if unstructuredList, isUnstructured := listObj.(*unstructured.UnstructuredList); isUnstructured {
		if apiVersion := unstructuredList.GetAPIVersion(); apiVersion != "" {
			return &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": apiVersion}}
		}
	}
	// Otherwise just instantiate an empty item
	elem := v.Type().Elem()
	return reflect.New(elem).Interface().(runtime.Object)
}
