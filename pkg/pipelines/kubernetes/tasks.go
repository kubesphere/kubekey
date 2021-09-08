package kubernetes

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/images"
	"github.com/kubesphere/kubekey/pkg/pipelines/kubernetes/templates"
	"github.com/kubesphere/kubekey/pkg/pipelines/kubernetes/templates/v1beta2"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

const (
	ClusterStatus = "ClusterStatus"
	ClusterExist  = "clusterExist"
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
		g.RootCache.Set(ClusterExist, false)
		return nil
	} else {
		g.RootCache.Set(ClusterExist, true)

		if v, ok := g.RootCache.Get(ClusterStatus); ok {
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
			if err := cluster.SearchJoinCmd(runtime); err != nil {
				return err
			}
			if err := cluster.SearchInfo(runtime); err != nil {
				return err
			}
			if err := cluster.SearchNodesInfo(runtime); err != nil {
				return err
			}
			g.RootCache.Set(ClusterStatus, cluster)
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
	binariesMapObj, ok := i.RootCache.Get(common.KubeBinaries)
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]files.KubeBinary)

	if err := SyncKubeBinaries(runtime, binariesMap); err != nil {
		return err
	}
	return nil
}

// SyncKubeBinaries is used to sync kubernetes' binaries to each node.
func SyncKubeBinaries(runtime connector.Runtime, binariesMap map[string]files.KubeBinary) error {
	_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("if [ -d %s ]; then rm -rf %s ;fi && mkdir -p %s",
		common.TmpDir, common.TmpDir, common.TmpDir), false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "reset tmp dir failed")
	}

	binaryList := []string{"kubeadm", "kubelet", "kubectl", "helm", "kubecni"}
	for _, name := range binaryList {
		binary, ok := binariesMap[name]
		if !ok {
			return fmt.Errorf("get kube binary %s info failed: no such key", name)
		}
		switch name {
		case "kubelet":
			if err := runtime.GetRunner().Scp(binary.Path, fmt.Sprintf("%s/%s", common.TmpDir, binary.Name)); err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
			}
		case "kubecni":
			dst := filepath.Join("/opt/cni/bin", binary.Name)
			if err := runtime.GetRunner().SudoScp(binary.Path, dst); err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
			}
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar -zxf %s", dst), false); err != nil {
				return err
			}
		default:
			dst := filepath.Join(common.BinDir, binary.Name)
			if err := runtime.GetRunner().SudoScp(binary.Path, dst); err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync kube binaries failed"))
			}
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", dst), false); err != nil {
				return err
			}
		}
	}
	return nil
}

type SyncKubelet struct {
	common.KubeAction
}

func (s *SyncKubelet) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("cp -f /tmp/kubekey/kubelet /usr/local/bin/kubelet "+
		"&& chmod +x /usr/local/bin/kubelet", false); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync kubelet service failed")
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
		},
	}

	templateAction.Init(nil, nil, runtime)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type GenerateKubeadmConfig struct {
	common.KubeAction
}

