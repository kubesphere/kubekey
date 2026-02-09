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
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/fsnotify/fsnotify"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	apistorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/klog/v2"
)

const (
	// incomingBufSize is the buffer size for incoming events to reduce context switches
	incomingBufSize = 100
	// outgoingBufSize is the buffer size for outgoing watch events
	outgoingBufSize = 100
)

// fileWatcher implements watch.Interface for local file system watching.
// It follows the pattern from k8s.io/apiserver/pkg/storage/etcd3/watcher.
type fileWatcher struct {
	// prefix is the root directory prefix for watching
	prefix string
	// clusterScoped indicates if the resource is cluster-scoped
	clusterScoped bool
	// codec is used to encode/decode objects
	codec runtime.Codec
	// newFunc creates new runtime objects
	newFunc func() runtime.Object
	// sendInitialEventsEnabled indicates whether to send initial events with bookmark
	sendInitialEventsEnabled bool
	// watcher is the underlying fsnotify watcher
	watcher *fsnotify.Watcher
	// ctx is the context for this watcher
	ctx context.Context
	// cancel is the cancel function for the context
	cancel context.CancelFunc
	// incomingEventChan receives events from the file system watcher
	incomingEventChan chan *fileEvent
	// resultChan is the channel for sending watch events to consumers
	resultChan chan watch.Event
	// stopCh is used to signal the watcher to stop
	stopCh chan struct{}
	// stopped indicates if the watcher has been stopped
	stopped bool
	// stopMux protects the stopped flag
	stopMux sync.Mutex
}

// fileEvent represents an internal event from the file system watcher
type fileEvent struct {
	event fsnotify.Event
	obj   runtime.Object
}

// newFileWatcher creates a new file watcher for the given path.
// This follows the k8s.io/apiserver/pkg/storage/etcd3 watcher pattern.
func newFileWatcher(prefix, path string, codec runtime.Codec, newFunc func() runtime.Object, opts apistorage.ListOptions, isClusterScoped bool) (watch.Interface, error) {
	// Validate and create the watch directory if it doesn't exist
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "failed to stat path %q", path)
		}
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return nil, errors.Wrapf(err, "failed to create dir %q", path)
		}
	}

	// Create the fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create file watcher for %q", path)
	}

	// Add the main path to watch
	if err := watcher.Add(path); err != nil {
		return nil, errors.Wrapf(err, "failed to add path to file watcher %q", path)
	}

	// Add namespace directories to watcher if watching at prefix level and resource is namespace-scoped
	if prefix == path && !isClusterScoped {
		entries, err := os.ReadDir(prefix)
		if err != nil {
			watcher.Close()
			return nil, errors.Wrapf(err, "failed to read dir %q", path)
		}
		for _, e := range entries {
			if e.IsDir() {
				nsPath := filepath.Join(prefix, e.Name())
				if err := watcher.Add(nsPath); err != nil {
					watcher.Close()
					return nil, errors.Wrapf(err, "failed to add namespace dir to file watcher %q", e.Name())
				}
			}
		}
	}

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	w := &fileWatcher{
		prefix:                   prefix,
		clusterScoped:            isClusterScoped,
		codec:                    codec,
		newFunc:                  newFunc,
		sendInitialEventsEnabled: opts.SendInitialEvents != nil && *opts.SendInitialEvents,
		watcher:                  watcher,
		ctx:                      ctx,
		cancel:                   cancel,
		incomingEventChan:        make(chan *fileEvent, incomingBufSize),
		resultChan:               make(chan watch.Event, outgoingBufSize),
		stopCh:                   make(chan struct{}),
	}
	// Start the watch loop first to ensure processEvents is ready
	go w.run()

	// Send initial ADDED events for all existing objects
	// This is required for informer cache to sync properly
	if err := w.sendInitialEvents(path); err != nil {
		watcher.Close()
		return nil, errors.Wrapf(err, "failed to send initial events for %q", path)
	}

	return w, nil
}

