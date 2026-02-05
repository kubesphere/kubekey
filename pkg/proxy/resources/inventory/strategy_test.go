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

package inventory

import (
	"context"
	"testing"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestInventoryStrategy_NamespaceScoped(t *testing.T) {
	if !Strategy.NamespaceScoped() {
		t.Errorf("Inventory must be namespace scoped")
	}
}

func TestInventoryStrategy_AllowCreateOnUpdate(t *testing.T) {
	if Strategy.AllowCreateOnUpdate() {
		t.Errorf("Inventory should not allow create on update")
	}
}

func TestInventoryStrategy_AllowUnconditionalUpdate(t *testing.T) {
	if !Strategy.AllowUnconditionalUpdate() {
		t.Errorf("Inventory should allow unconditional update")
	}
}

func TestInventoryStrategy_PrepareForCreate(t *testing.T) {
	inventory := &kkcorev1.Inventory{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-inventory",
			Namespace: "default",
		},
		Spec: kkcorev1.InventorySpec{
			Hosts: kkcorev1.InventoryHost{
				"node1": runtime.RawExtension{Raw: []byte(`{"address": "192.168.1.1"}`)},
			},
		},
	}

	Strategy.PrepareForCreate(context.Background(), inventory)

	// No specific preparation expected for Inventory
	if inventory.Name != "test-inventory" {
		t.Errorf("Expected name to remain 'test-inventory', got %v", inventory.Name)
	}
}

func TestInventoryStrategy_Validate(t *testing.T) {
	errs := Strategy.Validate(context.Background(), &kkcorev1.Inventory{})
	if len(errs) != 0 {
		t.Errorf("Expected no validation errors, got %v", errs)
	}
}

func TestInventoryStrategy_ValidateUpdate(t *testing.T) {
	oldInventory := &kkcorev1.Inventory{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-inventory",
			Namespace:       "default",
			ResourceVersion: "1",
		},
		Spec: kkcorev1.InventorySpec{
			Hosts: kkcorev1.InventoryHost{
				"node1": runtime.RawExtension{Raw: []byte(`{"address": "192.168.1.1"}`)},
			},
		},
	}

	newInventory := oldInventory.DeepCopy()
	newInventory.Spec.Hosts["node2"] = runtime.RawExtension{Raw: []byte(`{"address": "192.168.1.2"}`)}

	errs := Strategy.ValidateUpdate(context.Background(), newInventory, oldInventory)
	if len(errs) != 0 {
		t.Errorf("Expected no validation errors for update, got %v", errs)
	}
}

func TestInventoryStrategy_WarningsOnCreate(t *testing.T) {
	warnings := Strategy.WarningsOnCreate(context.Background(), &kkcorev1.Inventory{})
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings on create, got %v", warnings)
	}
}

func TestInventoryStrategy_WarningsOnUpdate(t *testing.T) {
	warnings := Strategy.WarningsOnUpdate(context.Background(), &kkcorev1.Inventory{}, &kkcorev1.Inventory{})
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings on update, got %v", warnings)
	}
}

func TestInventoryStrategy_GetResetFields(t *testing.T) {
	fields := Strategy.GetResetFields()
	if fields != nil {
		t.Errorf("Expected no reset fields, got %v", fields)
	}
}