func (g *GenerateKubeadmConfig) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	localConfig := filepath.Join(runtime.GetWorkDir(), "kubeadm-config.yaml")
	if util.IsExist(localConfig) {
		if err := runtime.GetRunner().SudoScp(localConfig, "/etc/kubernetes/kubeadm-config.yaml"); err != nil {
			return errors.Wrap(errors.WithStack(err), "scp local kubeadm config failed")
		}
	} else {
		// generate etcd configuration
		var externalEtcd kubekeyapiv1alpha1.ExternalEtcd
		var endpointsList []string
		var caFile, certFile, keyFile, containerRuntimeEndpoint string

		for _, host := range runtime.GetHostsByRole(common.ETCD) {
			endpoint := fmt.Sprintf("https://%s:%s", host.GetInternalAddress(), kubekeyapiv1alpha1.DefaultEtcdPort)
			endpointsList = append(endpointsList, endpoint)
		}
		externalEtcd.Endpoints = endpointsList

		caFile = "/etc/ssl/etcd/ssl/ca.pem"
		certFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", host.GetName())
		keyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", host.GetName())

		externalEtcd.CaFile = caFile
		externalEtcd.CertFile = certFile
		externalEtcd.KeyFile = keyFile

		_, ApiServerArgs := util.GetArgs(v1beta2.ApiServerArgs, g.KubeConf.Cluster.Kubernetes.ApiServerArgs)
		_, ControllerManagerArgs := util.GetArgs(v1beta2.ControllermanagerArgs, g.KubeConf.Cluster.Kubernetes.ControllerManagerArgs)
		_, SchedulerArgs := util.GetArgs(v1beta2.SchedulerArgs, g.KubeConf.Cluster.Kubernetes.SchedulerArgs)

		// generate cri configuration
		switch g.KubeConf.Cluster.Kubernetes.ContainerManager {
		case "docker":
			containerRuntimeEndpoint = ""
		case "crio":
			containerRuntimeEndpoint = kubekeyapiv1alpha1.DefaultCrioEndpoint
		case "containerd":
			containerRuntimeEndpoint = kubekeyapiv1alpha1.DefaultContainerdEndpoint
		case "isula":
			containerRuntimeEndpoint = kubekeyapiv1alpha1.DefaultIsulaEndpoint
		default:
			containerRuntimeEndpoint = ""
		}

		if g.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint != "" {
			containerRuntimeEndpoint = g.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint
		}

		templateAction := action.Template{
			Template: v1beta2.KubeadmConfig,
			Dst:      filepath.Join(common.KubeConfigDir, v1beta2.KubeadmConfig.Name()),
			Data: util.Data{
				"ImageRepo":              strings.TrimSuffix(images.GetImage(runtime, g.KubeConf, "kube-apiserver").ImageRepo(), "/kube-apiserver"),
				"CorednsRepo":            strings.TrimSuffix(images.GetImage(runtime, g.KubeConf, "coredns").ImageRepo(), "/coredns"),
				"CorednsTag":             images.GetImage(runtime, g.KubeConf, "coredns").Tag,
				"Version":                g.KubeConf.Cluster.Kubernetes.Version,
				"ClusterName":            g.KubeConf.Cluster.Kubernetes.ClusterName,
				"ControlPlanAddr":        g.KubeConf.Cluster.ControlPlaneEndpoint.Address,
				"ControlPlanPort":        g.KubeConf.Cluster.ControlPlaneEndpoint.Port,
				"ControlPlaneEndpoint":   fmt.Sprintf("%s:%d", g.KubeConf.Cluster.ControlPlaneEndpoint.Domain, g.KubeConf.Cluster.ControlPlaneEndpoint.Port),
				"PodSubnet":              g.KubeConf.Cluster.Network.KubePodsCIDR,
				"ServiceSubnet":          g.KubeConf.Cluster.Network.KubeServiceCIDR,
				"CertSANs":               g.KubeConf.Cluster.GenerateCertSANs(),
				"ExternalEtcd":           externalEtcd,
				"NodeCidrMaskSize":       g.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
				"CriSock":                containerRuntimeEndpoint,
				"InternalLBDisabled":     !g.KubeConf.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled(),
				"AdvertiseAddress":       host.GetInternalAddress(),
				"ApiServerArgs":          ApiServerArgs,
				"ControllerManagerArgs":  ControllerManagerArgs,
				"SchedulerArgs":          SchedulerArgs,
				"KubeletConfiguration":   v1beta2.GetKubeletConfiguration(runtime, g.KubeConf, containerRuntimeEndpoint),
				"KubeProxyConfiguration": v1beta2.GetKubeProxyConfiguration(g.KubeConf),
			},
		}

		templateAction.Init(nil, nil, runtime)
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
	if _, err := runtime.GetRunner().Cmd("sudo env PATH=$PATH /bin/sh -c \""+
		"/usr/local/bin/kubeadm init "+
		"--config=/etc/kubernetes/kubeadm-config.yaml "+
		"--ignore-preflight-errors=FileExisting-crictl\"", true); err != nil {
		// kubeadm reset and then retry
		_, _ = runtime.GetRunner().Cmd("sudo env PATH=$PATH /bin/sh -c \"/usr/local/bin/kubeadm reset -f\"", true)
		return errors.Wrap(errors.WithStack(err), "init kubernetes cluster failed")
	}
	return nil
}

type CopyKubeConfig struct {
	common.KubeAction
}

func (c *CopyKubeConfig) Execute(runtime connector.Runtime) error {
	createConfigDirCmd := "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	getKubeConfigCmd := "cp -f /etc/kubernetes/admin.conf /root/.kube/config"
	getKubeConfigCmdUsr := "cp -f /etc/kubernetes/admin.conf $HOME/.kube/config"
	chownKubeConfig := "chown $(id -u):$(id -g) $HOME/.kube/config"

	cmd := strings.Join([]string{createConfigDirCmd, getKubeConfigCmd, getKubeConfigCmdUsr, chownKubeConfig}, " && ")
	if _, err := runtime.GetRunner().SudoCmd(cmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "copy kube config failed")
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
		return errors.Wrap(errors.WithStack(err), "remove master taint failed")
	}
	return nil
}

