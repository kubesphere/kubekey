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
	binariesMapObj, ok := i.PipelineCache.Get(common.KubeBinaries)
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
	IsInitConfiguration bool
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
		var externalEtcd kubekeyapiv1alpha1.ExternalEtcd
		var endpointsList []string
		var caFile, certFile, keyFile string

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

		checkCgroupDriver, err := v1beta2.GetKubeletCgroupDriver(runtime, g.KubeConf)
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
			Template: v1beta2.KubeadmConfig,
			Dst:      filepath.Join(common.KubeConfigDir, v1beta2.KubeadmConfig.Name()),
			Data: util.Data{
				"IsInitCluster":          g.IsInitConfiguration,
				"ImageRepo":              strings.TrimSuffix(images.GetImage(runtime, g.KubeConf, "kube-apiserver").ImageRepo(), "/kube-apiserver"),
				"CorednsRepo":            strings.TrimSuffix(images.GetImage(runtime, g.KubeConf, "coredns").ImageRepo(), "/coredns"),
				"CorednsTag":             images.GetImage(runtime, g.KubeConf, "coredns").Tag,
				"Version":                g.KubeConf.Cluster.Kubernetes.Version,
				"ClusterName":            g.KubeConf.Cluster.Kubernetes.ClusterName,
				"AdvertiseAddress":       host.GetInternalAddress(),
				"ControlPlanPort":        g.KubeConf.Cluster.ControlPlaneEndpoint.Port,
				"ControlPlaneEndpoint":   fmt.Sprintf("%s:%d", g.KubeConf.Cluster.ControlPlaneEndpoint.Domain, g.KubeConf.Cluster.ControlPlaneEndpoint.Port),
				"PodSubnet":              g.KubeConf.Cluster.Network.KubePodsCIDR,
				"ServiceSubnet":          g.KubeConf.Cluster.Network.KubeServiceCIDR,
				"CertSANs":               g.KubeConf.Cluster.GenerateCertSANs(),
				"ExternalEtcd":           externalEtcd,
				"NodeCidrMaskSize":       g.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
				"CriSock":                g.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint,
				"ApiServerArgs":          ApiServerArgs,
				"ControllerManagerArgs":  ControllerManagerArgs,
				"SchedulerArgs":          SchedulerArgs,
				"KubeletConfiguration":   v1beta2.GetKubeletConfiguration(runtime, g.KubeConf, g.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint),
				"KubeProxyConfiguration": v1beta2.GetKubeProxyConfiguration(g.KubeConf),
				"IsControlPlane":         host.IsRole(common.Master),
				"CgroupDriver":           checkCgroupDriver,
				"BootstrapToken":         bootstrapToken,
				"CertificateKey":         certificateKey,
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
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubeadm init "+
		"--config=/etc/kubernetes/kubeadm-config.yaml "+
		"--ignore-preflight-errors=FileExisting-crictl", true); err != nil {
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

type JoinNode struct {
	common.KubeAction
}

func (j *JoinNode) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubeadm join --config=/etc/kubernetes/kubeadm-config.yaml",
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

type SyncKubeConfigToWorker struct {
	common.KubeAction
}

func (s *SyncKubeConfigToWorker) Execute(runtime connector.Runtime) error {
	if v, ok := s.PipelineCache.Get(common.ClusterStatus); ok {
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
