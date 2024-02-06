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

package config

import (
	"k8s.io/apimachinery/pkg/runtime"
	apigeneric "k8s.io/apiserver/pkg/registry/generic"
	apiregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	apirest "k8s.io/apiserver/pkg/registry/rest"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
)

type ConfigStorage struct {
	Config *REST
}

type REST struct {
	*apiregistry.Store
}

func NewStorage(optsGetter apigeneric.RESTOptionsGetter) (ConfigStorage, error) {
	store := &apiregistry.Store{
		NewFunc:                   func() runtime.Object { return &kubekeyv1.Config{} },
		NewListFunc:               func() runtime.Object { return &kubekeyv1.ConfigList{} },
		DefaultQualifiedResource:  kubekeyv1.SchemeGroupVersion.WithResource("configs").GroupResource(),
		SingularQualifiedResource: kubekeyv1.SchemeGroupVersion.WithResource("config").GroupResource(),

		CreateStrategy:      Strategy,
		UpdateStrategy:      Strategy,
		DeleteStrategy:      Strategy,
		ReturnDeletedObject: true,

		TableConvertor: apirest.NewDefaultTableConvertor(kubekeyv1.SchemeGroupVersion.WithResource("configs").GroupResource()),
	}
	options := &apigeneric.StoreOptions{
		RESTOptions: optsGetter,
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return ConfigStorage{}, err
	}

	return ConfigStorage{
		Config: &REST{store},
	}, nil
}
