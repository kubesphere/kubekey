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

package precheck

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/version/kubernetes"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/version/kubesphere"
)

type GreetingsTask struct {
	action.BaseAction
}

func (h *GreetingsTask) Execute(runtime connector.Runtime) error {
	hello, err := runtime.GetRunner().SudoCmd("echo 'Greetings, KubeKey!'", false)
	if err != nil {
		return err
	}
	logger.Log.Messagef(runtime.RemoteHost().GetName(), hello)
	return nil
}

type NodePreCheck struct {
	common.KubeAction
}

func (n *NodePreCheck) Execute(runtime connector.Runtime) error {
	var results = make(map[string]string)
	results["name"] = runtime.RemoteHost().GetName()
	for _, software := range baseSoftware {
		var (
			cmd string
		)

		switch software {
		case docker:
			cmd = "docker version --format '{{.Server.Version}}'"
		case containerd:
			cmd = "containerd --version | cut -d ' ' -f 3"
		default:
			cmd = fmt.Sprintf("which %s", software)
		}

		switch software {
		case sudo:
			// sudo skip sudo prefix
		default:
			cmd = connector.SudoPrefix(cmd)
		}

		res, err := runtime.GetRunner().Cmd(cmd, false)
		switch software {
		case showmount:
			software = nfs
		case rbd:
			software = ceph
		case glusterfs:
			software = glusterfs
		}
		if err != nil || strings.Contains(res, "not found") {
			results[software] = ""
		} else {
			// software in path
			if strings.Contains(res, "bin/") {
				results[software] = "y"
			} else {
				// get software version, e.g. docker, containerd, etc.
				results[software] = res
			}
		}
	}

	output, err := runtime.GetRunner().Cmd("date +\"%Z %H:%M:%S\"", false)
	if err != nil {
		results["time"] = ""
	} else {
		results["time"] = strings.TrimSpace(output)
	}

	host := runtime.RemoteHost()
	if res, ok := host.GetCache().Get(common.NodePreCheck); ok {
		m := res.(map[string]string)
		m = results
		host.GetCache().Set(common.NodePreCheck, m)
	} else {
		host.GetCache().Set(common.NodePreCheck, results)
	}
	return nil
}

type GetKubeConfig struct {
	common.KubeAction
}

func (g *GetKubeConfig) Execute(runtime connector.Runtime) error {
	if exist, err := runtime.GetRunner().FileExist("$HOME/.kube/config"); err != nil {
		return err
	} else {
		if exist {
			return nil
		} else {
			if exist, err := runtime.GetRunner().FileExist("/etc/kubernetes/admin.conf"); err != nil {
				return err
			} else {
				if exist {
					if _, err := runtime.GetRunner().Cmd("mkdir -p $HOME/.kube", false); err != nil {
						return err
					}
					if _, err := runtime.GetRunner().SudoCmd("cp /etc/kubernetes/admin.conf $HOME/.kube/config", false); err != nil {
						return err
					}
					userId, err := runtime.GetRunner().Cmd("echo $(id -u)", false)
					if err != nil {
						return errors.Wrap(errors.WithStack(err), "get user id failed")
					}

					userGroupId, err := runtime.GetRunner().Cmd("echo $(id -g)", false)
					if err != nil {
						return errors.Wrap(errors.WithStack(err), "get user group id failed")
					}

					chownKubeConfig := fmt.Sprintf("chown -R %s:%s $HOME/.kube", userId, userGroupId)
					if _, err := runtime.GetRunner().SudoCmd(chownKubeConfig, false); err != nil {
						return errors.Wrap(errors.WithStack(err), "chown user kube config failed")
					}
				}
			}
		}
	}
	return errors.New("kube config not found")
}

type GetAllNodesK8sVersion struct {
	common.KubeAction
}

func (g *GetAllNodesK8sVersion) Execute(runtime connector.Runtime) error {
	var nodeK8sVersion string
	kubeletVersionInfo, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubelet --version", false)
	if err != nil {
		return errors.Wrap(err, "get current kubelet version failed")
	}
	nodeK8sVersion = strings.Split(kubeletVersionInfo, " ")[1]

	host := runtime.RemoteHost()
	host.GetCache().Set(common.NodeK8sVersion, nodeK8sVersion)

	if host.IsRole(common.Master) {
		apiserverVersion, err := runtime.GetRunner().SudoCmd(
			"cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | rev | cut -d ':' -f1 | rev",
			false)
		if err != nil {
			return errors.Wrap(err, "get current kube-apiserver version failed")
		}

		apiserverSemanticVersion, err := versionutil.ParseSemantic(apiserverVersion)
		if err != nil {
			return errors.Wrap(err, "parse kube-apiserver version failed")
		}

		kubeletSemanticVersion, err := versionutil.ParseSemantic(nodeK8sVersion)
		if err != nil {
			return errors.Wrap(err, "parse kubelet version failed")
		}

		if apiserverSemanticVersion.LessThan(kubeletSemanticVersion) {
			host.GetCache().Set(common.NodeK8sVersion, apiserverVersion)
		}
	}

	return nil
}

type CalculateMinK8sVersion struct {
	common.KubeAction
}

