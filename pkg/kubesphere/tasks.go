/*
 Copyright 2021 The KubeSphere Authors.

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

package kubesphere

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	ksv2 "github.com/kubesphere/kubekey/pkg/kubesphere/v2"
	ksv3 "github.com/kubesphere/kubekey/pkg/kubesphere/v3"
	"github.com/kubesphere/kubekey/pkg/version/kubesphere"
	"github.com/kubesphere/kubekey/pkg/version/kubesphere/templates"
	"github.com/pkg/errors"
	yamlV2 "gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AddInstallerConfig struct {
	common.KubeAction
}

func (a *AddInstallerConfig) Execute(runtime connector.Runtime) error {
	configurationBase64 := base64.StdEncoding.EncodeToString([]byte(a.KubeConf.Cluster.KubeSphere.Configurations))
	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("echo %s | base64 -d >> /etc/kubernetes/addons/kubesphere.yaml", configurationBase64),
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), "add config to ks-installer manifests failed")
	}
	return nil
}

type CreateNamespace struct {
	common.KubeAction
}

func (c *CreateNamespace) Execute(runtime connector.Runtime) error {
	_, err := runtime.GetRunner().SudoCmd(`cat <<EOF | /usr/local/bin/kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: kubesphere-system
---
apiVersion: v1
kind: Namespace
metadata:
  name: kubesphere-monitoring-system
EOF
`, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "create namespace: kubesphere-system and kubesphere-monitoring-system")
	}
	return nil
}

type Setup struct {
	common.KubeAction
}

func (s *Setup) Execute(runtime connector.Runtime) error {
	filePath := filepath.Join(common.KubeAddonsDir, templates.KsInstaller.Name())

	var addrList []string
	for _, host := range runtime.GetHostsByRole(common.ETCD) {
		addrList = append(addrList, host.GetInternalAddress())
	}
	etcdEndPoint := strings.Join(addrList, ",")
	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("sed -i '/endpointIps/s/\\:.*/\\: %s/g' %s", etcdEndPoint, filePath),
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("update etcd endpoint failed"))
	}

	if s.KubeConf.Cluster.Registry.PrivateRegistry != "" {
		PrivateRegistry := strings.Replace(s.KubeConf.Cluster.Registry.PrivateRegistry, "/", "\\/", -1)
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("sed -i '/local_registry/s/\\:.*/\\: %s/g' %s", PrivateRegistry, filePath),
			false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("add private registry: %s failed", s.KubeConf.Cluster.Registry.PrivateRegistry))
		}
	} else {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("sed -i '/local_registry/d' %s", filePath), false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("remove private registry failed"))
		}
	}

	_, ok := kubesphere.CNSource[s.KubeConf.Cluster.KubeSphere.Version]
	if ok && (os.Getenv("KKZONE") == "cn" || s.KubeConf.Cluster.Registry.PrivateRegistry == "registry.cn-beijing.aliyuncs.com") {
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("sed -i '/zone/s/\\:.*/\\: %s/g' %s", "cn", filePath),
			false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("add kubekey zone: %s failed", s.KubeConf.Cluster.Registry.PrivateRegistry))
		}
	} else {
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("sed -i '/zone/d' %s", filePath),
			false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("remove kubekey zone failed"))
		}
	}

	switch s.KubeConf.Arg.ContainerManager {
	case "docker", "containerd", "crio":
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("sed -i '/containerruntime/s/\\:.*/\\: %s/g' /etc/kubernetes/addons/kubesphere.yaml", s.KubeConf.Arg.ContainerManager), false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("set container runtime: %s failed", s.KubeConf.Arg.ContainerManager))
		}
	default:
		logger.Log.Message(runtime.RemoteHost().GetName(),
			fmt.Sprintf("Currently, the logging function of KubeSphere does not support %s. If %s is used, the logging function will be unavailable.",
				s.KubeConf.Arg.ContainerManager, s.KubeConf.Arg.ContainerManager))
	}

	caFile := "/etc/ssl/etcd/ssl/ca.pem"
	certFile := fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", runtime.GetHostsByRole(common.ETCD)[0].GetName())
	keyFile := fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", runtime.GetHostsByRole(common.ETCD)[0].GetName())
	if output, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("/usr/local/bin/kubectl -n kubesphere-monitoring-system create secret generic kube-etcd-client-certs "+
			"--from-file=etcd-client-ca.crt=%s "+
			"--from-file=etcd-client.crt=%s "+
			"--from-file=etcd-client.key=%s", caFile, certFile, keyFile), true); err != nil {
		if !strings.Contains(output, "exists") {
			return err
		}
	}
	return nil
}

