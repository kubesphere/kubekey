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
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

func TestTransport(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	_const.SetWorkDir(cwd)
	// create runtimeDir
	if _, err := os.Stat(_const.RuntimeDir); err != nil && os.IsNotExist(err) {
		err = os.Mkdir(_const.RuntimeDir, os.ModePerm)
		assert.NoError(t, err)
	}
	defer os.RemoveAll(_const.RuntimeDir)

	cli, err := NewLocalClient()
	assert.NoError(t, err)

	testcases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "create task",
			fn: func() error {
				return cli.Create(context.Background(), &kubekeyv1alpha1.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: kubekeyv1alpha1.KubeKeyTaskSpec{
						Name: "test",
					},
				})
			},
		},
		{
			name: "get task",
			fn: func() error {
				task := &kubekeyv1alpha1.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: kubekeyv1alpha1.KubeKeyTaskSpec{
						Name: "test",
					},
				}
				cli.Create(context.Background(), task)
				return cli.Get(context.Background(), ctrlclient.ObjectKeyFromObject(task), task)
			},
		},
		{
			name: "list task",
			fn: func() error {
				cli.Create(context.Background(), &kubekeyv1alpha1.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: kubekeyv1alpha1.KubeKeyTaskSpec{
						Name: "test",
					},
				})
				tasklist := &kubekeyv1alpha1.TaskList{}
				return cli.List(context.Background(), tasklist)
			},
		},
		{
			name: "update task",
			fn: func() error {
				cli.Create(context.Background(), &kubekeyv1alpha1.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: kubekeyv1alpha1.KubeKeyTaskSpec{
						Name: "test",
					},
				})
				if err := cli.Update(context.Background(), &kubekeyv1alpha1.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: kubekeyv1alpha1.KubeKeyTaskSpec{
						Name: "test1",
					},
				}); err != nil && !strings.Contains(err.Error(), "spec is immutable") {
					return err
				}
				return nil
			},
		},
		{
			name: "delete task",
			fn: func() error {
				cli.Create(context.Background(), &kubekeyv1alpha1.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: kubekeyv1alpha1.KubeKeyTaskSpec{
						Name: "test",
					},
				})
				return cli.Delete(context.Background(), &kubekeyv1alpha1.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				})
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, tc.fn())
		})
	}
}
