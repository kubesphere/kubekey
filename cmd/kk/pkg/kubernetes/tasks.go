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

package kubernetes

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	kubekeyv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/task"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/etcd"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/files"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/images"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/kubernetes/templates"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/utils"
)

type GetClusterStatus struct {
	common.KubeAction
}

func (g *GetClusterStatus) Execute(runtime connector.Runtime) error {
	exist, err := runtime.GetRunner().FileExist("/etc/kubernetes/admin.conf")
	if err != nil {
		return err
	}

	if !exist {
		g.PipelineCache.Set(common.ClusterExist, false)
		return nil
	} else {
		g.PipelineCache.Set(common.ClusterExist, true)

		if v, ok := g.PipelineCache.Get(common.ClusterStatus); ok {
			cluster := v.(*KubernetesStatus)
			if err := cluster.SearchVersion(runtime); err != nil {
				return err
			}
			if err := cluster.SearchKubeConfig(runtime); err != nil {
				return err
			}
			if err := cluster.LoadKubeConfig(runtime, g.KubeConf); err != nil {
				return err
			}
			if err := cluster.SearchClusterInfo(runtime); err != nil {
				return err
			}
			if err := cluster.SearchNodesInfo(runtime); err != nil {
				return err
			}
			if err := cluster.SearchJoinInfo(runtime); err != nil {
				return err
			}

			g.PipelineCache.Set(common.ClusterStatus, cluster)
		} else {
			return errors.New("get kubernetes cluster status by pipeline cache failed")
		}
	}
	return nil
}

type SyncKubeBinary struct {
	common.KubeAction
}

func (i *SyncKubeBinary) Execute(runtime connector.Runtime) error {
	binariesMapObj, ok := i.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	if err := SyncKubeBinaries(i, runtime, binariesMap); err != nil {
		return err
	}
	return nil
}

// SyncKubeBinaries is used to sync kubernetes' binaries to each node.
func SyncKubeBinaries(i *SyncKubeBinary, runtime connector.Runtime, binariesMap map[string]*files.KubeBinary) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binaryList := []string{"kubeadm", "kubelet", "kubectl", "helm", "kubecni"}
	if i.KubeConf.Cluster.Network.Plugin == "calico" {
		binaryList = append(binaryList, "calicoctl")
	}
	for _, name := range binaryList {
		binary, ok := binariesMap[name]
		if !ok {
			return fmt.Errorf("get kube binary %s info failed: no such key", name)
		}

		fileName := binary.FileName
		switch name {
		//case "kubelet":
		//	if err := runtime.GetRunner().Scp(binary.Path, fmt.Sprintf("%s/%s", common.TmpDir, binary.Name)); err != nil {
		//		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
		//	}
		case "kubecni":
			dst := filepath.Join(common.TmpDir, fileName)
			if err := runtime.GetRunner().Scp(binary.Path(), dst); err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
			}
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar -zxf %s -C /opt/cni/bin", dst), false); err != nil {
				return err
			}
		default:
			dst := filepath.Join(common.BinDir, fileName)
			if err := runtime.GetRunner().SudoScp(binary.Path(), dst); err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
			}
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", dst), false); err != nil {
				return err
			}
		}
	}
	return nil
}

type ChmodKubelet struct {
	common.KubeAction
}

func (c *ChmodKubelet) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("chmod +x /usr/local/bin/kubelet", false); err != nil {
		return errors.Wrap(errors.WithStack(err), "change kubelet mode failed")
	}
	return nil
}

type EnableKubelet struct {
	common.KubeAction
}

func (e *EnableKubelet) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl disable kubelet "+
		"&& systemctl enable kubelet "+
		"&& ln -snf /usr/local/bin/kubelet /usr/bin/kubelet", false); err != nil {
		return errors.Wrap(errors.WithStack(err), "enable kubelet service failed")
	}
	return nil
}

type GenerateKubeletEnv struct {
	common.KubeAction
}

