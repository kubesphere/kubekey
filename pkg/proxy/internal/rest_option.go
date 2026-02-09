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
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/versioning"
	apigeneric "k8s.io/apiserver/pkg/registry/generic"
	apistorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
	cgtoolscache "k8s.io/client-go/tools/cache"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// NewFileRESTOptionsGetter return fileRESTOptionsGetter
func NewFileRESTOptionsGetter(runtimedir string, gv schema.GroupVersion, isClusterScoped bool) apigeneric.RESTOptionsGetter {
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, _const.Scheme, _const.Scheme, json.SerializerOptions{Yaml: true})
	return &fileRESTOptionsGetter{
		isClusterScoped: isClusterScoped,
		runtimedir:      runtimedir,
		gv:              gv,
		storageConfig: &storagebackend.Config{
			Type:      "",
			Prefix:    "/",
			Transport: storagebackend.TransportConfig{},
			Codec: versioning.NewDefaultingCodecForScheme(
				_const.Scheme,
				serializer,
				serializer,
				gv,
				gv,
			),
			EncodeVersioner: runtime.NewMultiGroupVersioner(gv),
		},
	}
}

// fileRESTOptionsGetter local rest info
type fileRESTOptionsGetter struct {
	runtimedir      string
	gv              schema.GroupVersion
	storageConfig   *storagebackend.Config
	isClusterScoped bool
}

// GetRESTOptions implements generic.RESTOptionsGetter.
func (f *fileRESTOptionsGetter) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (apigeneric.RESTOptions, error) {
	resourcePrefix := filepath.Join("/", f.gv.Group, f.gv.Version, resource.Resource)

	return apigeneric.RESTOptions{
		StorageConfig: f.storageConfig.ForResource(resource),
		Decorator: func(storageConfig *storagebackend.ConfigForResource, resourcePrefix string,
			keyFunc func(obj runtime.Object) (string, error),
			newFunc func() runtime.Object,
			newListFunc func() runtime.Object,
			getAttrsFunc apistorage.AttrFunc,
			triggerFuncs apistorage.IndexerFuncs,
			indexers *cgtoolscache.Indexers) (apistorage.Interface, factory.DestroyFunc, error) {
			return newFileStorage(f.runtimedir, resourcePrefix, resource, storageConfig.Codec, newFunc, f.isClusterScoped)
		},
		EnableGarbageCollection:   false,
		DeleteCollectionWorkers:   0,
		ResourcePrefix:            resourcePrefix,
		CountMetricPollPeriod:     0,
		StorageObjectCountTracker: nil,
	}, nil
}
