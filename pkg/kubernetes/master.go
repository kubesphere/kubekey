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

package kubernetes

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kubesphere/kubekey/pkg/kubernetes/config/v1beta2"

	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/plugins/dns"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
)

const (
	IsInitCluster = true
	Docker        = "docker"
	Conatinerd    = "containerd"
	Crio          = "crio"
	Isula         = "isula"
)

// ClusterStatus is used to store cluster status
type ClusterStatus struct {
	isExist        bool
	version        string
	allNodesInfo   map[string]string
	kubeconfig     string
	bootstrapToken string
	certificateKey string
}

// GetClusterStatus is used to fetch status and info from cluster.
func (s *ClusterStatus) GetClusterStatus(mgr *manager.Manager, _ *kubekeyapiv1alpha1.HostCfg) error {
	output, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"[ -f /etc/kubernetes/admin.conf ] && echo 'Cluster already exists.' || echo 'Cluster will be created.'\"", 0, true)
	if strings.Contains(output, "Cluster will be created") {
		s.isExist = false
		return nil
	} else {
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "Failed to find /etc/kubernetes/admin.conf")
		}

		s.isExist = true
		if output, err := mgr.Runner.ExecuteCmd("sudo cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | awk -F '[:]' '{print $(NF-0)}'", 0, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "Failed to find current version")
		} else {
			if !strings.Contains(output, "No such file or directory") {
				s.version = output
			}
		}

		if err := s.loadKubeConfig(mgr); err != nil {
			return err
		}

		if err := s.getClusterInfo(mgr); err != nil {
			return err
		}

		if err := s.getJoinInfo(mgr); err != nil {
			return err
		}
	}

	return nil
}

// InitKubernetesCluster is used to init a new cluster.
func (s *ClusterStatus) InitKubernetesCluster(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if mgr.Runner.Index == 0 && !s.isExist {
		if err := generateKubeadmConfig(mgr, node, IsInitCluster, "", ""); err != nil {
			return err
		}

		for i := 0; i < 3; i++ {
			_, err2 := mgr.Runner.ExecuteCmd("sudo env PATH=$PATH /bin/sh -c \"/usr/local/bin/kubeadm init --config=/etc/kubernetes/kubeadm-config.yaml --ignore-preflight-errors=FileExisting-crictl\"", 0, true)
			if err2 != nil {
				if i == 2 {
					return errors.Wrap(errors.WithStack(err2), "Failed to init kubernetes cluster")
				}
				_, _ = mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubeadm reset -f\"", 0, true)
			} else {
				break
			}
		}
		if err := GetKubeConfigForControlPlane(mgr); err != nil {
			return err
		}
		if err := removeMasterTaint(mgr, node); err != nil {
			return err
		}
		if err := addWorkerLabel(mgr, node); err != nil {
			return err
		}
		if err := dns.CreateClusterDns(mgr); err != nil {
			return err
		}
	}

	return nil
}

// JoinNodesToCluster is used to join node to Cluster.
func (s *ClusterStatus) JoinNodesToCluster(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if !s.existNode(node) {
		if err := generateKubeadmConfig(mgr, node, !IsInitCluster, s.bootstrapToken, s.certificateKey); err != nil {
			return err
		}

		for i := 0; i < 3; i++ {
			_, err := mgr.Runner.ExecuteCmd("sudo env PATH=$PATH /bin/sh -c \"/usr/local/bin/kubeadm join --config=/etc/kubernetes/kubeadm-config.yaml\"", 0, true)
			if err != nil {
				if i == 2 {
					return errors.Wrap(errors.WithStack(err), "Failed to add master to cluster")
				}
				_, _ = mgr.Runner.ExecuteCmd("sudo env PATH=$PATH /bin/sh -c \"/usr/local/bin/kubeadm reset -f\"", 0, true)
			} else {
				break
			}
		}

		if node.IsMaster {
			if err := GetKubeConfigForControlPlane(mgr); err != nil {
				return err
			}
			err1 := removeMasterTaint(mgr, node)
			if err1 != nil {
				return err1
			}
			err2 := addWorkerLabel(mgr, node)
			if err2 != nil {
				return err2
			}
		}

		if node.IsWorker && !node.IsMaster {
			if err := s.getKubeConfigForWorker(mgr); err != nil {
				return err
			}
		}

		if err := addWorkerLabel(mgr, node); err != nil {
			return err
		}

	}
	return nil
}

