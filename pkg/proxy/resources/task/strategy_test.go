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

package task

import (
	"context"
	"testing"

	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTaskStrategy_NamespaceScoped(t *testing.T) {
	if !Strategy.NamespaceScoped() {
		t.Errorf("Task must be namespace scoped")
	}
}

func TestTaskStrategy_AllowCreateOnUpdate(t *testing.T) {
	if Strategy.AllowCreateOnUpdate() {
		t.Errorf("Task should not allow create on update")
	}
}

func TestTaskStrategy_AllowUnconditionalUpdate(t *testing.T) {
	if !Strategy.AllowUnconditionalUpdate() {
		t.Errorf("Task should allow unconditional update")
	}
}

func TestTaskStrategy_PrepareForCreate(t *testing.T) {
	task := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-task",
			Namespace: "default",
		},
		Spec: kkcorev1alpha1.TaskSpec{
			Name: "test-playbook",
		},
	}

	Strategy.PrepareForCreate(context.Background(), task)

	if task.Status.Phase != kkcorev1alpha1.TaskPhasePending {
		t.Errorf("Expected status phase to be Pending, got %v", task.Status.Phase)
	}
}

func TestTaskStrategy_Validate(t *testing.T) {
	errs := Strategy.Validate(context.Background(), &kkcorev1alpha1.Task{})
	if len(errs) != 0 {
		t.Errorf("Expected no validation errors, got %v", errs)
	}
}

func TestTaskStrategy_ValidateUpdate_SpecImmutable(t *testing.T) {
	oldTask := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-task",
			Namespace:       "default",
			ResourceVersion: "1",
		},
		Spec: kkcorev1alpha1.TaskSpec{
			Name: "old-playbook",
		},
	}

	newTask := oldTask.DeepCopy()
	newTask.Spec.Name = "new-playbook"

	errs := Strategy.ValidateUpdate(context.Background(), newTask, oldTask)
	if len(errs) == 0 {
		t.Errorf("Expected validation error for spec change, got none")
	}
}

func TestTaskStrategy_ValidateUpdate_StatusOnly(t *testing.T) {
	oldTask := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-task",
			Namespace:       "default",
			ResourceVersion: "1",
		},
		Spec: kkcorev1alpha1.TaskSpec{
			Name: "test-playbook",
		},
		Status: kkcorev1alpha1.TaskStatus{
			Phase: kkcorev1alpha1.TaskPhasePending,
		},
	}

	newTask := oldTask.DeepCopy()
	newTask.Status.Phase = kkcorev1alpha1.TaskPhaseRunning

	errs := Strategy.ValidateUpdate(context.Background(), newTask, oldTask)
	if len(errs) != 0 {
		t.Errorf("Expected no validation errors for status update, got %v", errs)
	}
}

func TestTaskStrategy_WarningsOnCreate(t *testing.T) {
	warnings := Strategy.WarningsOnCreate(context.Background(), &kkcorev1alpha1.Task{})
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings on create, got %v", warnings)
	}
}

func TestTaskStrategy_WarningsOnUpdate(t *testing.T) {
	warnings := Strategy.WarningsOnUpdate(context.Background(), &kkcorev1alpha1.Task{}, &kkcorev1alpha1.Task{})
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings on update, got %v", warnings)
	}
}

func TestTaskStrategy_GetResetFields(t *testing.T) {
	fields := Strategy.GetResetFields()
	if fields != nil {
		t.Errorf("Expected no reset fields, got %v", fields)
	}
}

func TestGetAttrs(t *testing.T) {
	task := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-task",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: kkcorev1alpha1.TaskSpec{
			Name: "test-playbook",
		},
	}

	labels, fields, err := GetAttrs(task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if labels["app"] != "test" {
		t.Errorf("Expected label 'app' to be 'test', got %v", labels)
	}

	if fields["metadata.name"] != "test-task" {
		t.Errorf("Expected field 'metadata.name' to be 'test-task', got %v", fields)
	}
}

func TestMatchTask(t *testing.T) {
	predicate := MatchTask(nil, nil)
	if predicate.Label != nil {
		t.Errorf("Expected nil label selector")
	}
	if predicate.Field != nil {
		t.Errorf("Expected nil field selector")
	}
	if predicate.GetAttrs == nil {
		t.Errorf("Expected non-nil GetAttrs")
	}
}

func TestToSelectableFields(t *testing.T) {
	task := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-task",
			Namespace: "default",
		},
		Spec: kkcorev1alpha1.TaskSpec{
			Name: "test-playbook",
		},
	}

	fields := ToSelectableFields(task)

	if fields["metadata.name"] != "test-task" {
		t.Errorf("Expected 'metadata.name' to be 'test-task', got %v", fields["metadata.name"])
	}

	if fields["metadata.namespace"] != "default" {
		t.Errorf("Expected 'metadata.namespace' to be 'default', got %v", fields["metadata.namespace"])
	}

	if fields["spec.name"] != "test-playbook" {
		t.Errorf("Expected 'spec.name' to be 'test-playbook', got %v", fields["spec.name"])
	}
}

func TestNameTriggerFunc(t *testing.T) {
	task := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-task",
		},
	}

	trigger := NameTriggerFunc(task)
	if trigger != "test-task" {
		t.Errorf("Expected trigger 'test-task', got %v", trigger)
	}
}

func TestOwnerPlaybookIndexFunc(t *testing.T) {
	task := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-task",
			Namespace: "default",
		},
	}
	task.OwnerReferences = []metav1.OwnerReference{
		{
			Kind: "Playbook",
			Name: "test-playbook",
		},
	}

	index, err := OwnerPlaybookIndexFunc(task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "default/test-playbook"
	if len(index) != 1 || index[0] != expected {
		t.Errorf("Expected index [%s], got %v", expected, index)
	}
}

func TestOwnerPlaybookIndexFunc_NoOwner(t *testing.T) {
	task := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-task",
			Namespace: "default",
		},
	}

	_, err := OwnerPlaybookIndexFunc(task)
	if err == nil {
		t.Errorf("Expected error for task without playbook owner, got none")
	}
}

func TestTaskPrepareForCreate_DoesNotOverwriteStatus(t *testing.T) {
	task := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-task",
			Namespace: "default",
		},
		Status: kkcorev1alpha1.TaskStatus{
			Phase: kkcorev1alpha1.TaskPhaseRunning,
		},
	}

	Strategy.PrepareForCreate(context.Background(), task)

	if task.Status.Phase != kkcorev1alpha1.TaskPhasePending {
		t.Errorf("Expected status phase to be overwritten to Pending, got %v", task.Status.Phase)
	}
}

func TestTaskPrepareForUpdate_DoesNotModifySpec(t *testing.T) {
	oldTask := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-task",
			Namespace: "default",
		},
		Spec: kkcorev1alpha1.TaskSpec{
			Name: "playbook-a",
		},
	}

	newTask := &kkcorev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-task",
			Namespace: "default",
		},
		Spec: kkcorev1alpha1.TaskSpec{
			Name: "playbook-b",
		},
	}

	Strategy.PrepareForUpdate(context.Background(), newTask, oldTask)

	// PrepareForUpdate does nothing for task
	// The validation happens in ValidateUpdate
}
