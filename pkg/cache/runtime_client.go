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

package cache

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

type delegatingClient struct {
	client ctrlclient.Client
	scheme *runtime.Scheme
}

func NewDelegatingClient(client ctrlclient.Client) ctrlclient.Client {
	scheme := runtime.NewScheme()
	if err := kubekeyv1.AddToScheme(scheme); err != nil {
		klog.Errorf("failed to add scheme: %v", err)
	}
	kubekeyv1.SchemeBuilder.Register(&kubekeyv1alpha1.Task{}, &kubekeyv1alpha1.TaskList{})
	return &delegatingClient{
		client: client,
		scheme: scheme,
	}
}

func (d delegatingClient) Get(ctx context.Context, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
	resource := _const.ResourceFromObject(obj)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.Get(ctx, key, obj, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
	}

	path := filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, key.Namespace, resource, key.Name, key.Name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		klog.Errorf("failed to read yaml file: %v", err)
		return err
	}
	if err := yaml.Unmarshal(data, obj); err != nil {
		klog.Errorf("unmarshal file %s error %v", path, err)
		return err
	}
	return nil
}

func (d delegatingClient) List(ctx context.Context, list ctrlclient.ObjectList, opts ...ctrlclient.ListOption) error {
	resource := _const.ResourceFromObject(list)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.List(ctx, list, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", list.GetObjectKind().GroupVersionKind().String())
	}
	// read all runtime.Object
	var objects []runtime.Object
	runtimeDirEntries, err := os.ReadDir(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir))
	if err != nil && !os.IsNotExist(err) {
		klog.Errorf("readDir %s error %v", filepath.Join(_const.GetWorkDir(), _const.RuntimeDir), err)
		return err
	}
	for _, re := range runtimeDirEntries {
		if re.IsDir() {
			resourceDir := filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, re.Name(), resource)
			entries, err := os.ReadDir(resourceDir)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				klog.Errorf("readDir %s error %v", resourceDir, err)
				return err
			}
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				resourceFile := filepath.Join(resourceDir, e.Name(), e.Name()+".yaml")
				data, err := os.ReadFile(resourceFile)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					klog.Errorf("read file %s error: %v", resourceFile, err)
					return err
				}
				var obj runtime.Object
				switch resource {
				case _const.RuntimePipelineDir:
					obj = &kubekeyv1.Pipeline{}
				case _const.RuntimeInventoryDir:
					obj = &kubekeyv1.Inventory{}
				case _const.RuntimeConfigDir:
					obj = &kubekeyv1.Config{}
				case _const.RuntimePipelineTaskDir:
					obj = &kubekeyv1alpha1.Task{}
				}
				if err := yaml.Unmarshal(data, &obj); err != nil {
					klog.Errorf("unmarshal file %s error: %v", resourceFile, err)
					return err
				}
				objects = append(objects, obj)
			}
		}
	}

	o := ctrlclient.ListOptions{}
	o.ApplyOptions(opts)

	switch {
	case o.Namespace != "":
		for i := len(objects) - 1; i >= 0; i-- {
			if objects[i].(metav1.Object).GetNamespace() != o.Namespace {
				objects = append(objects[:i], objects[i+1:]...)
			}
		}
	}

	if err := apimeta.SetList(list, objects); err != nil {
		return err
	}
	return nil
}

func (d delegatingClient) Create(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
	resource := _const.ResourceFromObject(obj)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.Create(ctx, obj, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
	}

	data, err := yaml.Marshal(obj)
	if err != nil {
		klog.Errorf("failed to marshal object: %v", err)
		return err
	}
	if err := os.MkdirAll(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, obj.GetNamespace(), resource, obj.GetName()), fs.ModePerm); err != nil {
		klog.Errorf("create dir %s error: %v", filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, obj.GetNamespace(), resource, obj.GetName()), err)
		return err
	}
	return os.WriteFile(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, obj.GetNamespace(), resource, obj.GetName(), obj.GetName()+".yaml"), data, fs.ModePerm)
}

func (d delegatingClient) Delete(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.DeleteOption) error {
	resource := _const.ResourceFromObject(obj)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.Delete(ctx, obj, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
	}

	return os.RemoveAll(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, obj.GetNamespace(), resource, obj.GetName()))
}

func (d delegatingClient) Update(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.UpdateOption) error {
	resource := _const.ResourceFromObject(obj)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.Update(ctx, obj, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
	}

	data, err := yaml.Marshal(obj)
	if err != nil {
		klog.Errorf("failed to marshal object: %v", err)
		return err
	}
	return os.WriteFile(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, obj.GetNamespace(), resource, obj.GetName(), obj.GetName()+".yaml"), data, fs.ModePerm)
}

func (d delegatingClient) Patch(ctx context.Context, obj ctrlclient.Object, patch ctrlclient.Patch, opts ...ctrlclient.PatchOption) error {
	resource := _const.ResourceFromObject(obj)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.Patch(ctx, obj, patch, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
	}

	patchData, err := patch.Data(obj)
	if err != nil {
		klog.Errorf("failed to get patch data: %v", err)
		return err
	}
	if len(patchData) == 0 {
		klog.V(4).Infof("nothing to patch, skip")
		return nil
	}
	data, err := yaml.Marshal(obj)
	if err != nil {
		klog.Errorf("failed to marshal object: %v", err)
		return err
	}
	return os.WriteFile(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, obj.GetNamespace(), resource, obj.GetName(), obj.GetName()+".yaml"), data, fs.ModePerm)
}