// loadKubeConfig is used to download the kubeconfig to local and store it in the states.
func (s *ClusterStatus) loadKubeConfig(mgr *manager.Manager) error {
	kubeCfgBase64Cmd := "cat /etc/kubernetes/admin.conf | base64 --wrap=0"
	kubeConfigStr, err1 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", kubeCfgBase64Cmd), 1, false)
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to get cluster kubeconfig")
	}

	kubeConfigPath := filepath.Join(mgr.WorkDir, fmt.Sprintf("config-%s", mgr.ObjName))
	kubeconfigStr, err := base64.StdEncoding.DecodeString(kubeConfigStr)
	if err != nil {
		return err
	}

	oldServer := fmt.Sprintf("server: https://%s:%d", mgr.Cluster.ControlPlaneEndpoint.Domain, mgr.Cluster.ControlPlaneEndpoint.Port)
	newServer := fmt.Sprintf("server: https://%s:%d", mgr.Cluster.ControlPlaneEndpoint.Address, mgr.Cluster.ControlPlaneEndpoint.Port)
	newKubeconfigStr := strings.Replace(string(kubeconfigStr), oldServer, newServer, -1)

	if err := ioutil.WriteFile(kubeConfigPath, []byte(newKubeconfigStr), 0644); err != nil {
		return err
	}
	s.kubeconfig = newKubeconfigStr
	return nil
}

// getClusterInfo is used to fetch cluster's nodes info.
func (s *ClusterStatus) getClusterInfo(mgr *manager.Manager) error {
	var allNodesInfo map[string]string
	clusterInfo, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubectl --no-headers=true get nodes -o custom-columns=:metadata.name,:status.nodeInfo.kubeletVersion,:status.addresses\"", 5, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to get cluster info")
	}
	ipv4Regexp, err := regexp.Compile("[\\d]+\\.[\\d]+\\.[\\d]+\\.[\\d]+")
	if err != nil {
		return err
	}
	ipv6Regexp, err := regexp.Compile("[a-f0-9]{1,4}(:[a-f0-9]{1,4}){7}|[a-f0-9]{1,4}(:[a-f0-9]{1,4}){0,7}::[a-f0-9]{0,4}(:[a-f0-9]{1,4}){0,7}")
	if err != nil {
		return err
	}
	tmp := strings.Split(clusterInfo, "\r\n")
	if len(tmp) >= 1 {
		for i := 0; i < len(tmp); i++ {
			if ipv4 := ipv4Regexp.FindStringSubmatch(tmp[i]); len(ipv4) != 0 {
				allNodesInfo[ipv4[0]] = ipv4[0]
			}
			if ipv6 := ipv6Regexp.FindStringSubmatch(tmp[i]); len(ipv6) != 0 {
				allNodesInfo[ipv6[0]] = ipv6[0]
			}
			if len(strings.Fields(tmp[i])) > 3 {
				allNodesInfo[strings.Fields(tmp[i])[0]] = strings.Fields(tmp[i])[1]
			} else {
				allNodesInfo[strings.Fields(tmp[i])[0]] = ""
			}
		}
	}
	s.allNodesInfo = allNodesInfo
	return nil
}

// getJoinInfo is used to fetch parameters (bootstrapToken and certificateKey) information when adding nodes.
func (s *ClusterStatus) getJoinInfo(mgr *manager.Manager) error {
	uploadCertsCmd := "/usr/local/bin/kubeadm init phase upload-certs --config=/etc/kubernetes/kubeadm-config.yaml --upload-certs"
	output, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", uploadCertsCmd), 5, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to upload kubeadm certs")
	}
	reg := regexp.MustCompile("[0-9|a-z]{64}")
	s.certificateKey = reg.FindAllString(output, -1)[0]
	err1 := patchKubeadmSecret(mgr)
	if err1 != nil {
		return err1
	}

	tokenCreateMasterCmd := "/usr/local/bin/kubeadm token create"
	output, err2 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", tokenCreateMasterCmd), 5, true)
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to create kubeadm token")
	}
	reg = regexp.MustCompile("[0-9|a-z]{6}.[0-9|a-z]{16}")
	s.bootstrapToken = reg.FindAllString(output, -1)[0]

	return nil
}

// getKubeConfigForWorker is used to sync kubeconfig to workers' ~/.kube/config .
func (s *ClusterStatus) getKubeConfigForWorker(mgr *manager.Manager) error {
	createConfigDirCmd := "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	chownKubeConfig := "chown $(id -u):$(id -g) -R $HOME/.kube"
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", createConfigDirCmd), 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to create kube dir")
	}
	syncKubeconfigForRootCmd := fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > %s\"", s.kubeconfig, "/root/.kube/config")
	syncKubeconfigForUserCmd := fmt.Sprintf("echo %s | base64 -d > %s && %s", s.kubeconfig, "$HOME/.kube/config", chownKubeConfig)
	if _, err := mgr.Runner.ExecuteCmd(syncKubeconfigForRootCmd, 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to sync kube config")
	}
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", syncKubeconfigForUserCmd), 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to sync kube config")
	}
	return nil
}

