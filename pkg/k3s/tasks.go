package k3s

import (
	"fmt"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/k3s/templates"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

type GetClusterStatus struct {
	common.KubeAction
}

func (g *GetClusterStatus) Execute(runtime connector.Runtime) error {
	exist, err := runtime.GetRunner().FileExist("/etc/systemd/system/k3s.service")
	if err != nil {
		return err
	}

	if !exist {
		g.PipelineCache.Set(common.ClusterExist, false)
		return nil
	} else {
		g.PipelineCache.Set(common.ClusterExist, true)

		if v, ok := g.PipelineCache.Get(common.ClusterStatus); ok {
			cluster := v.(*K3sStatus)
			if err := cluster.SearchVersion(runtime); err != nil {
				return err
			}
			if err := cluster.SearchKubeConfig(runtime); err != nil {
				return err
			}
			if err := cluster.LoadKubeConfig(runtime, g.KubeConf); err != nil {
				return err
			}
			if err := cluster.SearchNodeToken(runtime); err != nil {
				return err
			}
			if err := cluster.SearchInfo(runtime); err != nil {
				return err
			}
			if err := cluster.SearchNodesInfo(runtime); err != nil {
				return err
			}
			g.PipelineCache.Set(common.ClusterStatus, cluster)
		} else {
			return errors.New("get k3s cluster status by pipeline cache failed")
		}
	}
	return nil
}

type SyncKubeBinary struct {
	common.KubeAction
}

func (s *SyncKubeBinary) Execute(runtime connector.Runtime) error {
	binariesMapObj, ok := s.PipelineCache.Get(common.KubeBinaries)
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

	binaryList := []string{"k3s", "helm", "kubecni"}
	for _, name := range binaryList {
		binary, ok := binariesMap[name]
		if !ok {
			return fmt.Errorf("get kube binary %s info failed: no such key", name)
		}

		switch name {
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

	binaries := []string{"kubectl", "crictl", "ctr"}
	var createLinkCMDs []string
	for _, binary := range binaries {
		createLinkCMDs = append(createLinkCMDs, fmt.Sprintf("ln -snf /usr/local/bin/k3s /usr/local/bin/%s", binary))
	}
	if _, err := runtime.GetRunner().SudoCmd(strings.Join(createLinkCMDs, " && "), false); err != nil {
		return errors.Wrap(errors.WithStack(err), "create ctl tool link failed")
	}

	return nil
}

type ChmodScript struct {
	common.KubeAction
}

func (c *ChmodScript) Execute(runtime connector.Runtime) error {
	killAllScript := filepath.Join("/usr/local/bin", templates.K3sKillallScript.Name())
	uninstallScript := filepath.Join("/usr/local/bin", templates.K3sUninstallScript.Name())

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", killAllScript),
		false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s", uninstallScript),
		false); err != nil {
		return err
	}
	return nil
}

type GenerateK3sService struct {
	common.KubeAction
}

func (g *GenerateK3sService) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	var server string
	if !host.IsRole(common.Master) {
		server = fmt.Sprintf("https://%s:%d", g.KubeConf.Cluster.ControlPlaneEndpoint.Domain, g.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	}

	defaultKubeletArs := map[string]string{
		"cni-conf-dir":    "/etc/cni/net.d",
		"cni-bin-dir":     "/opt/cni/bin",
		"kube-reserved":   "cpu=200m,memory=250Mi,ephemeral-storage=1Gi",
		"system-reserved": "cpu=200m,memory=250Mi,ephemeral-storage=1Gi",
		"eviction-hard":   "memory.available<5%,nodefs.available<10%",
	}
	defaultKubeProxyArgs := map[string]string{
		"proxy-mode": "ipvs",
	}

	kubeApiserverArgs, _ := util.GetArgs(map[string]string{}, g.KubeConf.Cluster.Kubernetes.ApiServerArgs)
	kubeControllerManager, _ := util.GetArgs(map[string]string{
		"pod-eviction-timeout":        "3m0s",
		"terminated-pod-gc-threshold": "5",
	}, g.KubeConf.Cluster.Kubernetes.ControllerManagerArgs)
	kubeSchedulerArgs, _ := util.GetArgs(map[string]string{}, g.KubeConf.Cluster.Kubernetes.SchedulerArgs)
	kubeletArgs, _ := util.GetArgs(defaultKubeletArs, g.KubeConf.Cluster.Kubernetes.KubeletArgs)
	kubeProxyArgs, _ := util.GetArgs(defaultKubeProxyArgs, g.KubeConf.Cluster.Kubernetes.KubeProxyArgs)

	templateAction := action.Template{
		Template: templates.K3sService,
		Dst:      filepath.Join("/etc/systemd/system/", templates.K3sService.Name()),
		Data: util.Data{
			"Server":            server,
			"IsMaster":          host.IsRole(common.Master),
			"NodeIP":            host.GetInternalAddress(),
			"HostName":          host.GetName(),
			"PodSubnet":         g.KubeConf.Cluster.Network.KubePodsCIDR,
			"ServiceSubnet":     g.KubeConf.Cluster.Network.KubeServiceCIDR,
			"ClusterDns":        g.KubeConf.Cluster.CorednsClusterIP(),
			"CertSANs":          g.KubeConf.Cluster.GenerateCertSANs(),
			"PauseImage":        images.GetImage(runtime, g.KubeConf, "pause").ImageName(),
			"ApiserverArgs":     kubeApiserverArgs,
			"ControllerManager": kubeControllerManager,
			"SchedulerArgs":     kubeSchedulerArgs,
			"KubeletArgs":       kubeletArgs,
			"KubeProxyArgs":     kubeProxyArgs,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type GenerateK3sServiceEnv struct {
	common.KubeAction
}

func (g *GenerateK3sServiceEnv) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	clusterStatus, ok := g.PipelineCache.Get(common.ClusterStatus)
	if !ok {
		return errors.New("get cluster status by pipeline cache failed")
	}
	cluster := clusterStatus.(*K3sStatus)

	var externalEtcd kubekeyapiv1alpha2.ExternalEtcd
	var endpointsList []string
	var caFile, certFile, keyFile string
	var token string

	for _, node := range runtime.GetHostsByRole(common.ETCD) {
		endpoint := fmt.Sprintf("https://%s:%s", node.GetInternalAddress(), kubekeyapiv1alpha2.DefaultEtcdPort)
		endpointsList = append(endpointsList, endpoint)
	}
	externalEtcd.Endpoints = endpointsList

	externalEtcdEndpoints := strings.Join(endpointsList, ",")
	caFile = "/etc/ssl/etcd/ssl/ca.pem"
	certFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", runtime.GetHostsByRole(common.Master)[0].GetName())
	keyFile = fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", runtime.GetHostsByRole(common.Master)[0].GetName())

	externalEtcd.CaFile = caFile
	externalEtcd.CertFile = certFile
	externalEtcd.KeyFile = keyFile

	if !host.IsRole(common.Master) {
		token = cluster.NodeToken
	}

	templateAction := action.Template{
		Template: templates.K3sServiceEnv,
		Dst:      filepath.Join("/etc/systemd/system/", templates.K3sServiceEnv.Name()),
		Data: util.Data{
			"DataStoreEndPoint": externalEtcdEndpoints,
			"DataStoreCaFile":   caFile,
			"DataStoreCertFile": certFile,
			"DataStoreKeyFile":  keyFile,
			"IsMaster":          host.IsRole(common.Master),
			"Token":             token,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type EnableK3sService struct {
	common.KubeAction
}

func (e *EnableK3sService) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl enable --now k3s",
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), "enable k3s failed")
	}
	return nil
}

type CopyK3sKubeConfig struct {
	common.KubeAction
}

func (c *CopyK3sKubeConfig) Execute(runtime connector.Runtime) error {
	createConfigDirCmd := "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	getKubeConfigCmd := "cp -f /etc/rancher/k3s/k3s.yaml /root/.kube/config"
	getKubeConfigCmdUsr := "cp -f /etc/rancher/k3s/k3s.yaml $HOME/.kube/config"
	chownKubeConfig := "chown $(id -u):$(id -g) $HOME/.kube/config"

	cmd := strings.Join([]string{createConfigDirCmd, getKubeConfigCmd, getKubeConfigCmdUsr, chownKubeConfig}, " && ")
	if _, err := runtime.GetRunner().SudoCmd(cmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "copy k3s kube config failed")
	}
	return nil
}

type AddMasterTaint struct {
	common.KubeAction
}

func (a *AddMasterTaint) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	cmd := fmt.Sprintf(
		"/usr/local/bin/kubectl taint nodes %s node-role.kubernetes.io/master=effect:NoSchedule --overwrite",
		host.GetName())

	if _, err := runtime.GetRunner().SudoCmd(cmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "add master NoSchedule taint failed")
	}
	return nil
}

