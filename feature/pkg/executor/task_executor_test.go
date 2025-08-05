package executor

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
)

func TestTaskExecutor(t *testing.T) {
	testcases := []struct {
		name  string
		hosts []string
		task  *kkcorev1alpha1.Task
	}{
		{
			name: "debug module in single host",
			task: &kkcorev1alpha1.Task{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
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
			name:  "debug module in single host with loop",
			hosts: []string{"node1"},
			task: &kkcorev1alpha1.Task{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: corev1.NamespaceDefault,
				},
				Spec: kkcorev1alpha1.TaskSpec{
					Hosts: []string{"node1"},
					Module: kkcorev1alpha1.Module{
						Name: "debug",
						Args: runtime.RawExtension{Raw: []byte(`{"msg":"hello"}`)},
					},
					Loop: runtime.RawExtension{
						Raw: []byte(string(`["a", "b"]`)),
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
					Name:      "test3",
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
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			o, err := newTestOption(tc.hosts)
			if err != nil {
				t.Fatal(err)
			}

			if err := (&taskExecutor{
				option:         o,
				task:           tc.task,
				taskRunTimeout: 10 * time.Second,
			}).Exec(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}
