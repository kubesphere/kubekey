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

package _const

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
)

func TestWorkDir(t *testing.T) {
	// should not get workdir before set
	assert.Empty(t, GetWorkDir())
	// set workdir
	SetWorkDir("/tmp")
	assert.Equal(t, "/tmp", GetWorkDir())
	// should not set workdir again
	SetWorkDir("/tmp2")
	assert.Equal(t, "/tmp", GetWorkDir())
}

func TestResourceFromObject(t *testing.T) {
	assert.Equal(t, RuntimePipelineDir, ResourceFromObject(&kubekeyv1.Pipeline{}))
	assert.Equal(t, RuntimePipelineDir, ResourceFromObject(&kubekeyv1.PipelineList{}))
	assert.Equal(t, RuntimeConfigDir, ResourceFromObject(&kubekeyv1.Config{}))
	assert.Equal(t, RuntimeConfigDir, ResourceFromObject(&kubekeyv1.ConfigList{}))
	assert.Equal(t, RuntimeInventoryDir, ResourceFromObject(&kubekeyv1.Inventory{}))
	assert.Equal(t, RuntimeInventoryDir, ResourceFromObject(&kubekeyv1.InventoryList{}))
	assert.Equal(t, RuntimePipelineTaskDir, ResourceFromObject(&kubekeyv1alpha1.Task{}))
	assert.Equal(t, RuntimePipelineTaskDir, ResourceFromObject(&kubekeyv1alpha1.TaskList{}))
	assert.Equal(t, "", ResourceFromObject(&unstructured.Unstructured{}))
}