func (d delegatingClient) DeleteAllOf(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.DeleteAllOfOption) error {
	resource := _const.ResourceFromObject(obj)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.DeleteAllOf(ctx, obj, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
	}
	return d.Delete(ctx, obj)
}

func (d delegatingClient) Status() ctrlclient.SubResourceWriter {
	if d.client != nil {
		return d.client.Status()
	}
	return &delegatingSubResourceWriter{client: d.client}
}

func (d delegatingClient) SubResource(subResource string) ctrlclient.SubResourceClient {
	if d.client != nil {
		return d.client.SubResource(subResource)
	}
	return nil
}

func (d delegatingClient) Scheme() *runtime.Scheme {
	if d.client != nil {
		return d.client.Scheme()
	}
	return d.scheme
}

func (d delegatingClient) RESTMapper() apimeta.RESTMapper {
	if d.client != nil {
		return d.client.RESTMapper()
	}
	return nil
}

func (d delegatingClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	if d.client != nil {
		return d.client.GroupVersionKindFor(obj)
	}
	return apiutil.GVKForObject(obj, d.scheme)
}

func (d delegatingClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	if d.client != nil {
		return d.client.IsObjectNamespaced(obj)
	}
	return true, nil
}

type delegatingSubResourceWriter struct {
	client ctrlclient.Client
}

func (d delegatingSubResourceWriter) Create(ctx context.Context, obj ctrlclient.Object, subResource ctrlclient.Object, opts ...ctrlclient.SubResourceCreateOption) error {
	resource := _const.ResourceFromObject(obj)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.Status().Create(ctx, obj, subResource, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
	}

	data, err := yaml.Marshal(obj)
	if err != nil {
		klog.Errorf("failed to marshal object: %v", err)
		return err
	}
	return os.WriteFile(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, obj.GetNamespace(), resource, obj.GetName(), obj.GetName()+".yaml"), data, fs.ModePerm)

}

func (d delegatingSubResourceWriter) Update(ctx context.Context, obj ctrlclient.Object, opts ...ctrlclient.SubResourceUpdateOption) error {
	resource := _const.ResourceFromObject(obj)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.Status().Update(ctx, obj, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
	}

	data, err := yaml.Marshal(obj)
	if err != nil {
		klog.Errorf("failed to marshal object: %v", err)
		return err
	}
	return os.WriteFile(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, obj.GetNamespace(), resource, obj.GetName(), obj.GetName()+".yaml"), data, fs.ModePerm)
}

func (d delegatingSubResourceWriter) Patch(ctx context.Context, obj ctrlclient.Object, patch ctrlclient.Patch, opts ...ctrlclient.SubResourcePatchOption) error {
	resource := _const.ResourceFromObject(obj)
	if d.client != nil && resource != _const.RuntimePipelineTaskDir {
		return d.client.Status().Patch(ctx, obj, patch, opts...)
	}
	if resource == "" {
		return fmt.Errorf("unsupported object type: %s", obj.GetObjectKind().GroupVersionKind().String())
	}

	patchData, err := patch.Data(obj)
	if err != nil {
		klog.Errorf("failed to get patch data: %v", err)
		return err
	}
	if len(patchData) == 0 {
		klog.V(4).Infof("nothing to patch, skip")
		return nil
	}
	data, err := yaml.Marshal(obj)
	if err != nil {
		klog.Errorf("failed to marshal object: %v", err)
		return err
	}
	return os.WriteFile(filepath.Join(_const.GetWorkDir(), _const.RuntimeDir, obj.GetNamespace(), resource, obj.GetName(), obj.GetName()+".yaml"), data, fs.ModePerm)
}

func getPatchedJSON(patchType types.PatchType, originalJS, patchJS []byte, gvk schema.GroupVersionKind, creater runtime.ObjectCreater) ([]byte, error) {
	switch patchType {
	case types.JSONPatchType:
		patchObj, err := jsonpatch.DecodePatch(patchJS)
		if err != nil {
			return nil, err
		}
		bytes, err := patchObj.Apply(originalJS)
		// TODO: This is pretty hacky, we need a better structured error from the json-patch
		if err != nil && strings.Contains(err.Error(), "doc is missing key") {
			msg := err.Error()
			ix := strings.Index(msg, "key:")
			key := msg[ix+5:]
			return bytes, fmt.Errorf("Object to be patched is missing field (%s)", key)
		}
		return bytes, err

	case types.MergePatchType:
		return jsonpatch.MergePatch(originalJS, patchJS)

	case types.StrategicMergePatchType:
		// get a typed object for this GVK if we need to apply a strategic merge patch
		obj, err := creater.New(gvk)
		if err != nil {
			return nil, fmt.Errorf("cannot apply strategic merge patch for %s locally, try --type merge", gvk.String())
		}
		return strategicpatch.StrategicMergePatch(originalJS, patchJS, obj)

	default:
		// only here as a safety net - go-restful filters content-type
		return nil, fmt.Errorf("unknown Content-Type header for patch: %v", patchType)
	}
}
