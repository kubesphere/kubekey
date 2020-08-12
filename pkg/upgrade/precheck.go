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
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"os"
	"strings"
)

var versionCheck = map[string]map[string]map[string]bool{
	"v3.0.0": {
		"k8s": {
			"v1.18": true,
			"v1.17": true,
			"v1.16": true,
			"v1.15": true,
		},
		"ks": {
			"v2.1.0": true,
			"v2.1.1": true,
		},
	},
	"v2.1.1": {
		"k8s": {
			"v1.18": true,
			"v1.17": true,
			"v1.16": true,
			"v1.15": true,
		},
	},
}

func GetClusterInfo(mgr *manager.Manager) error {
	return mgr.RunTaskOnMasterNodes(getClusterInfo, true)
}

func getClusterInfo(mgr *manager.Manager, _ *kubekeyapi.HostCfg) error {
	if mgr.Runner.Index == 0 {

		k8sVersionStr, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | cut -d ':' -f 3\"", 3, false)
		if err != nil {
			return errors.Wrap(err, "Failed to get current kube-apiserver version")
		}

		ksVersion, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubectl get deploy -n  kubesphere-system ks-console -o jsonpath='{.metadata.labels.version}'\"", 2, false)
		if err != nil {
			if mgr.Cluster.KubeSphere.Enabled {
				return err
			}
		} else {
			if mgr.Cluster.KubeSphere.Enabled {
				if _, ok := versionCheck[mgr.Cluster.KubeSphere.Version]; !ok {
					return errors.New(fmt.Sprintf("Unsupported version: %s", mgr.Cluster.KubeSphere.Version))
				}
				if _, ok := versionCheck[mgr.Cluster.KubeSphere.Version]["ks"][ksVersion]; !ok {
					return errors.New(fmt.Sprintf("Unsupported upgrade plan: %s to %s", strings.TrimSpace(ksVersion), mgr.Cluster.KubeSphere.Version))
				}
				K8sTargetVersion := versionutil.MustParseSemantic(mgr.Cluster.Kubernetes.Version)
				if _, ok := versionCheck[mgr.Cluster.KubeSphere.Version]["k8s"][fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())]; !ok {
					return errors.New(fmt.Sprintf("KubeSphere %s does not support running on Kubernetes %s", mgr.Cluster.KubeSphere.Version, fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())))
				}
			} else {
				if _, ok := versionCheck[ksVersion]; !ok {
					return errors.New(fmt.Sprintf("Unsupported version: %s", ksVersion))
				}
				K8sTargetVersion := versionutil.MustParseSemantic(mgr.Cluster.Kubernetes.Version)
				if _, ok := versionCheck[ksVersion]["k8s"][fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())]; !ok {
					return errors.New(fmt.Sprintf("KubeSphere %s does not support running on Kubernetes %s", ksVersion, fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())))
				}
			}
		}

		if err := getNodestatus(mgr); err != nil {
			return err
		}
		if err := getComponentStatus(mgr); err != nil {
			return err
		}

		fmt.Println("Upgrade Confirmation:")
		fmt.Printf("kubernetes version: %s to %s\n", k8sVersionStr, mgr.Cluster.Kubernetes.Version)
		if mgr.Cluster.KubeSphere.Enabled {
			fmt.Printf("kubesphere version: %s to %s\n\n", ksVersion, mgr.Cluster.KubeSphere.Version)
		} else {
			fmt.Println()
		}

		reader := bufio.NewReader(os.Stdin)
	Loop:
		for {
			fmt.Printf("Continue upgrading cluster? [yes/no]: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				mgr.Logger.Fatal(err)
			}
			input = strings.TrimSpace(input)

			switch input {
			case "yes":
				break Loop
			case "no":
				os.Exit(0)
			default:
				continue
			}
		}
	}
	return nil
}

func getComponentStatus(mgr *manager.Manager) error {
	componentStatusStr, err := mgr.Runner.ExecuteCmd("sudo -E /bin/bash -c \"/usr/local/bin/kubectl get componentstatus -o go-template='{{range .items}}{{ printf \\\"%s: \\\" .metadata.name}}{{range .conditions}}{{ printf \\\"%v\\n\\\" .message }}{{end}}{{end}}'\"", 1, false)
	if err != nil {
		return err
	}
	fmt.Println("Components Status:")
	fmt.Println(componentStatusStr + "\n")
	return nil
}

func getNodestatus(mgr *manager.Manager) error {
	nodestatus, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubectl get node\"", 2, false)
	if err != nil {
		return err
	}
	fmt.Println("Cluster nodes status:")
	fmt.Println(nodestatus + "\n")
	return nil
}
