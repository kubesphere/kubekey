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

package v1

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestConfigValueDecodesRaw(t *testing.T) {
	cfg := &Config{Spec: runtime.RawExtension{Raw: []byte(`{"helm_version":"v3.18.5","cri":{"container_manager":"containerd"}}`)}}
	val := cfg.Value()
	if val["helm_version"] != "v3.18.5" {
		t.Fatalf("helm_version lost, got %v", val)
	}
	if err := unstructured.SetNestedField(val, "v1.34.3", "kube_version"); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := unstructured.NestedString(cfg.Value(), "cri", "container_manager"); got != "containerd" {
		t.Fatalf("cri.container_manager lost, got %q", got)
	}
	if cfg.Value()["kube_version"] != "v1.34.3" {
		t.Fatalf("kube_version not kept, got %v", cfg.Value())
	}
}

func TestConfigValueEmpty(t *testing.T) {
	if val := (&Config{}).Value(); val == nil || len(val) != 0 {
		t.Fatalf("expected empty non-nil map, got %v", val)
	}
}