func (g *GenerateKubeletEnv) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	templateAction := action.Template{
		Template: templates.KubeletEnv,
		Dst:      filepath.Join("/etc/systemd/system/kubelet.service.d", templates.KubeletEnv.Name()),
		Data: util.Data{
			"NodeIP":           host.GetInternalAddress(),
			"Hostname":         host.GetName(),
			"ContainerRuntime": "",
			"KubeletArgs":      g.KubeConf.Cluster.Kubernetes.KubeletArgs,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type GenerateKubeadmConfig struct {
	common.KubeAction
	IsInitConfiguration     bool
	WithSecurityEnhancement bool
}

func (g *GenerateKubeadmConfig) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	localConfig := filepath.Join(runtime.GetWorkDir(), "kubeadm-config.yaml")
	if util.IsExist(localConfig) {
		// todo: if it is necessary?
		if err := runtime.GetRunner().SudoScp(localConfig, "/etc/kubernetes/kubeadm-config.yaml"); err != nil {
			return errors.Wrap(errors.WithStack(err), "scp local kubeadm config failed")
		}
	} else {
		// generate etcd configuration
		var externalEtcd kubekeyv1alpha2.ExternalEtcd
		var endpointsList, etcdCertSANs []string

		switch g.KubeConf.Cluster.Etcd.Type {
		case kubekeyv1alpha2.KubeKey:
			for _, host := range runtime.GetHostsByRole(common.ETCD) {
				endpoint := fmt.Sprintf("https://%s:%s", host.GetInternalIPv4Address(), kubekeyv1alpha2.DefaultEtcdPort)
				endpointsList = append(endpointsList, endpoint)
			}
			externalEtcd.Endpoints = endpointsList

			externalEtcd.CAFile = "/etc/ssl/etcd/ssl/ca.pem"
			externalEtcd.CertFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", host.GetName())
			externalEtcd.KeyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", host.GetName())
		case kubekeyv1alpha2.External:
			externalEtcd.Endpoints = g.KubeConf.Cluster.Etcd.External.Endpoints

			if len(g.KubeConf.Cluster.Etcd.External.CAFile) != 0 && len(g.KubeConf.Cluster.Etcd.External.CAFile) != 0 && len(g.KubeConf.Cluster.Etcd.External.CAFile) != 0 {
				externalEtcd.CAFile = fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(g.KubeConf.Cluster.Etcd.External.CAFile))
				externalEtcd.CertFile = fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(g.KubeConf.Cluster.Etcd.External.CertFile))
				externalEtcd.KeyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/%s", filepath.Base(g.KubeConf.Cluster.Etcd.External.KeyFile))
			}
		case kubekeyv1alpha2.Kubeadm:
			altNames := etcd.GenerateAltName(g.KubeConf, &runtime)
			etcdCertSANs = append(etcdCertSANs, altNames.DNSNames...)
			for _, ip := range altNames.IPs {
				etcdCertSANs = append(etcdCertSANs, string(ip))
			}
		}

		_, ApiServerArgs := util.GetArgs(templates.GetApiServerArgs(g.WithSecurityEnhancement, g.KubeConf.Cluster.Kubernetes.EnableAudit()), g.KubeConf.Cluster.Kubernetes.ApiServerArgs)
		_, ControllerManagerArgs := util.GetArgs(templates.GetControllermanagerArgs(g.KubeConf.Cluster.Kubernetes.Version, g.WithSecurityEnhancement), g.KubeConf.Cluster.Kubernetes.ControllerManagerArgs)
		_, SchedulerArgs := util.GetArgs(templates.GetSchedulerArgs(g.WithSecurityEnhancement), g.KubeConf.Cluster.Kubernetes.SchedulerArgs)

		checkCgroupDriver, err := templates.GetKubeletCgroupDriver(runtime, g.KubeConf)
		if err != nil {
			return err
		}

		var (
			bootstrapToken, certificateKey string
			// todo: if port needed
		)
		if !g.IsInitConfiguration {
			if v, ok := g.PipelineCache.Get(common.ClusterStatus); ok {
				cluster := v.(*KubernetesStatus)
				bootstrapToken = cluster.BootstrapToken
				certificateKey = cluster.CertificateKey
			} else {
				return errors.New("get kubernetes cluster status by pipeline cache failed")
			}
		}

		templateAction := action.Template{
			Template: templates.KubeadmConfig,
			Dst:      filepath.Join(common.KubeConfigDir, templates.KubeadmConfig.Name()),
			Data: util.Data{
				"IsInitCluster":          g.IsInitConfiguration,
				"ImageRepo":              strings.TrimSuffix(images.GetImage(runtime, g.KubeConf, "kube-apiserver").ImageRepo(), "/kube-apiserver"),
				"EtcdTypeIsKubeadm":      g.KubeConf.Cluster.Etcd.Type == kubekeyv1alpha2.Kubeadm,
				"EtcdCertSANs":           etcdCertSANs,
				"EtcdRepo":               strings.TrimSuffix(images.GetImage(runtime, g.KubeConf, "etcd").ImageRepo(), "/etcd"),
				"EtcdTag":                images.GetImage(runtime, g.KubeConf, "etcd").Tag,
				"CorednsRepo":            strings.TrimSuffix(images.GetImage(runtime, g.KubeConf, "coredns").ImageRepo(), "/coredns"),
				"CorednsTag":             images.GetImage(runtime, g.KubeConf, "coredns").Tag,
				"Version":                g.KubeConf.Cluster.Kubernetes.Version,
				"ClusterName":            g.KubeConf.Cluster.Kubernetes.ClusterName,
				"DNSDomain":              g.KubeConf.Cluster.Kubernetes.DNSDomain,
				"AdvertiseAddress":       host.GetInternalIPv4Address(),
				"BindPort":               kubekeyv1alpha2.DefaultApiserverPort,
				"ControlPlaneEndpoint":   fmt.Sprintf("%s:%d", g.KubeConf.Cluster.ControlPlaneEndpoint.Domain, g.KubeConf.Cluster.ControlPlaneEndpoint.Port),
				"PodSubnet":              g.KubeConf.Cluster.Network.KubePodsCIDR,
				"ServiceSubnet":          g.KubeConf.Cluster.Network.KubeServiceCIDR,
				"CertSANs":               g.KubeConf.Cluster.GenerateCertSANs(),
				"ExternalEtcd":           externalEtcd,
				"NodeCidrMaskSize":       g.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
				"CriSock":                g.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint,
				"ApiServerArgs":          templates.UpdateFeatureGatesConfiguration(ApiServerArgs, g.KubeConf),
				"EnableAudit":            g.KubeConf.Cluster.Kubernetes.EnableAudit(),
				"ControllerManagerArgs":  templates.UpdateFeatureGatesConfiguration(ControllerManagerArgs, g.KubeConf),
				"SchedulerArgs":          templates.UpdateFeatureGatesConfiguration(SchedulerArgs, g.KubeConf),
				"KubeletConfiguration":   templates.GetKubeletConfiguration(runtime, g.KubeConf, g.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint, g.WithSecurityEnhancement),
				"KubeProxyConfiguration": templates.GetKubeProxyConfiguration(g.KubeConf),
				"IsV1beta3":              versionutil.MustParseSemantic(g.KubeConf.Cluster.Kubernetes.Version).AtLeast(versionutil.MustParseSemantic("v1.22.0")),
				"IsControlPlane":         host.IsRole(common.Master),
				"CgroupDriver":           checkCgroupDriver,
				"BootstrapToken":         bootstrapToken,
				"CertificateKey":         certificateKey,
				"IPv6Support":            host.GetInternalIPv6Address() != "",
			},
		}

		templateAction.Init(nil, nil)
		if err := templateAction.Execute(runtime); err != nil {
			return err
		}
	}
	return nil
}