// sendInitialEvents sends ADDED events for all existing objects in the directory.
// This is required for informer cache to initialize properly.
//
// Directory structure:
//   - Namespace-scoped resources: rootPath/<namespace>/<resource>.yaml
//   - Cluster-scoped resources:   rootPath/<resource>.yaml
func (w *fileWatcher) sendInitialEvents(rootPath string) error {
	var initialObjects []runtime.Object

	// First, collect all existing objects
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read root dir %q", rootPath)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(rootPath, entry.Name())

		if w.clusterScoped {
			// Cluster-scoped resources: files directly under rootPath
			if entry.IsDir() {
				continue
			}
			path := entryPath
			if !w.isRelevantFile(path) {
				continue
			}

			obj, err := w.readFileObject(path)
			if err != nil {
				klog.V(6).ErrorS(err, "failed to read object for initial event", "path", path)
				continue
			}
			if obj == nil {
				continue
			}

			initialObjects = append(initialObjects, obj)
		} else {
			// Namespace-scoped resources: directories are namespace directories
			if !entry.IsDir() {
				continue
			}

			nsEntries, err := os.ReadDir(entryPath)
			if err != nil {
				klog.V(6).ErrorS(err, "failed to read namespace dir", "path", entryPath)
				continue
			}

			for _, nsEntry := range nsEntries {
				if nsEntry.IsDir() {
					continue
				}

				path := filepath.Join(entryPath, nsEntry.Name())
				if !w.isRelevantFile(path) {
					continue
				}

				obj, err := w.readFileObject(path)
				if err != nil {
					klog.V(6).ErrorS(err, "failed to read object for initial event", "path", path)
					continue
				}
				if obj == nil {
					continue
				}

				initialObjects = append(initialObjects, obj)
			}
		}
	}

	// Send ADDED events for all existing objects
	for _, obj := range initialObjects {
		event := &fileEvent{
			event: fsnotify.Event{
				Name: "", // Synthetic event, no file path
				Op:   fsnotify.Create,
			},
			obj: obj,
		}
		w.queueEvent(event)
	}
	klog.V(4).InfoS("finished queueing initial events", "count", len(initialObjects))

	// Send bookmark to signal initial sync is complete (only if SendInitialEvents was requested)
	if w.sendInitialEventsEnabled {
		bookmark := &watch.Event{
			Type:   watch.Bookmark,
			Object: w.newFunc(),
		}
		if err := apistorage.AnnotateInitialEventsEndBookmark(bookmark.Object); err != nil {
			return errors.Wrap(err, "failed to annotate initial events end bookmark")
		}
		w.sendEvent(bookmark)
	}

	return nil
}

// run is the main event loop that processes file system events.
// It follows the etcd3 watcher pattern with proper goroutine lifecycle management.
func (w *fileWatcher) run() {
	var wg sync.WaitGroup

	// Start watching for file system events
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.watchEvents()
	}()

	// Process events and send to resultChan
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.processEvents()
	}()

	// Wait for either stop signal or context cancellation
	select {
	case <-w.stopCh:
		// User called Stop(), cancel the context
		w.cancel()
	case <-w.ctx.Done():
		// Context was cancelled (e.g., Stop() was called elsewhere)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the result channel to signal that no more events will be sent
	w.stopMux.Lock()
	if !w.stopped {
		w.stopped = true
		close(w.resultChan)
	}
	w.stopMux.Unlock()
}

// watchEvents receives events from the file system watcher and queues them for processing.
func (w *fileWatcher) watchEvents() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case event := <-w.watcher.Events:
			w.handleFileEvent(event)
		case err := <-w.watcher.Errors:
			// Log error but continue watching (similar to etcd3 watcher handling transient errors)
			klog.V(6).ErrorS(err, "file watcher error")
		}
	}
}

// handleFileEvent processes a single file system event and queues it for transformation.
func (w *fileWatcher) handleFileEvent(event fsnotify.Event) {
	klog.V(6).InfoS("received watcher event", "event", event)

	// Skip events for non-resource files
	if !w.isRelevantFile(event.Name) {
		return
	}

	// Handle directory events (namespace directories)
	if w.isNamespaceDir(event.Name) {
		w.handleNamespaceEvent(event)
		return
	}

	// Process the resource file
	obj, err := w.readFileObject(event.Name)
	if err != nil {
		if !os.IsNotExist(err) {
			klog.V(6).ErrorS(err, "failed to read file for watch event", "event", event)
		}
		// If file doesn't exist, it was deleted
		if os.IsNotExist(err) {
			w.queueEvent(&fileEvent{event: event, obj: nil})
		}
		return
	}

	w.queueEvent(&fileEvent{event: event, obj: obj})
}

// readFileObject reads and decodes a resource file.
func (w *fileWatcher) readFileObject(path string) (runtime.Object, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	obj, _, err := w.codec.Decode(data, nil, w.newFunc())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode resource file %q", path)
	}

	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert to meta object %q", path)
	}

	// Skip objects without a name
	if metaObj.GetName() == "" && metaObj.GetGenerateName() == "" {
		return nil, nil
	}

	return obj, nil
}

// queueEvent adds an event to the incoming events channel.
func (w *fileWatcher) queueEvent(e *fileEvent) {
	select {
	case w.incomingEventChan <- e:
	case <-w.ctx.Done():
	}
}

// processEvents transforms internal events and sends them to the result channel.
func (w *fileWatcher) processEvents() {
	klog.V(4).InfoS("processEvents goroutine started", "prefix", w.prefix)
	defer klog.V(4).InfoS("processEvents goroutine stopped", "prefix", w.prefix)

	for {
		select {
		case e := <-w.incomingEventChan:
			res := w.transform(e)
			if res == nil {
				continue
			}
			klog.V(4).InfoS("sending event to resultChan", "type", res.Type, "prefix", w.prefix)
			if !w.sendEvent(res) {
				klog.V(4).InfoS("sendEvent returned false, stopping", "prefix", w.prefix)
				return
			}

			// After successfully sending deletion event, delete the file
			if res.Type == watch.Deleted && strings.HasSuffix(filepath.Base(e.event.Name), deleteTagSuffix) {
				go func() {
					if err := os.Remove(e.event.Name); err != nil && !os.IsNotExist(err) {
						klog.V(6).ErrorS(err, "failed to remove deletion marker file after event sent", "event", e.event)
					}
				}()
			}
		case <-w.ctx.Done():
			klog.V(4).InfoS("context cancelled, stopping processEvents", "prefix", w.prefix)
			return
		}
	}
}

