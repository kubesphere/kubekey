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
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	kubekeycontroller "github.com/kubesphere/kubekey/controllers/kubekey"
	"github.com/kubesphere/kubekey/pkg/cluster/kubernetes/tmpl"
	"github.com/kubesphere/kubekey/pkg/plugins/dns"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	clusterIsExist  = false
	allNodesName    = map[string]string{}
	allNodesVersion = map[string]string{}
	clusterStatus   = map[string]string{
		"version":       "",
		"joinMasterCmd": "",
		"joinWorkerCmd": "",
		"clusterInfo":   "",
	}
)

func GetClusterStatus(mgr *manager.Manager) error {
	if mgr.InCluster {
		if err := kubekeycontroller.UpdateClusterConditions(mgr, "Init control plane", metav1.Now(), metav1.Now(), false, 4); err != nil {
			return err
		}
	}

	mgr.Logger.Infoln("Get cluster status")

	return mgr.RunTaskOnMasterNodes(getClusterStatus, false)
}

func getClusterStatus(mgr *manager.Manager, _ *kubekeyapiv1alpha1.HostCfg) error {
	if mgr.Runner.Index == 0 {
		if clusterStatus["clusterInfo"] == "" {
			output, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"[ -f /etc/kubernetes/admin.conf ] && echo 'Cluster already exists.' || echo 'Cluster will be created.'\"", 0, true)
			if strings.Contains(output, "Cluster will be created") {
				clusterIsExist = false
			} else {
				if err != nil {
					return errors.Wrap(errors.WithStack(err), "Failed to find /etc/kubernetes/admin.conf")
				} else {
					clusterIsExist = true
					if output, err := mgr.Runner.ExecuteCmd("sudo cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | awk -F '[:]' '{print $(NF-0)}'", 0, true); err != nil {
						return errors.Wrap(errors.WithStack(err), "Failed to find current version")
					} else {
						if !strings.Contains(output, "No such file or directory") {
							clusterStatus["version"] = output
						}
					}
					kubeCfgBase64Cmd := "cat /etc/kubernetes/admin.conf | base64 --wrap=0"
					kubeConfigStr, err1 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", kubeCfgBase64Cmd), 1, false)
					if err1 != nil {
						return errors.Wrap(errors.WithStack(err1), "Failed to get cluster kubeconfig")
					}
					clusterStatus["kubeconfig"] = kubeConfigStr
					if err := loadKubeConfig(mgr); err != nil {
						return err
					}
					if err := getJoinNodesCmd(mgr); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func InitKubernetesCluster(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Initializing kubernetes cluster")

	return mgr.RunTaskOnMasterNodes(initKubernetesCluster, true)
}

func initKubernetesCluster(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if mgr.Runner.Index == 0 && !clusterIsExist {
		var kubeadmCfgBase64 string
		if util.IsExist(fmt.Sprintf("%s/kubeadm-config.yaml", mgr.WorkDir)) {
			output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cat %s/kubeadm-config.yaml | base64 --wrap=0", mgr.WorkDir)).CombinedOutput()
			if err != nil {
				fmt.Println(string(output))
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to read custom kubeadm config: %s/kubeadm-config.yaml", mgr.WorkDir))
			}
			kubeadmCfgBase64 = strings.TrimSpace(string(output))
		} else {
			kubeadmCfg, err := tmpl.GenerateKubeadmCfg(mgr)
			if err != nil {
				return err
			}
			kubeadmCfgBase64 = base64.StdEncoding.EncodeToString([]byte(kubeadmCfg))
		}

		_, err1 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/kubernetes && echo %s | base64 -d > /etc/kubernetes/kubeadm-config.yaml\"", kubeadmCfgBase64), 1, false)
		if err1 != nil {
			return errors.Wrap(errors.WithStack(err1), "Failed to generate kubeadm config")
		}

		for i := 0; i < 3; i++ {
			_, err2 := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubeadm init --config=/etc/kubernetes/kubeadm-config.yaml\"", 0, true)
			if err2 != nil {
				if i == 2 {
					return errors.Wrap(errors.WithStack(err2), "Failed to init kubernetes cluster")
				} else {
					_, _ = mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubeadm reset -f\"", 0, true)
				}
			} else {
				break
			}
		}

		if err := GetKubeConfig(mgr); err != nil {
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
		clusterIsExist = true
		if err := getJoinNodesCmd(mgr); err != nil {
			return err
		}
		if err := loadKubeConfig(mgr); err != nil {
			return err
		}
	}

	return nil
}

func GetKubeConfig(mgr *manager.Manager) error {
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

func addWorkerLabel(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if node.IsWorker {
		addWorkerLabelCmd := fmt.Sprintf("sudo -E /bin/sh -c \"/usr/local/bin/kubectl label --overwrite node %s node-role.kubernetes.io/worker=\"", node.Name)
		_, _ = mgr.Runner.ExecuteCmd(addWorkerLabelCmd, 5, true)
	}
	return nil
}

func getJoinNodesCmd(mgr *manager.Manager) error {
	if err := getJoinCmd(mgr); err != nil {
		return err
	}
	return nil
}

func getJoinCmd(mgr *manager.Manager) error {
	uploadCertsCmd := "/usr/local/bin/kubeadm init phase upload-certs --upload-certs"
	output, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", uploadCertsCmd), 5, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to upload kubeadm certs")
	}
	reg := regexp.MustCompile("[0-9|a-z]{64}")
	certificateKey := reg.FindAllString(output, -1)[0]
	err1 := PatchKubeadmSecret(mgr)
	if err1 != nil {
		return err1
	}

	tokenCreateMasterCmd := "/usr/local/bin/kubeadm token create --print-join-command"
	output, err2 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", tokenCreateMasterCmd), 5, true)
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to get join node cmd")
	}

	joinWorkerStrList := strings.Split(output, "kubeadm join")
	clusterStatus["joinWorkerCmd"] = fmt.Sprintf("/usr/local/bin/kubeadm join %s", joinWorkerStrList[1])
	clusterStatus["joinMasterCmd"] = fmt.Sprintf("%s --control-plane --certificate-key %s", clusterStatus["joinWorkerCmd"], certificateKey)

	output, err3 := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubectl get nodes\"", 5, true)
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "Failed to get cluster info")
	}
	clusterStatus["clusterInfo"] = output
	tmp := strings.Split(clusterStatus["clusterInfo"], "\r\n")
	if len(tmp) > 1 {
		for i := 0; i < len(tmp); i++ {
			allNodesName[strings.Fields(tmp[i])[0]] = strings.Fields(tmp[i])[0]
			if len(strings.Fields(tmp[i])) < 5 {
				allNodesVersion[strings.Fields(tmp[i])[0]] = ""
			} else {
				allNodesVersion[strings.Fields(tmp[i])[0]] = strings.Fields(tmp[i])[4]
			}
		}
	}

	kubeCfgBase64Cmd := "cat /etc/kubernetes/admin.conf | base64 --wrap=0"
	output, err4 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", kubeCfgBase64Cmd), 1, false)
	if err4 != nil {
		return errors.Wrap(errors.WithStack(err4), "Failed to get cluster kubeconfig")
	}
	clusterStatus["kubeconfig"] = output
	return nil
}

