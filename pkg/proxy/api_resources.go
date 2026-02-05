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

package proxy

import (
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/endpoints/discovery"
	"k8s.io/apiserver/pkg/endpoints/handlers/negotiation"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/apiserver/pkg/features"
	apirest "k8s.io/apiserver/pkg/registry/rest"
	utilfeature "k8s.io/apiserver/pkg/util/feature"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// Default request timeout (30 minutes)
const defaultMinRequestTimeout = 1800 * time.Second

// apiResources represents a collection of resources in an API group.
// It manages resource registration information and route handlers.
type apiResources struct {
	gv                schema.GroupVersion // API group version
	prefix            string              // API path prefix (e.g., /apis/kubekey.kubesphere.io/v1alpha1)
	minRequestTimeout time.Duration       // Request timeout duration

	resourceOptions []resourceOptions            // Resource options list
	list            []metav1.APIResource         // API resource list (for discovery)
	typer           runtime.ObjectTyper          // Type resolver
	serializer      runtime.NegotiatedSerializer // Serializer
}

// resourceOptions defines registration options for a single resource
type resourceOptions struct {
	path         string              // Resource path (e.g., "tasks" or "tasks/status")
	resource     string              // Resource name (parsed from path)
	subresource  string              // Subresource name (parsed from path)
	resourcePath string              // Resource path template (e.g., /namespaces/{namespace}/tasks/{name})
	itemPath     string              // Resource item path template (e.g., /namespaces/{namespace}/tasks/{name})
	storage      apirest.Storage     // REST storage
	admit        admission.Interface // Admission control
}

// init initializes resource options and calculates path-related fields
func (o *resourceOptions) init() error {
	// Determine path prefix (namespace-scoped or cluster-scoped)
	var prefix string
	scoper, ok := o.storage.(apirest.Scoper)
	if !ok {
		return errors.Errorf("%q must implement scoper", o.path)
	}
	if scoper.NamespaceScoped() {
		prefix = "/namespaces/{namespace}/"
	} else {
		prefix = "/"
	}

	// Parse path (supports main resource and subresource)
	switch parts := strings.Split(o.path, "/"); len(parts) {
	case 2:
		// Subresource (e.g., tasks/status)
		o.resource, o.subresource = parts[0], parts[1]
		o.resourcePath = prefix + o.resource + "/{name}/" + o.subresource
		o.itemPath = prefix + o.resource + "/{name}/" + o.subresource
	case 1:
		// Main resource (e.g., tasks)
		o.resource = parts[0]
		o.resourcePath = prefix + o.resource
		o.itemPath = prefix + o.resource + "/{name}"
	default:
		return errors.New("api_installer allows only one or two segment paths (resource or resource/subresource)")
	}

	// Set default admission control
	if o.admit == nil {
		o.admit = newAlwaysAdmit()
	}

	return nil
}

// newAPIResources creates a new API resources collection
func newAPIResources(gv schema.GroupVersion) *apiResources {
	return &apiResources{
		gv:                gv,
		prefix:            "/apis/" + gv.String(),
		minRequestTimeout: defaultMinRequestTimeout,

		typer:      _const.Scheme,
		serializer: _const.CodecFactory,
	}
}

// AddResource adds a resource to the API resources collection
func (r *apiResources) AddResource(o resourceOptions) error {
	if err := o.init(); err != nil {
		return err
	}
	r.resourceOptions = append(r.resourceOptions, o)

	// Get storage version information (for StorageVersionHash)
	storageVersionProvider, isStorageVersionProvider := o.storage.(apirest.StorageVersionProvider)
	var apiResource metav1.APIResource
	if isStorageVersionProvider &&
		utilfeature.DefaultFeatureGate.Enabled(features.StorageVersionHash) &&
		storageVersionProvider.StorageVersion() != nil {
		versioner := storageVersionProvider.StorageVersion()
		gvk, err := getStorageVersionKind(versioner, o.storage, r.typer)
		if err != nil {
			return err
		}
		apiResource.Group = gvk.Group
		apiResource.Version = gvk.Version
		apiResource.Kind = gvk.Kind
		apiResource.StorageVersionHash = discovery.StorageVersionHash(gvk.Group, gvk.Version, gvk.Kind)
	}

	// Set basic resource information
	apiResource.Name = o.path
	apiResource.Namespaced = true
	apiResource.Verbs = []string{"*"} // Supports all REST operations

	// Get short names and categories
	if shortNamesProvider, ok := o.storage.(apirest.ShortNamesProvider); ok {
		apiResource.ShortNames = shortNamesProvider.ShortNames()
	}
	if categoriesProvider, ok := o.storage.(apirest.CategoriesProvider); ok {
		apiResource.Categories = categoriesProvider.Categories()
	}

	// Main resource must provide singular name
	if o.subresource == "" {
		singularNameProvider, ok := o.storage.(apirest.SingularNameProvider)
		if !ok {
			return errors.Errorf("resource %s must implement SingularNameProvider", o.path)
		}
		apiResource.SingularName = singularNameProvider.GetSingularName()
	}
	r.list = append(r.list, apiResource)

	return nil
}

// handlerAPIResources returns the Handler for /apis/{group}/{version} requests
func (r *apiResources) handlerAPIResources() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		responsewriters.WriteObjectNegotiated(r.serializer, negotiation.DefaultEndpointRestrictions, schema.GroupVersion{}, writer, request, http.StatusOK,
			&metav1.APIResourceList{GroupVersion: r.gv.String(), APIResources: r.list}, false)
	}
}

// getStorageVersionKind calculates the storage GVK
// Objects in storage are converted before being persisted to etcd
func getStorageVersionKind(storageVersioner runtime.GroupVersioner, storage apirest.Storage, typer runtime.ObjectTyper) (schema.GroupVersionKind, error) {
	object := storage.New()
	fqKinds, _, err := typer.ObjectKinds(object)
	if err != nil {
		return schema.GroupVersionKind{}, errors.Wrap(err, "failed to get object kind")
	}
	gvk, ok := storageVersioner.KindForGroupVersionKinds(fqKinds)
	if !ok {
		return schema.GroupVersionKind{}, errors.Errorf("failed to find the storage version kind for %v", reflect.TypeOf(object))
	}

	return gvk, nil
}