// transform converts an internal fileEvent to a watch.Event.
func (w *fileWatcher) transform(e *fileEvent) *watch.Event {
	event := e.event
	var eventType watch.EventType

	switch event.Op {
	case fsnotify.Create:
		if strings.HasSuffix(filepath.Base(event.Name), deleteTagSuffix) {
			eventType = watch.Deleted
		} else {
			eventType = watch.Added
		}
	case fsnotify.Write, fsnotify.Chmod:
		eventType = watch.Modified
	case fsnotify.Rename:
		// For rename events, check if the file still exists
		if _, err := os.Stat(event.Name); err == nil {
			// File was renamed to something else, treat as modified
			eventType = watch.Modified
		} else {
			// File was renamed away (likely to a temp location), treat as deleted
			eventType = watch.Deleted
		}
	default:
		return nil
	}

	obj := e.obj
	// For deletion events without an object, try to read the old object
	if eventType == watch.Deleted && obj == nil {
		var err error
		obj, err = w.readFileObject(event.Name)
		if err != nil {
			klog.V(6).ErrorS(err, "failed to read deleted object", "event", event)
			// Create a minimal object for the deletion event
			obj = w.newFunc()
		}
	}

	result := &watch.Event{
		Type:   eventType,
		Object: obj,
	}

	return result
}

// sendEvent sends a watch event to the result channel.
// Returns true if the event was sent successfully, false if the context was cancelled.
func (w *fileWatcher) sendEvent(event *watch.Event) bool {
	// Check if the result channel is full and log a warning
	if len(w.resultChan) == cap(w.resultChan) {
		klog.V(3).InfoS("Fast watcher, slow processing. Probably caused by slow dispatching events to watchers",
			"outgoingEvents", outgoingBufSize, "prefix", w.prefix)
	}

	klog.V(4).InfoS("attempting to send event", "type", event.Type, "prefix", w.prefix, "channelLen", len(w.resultChan), "channelCap", cap(w.resultChan))

	select {
	case w.resultChan <- *event:
		klog.V(4).InfoS("event sent successfully", "type", event.Type, "prefix", w.prefix)
		return true
	case <-w.ctx.Done():
		klog.V(4).InfoS("context cancelled, failed to send event", "type", event.Type, "prefix", w.prefix)
		return false
	}
}

// Stop stops the watcher.
// This follows the etcd3 watcher pattern where Stop() is idempotent.
func (w *fileWatcher) Stop() {
	w.stopMux.Lock()
	if w.stopped {
		w.stopMux.Unlock()
		return
	}
	w.stopped = true
	w.stopMux.Unlock()

	// Close the stop channel to trigger graceful shutdown
	close(w.stopCh)

	// Close the fsnotify watcher
	if err := w.watcher.Close(); err != nil {
		klog.ErrorS(err, "failed to close file watcher")
	}
}

// ResultChan returns the channel for receiving watch events.
// The returned channel will be closed when Stop() is called.
func (w *fileWatcher) ResultChan() <-chan watch.Event {
	return w.resultChan
}

// isRelevantFile checks if the file is a relevant resource file.
func (w *fileWatcher) isRelevantFile(name string) bool {
	return strings.HasSuffix(name, dataFileSuffix) ||
		strings.HasSuffix(name, dataFileSuffix+deleteTagSuffix)
}

// isNamespaceDir checks if the path is a namespace directory.
func (w *fileWatcher) isNamespaceDir(name string) bool {
	entry, err := os.Stat(name)
	if err != nil {
		return false
	}
	if !entry.IsDir() {
		return false
	}
	// Check if it's exactly one level below the prefix
	relPath := strings.TrimPrefix(name, w.prefix)
	parts := filepath.SplitList(relPath)
	return len(parts) == 1 && parts[0] != ""
}

// handleNamespaceEvent handles events for namespace directories.
func (w *fileWatcher) handleNamespaceEvent(event fsnotify.Event) {
	switch event.Op {
	case fsnotify.Create:
		if err := w.watcher.Add(event.Name); err != nil {
			klog.V(6).ErrorS(err, "failed to add namespace dir to file watcher", "event", event)
		}
	case fsnotify.Remove:
		if err := w.watcher.Remove(event.Name); err != nil {
			klog.V(6).ErrorS(err, "failed to remove namespace dir from file watcher", "event", event)
		}
	case fsnotify.Rename:
		// Handle rename by removing the old path and adding the new one if it exists
		if _, err := os.Stat(event.Name); err == nil {
			if err := w.watcher.Add(event.Name); err != nil {
				klog.V(6).ErrorS(err, "failed to add renamed dir to file watcher", "event", event)
			}
		}
	}
}