type KubeadmInit struct {
	common.KubeAction
}

func (k *KubeadmInit) Execute(runtime connector.Runtime) error {
	initCmd := "/usr/local/bin/kubeadm init --config=/etc/kubernetes/kubeadm-config.yaml --ignore-preflight-errors=FileExisting-crictl,ImagePull"

	if k.KubeConf.Cluster.Kubernetes.DisableKubeProxy {
		initCmd = initCmd + " --skip-phases=addon/kube-proxy"
	}

	if _, err := runtime.GetRunner().SudoCmd(initCmd, true); err != nil {
		// kubeadm reset and then retry
		resetCmd := "/usr/local/bin/kubeadm reset -f"
		if k.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint != "" {
			resetCmd = resetCmd + " --cri-socket " + k.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint
		}
		_, _ = runtime.GetRunner().SudoCmd(resetCmd, true)
		return errors.Wrap(errors.WithStack(err), "init kubernetes cluster failed")
	}
	return nil
}

type CopyKubeConfigForControlPlane struct {
	common.KubeAction
}

func (c *CopyKubeConfigForControlPlane) Execute(runtime connector.Runtime) error {
	createConfigDirCmd := "mkdir -p /root/.kube"
	getKubeConfigCmd := "cp -f /etc/kubernetes/admin.conf /root/.kube/config"
	cmd := strings.Join([]string{createConfigDirCmd, getKubeConfigCmd}, " && ")
	if _, err := runtime.GetRunner().SudoCmd(cmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "copy kube config failed")
	}

	userMkdir := "mkdir -p $HOME/.kube"
	if _, err := runtime.GetRunner().Cmd(userMkdir, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "user mkdir $HOME/.kube failed")
	}

	userCopyKubeConfig := "cp -f /etc/kubernetes/admin.conf $HOME/.kube/config"
	if _, err := runtime.GetRunner().SudoCmd(userCopyKubeConfig, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "user copy /etc/kubernetes/admin.conf to $HOME/.kube/config failed")
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
	return nil
}

type RemoveMasterTaint struct {
	common.KubeAction
}

func (r *RemoveMasterTaint) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl taint nodes %s node-role.kubernetes.io/master=:NoSchedule-",
		runtime.RemoteHost().GetName()), true); err != nil {
		logger.Log.Warning(err.Error())
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl taint nodes %s node-role.kubernetes.io/control-plane=:NoSchedule-",
		runtime.RemoteHost().GetName()), true); err != nil {
		logger.Log.Warningf(err.Error())
	}
	return nil
}

type AddWorkerLabel struct {
	common.KubeAction
}

func (a *AddWorkerLabel) Execute(runtime connector.Runtime) error {
	for _, host := range runtime.GetAllHosts() {
		if host.IsRole(common.Worker) {
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
				"/usr/local/bin/kubectl label --overwrite node %s node-role.kubernetes.io/worker=",
				host.GetName()), true); err != nil {
				return errors.Wrap(errors.WithStack(err), "add worker label failed")
			}
		}
	}

	return nil
}

type JoinNode struct {
	common.KubeAction
}

func (j *JoinNode) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubeadm join --config=/etc/kubernetes/kubeadm-config.yaml --ignore-preflight-errors=FileExisting-crictl,ImagePull",
		true); err != nil {
		resetCmd := "/usr/local/bin/kubeadm reset -f"
		if j.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint != "" {
			resetCmd = resetCmd + " --cri-socket " + j.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint
		}
		_, _ = runtime.GetRunner().SudoCmd(resetCmd, true)
		return errors.Wrap(errors.WithStack(err), "join node failed")
	}
	return nil
}

type KubeadmReset struct {
	common.KubeAction
}

func (k *KubeadmReset) Execute(runtime connector.Runtime) error {
	resetCmd := "/usr/local/bin/kubeadm reset -f"
	if k.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint != "" {
		resetCmd = resetCmd + " --cri-socket " + k.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint
	}
	_, _ = runtime.GetRunner().SudoCmd(resetCmd, true)
	return nil
}

type FilterFirstMaster struct {
	common.KubeAction
}

func filterString(filter string, nodes []string) []string {
	j := 0
	for _, v := range nodes {
		if v != filter {
			nodes[j] = v
			j++
		}
	}
	resArr := nodes[:j]
	return resArr
}

