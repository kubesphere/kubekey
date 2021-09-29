package precheck

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"regexp"
	"strings"
)

type NodePreCheck struct {
	common.KubeAction
}

func (n *NodePreCheck) Execute(runtime connector.Runtime) error {
	var results = make(map[string]string)
	results["name"] = runtime.RemoteHost().GetName()
	for _, software := range baseSoftware {
		_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("which %s", software), false)
		switch software {
		case showmount:
			software = nfs
		case rbd:
			software = ceph
		case glusterfs:
			software = glusterfs
		}
		if err != nil {
			results[software] = ""
			logger.Log.Debugf("exec cmd 'which %s' got err return: %v", software, err)
		} else {
			results[software] = "y"
			if software == docker {
				dockerVersion, err := runtime.GetRunner().SudoCmd("docker version --format '{{.Server.Version}}'", false)
				if err != nil {
					results[software] = UnknownVersion
				} else {
					results[software] = dockerVersion
				}
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
					if _, err := runtime.GetRunner().SudoCmd(
						"mkdir -p $HOME/.kube "+
							"&& cp /etc/kubernetes/admin.conf $HOME/.kube/config "+
							"&& chown $(id -u):$(id -g) -R $HOME/.kube", false); err != nil {
						return err
					}
				}
			}
		}
	}
	return errors.New("kube config not found")
}

type K8sVersionCheck struct {
	common.KubeAction
}

func (k *K8sVersionCheck) Execute(runtime connector.Runtime) error {
	k8sVersionStr, err := runtime.GetRunner().SudoCmd(
		"cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | rev | cut -d ':' -f1 | rev",
		false)
	if err != nil {
		return errors.Wrap(err, "get current kube-apiserver version failed")
	}
	k.PipelineCache.Set(common.K8sVersion, k8sVersionStr)
	return nil
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

func (k *CheckDesiredK8sVersion) Execute(runtime connector.Runtime) error {
	K8sTargetVersion := versionutil.MustParseSemantic(k.KubeConf.Cluster.Kubernetes.Version)
	if _, ok := matrix["other"]["k8s"][fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())]; !ok {
		return errors.New(fmt.Sprintf("does not support upgrade to Kubernetes %s",
			fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())))
	}
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
		if k.KubeConf.Arg.KsEnable {
			return errors.Wrap(err, "get kubeSphere version failed")
		} else {
			ksVersionStr = ""
		}
	}
	k.PipelineCache.Set(common.KubeSphereVersion, ksVersionStr)
	return nil
}

type DependencyCheck struct {
	common.KubeAction
}

func (d *DependencyCheck) Execute(runtime connector.Runtime) error {
	currentKsVersion, ok := d.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		return errors.New("get current KubeSphere version failed by pipeline cache")
	}
	desiredVersion := d.KubeConf.Cluster.KubeSphere.Version

	if d.KubeConf.Arg.KsEnable {
		var version string
		if strings.Contains(desiredVersion, "latest") ||
			strings.Contains(desiredVersion, "nightly") {
			version = Latest
		} else {
			r := regexp.MustCompile("v(\\d+\\.)?(\\d+\\.)?(\\*|\\d+)")
			version = r.FindString(desiredVersion)
		}

		if _, ok := matrix[version]; !ok {
			return errors.New(fmt.Sprintf("Unsupported version: %s", desiredVersion))
		}
		if _, ok := matrix[version]["ks"][currentKsVersion]; !ok {
			return errors.New(fmt.Sprintf("Unsupported upgrade plan: %s to %s", currentKsVersion, desiredVersion))
		}
		K8sTargetVersion := versionutil.MustParseSemantic(d.KubeConf.Cluster.Kubernetes.Version)
		if _, ok := matrix[version]["k8s"][fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())]; !ok {
			return errors.New(fmt.Sprintf("KubeSphere %s does not support running on Kubernetes %s",
				version, fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())))
		}
	} else {
		K8sTargetVersion := versionutil.MustParseSemantic(d.KubeConf.Cluster.Kubernetes.Version)
		if _, ok := matrix[currentKsVersion]["k8s"][fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())]; !ok {
			return errors.New(fmt.Sprintf("KubeSphere %s does not support running on Kubernetes %s",
				currentKsVersion, fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor())))
		}
	}
	return nil
}

type GetKubernetesNodesStatus struct {
	common.KubeAction
}

func (g *GetKubernetesNodesStatus) Execute(runtime connector.Runtime) error {
	nodeStatus, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get node", false)
	if err != nil {
		return err
	}

	g.PipelineCache.Set(common.ClusterNodeStatus, nodeStatus)
	return nil
}