func PatchKubeadmSecret(mgr *manager.Manager) error {
	externalEtcdCerts := []string{"external-etcd-ca.crt", "external-etcd.crt", "external-etcd.key"}
	for _, cert := range externalEtcdCerts {
		_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"/usr/local/bin/kubectl patch -n kube-system secret kubeadm-certs -p '{\\\"data\\\": {\\\"%s\\\": \\\"\\\"}}'\"", cert), 5, true)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "Failed to patch kubeadm secret")
		}
	}
	return nil
}

func JoinNodesToCluster(mgr *manager.Manager) error {
	if mgr.InCluster {
		if err := kubekeycontroller.UpdateClusterConditions(mgr, "Join nodes", metav1.Now(), metav1.Now(), false, 5); err != nil {
			return err
		}
	}

	mgr.Logger.Infoln("Joining nodes to cluster")

	if err := mgr.RunTaskOnK8sNodes(joinNodesToCluster, true); err != nil {
		return err
	}

	if mgr.InCluster {
		if err := kubekeycontroller.UpdateClusterConditions(mgr, "Join nodes", mgr.Conditions[4].StartTime, metav1.Now(), true, 5); err != nil {
			return err
		}
	}

	return nil
}

func joinNodesToCluster(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if !ExistNodeName(node.Name) || !ExistNodeVersion(node.Name) {
		if node.IsMaster {
			err := addMaster(mgr)
			if err != nil {
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
			err := addWorker(mgr)
			if err != nil {
				return err
			}
			err1 := addWorkerLabel(mgr, node)
			if err1 != nil {
				return err1
			}
		}
	}
	return nil
}

func addMaster(mgr *manager.Manager) error {
	for i := 0; i < 3; i++ {
		_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", clusterStatus["joinMasterCmd"]), 0, true)
		if err != nil {
			if i == 2 {
				return errors.Wrap(errors.WithStack(err), "Failed to add master to cluster")
			} else {
				_, _ = mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubeadm reset -f\"", 0, true)
			}
		} else {
			break
		}
	}

	if err := GetKubeConfig(mgr); err != nil {
		return err
	}
	return nil
}

func addWorker(mgr *manager.Manager) error {
	for i := 0; i < 3; i++ {
		_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", clusterStatus["joinWorkerCmd"]), 0, true)
		if err != nil {
			if i == 2 {
				return errors.Wrap(errors.WithStack(err), "Failed to add worker to cluster")
			} else {
				_, _ = mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubeadm reset -f\"", 0, true)
			}
		} else {
			break
		}
	}

	createConfigDirCmd := "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	chownKubeConfig := "chown $(id -u):$(id -g) -R $HOME/.kube"
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", createConfigDirCmd), 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to create kube dir")
	}
	syncKubeconfigForRootCmd := fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > %s\"", clusterStatus["kubeconfig"], "/root/.kube/config")
	syncKubeconfigForUserCmd := fmt.Sprintf("echo %s | base64 -d > %s && %s", clusterStatus["kubeconfig"], "$HOME/.kube/config", chownKubeConfig)
	if _, err := mgr.Runner.ExecuteCmd(syncKubeconfigForRootCmd, 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to sync kube config")
	}
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", syncKubeconfigForUserCmd), 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to sync kube config")
	}
	return nil
}

func loadKubeConfig(mgr *manager.Manager) error {
	kubeConfigPath := filepath.Join(mgr.WorkDir, fmt.Sprintf("config-%s", mgr.ObjName))
	kubeconfigStr, err := base64.StdEncoding.DecodeString(clusterStatus["kubeconfig"])
	if err != nil {
		return err
	}

	oldServer := fmt.Sprintf("server: https://%s:%d", mgr.Cluster.ControlPlaneEndpoint.Domain, mgr.Cluster.ControlPlaneEndpoint.Port)
	newServer := fmt.Sprintf("server: https://%s:%d", mgr.Cluster.ControlPlaneEndpoint.Address, mgr.Cluster.ControlPlaneEndpoint.Port)
	newKubeconfigStr := strings.Replace(string(kubeconfigStr), oldServer, newServer, -1)

	if err := ioutil.WriteFile(kubeConfigPath, []byte(newKubeconfigStr), 0644); err != nil {
		return err
	}

	return nil
}