func (f *FilterFirstMaster) Execute(runtime connector.Runtime) error {
	firstMaster := runtime.GetHostsByRole(common.Master)[0].GetName()
	//kubectl get node
	var nodes []string
	res, err := runtime.GetRunner().Cmd(
		"sudo -E /usr/local/bin/kubectl get nodes | awk '{print $1}'",
		true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl get nodes failed")
	}

	if !strings.Contains(res, "\r\n") {
		nodes = append(nodes, res)
	} else {
		nodes = strings.Split(res, "\r\n")
	}
	//nodes filter first master
	resArr := filterString(firstMaster, nodes)
	//nodes filter etcd nodes
	for i := 0; i < len(runtime.GetHostsByRole(common.ETCD)); i++ {
		etcdName := runtime.GetHostsByRole(common.ETCD)[i].GetName()
		resArr = filterString(etcdName, resArr)
	}
	workerName := make(map[string]struct{})
	for j := 0; j < len(runtime.GetHostsByRole(common.Worker)); j++ {
		workerName[runtime.GetHostsByRole(common.Worker)[j].GetName()] = struct{}{}
	}
	//make sure node is not the first master and etcd node name
	var node string
	for i := 0; i < len(resArr); i++ {
		if _, ok := workerName[resArr[i]]; ok && resArr[i] == f.KubeConf.Arg.NodeName {
			node = resArr[i]
			break
		}
	}

	if node == "" {
		return errors.New("" +
			"1. check the node name in the config-sample.yaml\n" +
			"2. check the node name in the Kubernetes cluster\n" +
			"3. check the node name is the first master and etcd node name\n")
	}

	f.PipelineCache.Set("dstNode", node)
	return nil
}

type FindNode struct {
	common.KubeAction
}

func (f *FindNode) Execute(runtime connector.Runtime) error {
	var resArr []string
	res, err := runtime.GetRunner().Cmd(
		"sudo -E /usr/local/bin/kubectl get nodes | awk '$3 !~ /master|control-plane|ROLES/ {print $1}'",
		true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl get nodes failed")
	}

	if !strings.Contains(res, "\r\n") {
		resArr = append(resArr, res)
	} else {
		resArr = strings.Split(res, "\r\n")
	}

	workerName := make(map[string]struct{})
	for j := 0; j < len(runtime.GetHostsByRole(common.Worker)); j++ {
		workerName[runtime.GetHostsByRole(common.Worker)[j].GetName()] = struct{}{}
	}

	var node string
	for i := 0; i < len(resArr); i++ {
		if _, ok := workerName[resArr[i]]; ok && resArr[i] == f.KubeConf.Arg.NodeName {
			node = resArr[i]
			break
		}
	}

	if node == "" {
		return errors.New("" +
			"1. check the node name in the config-sample.yaml\n" +
			"2. check the node name in the Kubernetes cluster\n" +
			"3. only support to delete a worker\n")
	}

	f.PipelineCache.Set("dstNode", node)
	return nil
}

type DrainNode struct {
	common.KubeAction
}

func (d *DrainNode) Execute(runtime connector.Runtime) error {
	nodeName, ok := d.PipelineCache.Get("dstNode")
	if !ok {
		return errors.New("get dstNode failed by pipeline cache")
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl drain %s --delete-emptydir-data --ignore-daemonsets --timeout=2m --force", nodeName),
		true); err != nil {
		return errors.Wrap(err, "drain the node failed")
	}
	return nil
}

type KubectlDeleteNode struct {
	common.KubeAction
}

func (k *KubectlDeleteNode) Execute(runtime connector.Runtime) error {
	nodeName, ok := k.PipelineCache.Get("dstNode")
	if !ok {
		return errors.New("get dstNode failed by pipeline cache")
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl delete node %s", nodeName),
		true); err != nil {
		return errors.Wrap(err, "delete the node failed")
	}
	return nil
}

type SetUpgradePlan struct {
	common.KubeAction
	Step UpgradeStep
}

func (s *SetUpgradePlan) Execute(_ connector.Runtime) error {
	currentVersion, ok := s.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return errors.New("get current Kubernetes version failed by pipeline cache")
	}

	desiredVersion, ok := s.PipelineCache.GetMustString(common.DesiredK8sVersion)
	if !ok {
		return errors.New("get desired Kubernetes version failed by pipeline cache")
	}
	if cmp, err := versionutil.MustParseSemantic(currentVersion).Compare(desiredVersion); err != nil {
		return err
	} else if cmp == 1 {
		logger.Log.Messagef(
			common.LocalHost,
			"The current version (%s) is greater than the target version (%s)",
			currentVersion, desiredVersion)
		os.Exit(0)
	}

	if s.Step == ToV121 {
		v122 := versionutil.MustParseSemantic("v1.22.0")
		atLeast := versionutil.MustParseSemantic(desiredVersion).AtLeast(v122)
		cmp, err := versionutil.MustParseSemantic(currentVersion).Compare("v1.21.5")
		if err != nil {
			return err
		}
		if atLeast && cmp <= 0 {
			desiredVersion = "v1.21.5"
		}
	}

	s.PipelineCache.Set(common.PlanK8sVersion, desiredVersion)
	return nil
}

type CalculateNextVersion struct {
	common.KubeAction
}

func (c *CalculateNextVersion) Execute(_ connector.Runtime) error {
	currentVersion, ok := c.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return errors.New("get current Kubernetes version failed by pipeline cache")
	}
	planVersion, ok := c.PipelineCache.GetMustString(common.PlanK8sVersion)
	if !ok {
		return errors.New("get upgrade plan Kubernetes version failed by pipeline cache")
	}
	nextVersionStr, err := calculateNextStr(currentVersion, planVersion)
	if err != nil {
		return errors.Wrap(err, "calculate next version failed")
	}
	c.KubeConf.Cluster.Kubernetes.Version = nextVersionStr
	return nil
}