// getKubeConfigForControlPlane is used to copy admin.conf to ~/.kube/config .
func GetKubeConfigForControlPlane(mgr *manager.Manager) error {
	createConfigDirCmd := "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	getKubeConfigCmd := "cp -f /etc/kubernetes/admin.conf /root/.kube/config"
	getKubeConfigCmdUsr := "cp -f /etc/kubernetes/admin.conf $HOME/.kube/config"
	chownKubeConfig := "chown $(id -u):$(id -g) $HOME/.kube/config"

	cmd := strings.Join([]string{createConfigDirCmd, getKubeConfigCmd, getKubeConfigCmdUsr, chownKubeConfig}, " && ")
	_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", cmd), 2, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to init kubernetes cluster")
	}
	return nil
}

// removeMasterTaint is used to remove taint when current node both in controlPlane and worker.
func removeMasterTaint(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if node.IsWorker {
		removeMasterTaintCmd := fmt.Sprintf("sudo -E /bin/sh -c \"/usr/local/bin/kubectl taint nodes %s node-role.kubernetes.io/master=:NoSchedule-\"", node.Name)
		_, err := mgr.Runner.ExecuteCmd(removeMasterTaintCmd, 5, true)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "Failed to remove master taint")
		}
	}
	return nil
}

// addWorkerLabel is used to add woker label (node-role.kubernetes.io/worker=) when current node in worker.
func addWorkerLabel(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if node.IsWorker {
		addWorkerLabelCmd := fmt.Sprintf("sudo -E /bin/sh -c \"/usr/local/bin/kubectl label --overwrite node %s node-role.kubernetes.io/worker=\"", node.Name)
		_, _ = mgr.Runner.ExecuteCmd(addWorkerLabelCmd, 5, true)
	}
	return nil
}

// patchKubeadmSecret is used to patch etcd's certs for kubeadm-certs secret.
func patchKubeadmSecret(mgr *manager.Manager) error {
	externalEtcdCerts := []string{"external-etcd-ca.crt", "external-etcd.crt", "external-etcd.key"}
	for _, cert := range externalEtcdCerts {
		_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"/usr/local/bin/kubectl patch -n kube-system secret kubeadm-certs -p '{\\\"data\\\": {\\\"%s\\\": \\\"\\\"}}'\"", cert), 5, true)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "Failed to patch kubeadm secret")
		}
	}
	return nil
}

// generateKubeadmConfig is used to generate kubeadm config for all nodes.
func generateKubeadmConfig(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg, isInitCluster bool, bootstrapToken, certificateKey string) error {
	var kubeadmCfgBase64 string
	if util.IsExist(fmt.Sprintf("%s/kubeadm-config.yaml", mgr.WorkDir)) {
		output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cat %s/kubeadm-config.yaml | base64 --wrap=0", mgr.WorkDir)).CombinedOutput()
		if err != nil {
			fmt.Println(string(output))
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to read custom kubeadm config: %s/kubeadm-config.yaml", mgr.WorkDir))
		}
		kubeadmCfgBase64 = strings.TrimSpace(string(output))
	} else {
		kubeadmCfg, err := v1beta2.GenerateKubeadmCfg(mgr, node, isInitCluster, bootstrapToken, certificateKey)
		if err != nil {
			return err
		}
		kubeadmCfgBase64 = base64.StdEncoding.EncodeToString([]byte(kubeadmCfg))
	}

	_, err1 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/kubernetes && echo %s | base64 -d > /etc/kubernetes/kubeadm-config.yaml\"", kubeadmCfgBase64), 1, false)
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), fmt.Sprintf("Failed to generate kubeadm config for %s", node.Name))
	}

	return nil
}

// AddLabelsForNodes is used to add pre-configured labels.
func AddLabelsForNodes(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	for k, v := range node.Labels {
		addLabelCmd := fmt.Sprintf("sudo -E /bin/sh -c \"/usr/local/bin/kubectl label --overwrite node %s %s=%s\"", node.Name, k, v)
		_, _ = mgr.Runner.ExecuteCmd(addLabelCmd, 5, true)
	}

	return nil
}