type Apply struct {
	common.KubeAction
}

func (a *Apply) Execute(runtime connector.Runtime) error {
	filePath := filepath.Join(common.KubeAddonsDir, templates.KsInstaller.Name())

	deployKubesphereCmd := fmt.Sprintf("/usr/local/bin/kubectl apply -f %s --force", filePath)
	if _, err := runtime.GetRunner().SudoCmd(deployKubesphereCmd, true); err != nil {
		return errors.Wrapf(errors.WithStack(err), "deploy %s failed", filePath)
	}
	return nil
}

type Check struct {
	common.KubeAction
}

func (c *Check) Execute(runtime connector.Runtime) error {
	var (
		position = 1
		notes    = "Please wait for the installation to complete: "
	)

	ch := make(chan string)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go CheckKubeSphereStatus(ctx, runtime, ch)

	stop := false
	for !stop {
		select {
		case res := <-ch:
			fmt.Printf("\033[%dA\033[K", position)
			fmt.Println(res)
			stop = true
		default:
			for i := 0; i < 10; i++ {
				if i < 5 {
					fmt.Printf("\033[%dA\033[K", position)

					output := fmt.Sprintf(
						"%s%s%s",
						notes,
						strings.Repeat(" ", i),
						">>--->",
					)

					fmt.Printf("%s \033[K\n", output)
					time.Sleep(time.Duration(200) * time.Millisecond)
				} else {
					fmt.Printf("\033[%dA\033[K", position)

					output := fmt.Sprintf(
						"%s%s%s",
						notes,
						strings.Repeat(" ", 10-i),
						"<---<<",
					)

					fmt.Printf("%s \033[K\n", output)
					time.Sleep(time.Duration(200) * time.Millisecond)
				}
			}
		}
	}
	return nil
}

func CheckKubeSphereStatus(ctx context.Context, runtime connector.Runtime, stopChan chan string) {
	defer close(stopChan)
	for {
		select {
		case <-ctx.Done():
			stopChan <- ""
		default:
			_, err := runtime.GetRunner().SudoCmd(
				"/usr/local/bin/kubectl exec -n kubesphere-system "+
					"$(kubectl get pod -n kubesphere-system -l app=ks-install -o jsonpath='{.items[0].metadata.name}') "+
					"-- ls /kubesphere/playbooks/kubesphere_running", false)
			if err == nil {
				output, err := runtime.GetRunner().SudoCmd(
					"/usr/local/bin/kubectl exec -n kubesphere-system "+
						"$(kubectl get pod -n kubesphere-system -l app=ks-install -o jsonpath='{.items[0].metadata.name}') "+
						"-- cat /kubesphere/playbooks/kubesphere_running", false)
				if err == nil && output != "" {
					stopChan <- output
					break
				}
			}
		}
	}
}

type ConvertV2ToV3 struct {
	common.KubeAction
}

func (c *ConvertV2ToV3) Execute(runtime connector.Runtime) error {
	configV2Str, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl get cm -n kubesphere-system ks-installer -o jsonpath='{.data.ks-config\\.yaml}'",
		false)
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
	c.KubeConf.Cluster.KubeSphere.Configurations = configV3
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
	//v3.Monitoring.PrometheusReplicas = v2.Monitoring.PrometheusReplicas
	v3.Monitoring.PrometheusVolumeSize = v2.Monitoring.PrometheusVolumeSize
	//v3.Monitoring.AlertmanagerReplicas = 1

	v3.Common.EtcdVolumeSize = v2.Common.EtcdVolumeSize
	v3.Common.MinioVolumeSize = v2.Common.MinioVolumeSize
	v3.Common.MysqlVolumeSize = v2.Common.MysqlVolumeSize
	v3.Common.OpenldapVolumeSize = v2.Common.OpenldapVolumeSize
	v3.Common.RedisVolumSize = v2.Common.RedisVolumSize
	//v3.Common.ES.ElasticsearchDataReplicas = v2.Logging.ElasticsearchDataReplicas
	//v3.Common.ES.ElasticsearchMasterReplicas = v2.Logging.ElasticsearchMasterReplicas
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
