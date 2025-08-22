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
	"sync"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/versioning"
	apigeneric "k8s.io/apiserver/pkg/registry/generic"
	apistorage "k8s.io/apiserver/pkg/storage"
	cacherstorage "k8s.io/apiserver/pkg/storage/cacher"
	"k8s.io/apiserver/pkg/storage/storagebackend"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
	cgtoolscache "k8s.io/client-go/tools/cache"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// NewFileRESTOptionsGetter return fileRESTOptionsGetter
func NewFileRESTOptionsGetter(runtimedir string, gv schema.GroupVersion) apigeneric.RESTOptionsGetter {
	return &fileRESTOptionsGetter{
		runtimedir: runtimedir,
		gv:         gv,
		storageConfig: &storagebackend.Config{
			Type:            "",
			Prefix:          "/",
			Transport:       storagebackend.TransportConfig{},
			Codec:           newYamlCodec(gv),
			EncodeVersioner: runtime.NewMultiGroupVersioner(gv),
		},
	}
}

func newYamlCodec(gv schema.GroupVersion) runtime.Codec {
	yamlSerializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, _const.Scheme, _const.Scheme, json.SerializerOptions{Yaml: true})

	return versioning.NewDefaultingCodecForScheme(
		_const.Scheme,
		yamlSerializer,
		yamlSerializer,
		gv,
		gv,
	)
}

// fileRESTOptionsGetter local rest info
type fileRESTOptionsGetter struct {
	runtimedir    string
	gv            schema.GroupVersion
	storageConfig *storagebackend.Config
}

// GetRESTOptions implements generic.RESTOptionsGetter.
func (f *fileRESTOptionsGetter) GetRESTOptions(resource schema.GroupResource, example runtime.Object) (apigeneric.RESTOptions, error) {
	prefix := filepath.Join(f.gv.Group, f.gv.Version, resource.Resource)

	return apigeneric.RESTOptions{
		StorageConfig: f.storageConfig.ForResource(resource),
		Decorator: func(storageConfig *storagebackend.ConfigForResource, resourcePrefix string,
			keyFunc func(obj runtime.Object) (string, error),
			newFunc func() runtime.Object,
			newListFunc func() runtime.Object,
			getAttrsFunc apistorage.AttrFunc,
			triggerFuncs apistorage.IndexerFuncs,
			indexers *cgtoolscache.Indexers) (apistorage.Interface, factory.DestroyFunc, error) {

			s, d := newFileStorage(prefix, resource, storageConfig.Codec, newFunc)

			resourcePrefix = filepath.Join(f.runtimedir, resourcePrefix)

			cacherConfig := cacherstorage.Config{
				Storage:        s,
				Versioner:      apistorage.APIObjectVersioner{},
				GroupResource:  storageConfig.GroupResource,
				ResourcePrefix: resourcePrefix,
				KeyFunc:        keyFunc,
				NewFunc:        newFunc,
				NewListFunc:    newListFunc,
				GetAttrsFunc:   getAttrsFunc,
				IndexerFuncs:   triggerFuncs,
				Indexers:       indexers,
				Codec:          storageConfig.Codec,
			}
			cacher, err := cacherstorage.NewCacherFromConfig(cacherConfig)
			if err != nil {
				return nil, func() {}, errors.Wrap(err, "failed to new cache")
			}
			var once sync.Once
			destroyFunc := func() {
				once.Do(func() {
					cacher.Stop()
					d()
				})
			}

			return cacher, destroyFunc, nil
		},
		EnableGarbageCollection:   false,
		DeleteCollectionWorkers:   0,
		ResourcePrefix:            prefix,
		CountMetricPollPeriod:     0,
		StorageObjectCountTracker: nil,
	}, nil
}
