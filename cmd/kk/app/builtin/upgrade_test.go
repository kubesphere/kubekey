//go:build builtin
// +build builtin

package builtin

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNodesAtTarget(t *testing.T) {
	tests := []struct {
		name     string
		kubelets []string
		target   string
		want     bool
		wantErr  bool
	}{
		{
			name:     "all nodes exactly at target",
			kubelets: []string{"v1.34.3", "v1.34.3", "v1.34.3"},
			target:   "v1.34.3",
			want:     true,
		},
		{
			name:     "all nodes above target",
			kubelets: []string{"v1.34.3"},
			target:   "v1.33.7",
			want:     true,
		},
		{
			name:     "same-minor patch bump is real work",
			kubelets: []string{"v1.34.1"},
			target:   "v1.34.3",
			want:     false,
		},
		{
			name:     "minor behind",
			kubelets: []string{"v1.33.7"},
			target:   "v1.34.3",
			want:     false,
		},
		{
			name:     "one lagging node keeps the upgrade alive",
			kubelets: []string{"v1.34.3", "v1.34.3", "v1.33.7"},
			target:   "v1.34.3",
			want:     false,
		},
		{
			name:     "unreadable kubelet version is not evidence of being at target",
			kubelets: []string{""},
			target:   "v1.34.3",
			want:     false,
		},
		{
			name:     "no nodes is an error",
			kubelets: []string{},
			target:   "v1.34.3",
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, len(tt.kubelets))
			for i, kv := range tt.kubelets {
				objs = append(objs, &corev1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("node-%d", i)},
					Status: corev1.NodeStatus{
						NodeInfo: corev1.NodeSystemInfo{KubeletVersion: kv},
					},
				})
			}
			cs := fake.NewSimpleClientset(objs...)

			got, err := nodesAtTarget(cs, tt.target)
			if (err != nil) != tt.wantErr {
				t.Fatalf("nodesAtTarget() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("nodesAtTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}
