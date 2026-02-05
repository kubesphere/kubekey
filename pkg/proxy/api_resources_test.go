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
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// mockStorage is a mock implementation of apirest.Storage for testing
type mockStorage struct {
	clusterScoped bool
	newFunc       func() runtime.Object
}

func (m *mockStorage) New() runtime.Object {
	if m.newFunc != nil {
		return m.newFunc()
	}
	return nil
}

func (m *mockStorage) Destroy() {}

// NamespaceScoped implements apirest.Scoper
func (m *mockStorage) NamespaceScoped() bool {
	return !m.clusterScoped
}

// GetSingularName implements apirest.SingularNameProvider
func (m *mockStorage) GetSingularName() string {
	return "task"
}

func TestResourceOptions_Init_NamespaceScoped(t *testing.T) {
	testCases := []struct {
		name                 string
		path                 string
		clusterScoped        bool
		expectedResource     string
		expectedResourcePath string
		expectedItemPath     string
		description          string
	}{
		{
			name:                 "namespaced resource",
			path:                 "tasks",
			clusterScoped:        false,
			expectedResource:     "tasks",
			expectedResourcePath: "/namespaces/{namespace}/tasks",
			expectedItemPath:     "/namespaces/{namespace}/tasks/{name}",
			description:          "should set correct paths for namespace-scoped resource",
		},
		{
			name:                 "cluster-scoped resource",
			path:                 "clustertasks",
			clusterScoped:        true,
			expectedResource:     "clustertasks",
			expectedResourcePath: "/clustertasks",
			expectedItemPath:     "/clustertasks/{name}",
			description:          "should set correct paths for cluster-scoped resource",
		},
		{
			name:                 "namespaced subresource",
			path:                 "tasks/status",
			clusterScoped:        false,
			expectedResource:     "tasks",
			expectedResourcePath: "/namespaces/{namespace}/tasks/{name}/status",
			expectedItemPath:     "/namespaces/{namespace}/tasks/{name}/status",
			description:          "should set correct paths for namespace-scoped subresource",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := resourceOptions{
				path:    tc.path,
				storage: &mockStorage{clusterScoped: tc.clusterScoped},
			}

			err := opts.init()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if opts.resource != tc.expectedResource {
				t.Errorf("expected resource=%q, got %q", tc.expectedResource, opts.resource)
			}

			if opts.resourcePath != tc.expectedResourcePath {
				t.Errorf("expected resourcePath=%q, got %q", tc.expectedResourcePath, opts.resourcePath)
			}

			if opts.itemPath != tc.expectedItemPath {
				t.Errorf("expected itemPath=%q, got %q", tc.expectedItemPath, opts.itemPath)
			}
		})
	}
}

func TestResourceOptions_Init_InvalidPath(t *testing.T) {
	testCases := []struct {
		name          string
		path          string
		clusterScoped bool
		expectError   bool
		description   string
	}{
		{
			name:          "too many segments",
			path:          "a/b/c/d",
			clusterScoped: false,
			expectError:   true,
			description:   "should return error for path with too many segments",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := resourceOptions{
				path:    tc.path,
				storage: &mockStorage{clusterScoped: tc.clusterScoped},
			}

			err := opts.init()
			if tc.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestResourceOptions_Init_DefaultAdmit(t *testing.T) {
	testCases := []struct {
		name        string
		description string
	}{
		{
			name:        "should set default admission handler",
			description: "when admit is nil, should set newAlwaysAdmit as default",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := resourceOptions{
				path:    "tasks",
				storage: &mockStorage{clusterScoped: false},
				admit:   nil,
			}

			err := opts.init()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if opts.admit == nil {
				t.Error("expected admit to be set")
			}
		})
	}
}

func TestNewAPIResources(t *testing.T) {
	testCases := []struct {
		name           string
		gv             schema.GroupVersion
		expectedPrefix string
		description    string
	}{
		{
			name:           "v1alpha1 group version",
			gv:             schema.GroupVersion{Group: "kubekey.kubesphere.io", Version: "v1alpha1"},
			expectedPrefix: "/apis/kubekey.kubesphere.io/v1alpha1",
			description:    "should create apiResources with correct prefix",
		},
		{
			name:           "core v1 group version",
			gv:             schema.GroupVersion{Group: "", Version: "v1"},
			expectedPrefix: "/apis/v1",
			description:    "should handle core group version",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiRes := newAPIResources(tc.gv)

			if apiRes.gv != tc.gv {
				t.Errorf("expected gv=%v, got %v", tc.gv, apiRes.gv)
			}

			if apiRes.prefix != tc.expectedPrefix {
				t.Errorf("expected prefix=%q, got %q", tc.expectedPrefix, apiRes.prefix)
			}

			if apiRes.minRequestTimeout == 0 {
				t.Error("expected minRequestTimeout to be set")
			}

			if apiRes.typer == nil {
				t.Error("expected typer to be set")
			}

			if apiRes.serializer == nil {
				t.Error("expected serializer to be set")
			}
		})
	}
}

func TestAPIResources_AddResource(t *testing.T) {
	testCases := []struct {
		name          string
		path          string
		clusterScoped bool
		expectError   bool
		description   string
	}{
		{
			name:          "valid namespace-scoped resource",
			path:          "tasks",
			clusterScoped: false,
			expectError:   false,
			description:   "should add resource without error",
		},
		{
			name:          "valid cluster-scoped resource",
			path:          "clustertasks",
			clusterScoped: true,
			expectError:   false,
			description:   "should add cluster-scoped resource without error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiRes := newAPIResources(schema.GroupVersion{Group: "test.io", Version: "v1"})

			opts := resourceOptions{
				path:    tc.path,
				storage: &mockStorage{clusterScoped: tc.clusterScoped},
			}

			err := apiRes.AddResource(opts)
			if tc.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
