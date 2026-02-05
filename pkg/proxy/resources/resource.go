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

package resources

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	apirest "k8s.io/apiserver/pkg/registry/rest"
)

// ResourceStorage defines the storage and metadata interface for a REST resource.
// Each resource type (Task, Inventory, Playbook) must implement this interface.
type ResourceStorage interface {
	// GVK returns the GroupVersionKind of the resource
	GVK() schema.GroupVersionKind
	// GVRs returns all GroupVersionResources of the resource (including subresources)
	// For example: tasks and tasks/status
	GVRs() []schema.GroupVersionResource
	// Storage returns the REST storage for the given GVR
	// Main resource and subresource may use different storage implementations
	Storage(gvr schema.GroupVersionResource) apirest.Storage
	// IsAlwaysLocal returns true if the resource should always use local storage
	// Task resources are always locally stored because they are running data
	IsAlwaysLocal() bool
}