func calculateNextStr(currentVersion, desiredVersion string) (string, error) {
	current := versionutil.MustParseSemantic(currentVersion)
	target := versionutil.MustParseSemantic(desiredVersion)
	var nextVersionMinor uint
	if target.Minor() == current.Minor() {
		nextVersionMinor = current.Minor()
	} else {
		nextVersionMinor = current.Minor() + 1
	}

	if nextVersionMinor == target.Minor() {
		if _, ok := files.FileSha256["kubeadm"]["amd64"][desiredVersion]; !ok {
			return "", errors.Errorf("the target version %s is not supported", desiredVersion)
		}
		return desiredVersion, nil
	} else {
		nextVersionPatchList := make([]int, 0)
		for supportVersionStr := range files.FileSha256["kubeadm"]["amd64"] {
			supportVersion := versionutil.MustParseSemantic(supportVersionStr)
			if supportVersion.Minor() == nextVersionMinor {
				nextVersionPatchList = append(nextVersionPatchList, int(supportVersion.Patch()))
			}
		}
		sort.Ints(nextVersionPatchList)

		nextVersion := current.WithMinor(nextVersionMinor)
		if len(nextVersionPatchList) == 0 {
			return "", errors.Errorf("Kubernetes minor version v%d.%d.x is not supported", nextVersion.Major(), nextVersion.Minor())
		}
		nextVersion = nextVersion.WithPatch(uint(nextVersionPatchList[len(nextVersionPatchList)-1]))

		return fmt.Sprintf("v%s", nextVersion.String()), nil
	}
}

type RestartKubelet struct {
	common.KubeAction
	ModuleName string
}

func (r *RestartKubelet) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	if _, err := runtime.GetRunner().SudoCmd("systemctl stop kubelet", false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("stop kubelet failed: %s", host.GetName()))
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl restart kubelet", true); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("restart kubelet failed: %s", host.GetName()))
	}
	time.Sleep(10 * time.Second)
	return nil
}

type UpgradeKubeMaster struct {
	common.KubeAction
	ModuleName string
}

func (u *UpgradeKubeMaster) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	if err := KubeadmUpgradeTasks(runtime, u); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("upgrade cluster using kubeadm failed: %s", host.GetName()))
	}

	if _, err := runtime.GetRunner().SudoCmd("systemctl stop kubelet", false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("stop kubelet failed: %s", host.GetName()))
	}

	if u.KubeConf.Cluster.Kubernetes.IsAtLeastV124() {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("echo 'KUBELET_KUBEADM_ARGS=\\\"--container-runtime-endpoint=%s\\\"' > /var/lib/kubelet/kubeadm-flags.env", u.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint), false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("update kubelet config failed: %s", host.GetName()))
		}
	}

	if err := SetKubeletTasks(runtime, u.KubeAction); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("set kubelet failed: %s", host.GetName()))
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl restart kubelet", true); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("restart kubelet failed: %s", host.GetName()))
	}

	time.Sleep(10 * time.Second)
	return nil
}

type UpgradeKubeWorker struct {
	common.KubeAction
	ModuleName string
}

func (u *UpgradeKubeWorker) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubeadm upgrade node", true); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("upgrade node using kubeadm failed: %s", host.GetName()))
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl stop kubelet", true); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("stop kubelet failed: %s", host.GetName()))
	}
	if u.KubeConf.Cluster.Kubernetes.IsAtLeastV124() {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("echo 'KUBELET_KUBEADM_ARGS=\\\"--container-runtime-endpoint=%s\\\"' > /var/lib/kubelet/kubeadm-flags.env", u.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint), false); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("update kubelet config failed: %s", host.GetName()))
		}
	}
	if err := SetKubeletTasks(runtime, u.KubeAction); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("set kubelet failed: %s", host.GetName()))
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl restart kubelet", true); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("restart kubelet failed: %s", host.GetName()))
	}
	time.Sleep(10 * time.Second)
	return nil
}

func KubeadmUpgradeTasks(runtime connector.Runtime, u *UpgradeKubeMaster) error {
	host := runtime.RemoteHost()

	kubeadmUpgrade := &task.RemoteTask{
		Name:     "KubeadmUpgrade",
		Desc:     "Upgrade cluster using kubeadm",
		Hosts:    []connector.Host{host},
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(KubeadmUpgrade),
		Parallel: false,
		Retry:    3,
	}

	copyKubeConfig := &task.RemoteTask{
		Name:     "CopyKubeConfig",
		Desc:     "Copy admin.conf to ~/.kube/config",
		Hosts:    []connector.Host{host},
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(CopyKubeConfigForControlPlane),
		Parallel: false,
		Retry:    2,
	}

	tasks := []task.Interface{
		kubeadmUpgrade,
		copyKubeConfig,
	}

	for i := range tasks {
		t := tasks[i]
		t.Init(runtime, u.ModuleCache, u.PipelineCache)
		if res := t.Execute(); res.IsFailed() {
			return res.CombineErr()
		}
	}
	return nil
}

type KubeadmUpgrade struct {
	common.KubeAction
}

func (k *KubeadmUpgrade) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	fmt.Println(k.KubeConf.Cluster.Kubernetes.Version)
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"timeout -k 600s 600s /usr/local/bin/kubeadm upgrade apply %s -y "+
			"--ignore-preflight-errors=all "+
			"--allow-experimental-upgrades "+
			"--allow-release-candidate-upgrades "+
			"--etcd-upgrade=false "+
			"--certificate-renewal=true ",
		k.KubeConf.Cluster.Kubernetes.Version), false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("upgrade master failed: %s", host.GetName()))
	}
	return nil
}

