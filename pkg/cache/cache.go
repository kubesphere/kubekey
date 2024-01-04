/*
Copyright 2023 The KubeSphere Authors.

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

package cache

import (
	"sync"
)

// Cache is the interface for cache.
type Cache interface {
	// Name of pool
	Name() string
	// Put the cached value for the given key.
	Put(key string, value any)
	// Get the cached value for the given key.
	Get(key string) (any, bool)
	// Release the cached value for the given id.
	Release(id string)
	// Clean all cached value
	Clean()
}

type local struct {
	name  string
	cache map[string]any

	sync.Mutex
}

func (p *local) Name() string {
	return p.name
}

func (p *local) Put(key string, value any) {
	p.Lock()
	defer p.Unlock()

	p.cache[key] = value
}

func (p *local) Get(key string) (any, bool) {
	v, ok := p.cache[key]
	if ok {
		return v, ok
	}
	return v, false
}

func (p *local) Release(id string) {
	p.Lock()
	defer p.Unlock()

	delete(p.cache, id)
}

func (p *local) Clean() {
	p.Lock()
	defer p.Unlock()
	for id := range p.cache {
		delete(p.cache, id)
	}
}

// NewLocalCache return a local cache
func NewLocalCache(name string) Cache {
	return &local{
		name:  name,
		cache: make(map[string]any),
	}
}

var (
	// LocalVariable is a local cache for variable.Variable
	LocalVariable = NewLocalCache("variable")
)
