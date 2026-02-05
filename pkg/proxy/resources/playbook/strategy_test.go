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

package playbook

import (
	"context"
	"testing"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPlaybookStrategy_NamespaceScoped(t *testing.T) {
	if !Strategy.NamespaceScoped() {
		t.Errorf("Playbook must be namespace scoped")
	}
}

func TestPlaybookStrategy_AllowCreateOnUpdate(t *testing.T) {
	if Strategy.AllowCreateOnUpdate() {
		t.Errorf("Playbook should not allow create on update")
	}
}

func TestPlaybookStrategy_AllowUnconditionalUpdate(t *testing.T) {
	if !Strategy.AllowUnconditionalUpdate() {
		t.Errorf("Playbook should allow unconditional update")
	}
}

func TestPlaybookStrategy_PrepareForCreate(t *testing.T) {
	playbook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-playbook",
			Namespace: "default",
		},
		Spec: kkcorev1.PlaybookSpec{
			Playbook: "site.yml",
		},
	}

	Strategy.PrepareForCreate(context.Background(), playbook)

	// No specific preparation expected for Playbook
	if playbook.Name != "test-playbook" {
		t.Errorf("Expected name to remain 'test-playbook', got %v", playbook.Name)
	}
}

func TestPlaybookStrategy_Validate(t *testing.T) {
	errs := Strategy.Validate(context.Background(), &kkcorev1.Playbook{})
	if len(errs) != 0 {
		t.Errorf("Expected no validation errors, got %v", errs)
	}
}

func TestPlaybookStrategy_ValidateUpdate_SpecImmutable(t *testing.T) {
	oldPlaybook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-playbook",
			Namespace:       "default",
			ResourceVersion: "1",
		},
		Spec: kkcorev1.PlaybookSpec{
			Playbook: "old-site.yml",
		},
	}

	newPlaybook := oldPlaybook.DeepCopy()
	newPlaybook.Spec.Playbook = "new-site.yml"

	errs := Strategy.ValidateUpdate(context.Background(), newPlaybook, oldPlaybook)
	if len(errs) == 0 {
		t.Errorf("Expected validation error for spec change, got none")
	}
}

func TestPlaybookStrategy_ValidateUpdate_StatusOnly(t *testing.T) {
	oldPlaybook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-playbook",
			Namespace:       "default",
			ResourceVersion: "1",
		},
		Spec: kkcorev1.PlaybookSpec{
			Playbook: "site.yml",
		},
		Status: kkcorev1.PlaybookStatus{
			Phase: kkcorev1.PlaybookPhasePending,
		},
	}

	newPlaybook := oldPlaybook.DeepCopy()
	newPlaybook.Status.Phase = kkcorev1.PlaybookPhaseRunning

	errs := Strategy.ValidateUpdate(context.Background(), newPlaybook, oldPlaybook)
	if len(errs) != 0 {
		t.Errorf("Expected no validation errors for status update, got %v", errs)
	}
}

func TestPlaybookStrategy_WarningsOnCreate(t *testing.T) {
	warnings := Strategy.WarningsOnCreate(context.Background(), &kkcorev1.Playbook{})
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings on create, got %v", warnings)
	}
}

func TestPlaybookStrategy_WarningsOnUpdate(t *testing.T) {
	warnings := Strategy.WarningsOnUpdate(context.Background(), &kkcorev1.Playbook{}, &kkcorev1.Playbook{})
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings on update, got %v", warnings)
	}
}

func TestPlaybookStrategy_GetResetFields(t *testing.T) {
	fields := Strategy.GetResetFields()
	if fields != nil {
		t.Errorf("Expected no reset fields, got %v", fields)
	}
}

func TestPlaybookPrepareForUpdate_DoesNotModifySpec(t *testing.T) {
	oldPlaybook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-playbook",
			Namespace: "default",
		},
		Spec: kkcorev1.PlaybookSpec{
			Playbook: "playbook-a.yml",
		},
	}

	newPlaybook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-playbook",
			Namespace: "default",
		},
		Spec: kkcorev1.PlaybookSpec{
			Playbook: "playbook-b.yml",
		},
	}

	Strategy.PrepareForUpdate(context.Background(), newPlaybook, oldPlaybook)

	// PrepareForUpdate does nothing for playbook
	// The validation happens in ValidateUpdate
}