type AddWorkerLabel struct {
	common.KubeAction
}

func (a *AddWorkerLabel) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()

	cmd := fmt.Sprintf(
		"/usr/local/bin/kubectl label --overwrite node %s node-role.kubernetes.io/worker=",
		host.GetName())

	if _, err := runtime.GetRunner().SudoCmd(cmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "add master NoSchedule taint failed")
	}
	return nil
}

type SyncKubeConfigToWorker struct {
	common.KubeAction
}

func (s *SyncKubeConfigToWorker) Execute(runtime connector.Runtime) error {
	if v, ok := s.PipelineCache.Get(common.ClusterStatus); ok {
		cluster := v.(*K3sStatus)

		createConfigDirCmd := "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
		if _, err := runtime.GetRunner().SudoCmd(createConfigDirCmd, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "create .kube dir failed")
		}

		oldServer := "server: https://127.0.0.1:6443"
		newServer := fmt.Sprintf("server: https://%s:%d",
			s.KubeConf.Cluster.ControlPlaneEndpoint.Domain,
			s.KubeConf.Cluster.ControlPlaneEndpoint.Port)
		newKubeConfig := strings.Replace(cluster.KubeConfig, oldServer, newServer, -1)

		syncKubeConfigForRootCmd := fmt.Sprintf("echo '%s' > %s", newKubeConfig, "/root/.kube/config")
		if _, err := runtime.GetRunner().SudoCmd(syncKubeConfigForRootCmd, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "sync kube config for root failed")
		}

		chownKubeConfig := "chown $(id -u):$(id -g) -R $HOME/.kube"
		syncKubeConfigForUserCmd := fmt.Sprintf("echo '%s' > %s && %s", newKubeConfig, "$HOME/.kube/config", chownKubeConfig)
		if _, err := runtime.GetRunner().SudoCmd(syncKubeConfigForUserCmd, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "sync kube config for normal user failed")
		}
	}
	return nil
}

type ExecUninstallScript struct {
	common.KubeAction
}

func (e *ExecUninstallScript) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && /usr/local/bin/k3s-killall.sh",
		true); err != nil {
		return errors.Wrap(errors.WithStack(err), "add master NoSchedule taint failed")
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && /usr/local/bin/k3s-uninstall.sh",
		true); err != nil {
		return errors.Wrap(errors.WithStack(err), "add master NoSchedule taint failed")
	}
	return nil
}
