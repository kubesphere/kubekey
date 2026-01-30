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
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/fsnotify/fsnotify"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strings"
)

// fileWatcher watcher local dir resource files.
type fileWatcher struct {
	prefix       string
	codec        runtime.Codec
	newFunc      func() runtime.Object
	watcher      *fsnotify.Watcher
	watchEvents  chan watch.Event
	cachedObject map[string]runtime.Object
}

// newFileWatcher get fileWatcher
func newFileWatcher(prefix, path string, codec runtime.Codec, newFunc func() runtime.Object, newList runtime.Object) (watch.Interface, error) {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "failed to stat path %q", path)
		}
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return nil, errors.Wrapf(err, "failed to create dir %q", path)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create file watcher %q", path)
	}
	if err := watcher.Add(path); err != nil {
		return nil, errors.Wrapf(err, "failed to add path to file watcher %q", path)
	}
	// add namespace dir to watcher
	if prefix == path {
		entry, err := os.ReadDir(prefix)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read dir %q", path)
		}
		for _, e := range entry {
			if e.IsDir() {
				p := filepath.Join(prefix, e.Name())
				if err := watcher.Add(p); err != nil {
					return nil, errors.Wrapf(err, "failed to add namespace dir to file watcher %q", e.Name())
				}
			}
		}
	}

	var newObject = make(map[string]runtime.Object)
	items, _ := meta.ExtractList(newList)
	var gvk schema.GroupVersionKind
	for _, item := range items {
		if itemGvk := item.GetObjectKind().GroupVersionKind(); !itemGvk.Empty() && gvk.Empty() {
			gvk = itemGvk
		}
		metaObj, err := meta.Accessor(item)
		if err != nil {
			continue
		}
		newObject[fmt.Sprintf("%s/%s", metaObj.GetNamespace(), metaObj.GetName())] = item
	}

	w := &fileWatcher{
		prefix:       prefix,
		codec:        codec,
		watcher:      watcher,
		newFunc:      newFunc,
		watchEvents:  make(chan watch.Event, 10),
		cachedObject: newObject,
	}

	go w.watch()

	w.handleBookMark(newList, items, gvk)

	return w, nil
}

func (w *fileWatcher) handleBookMark(listObj runtime.Object, items []runtime.Object, gvk schema.GroupVersionKind) {
	evt, err := handleBookmark(listObj, gvk)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.watchEvents <- *evt
}

func handleBookmark(obj runtime.Object, gvk schema.GroupVersionKind) (*watch.Event, error) {

	// 创建Bookmark对象
	bookmarkObj, err := createInitialEventsEndBookmark(gvk)
	if err != nil {
		return nil, fmt.Errorf("failed to create bookmark object: %w", err)
	}

	// 创建watch事件
	event := &watch.Event{
		Type:   watch.Bookmark,
		Object: bookmarkObj,
	}

	return event, nil
}

// createInitialEventsEndBookmark 创建带initial-events-end注解的Bookmark对象
func createInitialEventsEndBookmark(gvk schema.GroupVersionKind) (*unstructured.Unstructured, error) {
	// 创建Bookmark对象
	bookmarkObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": gvk.GroupVersion().String(),
			"kind":       gvk.Kind,
			"metadata": map[string]interface{}{
				"resourceVersion": "1",
				"annotations": map[string]interface{}{
					"k8s.io/initial-events-end": "true",
				},
				// 添加必要的元数据字段
				"creationTimestamp": metav1.Now(),
			},
		},
	}

	// 设置GVK
	bookmarkObj.SetGroupVersionKind(gvk)

	// 为Bookmark对象生成一个合理的UID
	bookmarkObj.SetUID(generateBookmarkUID(gvk, gvk.Version))

	return bookmarkObj, nil
}

// generateBookmarkUID 为Bookmark生成UID
func generateBookmarkUID(gvk schema.GroupVersionKind, resourceVersion string) types.UID {
	// 使用GVK和resourceVersion生成一个确定的UID
	uidStr := fmt.Sprintf("bookmark-%s-%s-%s-%s",
		gvk.Group,
		gvk.Version,
		gvk.Kind,
		resourceVersion)
	return types.UID(uidStr)
}

// Stop watch
func (w *fileWatcher) Stop() {
	close(w.watchEvents)
	if err := w.watcher.Close(); err != nil {
		klog.ErrorS(err, "failed to close file watcher")
	}
}

// ResultChan get watch event
func (w *fileWatcher) ResultChan() <-chan watch.Event {
	return w.watchEvents
}

func (w *fileWatcher) watch() {
	for {
		select {
		case event := <-w.watcher.Events:
			klog.V(6).InfoS("receive watcher event", "event", event)
			// Adjust the listening range. a watcher for a namespace.
			// the watcher contains all resources in the namespace.
			entry, err := os.Stat(event.Name)
			if err != nil {
				klog.V(6).ErrorS(err, "failed to stat resource file", "event", event)

				continue
			}
			if entry.IsDir() && len(filepath.SplitList(strings.TrimPrefix(event.Name, w.prefix))) == 1 {
				// the dir is namespace dir
				switch event.Op {
				case fsnotify.Create:
					if err := w.watcher.Add(event.Name); err != nil {
						klog.V(6).ErrorS(err, "failed to add namespace dir to file watcher", "event", event)
					}
				case fsnotify.Remove:
					if err := w.watcher.Remove(event.Name); err != nil {
						klog.V(6).ErrorS(err, "failed to remove namespace dir to file watcher", "event", event)
					}
				default:
					// do nothing
				}

				continue
			}

			if err := w.watchFile(event); err != nil {
				klog.V(6).ErrorS(err, "watch resource file error")
			}

		case err := <-w.watcher.Errors:
			klog.V(6).ErrorS(err, "file watcher error")
			fmt.Println(err, "file watcher error")

			return
		}
	}
}

// watchFile for resource.
func (w *fileWatcher) watchFile(event fsnotify.Event) error {
	if !strings.HasSuffix(event.Name, yamlSuffix) && !strings.HasSuffix(event.Name, yamlSuffix+deleteTagSuffix) {
		return nil
	}
	data, err := os.ReadFile(event.Name)
	if err != nil {
		return errors.Wrapf(err, "failed to read resource file %q", event.Name)
	}
	obj, _, err := w.codec.Decode(data, nil, w.newFunc())
	if err != nil {
		return errors.Wrapf(err, "failed to decode resource file %q", event.Name)
	}
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return errors.Wrapf(err, "failed to dconvert to meta object %q", event.Name)
	}
	if metaObj.GetName() == "" && metaObj.GetGenerateName() == "" { // ignore unknown file
		klog.V(6).InfoS("name is empty. ignore", "event", event)

		return nil
	}

	switch event.Op {
	case fsnotify.Create:
		if strings.HasSuffix(filepath.Base(event.Name), deleteTagSuffix) {
			// delete event
			w.watchEvents <- watch.Event{
				Type:   watch.Deleted,
				Object: obj,
			}
			if err := os.Remove(event.Name); err != nil {
				klog.ErrorS(err, "failed to remove file", "event", event)
			}
		} else {
			w.watchEvents <- watch.Event{
				Type:   watch.Added,
				Object: obj,
			}
		}
	case fsnotify.Write:
		// update event
		w.watchEvents <- watch.Event{
			Type:   watch.Modified,
			Object: obj,
		}
	}

	return nil
}