func SetKubeletTasks(runtime connector.Runtime, kubeAction common.KubeAction) error {
	host := runtime.RemoteHost()
	chmodKubelet := &task.RemoteTask{
		Name:     "ChownKubelet",
		Desc:     "Change kubelet owner",
		Hosts:    []connector.Host{host},
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(ChmodKubelet),
		Parallel: false,
		Retry:    2,
	}

	enableKubelet := &task.RemoteTask{
		Name:     "EnableKubelet",
		Desc:     "enable kubelet service",
		Hosts:    []connector.Host{host},
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(EnableKubelet),
		Parallel: false,
		Retry:    5,
	}

	tasks := []task.Interface{
		chmodKubelet,
		enableKubelet,
	}

	for i := range tasks {
		t := tasks[i]
		t.Init(runtime, kubeAction.ModuleCache, kubeAction.PipelineCache)
		if res := t.Execute(); res.IsFailed() {
			return res.CombineErr()
		}
	}
	return nil
}

//type UpgradeDNS struct {
//	common.KubeAction
//	ModuleName string
//}
//
//func (r *UpgradeDNS) Execute(runtime connector.Runtime) error {
//	if err := UpgradeCoredns(runtime, r.KubeAction); err != nil {
//		return errors.Wrap(errors.WithStack(err), "re-config coredns failed")
//	}
//	return nil
//}
//
//func UpgradeCoredns(runtime connector.Runtime, kubeAction common.KubeAction) error {
//	host := runtime.RemoteHost()
//
//	generateCoreDNSSvc := &task.RemoteTask{
//		Name:  "GenerateCoreDNS",
//		Desc:  "generate coredns manifests",
//		Hosts: []connector.Host{host},
//		Prepare: &prepare.PrepareCollection{
//			new(common.OnlyFirstMaster),
//		},
//		Action: &action.Template{
//			Template: dnsTemplates.Coredns,
//			Dst:      filepath.Join(common.KubeConfigDir, dnsTemplates.Coredns.Name()),
//			Data: util.Data{
//				"ClusterIP":    kubeAction.KubeConf.Cluster.CorednsClusterIP(),
//				"CorednsImage": images.GetImage(runtime, kubeAction.KubeConf, "coredns").ImageName(),
//				"DNSEtchHsts":  kubeAction.KubeConf.Cluster.DNS.DNSEtcHosts,
//			},
//		},
//		Parallel: true,
//	}
//
//	override := &task.RemoteTask{
//		Name:  "UpgradeCoreDNS",
//		Desc:  "upgrade coredns",
//		Hosts: []connector.Host{host},
//		Prepare: &prepare.PrepareCollection{
//			new(common.OnlyFirstMaster),
//		},
//		Action:   new(dns.DeployCoreDNS),
//		Parallel: false,
//	}
//
//	generateNodeLocalDNS := &task.RemoteTask{
//		Name:  "GenerateNodeLocalDNS",
//		Desc:  "generate nodelocaldns",
//		Hosts: []connector.Host{host},
//		Prepare: &prepare.PrepareCollection{
//			new(common.OnlyFirstMaster),
//			new(dns.EnableNodeLocalDNS),
//		},
//		Action: &action.Template{
//			Template: dnsTemplates.NodeLocalDNSService,
//			Dst:      filepath.Join(common.KubeConfigDir, dnsTemplates.NodeLocalDNSService.Name()),
//			Data: util.Data{
//				"NodelocaldnsImage": images.GetImage(runtime, kubeAction.KubeConf, "k8s-dns-node-cache").ImageName(),
//			},
//		},
//		Parallel: true,
//	}
//
//	applyNodeLocalDNS := &task.RemoteTask{
//		Name:  "DeployNodeLocalDNS",
//		Desc:  "deploy nodelocaldns",
//		Hosts: []connector.Host{host},
//		Prepare: &prepare.PrepareCollection{
//			new(common.OnlyFirstMaster),
//			new(dns.EnableNodeLocalDNS),
//		},
//		Action:   new(dns.DeployNodeLocalDNS),
//		Parallel: true,
//		Retry:    5,
//	}
//
//	generateNodeLocalDNSConfigMap := &task.RemoteTask{
//		Name:  "GenerateNodeLocalDNSConfigMap",
//		Desc:  "generate nodelocaldns configmap",
//		Hosts: []connector.Host{host},
//		Prepare: &prepare.PrepareCollection{
//			new(common.OnlyFirstMaster),
//			new(dns.EnableNodeLocalDNS),
//			new(dns.NodeLocalDNSConfigMapNotExist),
//		},
//		Action:   new(dns.GenerateNodeLocalDNSConfigMap),
//		Parallel: true,
//	}
//
//	applyNodeLocalDNSConfigMap := &task.RemoteTask{
//		Name:  "ApplyNodeLocalDNSConfigMap",
//		Desc:  "apply nodelocaldns configmap",
//		Hosts: []connector.Host{host},
//		Prepare: &prepare.PrepareCollection{
//			new(common.OnlyFirstMaster),
//			new(dns.EnableNodeLocalDNS),
//			new(dns.NodeLocalDNSConfigMapNotExist),
//		},
//		Action:   new(dns.ApplyNodeLocalDNSConfigMap),
//		Parallel: true,
//		Retry:    5,
//	}
//
//	tasks := []task.Interface{
//		override,
//		generateCoreDNSSvc,
//		override,
//		generateNodeLocalDNS,
//		applyNodeLocalDNS,
//		generateNodeLocalDNSConfigMap,
//		applyNodeLocalDNSConfigMap,
//	}
//
//	for i := range tasks {
//		t := tasks[i]
//		t.Init(runtime, kubeAction.ModuleCache, kubeAction.PipelineCache)
//		if res := t.Execute(); res.IsFailed() {
//			return res.CombineErr()
//		}
//	}
//	return nil
//}

