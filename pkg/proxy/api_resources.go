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
	"fmt"
	"net/http"
	"reflect"
	"time"

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
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

const defaultMinRequestTimeout = 1800 * time.Second

type apiResources struct {
	gv                schema.GroupVersion
	prefix            string
	minRequestTimeout time.Duration

	resourceOptions []resourceOptions
	list            []metav1.APIResource
	typer           runtime.ObjectTyper
	serializer      runtime.NegotiatedSerializer
}

type resourceOptions struct {
	path    string
	storage apirest.Storage
	admit   admission.Interface
}

func newApiIResources(gv schema.GroupVersion) *apiResources {
	return &apiResources{
		gv:                gv,
		prefix:            "/apis/" + gv.String(),
		minRequestTimeout: defaultMinRequestTimeout,

		typer:      _const.Scheme,
		serializer: _const.Codecs,
	}
}

func (r *apiResources) AddResource(o resourceOptions) error {
	if o.admit == nil {
		// set default admit
		o.admit = newAlwaysAdmit()
	}
	r.resourceOptions = append(r.resourceOptions, o)
	storageVersionProvider, isStorageVersionProvider := o.storage.(apirest.StorageVersionProvider)
	var apiResource metav1.APIResource
	if utilfeature.DefaultFeatureGate.Enabled(features.StorageVersionHash) &&
		isStorageVersionProvider &&
		storageVersionProvider.StorageVersion() != nil {
		versioner := storageVersionProvider.StorageVersion()
		gvk, err := getStorageVersionKind(versioner, o.storage, r.typer)
		if err != nil {
			klog.V(4).ErrorS(err, "failed to get storage version kind", "storage", reflect.TypeOf(o.storage))
			return err
		}
		apiResource.Group = gvk.Group
		apiResource.Version = gvk.Version
		apiResource.Kind = gvk.Kind
		apiResource.StorageVersionHash = discovery.StorageVersionHash(gvk.Group, gvk.Version, gvk.Kind)
	}
	apiResource.Name = o.path
	apiResource.Namespaced = true
	apiResource.Verbs = []string{"*"}
	if shortNamesProvider, ok := o.storage.(apirest.ShortNamesProvider); ok {
		apiResource.ShortNames = shortNamesProvider.ShortNames()
	}
	if categoriesProvider, ok := o.storage.(apirest.CategoriesProvider); ok {
		apiResource.Categories = categoriesProvider.Categories()
	}
	_, subResource, err := splitSubresource(o.path)
	if err != nil {
		return err
	}
	if len(subResource) == 0 {
		singularNameProvider, ok := o.storage.(apirest.SingularNameProvider)
		if !ok {
			return fmt.Errorf("resource %s must implement SingularNameProvider", o.path)
		}
		apiResource.SingularName = singularNameProvider.GetSingularName()
	}
	r.list = append(r.list, apiResource)
	return nil
}

func (r *apiResources) handlerApiResources() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		responsewriters.WriteObjectNegotiated(r.serializer, negotiation.DefaultEndpointRestrictions, schema.GroupVersion{}, writer, request, http.StatusOK,
			&metav1.APIResourceList{GroupVersion: r.gv.String(), APIResources: r.list}, false)
	}
}

// calculate the storage gvk, the gvk objects are converted to before persisted to the etcd.
func getStorageVersionKind(storageVersioner runtime.GroupVersioner, storage apirest.Storage, typer runtime.ObjectTyper) (schema.GroupVersionKind, error) {
	object := storage.New()
	fqKinds, _, err := typer.ObjectKinds(object)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	gvk, ok := storageVersioner.KindForGroupVersionKinds(fqKinds)
	if !ok {
		return schema.GroupVersionKind{}, fmt.Errorf("cannot find the storage version kind for %v", reflect.TypeOf(object))
	}
	return gvk, nil
}