func (g *CalculateMinK8sVersion) Execute(runtime connector.Runtime) error {
	versionList := make([]*versionutil.Version, 0, len(runtime.GetHostsByRole(common.K8s)))
	for _, host := range runtime.GetHostsByRole(common.K8s) {
		version, ok := host.GetCache().GetMustString(common.NodeK8sVersion)
		if !ok {
			return errors.Errorf("get node %s Kubernetes version failed by host cache", host.GetName())
		}
		if versionObj, err := versionutil.ParseSemantic(version); err != nil {
			return errors.Wrap(err, "parse node version failed")
		} else {
			versionList = append(versionList, versionObj)
		}
	}

	minVersion := versionList[0]
	for _, version := range versionList {
		if !minVersion.LessThan(version) {
			minVersion = version
		}
	}
	g.PipelineCache.Set(common.K8sVersion, fmt.Sprintf("v%s", minVersion))
	return nil
}

type CheckDesiredK8sVersion struct {
	common.KubeAction
}

func (k *CheckDesiredK8sVersion) Execute(_ connector.Runtime) error {
	if ok := kubernetes.VersionSupport(k.KubeConf.Cluster.Kubernetes.Version); !ok {
		return errors.New(fmt.Sprintf("does not support upgrade to Kubernetes %s",
			k.KubeConf.Cluster.Kubernetes.Version))
	}
	k.PipelineCache.Set(common.DesiredK8sVersion, k.KubeConf.Cluster.Kubernetes.Version)
	return nil
}

type KsVersionCheck struct {
	common.KubeAction
}

func (k *KsVersionCheck) Execute(runtime connector.Runtime) error {
	ksVersionStr, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl get deploy -n  kubesphere-system ks-console -o jsonpath='{.metadata.labels.version}'",
		false)
	if err != nil {
		if k.KubeConf.Cluster.KubeSphere.Enabled {
			return errors.Wrap(err, "get kubeSphere version failed")
		} else {
			ksVersionStr = ""
		}
	}

	ccKsVersionStr, ccErr := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl get ClusterConfiguration ks-installer -n  kubesphere-system  -o jsonpath='{.metadata.labels.version}'",
		false)

	if ccErr == nil && versionutil.MustParseSemantic(ccKsVersionStr).AtLeast(versionutil.MustParseSemantic("v3.1.0")) {
		ksVersionStr = ccKsVersionStr
	}

	k.PipelineCache.Set(common.KubeSphereVersion, ksVersionStr)
	return nil
}

type DependencyCheck struct {
	common.KubeAction
}

func (d *DependencyCheck) Execute(_ connector.Runtime) error {
	currentKsVersion, ok := d.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		return errors.New("get current KubeSphere version failed by pipeline cache")
	}
	desiredVersion := d.KubeConf.Cluster.KubeSphere.Version

	if d.KubeConf.Cluster.KubeSphere.Enabled {
		var version string
		if latest, ok := kubesphere.LatestRelease(desiredVersion); ok {
			version = latest.Version
		} else if ks, ok := kubesphere.DevRelease(desiredVersion); ok {
			version = ks.Version
		} else {
			r := regexp.MustCompile("v(\\d+\\.)?(\\d+\\.)?(\\*|\\d+)")
			version = r.FindString(desiredVersion)
		}

		ksInstaller, ok := kubesphere.VersionMap[version]
		if !ok {
			return errors.New(fmt.Sprintf("Unsupported version: %s", desiredVersion))
		}

		if currentKsVersion != desiredVersion {
			if ok := ksInstaller.UpgradeSupport(currentKsVersion); !ok {
				return errors.New(fmt.Sprintf("Unsupported upgrade plan: %s to %s", currentKsVersion, desiredVersion))
			}
		}

		if ok := ksInstaller.K8sSupport(d.KubeConf.Cluster.Kubernetes.Version); !ok {
			return errors.New(fmt.Sprintf("KubeSphere %s does not support running on Kubernetes %s",
				version, d.KubeConf.Cluster.Kubernetes.Version))
		}
	} else {
		ksInstaller, ok := kubesphere.VersionMap[currentKsVersion]
		if !ok {
			return errors.New(fmt.Sprintf("Unsupported version: %s", currentKsVersion))
		}

		if ok := ksInstaller.K8sSupport(d.KubeConf.Cluster.Kubernetes.Version); !ok {
			return errors.New(fmt.Sprintf("KubeSphere %s does not support running on Kubernetes %s",
				currentKsVersion, d.KubeConf.Cluster.Kubernetes.Version))
		}
	}
	return nil
}

type GetKubernetesNodesStatus struct {
	common.KubeAction
}

func (g *GetKubernetesNodesStatus) Execute(runtime connector.Runtime) error {
	nodeStatus, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get node -o wide", false)
	if err != nil {
		return err
	}
	g.PipelineCache.Set(common.ClusterNodeStatus, nodeStatus)

	cri, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get node -o jsonpath=\"{.items[*].status.nodeInfo.containerRuntimeVersion}\"", false)
	if err != nil {
		return err
	}
	g.PipelineCache.Set(common.ClusterNodeCRIRuntimes, cri)
	return nil
}
