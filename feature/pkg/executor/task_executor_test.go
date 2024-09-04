package executor

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	kkcorev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1alpha1"
)

func TestTaskExecutor(t *testing.T) {
	testcases := []struct {
		name string
		task *kkcorev1alpha1.Task
	}{
		{
			name: "debug module in single host",
			task: &kkcorev1alpha1.Task{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: corev1.NamespaceDefault,
				},
				Spec: kkcorev1alpha1.TaskSpec{
					Hosts: []string{"node1"},
					Module: kkcorev1alpha1.Module{
						Name: "debug",
						Args: runtime.RawExtension{Raw: []byte(`{"msg":"hello"}`)},
					},
				},
				Status: kkcorev1alpha1.TaskStatus{},
			},
		},
		{
			name: "debug module in multiple hosts",
			task: &kkcorev1alpha1.Task{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: corev1.NamespaceDefault,
				},
				Spec: kkcorev1alpha1.TaskSpec{
					Hosts: []string{"node1", "n2"},
					Module: kkcorev1alpha1.Module{
						Name: "debug",
						Args: runtime.RawExtension{Raw: []byte(`{"msg":"hello"}`)},
					},
				},
				Status: kkcorev1alpha1.TaskStatus{},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			o, err := newTestOption()
			if err != nil {
				t.Fatal(err)
			}

			if err := (&taskExecutor{
				option: o,
				task:   tc.task,
			}).Exec(context.TODO()); err != nil {
				t.Fatal(err)
			}
		})
	}
}