type SetCurrentK8sVersion struct {
	common.KubeAction
}

func (s *SetCurrentK8sVersion) Execute(_ connector.Runtime) error {
	s.PipelineCache.Set(common.K8sVersion, s.KubeConf.Cluster.Kubernetes.Version)
	return nil
}

type SaveKubeConfig struct {
	common.KubeAction
}

func (s *SaveKubeConfig) Execute(runtime connector.Runtime) error {
	status, ok := s.PipelineCache.Get(common.ClusterStatus)
	if !ok {
		return errors.New("get kubernetes status failed by pipeline cache")
	}
	cluster := status.(*KubernetesStatus)
	kubeConfigStr := cluster.KubeConfig

	clusterPublicAddress := s.KubeConf.Cluster.ControlPlaneEndpoint.Address
	master1 := runtime.GetHostsByRole(common.Master)[0]
	if clusterPublicAddress == master1.GetInternalAddress() || clusterPublicAddress == "" {
		clusterPublicAddress = master1.GetAddress()
	}

	oldServer := fmt.Sprintf("https://%s:%d", s.KubeConf.Cluster.ControlPlaneEndpoint.Domain, s.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	newServer := fmt.Sprintf("https://%s:%d", clusterPublicAddress, s.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	newKubeConfigStr := strings.Replace(kubeConfigStr, oldServer, newServer, -1)
	kubeConfigBase64 := base64.StdEncoding.EncodeToString([]byte(newKubeConfigStr))

	config, err := clientcmd.NewClientConfigFromBytes([]byte(newKubeConfigStr))
	if err != nil {
		return err
	}
	restConfig, err := config.ClientConfig()
	if err != nil {
		return err
	}
	clientsetForCluster, err := kube.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubekey-system",
		},
	}
	if _, err := clientsetForCluster.
		CoreV1().
		Namespaces().
		Get(context.TODO(), namespace.Name, metav1.GetOptions{}); kubeerrors.IsNotFound(err) {
		if _, err := clientsetForCluster.
			CoreV1().
			Namespaces().
			Create(context.TODO(), namespace, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-kubeconfig", s.KubeConf.ClusterName),
		},
		Data: map[string]string{
			"kubeconfig": kubeConfigBase64,
		},
	}

	if _, err := clientsetForCluster.
		CoreV1().
		ConfigMaps("kubekey-system").
		Get(context.TODO(), cm.Name, metav1.GetOptions{}); kubeerrors.IsNotFound(err) {
		if _, err := clientsetForCluster.
			CoreV1().
			ConfigMaps("kubekey-system").
			Create(context.TODO(), cm, metav1.CreateOptions{}); err != nil {
			return err
		}
	} else {
		if _, err := clientsetForCluster.
			CoreV1().
			ConfigMaps("kubekey-system").
			Update(context.TODO(), cm, metav1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}

type ConfigureKubernetes struct {
	common.KubeAction
}

func (c *ConfigureKubernetes) Execute(runtime connector.Runtime) error {
	hosts := runtime.GetHostsByRole(common.K8s)

	for j := 0; j < len(hosts); j++ {
		kubeHost := hosts[j].(*kubekeyv1alpha2.KubeHost)
		for k, v := range kubeHost.Labels {
			labelCmd := fmt.Sprintf("/usr/local/bin/kubectl label --overwrite node %s %s=%s", hosts[j].GetName(), k, v)
			_, err := runtime.GetRunner().SudoCmd(labelCmd, true)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type EtcdSecurityEnhancemenAction struct {
	common.KubeAction
	ModuleName string
}

func (s *EtcdSecurityEnhancemenAction) Execute(runtime connector.Runtime) error {
	chmodEtcdCertsDirCmd := "chmod 700 /etc/ssl/etcd/ssl"
	chmodEtcdCertsCmd := "chmod 600 /etc/ssl/etcd/ssl/*"
	chmodEtcdDataDirCmd := "chmod 700 /var/lib/etcd"
	chmodEtcdCmd := "chmod 550 /usr/local/bin/etcd*"

	chownEtcdCertsDirCmd := "chown root:root /etc/ssl/etcd/ssl"
	chownEtcdCertsCmd := "chown root:root /etc/ssl/etcd/ssl/*"
	chownEtcdDataDirCmd := "chown etcd:etcd /var/lib/etcd"
	chownEtcdCmd := "chown root:root /usr/local/bin/etcd*"

	ETCDcmds := []string{chmodEtcdCertsDirCmd, chmodEtcdCertsCmd, chmodEtcdDataDirCmd, chmodEtcdCmd, chownEtcdCertsDirCmd, chownEtcdCertsCmd, chownEtcdDataDirCmd, chownEtcdCmd}

	if _, err := runtime.GetRunner().SudoCmd(strings.Join(ETCDcmds, " && "), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	return nil
}

type MasterSecurityEnhancemenAction struct {
	common.KubeAction
	ModuleName string
}

func (k *MasterSecurityEnhancemenAction) Execute(runtime connector.Runtime) error {
	// Control-plane Security Enhancemen
	chmodKubernetesDirCmd := "chmod 644 /etc/kubernetes"
	chownKubernetesDirCmd := "chown root:root /etc/kubernetes"

	chmodKubernetesConfigCmd := "chmod 600 -R /etc/kubernetes"
	chownKubernetesConfigCmd := "chown root:root -R /etc/kubernetes/*"

	chmodKubernetesManifestsDirCmd := "chmod 644 /etc/kubernetes/manifests"
	chownKubernetesManifestsDirCmd := "chown root:root /etc/kubernetes/manifests"

	chmodKubernetesCertsDirCmd := "chmod 644 /etc/kubernetes/pki"
	chownKubernetesCertsDirCmd := "chown root:root /etc/kubernetes/pki"

	// node Security Enhancemen
	chmodCniConfigDir := "chmod 600 -R /etc/cni/net.d"
	chownCniConfigDir := "chown root:root -R /etc/cni/net.d"

	chmodBinDir := "chmod 550 /usr/local/bin/"
	chownBinDir := "chown root:root /usr/local/bin/"

	chmodKubeCmd := "chmod 550 -R /usr/local/bin/kube*"
	chownKubeCmd := "chown root:root -R /usr/local/bin/kube*"

	chmodHelmCmd := "chmod 550 /usr/local/bin/helm"
	chownHelmCmd := "chown root:root /usr/local/bin/helm"

	chmodCniDir := "chmod 550 -R /opt/cni/bin"
	chownCniDir := "chown root:root -R /opt/cni/bin"

	chmodKubeletConfig := "chmod 640 /var/lib/kubelet/config.yaml && chmod 640 -R /etc/systemd/system/kubelet.service*"
	chownKubeletConfig := "chown root:root /var/lib/kubelet/config.yaml && chown root:root -R /etc/systemd/system/kubelet.service*"

	chmodCertsRenew := "chmod 640 /etc/systemd/system/k8s-certs-renew*"
	chownCertsRenew := "chown root:root /etc/systemd/system/k8s-certs-renew*"

	chmodMasterCmds := []string{chmodKubernetesConfigCmd, chmodKubernetesDirCmd, chmodKubernetesManifestsDirCmd, chmodKubernetesCertsDirCmd}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chmodMasterCmds, " && "), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}
	chownMasterCmds := []string{chownKubernetesConfigCmd, chownKubernetesDirCmd, chownKubernetesManifestsDirCmd, chownKubernetesCertsDirCmd}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chownMasterCmds, " && "), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	chmodNodesCmds := []string{chmodBinDir, chmodKubeCmd, chmodHelmCmd, chmodCniDir, chmodCniConfigDir, chmodKubeletConfig, chmodCertsRenew}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chmodNodesCmds, " && "), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}
	chownNodesCmds := []string{chownBinDir, chownKubeCmd, chownHelmCmd, chownCniDir, chownCniConfigDir, chownKubeletConfig, chownCertsRenew}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chownNodesCmds, " && "), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	return nil
}

