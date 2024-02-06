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
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

type fileWatcher struct {
	resource    schema.GroupResource
	codec       runtime.Codec
	newFunc     func() runtime.Object
	watcher     *fsnotify.Watcher
	watchEvents chan watch.Event
}

func newFileWatcher(resource schema.GroupResource, codec runtime.Codec, path string) (watch.Interface, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return nil, err
			}
		} else {
			klog.V(4).ErrorS(err, "failed to stat path", "path", path)
			return nil, err
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.V(4).ErrorS(err, "failed to create file watcher", "path", path)
		return nil, err
	}
	if err := watcher.Add(path); err != nil {
		klog.V(4).ErrorS(err, "failed to add path to file watcher", "path", path)
		return nil, err
	}
	w := &fileWatcher{
		resource: resource,
		codec:    codec,
		watcher:  watcher,
	}

	go w.watch()
	return w, nil
}

func (w *fileWatcher) Stop() {
	if err := w.watcher.Close(); err != nil {
		klog.V(4).ErrorS(err, "failed to close file watcher")
	}
}

func (w *fileWatcher) ResultChan() <-chan watch.Event {
	return w.watchEvents
}

func (w *fileWatcher) watch() {
	// stop the watcher
	//defer f.Stop()
	for {
		select {
		case event := <-w.watcher.Events:
			// filter resource type
			klog.V(4).InfoS("receive watcher event", "event", event)
			relPath, err := filepath.Rel(_const.GetRuntimeDir(), event.Name)
			if err != nil {
				continue
			}
			// the second element is the resource name
			pl := filepath.SplitList(relPath)
			if len(pl) < 2 || pl[1] != w.resource.Resource {
				continue
			}
			data, err := os.ReadFile(event.Name)
			if err != nil {
				klog.V(4).ErrorS(err, "failed to read resource file", "event", event)
				continue
			}
			switch event.Op {
			case fsnotify.Create:
				obj, _, err := w.codec.Decode(data, nil, w.newFunc())
				if err != nil {
					klog.V(4).ErrorS(err, "failed to decode resource file", "event", event)
					continue
				}
				w.watchEvents <- watch.Event{
					Type:   watch.Added,
					Object: obj,
				}
			case fsnotify.Write:
				obj, _, err := w.codec.Decode(data, nil, w.newFunc())
				if err != nil {
					klog.V(4).ErrorS(err, "failed to decode resource file", "event", event)
					continue
				}
				if strings.HasSuffix(filepath.Base(event.Name), deleteTag) {
					// delete event
					w.watchEvents <- watch.Event{
						Type:   watch.Deleted,
						Object: obj,
					}
					if err := os.Remove(event.Name); err != nil {
						klog.ErrorS(err, "failed to remove file", "event", event)
					}
				} else {
					// update event
					w.watchEvents <- watch.Event{
						Type:   watch.Modified,
						Object: obj,
					}
				}
			}
		case err := <-w.watcher.Errors:
			klog.V(4).ErrorS(err, "file watcher error")
			return
		}
	}
}