type AddWorkerLabel struct {
	common.KubeAction
}

func (a *AddWorkerLabel) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"/usr/local/bin/kubectl label --overwrite node %s node-role.kubernetes.io/worker=",
		runtime.RemoteHost().GetName()), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "add worker label failed")
	}
	return nil
}

type GetJoinCmd struct {
	common.KubeAction
}

func (g *GetJoinCmd) Execute(runtime connector.Runtime) error {
	if v, ok := g.RootCache.Get(ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)
		if err := cluster.SearchJoinCmd(runtime); err != nil {
			return err
		}
		g.RootCache.Set(ClusterStatus, cluster)
	} else {
		return errors.New("get kubernetes cluster status by pipeline cache failed")
	}
	return nil
}

type GetKubeConfig struct {
	common.KubeAction
}

func (g *GetKubeConfig) Execute(runtime connector.Runtime) error {
	if v, ok := g.RootCache.Get(ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)
		if err := cluster.SearchKubeConfig(runtime); err != nil {
			return err
		}
		g.RootCache.Set(ClusterStatus, cluster)
	} else {
		return errors.New("get kubernetes cluster status by pipeline cache failed")
	}
	return nil
}

type LoadKubeConfig struct {
	common.KubeAction
}

func (l *LoadKubeConfig) Execute(runtime connector.Runtime) error {
	if v, ok := l.RootCache.Get(ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)
		if err := cluster.LoadKubeConfig(runtime, l.KubeConf); err != nil {
			return err
		}
		l.RootCache.Set(ClusterStatus, cluster)
	} else {
		return errors.New("get kubernetes cluster status by pipeline cache failed")
	}
	return nil
}

type AddMasterNode struct {
	common.KubeAction
}

func (a *AddMasterNode) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	if v, ok := a.RootCache.Get(ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)
		if _, err := runtime.GetRunner().Cmd(fmt.Sprintf(
			"sudo env PATH=$PATH /bin/sh -c \"%s\"", fmt.Sprintf("%s --apiserver-advertise-address %s", cluster.JoinMasterCmd, host.GetInternalAddress())),
			true); err != nil {
			_, _ = runtime.GetRunner().Cmd("sudo env PATH=$PATH /bin/sh -c \"/usr/local/bin/kubeadm reset -f\"", true)
			return errors.Wrap(errors.WithStack(err), "join master failed")
		}
	}
	return nil
}

type AddWorkerNode struct {
	common.KubeAction
}

func (a *AddWorkerNode) Execute(runtime connector.Runtime) error {
	if v, ok := a.RootCache.Get(ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)
		if _, err := runtime.GetRunner().Cmd(fmt.Sprintf("sudo env PATH=$PATH /bin/sh -c \"%s\"", cluster.JoinWorkerCmd), true); err != nil {
			_, _ = runtime.GetRunner().Cmd("sudo env PATH=$PATH /bin/sh -c \"/usr/local/bin/kubeadm reset -f\"", true)
			return errors.Wrap(errors.WithStack(err), "join master failed")
		}
	}
	return nil
}

type SyncKubeConfig struct {
	common.KubeAction
}

func (s *SyncKubeConfig) Execute(runtime connector.Runtime) error {
	if v, ok := s.RootCache.Get(ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)

		createConfigDirCmd := "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
		if _, err := runtime.GetRunner().SudoCmd(createConfigDirCmd, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "create .kube dir failed")
		}

		syncKubeConfigForRootCmd := fmt.Sprintf("echo %s | base64 -d > %s", cluster.KubeConfig, "/root/.kube/config")
		if _, err := runtime.GetRunner().SudoCmd(syncKubeConfigForRootCmd, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "sync kube config for root failed")
		}

		chownKubeConfig := "chown $(id -u):$(id -g) -R $HOME/.kube"
		syncKubeConfigForUserCmd := fmt.Sprintf("echo %s | base64 -d > %s && %s", cluster.KubeConfig, "$HOME/.kube/config", chownKubeConfig)
		if _, err := runtime.GetRunner().SudoCmd(syncKubeConfigForUserCmd, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "sync kube config for normal user failed")
		}
	}
	return nil
}