type NodesSecurityEnhancemenAction struct {
	common.KubeAction
	ModuleName string
}

func (n *NodesSecurityEnhancemenAction) Execute(runtime connector.Runtime) error {
	// Control-plane Security Enhancemen
	chmodKubernetesDirCmd := "chmod 644 /etc/kubernetes"
	chownKubernetesDirCmd := "chown root:root /etc/kubernetes"

	chmodKubernetesConfigCmd := "chmod 600 -R /etc/kubernetes"
	chownKubernetesConfigCmd := "chown root:root -R /etc/kubernetes"

	chmodKubernetesManifestsDirCmd := "chmod 644 /etc/kubernetes/manifests"
	chownKubernetesManifestsDirCmd := "chown root:root /etc/kubernetes/manifests"

	chmodKubernetesCertsDirCmd := "chmod 644 /etc/kubernetes/pki"
	chownKubernetesCertsDirCmd := "chown root:root /etc/kubernetes/pki"

	// node Security Enhancemen
	chmodCniConfigDir := "chmod 600 -R /etc/cni/net.d"
	chownCniConfigDir := "chown root:root -R /etc/cni/net.d"

	chmodBinDir := "chmod 550 /usr/local/bin/"
	chownBinDir := "chown root:root /usr/local/bin/"

	chmodKubeCmd := "chmod 550 -R /usr/local/bin/kube*"
	chownKubeCmd := "chown root:root -R /usr/local/bin/kube*"

	chmodHelmCmd := "chmod 550 /usr/local/bin/helm"
	chownHelmCmd := "chown root:root /usr/local/bin/helm"

	chmodCniDir := "chmod 550 -R /opt/cni/bin"
	chownCniDir := "chown root:root -R /opt/cni/bin"

	chmodKubeletConfig := "chmod 640 /var/lib/kubelet/config.yaml && chmod 640 -R /etc/systemd/system/kubelet.service*"
	chownKubeletConfig := "chown root:root /var/lib/kubelet/config.yaml && chown root:root -R /etc/systemd/system/kubelet.service*"

	chmodMasterCmds := []string{chmodKubernetesConfigCmd, chmodKubernetesDirCmd, chmodKubernetesManifestsDirCmd, chmodKubernetesCertsDirCmd}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chmodMasterCmds, " && "), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}
	chownMasterCmds := []string{chownKubernetesConfigCmd, chownKubernetesDirCmd, chownKubernetesManifestsDirCmd, chownKubernetesCertsDirCmd}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chownMasterCmds, " && "), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	chmodNodesCmds := []string{chmodBinDir, chmodKubeCmd, chmodHelmCmd, chmodCniDir, chmodCniConfigDir, chmodKubeletConfig}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chmodNodesCmds, " && "), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}
	chownNodesCmds := []string{chownBinDir, chownKubeCmd, chownHelmCmd, chownCniDir, chownCniConfigDir, chownKubeletConfig}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(chownNodesCmds, " && "), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Updating permissions failed.")
	}

	return nil
}
