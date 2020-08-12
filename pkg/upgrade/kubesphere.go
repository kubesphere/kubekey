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

package upgrade

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/deploy"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
	ksv2 "github.com/kubesphere/kubekey/pkg/kubesphere/v2"
	ksv3 "github.com/kubesphere/kubekey/pkg/kubesphere/v3"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	yamlV2 "gopkg.in/yaml.v2"
	kubeErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"strings"
)

func KsToV3(version, repo, kubeconfig string) error {
	clientset, err := util.NewClient(kubeconfig)
	if err != nil {
		return err
	}

	clusterConfigMap, err1 := clientset.CoreV1().ConfigMaps("kubesphere-system").Get(context.TODO(), "ks-installer", metav1.GetOptions{})
	if err1 != nil {
		return err1
	}

	clusterCfgV2 := ksv2.V2{}
	clusterCfgV3 := ksv3.V3{}
	if err := yamlV2.Unmarshal([]byte(clusterConfigMap.Data["ks-config.yaml"]), &clusterCfgV2); err != nil {
		return err
	}

	configV3, err := MigrateConfig2to3(&clusterCfgV2, &clusterCfgV3)
	if err != nil {
		return err
	}

	fmt.Println(configV3)

	var kubesphereConfig, installerYaml string

	switch version {
	case "":
		kubesphereConfig = configV3
		str, err := kubesphere.GenerateKubeSphereYaml(repo, "latest")
		if err != nil {
			return err
		}
		installerYaml = str
	case "v3.0.0":
		kubesphereConfig = configV3
		str, err := kubesphere.GenerateKubeSphereYaml(repo, "latest")
		if err != nil {
			return err
		}
		installerYaml = str
	default:
		return errors.New(fmt.Sprintf("Unsupported version: %s", strings.TrimSpace(version)))
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
			if err := deploy.CreateObject(clientset, j2, deploy.CustomResourceDefinition); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})

			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "Namespace":
			if err := deploy.CreateObject(clientset, j2, deploy.Namespaces); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})
			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "ServiceAccount":
			if err := deploy.CreateObject(clientset, j2, deploy.ServiceAccount); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})
			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "ClusterRole":
			if err := deploy.CreateObject(clientset, j2, deploy.ClusterRole); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})
			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "ClusterRoleBinding":
			if err := deploy.CreateObject(clientset, j2, deploy.ClusterRoleBinding); err != nil {
				if !kubeErr.IsAlreadyExists(err) {
					return err
				}
			}

			metadata := result["metadata"].(map[string]interface{})
			fmt.Printf("%s/%s  created\n", result["kind"], metadata["name"])
		case "Deployment":
			if err := deploy.CreateObject(clientset, j2, deploy.Deployment); err != nil {
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

	if err := deploy.CreateObject(clientset, j2, deploy.ClusterConfiguration); err != nil {
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

func MigrateConfig2to3(v2 *ksv2.V2, v3 *ksv3.V3) (string, error) {
	v3.Etcd = ksv3.Etcd(v2.Etcd)
	v3.Persistence = ksv3.Persistence(v2.Persistence)
	v3.Alerting = ksv3.Alerting(v2.Alerting)
	v3.Notification = ksv3.Notification(v2.Notification)
	v3.LocalRegistry = v2.LocalRegistry
	v3.Servicemesh = ksv3.Servicemesh(v2.Servicemesh)
	v3.Devops = ksv3.Devops(v2.Devops)
	v3.Openpitrix = ksv3.Openpitrix(v2.Openpitrix)
	v3.Console = ksv3.Console(v2.Console)

	if v2.MetricsServerNew.Enabled == "" {
		if v2.MetricsServerOld.Enabled == "true" || v2.MetricsServerOld.Enabled == "True" {
			v3.MetricsServer.Enabled = true
		} else {
			v3.MetricsServer.Enabled = false
		}
	} else {
		if v2.MetricsServerNew.Enabled == "true" || v2.MetricsServerNew.Enabled == "True" {
			v3.MetricsServer.Enabled = true
		} else {
			v3.MetricsServer.Enabled = false
		}
	}

	v3.Monitoring.PrometheusMemoryRequest = v2.Monitoring.PrometheusMemoryRequest
	v3.Monitoring.PrometheusReplicas = v2.Monitoring.PrometheusReplicas
	v3.Monitoring.PrometheusVolumeSize = v2.Monitoring.PrometheusVolumeSize
	v3.Monitoring.AlertmanagerReplicas = 1

	v3.Common.EtcdVolumeSize = v2.Common.EtcdVolumeSize
	v3.Common.MinioVolumeSize = v2.Common.MinioVolumeSize
	v3.Common.MysqlVolumeSize = v2.Common.MysqlVolumeSize
	v3.Common.OpenldapVolumeSize = v2.Common.OpenldapVolumeSize
	v3.Common.RedisVolumSize = v2.Common.RedisVolumSize
	v3.Common.ES.ElasticsearchDataReplicas = v2.Logging.ElasticsearchDataReplicas
	v3.Common.ES.ElasticsearchMasterReplicas = v2.Logging.ElasticsearchMasterReplicas
	v3.Common.ES.ElkPrefix = v2.Logging.ElkPrefix
	v3.Common.ES.LogMaxAge = v2.Logging.LogMaxAge
	if v2.Logging.ElasticsearchVolumeSize == "" {
		v3.Common.ES.ElasticsearchDataVolumeSize = v2.Logging.ElasticsearchDataVolumeSize
		v3.Common.ES.ElasticsearchMasterVolumeSize = v2.Logging.ElasticsearchMasterVolumeSize
	} else {
		v3.Common.ES.ElasticsearchMasterVolumeSize = "4Gi"
		v3.Common.ES.ElasticsearchDataVolumeSize = v2.Logging.ElasticsearchVolumeSize
	}

	v3.Logging.Enabled = v2.Logging.Enabled
	v3.Logging.LogsidecarReplicas = v2.Logging.LogsidecarReplicas

	v3.Authentication.JwtSecret = ""
	v3.Multicluster.ClusterRole = "none"
	v3.Events.Ruler.Replicas = 2

	var clusterConfiguration = ksv3.ClusterConfig{
		ApiVersion: "installer.kubesphere.io/v1alpha1",
		Kind:       "ClusterConfiguration",
		Metadata: ksv3.Metadata{
			Name:      "ks-installer",
			Namespace: "kubesphere-system",
			Label:     ksv3.Label{Version: "v3.0.0"},
		},
		Spec: v3,
	}

	configV3, err := yamlV2.Marshal(clusterConfiguration)
	if err != nil {
		return "", err
	}

	return string(configV3), nil
}

func SyncConfiguration(mgr *manager.Manager) error {
	if err := mgr.RunTaskOnMasterNodes(syncConfiguration, true); err != nil {
		return err
	}
	return nil
}

func syncConfiguration(mgr *manager.Manager, _ *kubekeyapi.HostCfg) error {
	if mgr.Runner.Index == 0 {
		configV2Str, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubectl get cm -n kubesphere-system ks-installer -o jsonpath='{.data.ks-config\\.yaml}'\"", 2, false)
		if err != nil {
			return err
		}

		clusterCfgV2 := ksv2.V2{}
		clusterCfgV3 := ksv3.V3{}
		if err := yamlV2.Unmarshal([]byte(configV2Str), &clusterCfgV2); err != nil {
			return err
		}

		configV3, err := MigrateConfig2to3(&clusterCfgV2, &clusterCfgV3)
		if err != nil {
			return err
		}

		mgr.Cluster.KubeSphere.Configurations = "---\n" + configV3

	}

	return nil
}
