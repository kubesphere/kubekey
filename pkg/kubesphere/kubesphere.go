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

package kubesphere

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
)

var stopChan = make(chan string, 1)

func DeployKubeSphere(mgr *manager.Manager) error {
	configMap := fmt.Sprintf("%s/ks-installer-configmap.yaml", mgr.WorkDir)
	deployment := fmt.Sprintf("%s/ks-installer-deployment.yaml", mgr.WorkDir)
	if util.IsExist(configMap) && util.IsExist(deployment) {
		mgr.Logger.Infoln("Deploying KubeSphere ...")
		if err := mgr.RunTaskOnMasterNodes(deployKubeSphere, true); err != nil {
			return err
		}
		ResultNotes()
	}

	return nil
}

func deployKubeSphere(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	if mgr.Runner.Index == 0 {
		if err := DeployKubeSphereStep(mgr); err != nil {
			return err
		}
	}

	return nil
}

func DeployKubeSphereStep(mgr *manager.Manager) error {
	configMap := fmt.Sprintf("%s/ks-installer-configmap.yaml", mgr.WorkDir)
	deployment := fmt.Sprintf("%s/ks-installer-deployment.yaml", mgr.WorkDir)

	if util.IsExist(configMap) && util.IsExist(deployment) {
		result := make(map[string]interface{})
		content, err := ioutil.ReadFile(configMap)
		if err != nil {
			return errors.Wrap(err, "Unable to read the given configmap")
		}
		if err := yaml.Unmarshal(content, &result); err != nil {
			return errors.Wrap(err, "Unable to convert file to yaml")
		}
		metadata := result["metadata"].(map[interface{}]interface{})
		labels := metadata["labels"].(map[interface{}]interface{})
		_, ok := labels["version"]
		if ok {
			switch labels["version"] {
			case "v2.1.1":
				err := mgr.Runner.ScpFile(fmt.Sprintf("%s/%s/%s", mgr.WorkDir, mgr.Cluster.Kubernetes.Version, "helm2"), fmt.Sprintf("%s/%s", "/tmp/kubekey", "helm2"))
				if err != nil {
					return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to sync helm2"))
				}
				_, err1 := mgr.Runner.RunCmdOutput(fmt.Sprintf("sudo -E /bin/sh -c \"cp /tmp/kubekey/helm2  /usr/local/bin/helm2  && chmod +x /usr/local/bin/helm2\""))
				if err1 != nil {
					return errors.Wrap(errors.WithStack(err1), fmt.Sprintf("Failed to sync helm2"))
				}
				_, err2 := mgr.Runner.RunCmdOutput(`cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tiller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tiller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: tiller
    namespace: kube-system
EOF
`)
				if err2 != nil {
					return errors.Wrap(errors.WithStack(err2), fmt.Sprintf("Failed to create helm rbac"))
				}
				var tillerRepo string
				if mgr.Cluster.Registry.PrivateRegistry != "" {
					tillerRepo = fmt.Sprintf("%s/kubesphere", mgr.Cluster.Registry.PrivateRegistry)
				} else {
					tillerRepo = "kubesphere"
				}
				_, err3 := mgr.Runner.RunCmdOutput(fmt.Sprintf("sudo -E /bin/sh -c \"/usr/local/bin/helm2 init --service-account=tiller --skip-refresh --tiller-image=%s/tiller:v2.16.2\"", tillerRepo))
				if err3 != nil {
					return errors.Wrap(errors.WithStack(err3), fmt.Sprintf("Failed to sync helm2"))
				}

			}
		}
	}

	if mgr.Cluster.Registry.PrivateRegistry != "" {
		configMapBase64, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sed -i \"/local_registry/s/\\:.*/\\: %s/g\" %s", mgr.Cluster.Registry.PrivateRegistry, configMap)).CombinedOutput()
		if err != nil {
			fmt.Println(string(configMapBase64))
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to add private registry: %s", mgr.Cluster.Registry.PrivateRegistry))
		}
	}

	configMapBase64, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cat %s | base64 --wrap=0", configMap)).CombinedOutput()
	if err != nil {
		fmt.Println(string(configMapBase64))
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to read kubesphere configmap: %s", configMap))
	}
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/addons/kubesphere.yaml\"", strings.TrimSpace(string(configMapBase64))))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate kubesphere manifests")
	}

	deployBase64, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cat %s | base64 --wrap=0", deployment)).CombinedOutput()
	if err != nil {
		fmt.Println(string(deployBase64))
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to read kubesphere deployment: %s", deployment))
	}
	_, err2 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d >> /etc/kubernetes/addons/kubesphere.yaml\"", strings.TrimSpace(string(deployBase64))))
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate kubesphere manifests")
	}

	_, err3 := mgr.Runner.RunCmdOutput(`cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: kubesphere-system
EOF
`)
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "Failed to create namespace: kubesphere-system")
	}
	_, err4 := mgr.Runner.RunCmdOutput("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/kubesphere.yaml")
	if err4 != nil {
		return errors.Wrap(errors.WithStack(err4), "Failed to deploy /etc/kubernetes/addons/kubesphere.yaml")
	}
	go CheckKubeSphereStatus(mgr)
	return nil
}

func CheckKubeSphereStatus(mgr *manager.Manager) {
	for i := 30; i > 0; i-- {
		time.Sleep(10 * time.Second)
		_, err := mgr.Runner.RunCmd(
			"/usr/local/bin/kubectl exec -n kubesphere-system $(kubectl get pod -n kubesphere-system -l app=ks-install -o jsonpath='{.items[0].metadata.name}') ls kubesphere/playbooks/kubesphere_running",
		)
		if err == nil {
			output, err := mgr.Runner.RunCmd(
				"/usr/local/bin/kubectl exec -n kubesphere-system $(kubectl get pod -n kubesphere-system -l app=ks-install -o jsonpath='{.items[0].metadata.name}') cat kubesphere/playbooks/kubesphere_running",
			)
			if err == nil && output != "" {
				stopChan <- output
				break
			}
		}
	}
	stopChan <- ""
}

func ResultNotes() {
	var (
		position = 1
		notes    = "Please wait for the installation to complete: "
	)
	fmt.Println("\n")
Loop:
	for {
		select {
		case result := <-stopChan:
			fmt.Printf("\033[%dA\033[K", position)
			fmt.Println(result)
			break Loop
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
}
