package executor

import (
	"context"
	"encoding/json"
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

func TestNormalizeJSONNumbers(t *testing.T) {
	input := map[string]any{
		"int":    json.Number("24"),
		"float":  json.Number("3.14"),
		"nested": map[string]any{"value": json.Number("9007199254740993")},
		"list":   []any{json.Number("1"), json.Number("2.5")},
		"str":    "foo",
	}
	out := normalizeJSONNumbers(input).(map[string]any)

	if _, ok := out["int"].(int64); !ok {
		t.Fatalf("expected int to be int64, got %T", out["int"])
	}
	if _, ok := out["float"].(float64); !ok {
		t.Fatalf("expected float to be float64, got %T", out["float"])
	}
	if _, ok := out["nested"].(map[string]any)["value"].(int64); !ok {
		t.Fatalf("expected large int to be int64, got %T", out["nested"].(map[string]any)["value"])
	}
	list := out["list"].([]any)
	if _, ok := list[0].(int64); !ok {
		t.Fatalf("expected list[0] to be int64, got %T", list[0])
	}
	if _, ok := list[1].(float64); !ok {
		t.Fatalf("expected list[1] to be float64, got %T", list[1])
	}
}
