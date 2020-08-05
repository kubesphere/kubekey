/*
Copyright 2020 The KubeSphere Authors.

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

package deploy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/ghodss/yaml"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
	"github.com/pkg/errors"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const (
	customResourceDefinition = "/apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions"
	namespaces               = "/api/v1/namespaces"
	serviceAccount           = "/api/v1/namespaces/kubesphere-system/serviceaccounts"
	clusterRole              = "/apis/rbac.authorization.k8s.io/v1/clusterroles"
	clusterRoleBinding       = "/apis/rbac.authorization.k8s.io/v1/clusterrolebindings"
	deployment               = "/apis/apps/v1/namespaces/kubesphere-system/deployments"
	clusterConfiguration     = "/apis/installer.kubesphere.io/v1alpha1/namespaces/kubesphere-system/clusterconfigurations"
)

func DeployKubeSphere(version, repo, kubeconfig string) error {
	var kubesphereConfig, installerYaml string

	switch version {
	case "":
		kubesphereConfig = kubesphere.V3_0_0
		str, err := kubesphere.GenerateKubeSphereYaml(repo, "latest")
		if err != nil {
			return err
		}
		installerYaml = str
	case "v3.0.0":
		kubesphereConfig = kubesphere.V3_0_0
		str, err := kubesphere.GenerateKubeSphereYaml(repo, "latest")
		if err != nil {
			return err
		}
		installerYaml = str
	case "v2.1.1":
		kubesphereConfig = kubesphere.V2_1_1
		str, err := kubesphere.GenerateKubeSphereYaml(repo, "v2.1.1")
		if err != nil {
			return err
		}
		installerYaml = str
	default:
		return errors.New(fmt.Sprintf("Unsupported version: %s", strings.TrimSpace(version)))
	}

	clientset, err := newClient(kubeconfig)
	if err != nil {
		return err
	}

	b1 := bufio.NewReader(bytes.NewReader([]byte(installerYaml)))
	for {
		result := make(map[string]interface{})
		content, err := k8syaml.NewYAMLReader(b1).Read()
		if len(content) == 0 {
			break
		}
		if err != nil {
			return errors.Wrap(err, "Unable to read the manifests")
		}

		err = yaml.Unmarshal(content, &result)
		if err != nil {
			return errors.Wrap(err, "Unable to unmarshal the manifests")
		}

		j2, err1 := yaml.YAMLToJSON(content)
		if err1 != nil {
			return err
		}

		switch result["kind"] {
		case "CustomResourceDefinition":
			if err := createObject(clientset, j2, customResourceDefinition); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})

			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "Namespace":
			if err := createObject(clientset, j2, namespaces); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})
			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "ServiceAccount":
			if err := createObject(clientset, j2, serviceAccount); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})
			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "ClusterRole":
			if err := createObject(clientset, j2, clusterRole); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})
			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "ClusterRoleBinding":
			if err := createObject(clientset, j2, clusterRoleBinding); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})
			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "Deployment":
			if err := createObject(clientset, j2, deployment); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})
			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		}
	}

	j2, err1 := yaml.YAMLToJSON([]byte(kubesphereConfig))
	if err1 != nil {
		return err
	}

	patchJSON := []byte(`[
		{"op": "replace", "path": "/spec/etcd/monitoring", "value": "false"}
	]`)
	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		return err
	}

	modified, err := patch.Apply(j2)
	if err != nil {
		return err
	}

	if err := createObject(clientset, modified, clusterConfiguration); err != nil {
		if !kubeErr.IsAlreadyExists(err) {
			return err
		}
	}
	result := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(kubesphereConfig), &result)
	if err != nil {
		return errors.Wrap(err, "Unable to unmarshal the manifests")
	}
	metadata := result["metadata"].(map[string]interface{})
	fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
	return nil
}

func createObject(clientset *kubernetes.Clientset, body []byte, request string) error {
	if err := clientset.
		RESTClient().Post().
		AbsPath(request).
		Body(body).
		Do(context.TODO()).Error(); err != nil {
		if !kubeErr.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}
