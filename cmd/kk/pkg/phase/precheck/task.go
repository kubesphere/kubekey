/*
 Copyright 2022 The KubeSphere Authors.

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
	"strings"

	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/version/kubesphere"
)

type CheckUpgradeK8sVersion struct {
	common.KubeAction
}

func compareVersionsForUpgrade(currentVersion *versionutil.Version, currentMaxVersion *versionutil.Version, desiredVersion *versionutil.Version) error {
	if desiredVersion.LessThan(currentMaxVersion) {
		return errors.New("desired version to upgrade is less than the max version in cluster")
	}
	if desiredVersion.Major()-currentVersion.Major() != 0 || desiredVersion.Minor()-currentVersion.Minor() > 1 {
		return errors.New("skipping MINOR versions when upgrading is unsupported")
	}
	return nil
}

func (k *CheckUpgradeK8sVersion) Execute(_ connector.Runtime) error {
	minK8sVersion, ok := k.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return errors.New("get current k8s version failed by pipeline cache")
	}
	minVersionObj, err := versionutil.ParseSemantic(minK8sVersion)
	if err != nil {
		return errors.Wrap(err, "parse min k8s version failed")
	}

	maxK8sVersion, ok := k.PipelineCache.GetMustString(common.MaxK8sVersion)
	if !ok {
		return errors.New("get max k8s version failed by pipeline cache")
	}
	maxVersionObj, err := versionutil.ParseSemantic(maxK8sVersion)
	if err != nil {
		return errors.Wrap(err, "parse max k8s version failed")
	}

	desiredVersion, ok := k.PipelineCache.GetMustString(common.DesiredK8sVersion)
	if !ok {
		return errors.New("get desired k8s version failed by pipeline cache")
	}
	desiredVersionObj, err := versionutil.ParseSemantic(desiredVersion)
	if err != nil {
		return errors.Wrap(err, "parse desired k8s version failed")
	}

	if err := compareVersionsForUpgrade(minVersionObj, maxVersionObj, desiredVersionObj); err != nil {
		return err
	}

	k.PipelineCache.Set(common.PlanK8sVersion, desiredVersion)
	return nil
}

type CalculateMaxK8sVersion struct {
	common.KubeAction
}

func (g *CalculateMaxK8sVersion) Execute(runtime connector.Runtime) error {
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

	maxVersion := versionList[0]
	for _, version := range versionList {
		if maxVersion.LessThan(version) {
			maxVersion = version
		}
	}
	g.PipelineCache.Set(common.MaxK8sVersion, fmt.Sprintf("v%s", maxVersion))
	return nil
}

type GetAllNodesK8sVersion struct {
	common.KubeAction
}

func (g *GetAllNodesK8sVersion) Execute(runtime connector.Runtime) error {
	var nodeK8sVersion string
	host := runtime.RemoteHost()
	hostName := host.GetName()

	kubectlCmd := fmt.Sprintf("/usr/local/bin/kubectl get nodes %s", hostName)
	kubectlVersionInfo, err := runtime.GetRunner().SudoCmd(kubectlCmd, false)
	if err != nil {
		errorMsg := fmt.Sprintf("get current version of %s failed", hostName)
		return errors.Wrap(err, errorMsg)
	}

	kubectlCmdGroup := strings.Split(kubectlVersionInfo, " ")
	nodeK8sVersion = kubectlCmdGroup[len(kubectlCmdGroup)-1]

	if host.IsRole(common.Master) {
		apiserverVersion, err := runtime.GetRunner().SudoCmd(
			"cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | rev | cut -d ':' -f1 | rev",
			false)
		if err != nil {
			return errors.Wrap(err, "get current kube-apiserver version failed")
		}
		nodeK8sVersion = apiserverVersion
	}
	host.GetCache().Set(common.NodeK8sVersion, nodeK8sVersion)
	return nil
}

type GetMasterK8sVersion struct {
	common.KubeAction
}

func (g *GetMasterK8sVersion) Execute(runtime connector.Runtime) error {
	apiserverVersion, err := runtime.GetRunner().SudoCmd(
		"cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | rev | cut -d ':' -f1 | rev",
		false)
	if err != nil {
		return errors.Wrap(err, "get current kube-apiserver version failed")
	}
	g.PipelineCache.Set(common.K8sVersion, apiserverVersion)
	return nil
}

type KsPhaseDependencyCheck struct {
	common.KubeAction
}

func (d *KsPhaseDependencyCheck) Execute(_ connector.Runtime) error {
	currentKsVersion, ok := d.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		return errors.New("get current KubeSphere version failed by pipeline cache")
	}
	currentK8sVersion, ok := d.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return errors.New("get current K8s version failed by pipeline cache")
	}
	desiredVersion := d.KubeConf.Cluster.KubeSphere.Version

	if d.KubeConf.Cluster.KubeSphere.Enabled {
		KsInstaller, ok := kubesphere.VersionMap[desiredVersion]
		if !ok {
			return errors.New(fmt.Sprintf("Unsupported version: %s", desiredVersion))
		}
		if currentKsVersion != desiredVersion {
			if ok := KsInstaller.UpgradeSupport(currentKsVersion); !ok {
				return errors.New(fmt.Sprintf("Unsupported upgrade plan: %s to %s", currentKsVersion, desiredVersion))
			}
		}
		if ok := KsInstaller.K8sSupport(currentK8sVersion); !ok {
			return errors.New(fmt.Sprintf("KubeSphere %s does not support running on Kubernetes %s",
				currentK8sVersion, d.KubeConf.Cluster.Kubernetes.Version))
		}
	}

	return nil
}
