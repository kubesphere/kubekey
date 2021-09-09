package upgrade

import (
	"encoding/base64"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
	"github.com/kubesphere/kubekey/pkg/kubernetes/config/v1beta2"
	"github.com/kubesphere/kubekey/pkg/kubernetes/preinstall"
	"github.com/kubesphere/kubekey/pkg/plugins/dns"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

type UpgradeStatus struct {
	currentVersions   map[string]string
	currentVersionStr string
	nextVersionStr    string
	mu                sync.Mutex
	kubeConfig        string
}

func UpgradeKubeCluster(mgr *manager.Manager) error {
	var currentClusterUpgradeStatus *UpgradeStatus

	mgr.Logger.Infoln("Get current version")
	if err := mgr.RunTaskOnK8sNodes(currentClusterUpgradeStatus.getCurrentVersion, true); err != nil {
		return err
	}

	return currentClusterUpgradeStatus.UpgradeCurrentKubeCluster(mgr)
}

func (s *UpgradeStatus) getCurrentVersion(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	kubeletVersionInfo, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubelet --version\"", 3, false)
	if err != nil {
		return errors.Wrap(err, "Failed to get current kubelet version")
	}
	kubeletVersionStr := strings.Split(kubeletVersionInfo, " ")[1]
	s.mu.Lock()
	s.currentVersions[kubeletVersionStr] = kubeletVersionStr
	if minVersion, err := getMinVersion(s.currentVersions); err != nil {
		return err
	} else {
		s.currentVersions = make(map[string]string)
		s.currentVersions[minVersion] = minVersion
		s.currentVersionStr = fmt.Sprintf("v%s", minVersion)
	}
	s.mu.Unlock()

	if node.IsMaster {
		apiserverVersionStr, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | rev | cut -d ':' -f1 | rev\"", 3, false)
		if err != nil {
			return errors.Wrap(err, "Failed to get current kube-apiserver version")
		}
		s.mu.Lock()
		s.currentVersions[apiserverVersionStr] = apiserverVersionStr
		if minVersion, err := getMinVersion(s.currentVersions); err != nil {
			return err
		} else {
			s.currentVersions = make(map[string]string)
			s.currentVersions[minVersion] = minVersion
			s.currentVersionStr = fmt.Sprintf("v%s", minVersion)
		}
		s.mu.Unlock()
	}

	return nil
}

func getMinVersion(versionsMap map[string]string) (string, error) {
	versionList := []*versionutil.Version{}

	for version := range versionsMap {
		if versionStr, err := versionutil.ParseSemantic(version); err == nil {
			versionList = append(versionList, versionStr)
		} else {
			return "", err
		}
	}

	if len(versionList) > 1 {
		minVersion := versionList[0]
		for _, version := range versionList {
			if minVersion.AtLeast(version) {
				minVersion = version
			}
		}

		return minVersion.String(), nil
	} else {
		return versionList[0].String(), nil
	}
}

func (s *UpgradeStatus) upgradeKubeMasters(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	kubeletVersion, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubelet --version\"", 3, false)
	if err != nil {
		return errors.Wrap(err, "Failed to get current kubelet version")
	}
	kubeApiserverVersion, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | rev | cut -d ':' -f1 | rev\"", 3, false)
	if err != nil {
		return errors.Wrap(err, "Failed to get current kubelet version")
	}
	if strings.Split(kubeletVersion, " ")[1] != mgr.Cluster.Kubernetes.Version || strings.TrimSpace(kubeApiserverVersion) != mgr.Cluster.Kubernetes.Version {
		mgr.Logger.Infof("Upgrading %s [%s]\n", node.Name, node.InternalAddress)
		if err := kubernetes.SyncKubeBinaries(mgr, node); err != nil {
			return err
		}

		var kubeadmCfgBase64 string
		if util.IsExist(fmt.Sprintf("%s/kubeadm-config.yaml", mgr.WorkDir)) {
			output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cat %s/kubeadm-config.yaml | base64 --wrap=0", mgr.WorkDir)).CombinedOutput()
			if err != nil {
				fmt.Println(string(output))
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to read custom kubeadm config: %s/kubeadm-config.yaml", mgr.WorkDir))
			}
			kubeadmCfgBase64 = strings.TrimSpace(string(output))
		} else {
			kubeadmCfg, err := v1beta2.GenerateKubeadmCfg(mgr, node, kubernetes.IsInitCluster, "", "")
			if err != nil {
				return err
			}
			kubeadmCfgBase64 = base64.StdEncoding.EncodeToString([]byte(kubeadmCfg))
		}

		_, err1 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/kubernetes && echo %s | base64 -d > /etc/kubernetes/kubeadm-config.yaml\"", kubeadmCfgBase64), 1, false)
		if err1 != nil {
			return errors.Wrap(errors.WithStack(err1), "Failed to generate kubeadm config")
		}

		// Modify the value of 'advertiseAddress' to the current node's address.
		// Make sure the current master node's address will be registered in kubernetes endpoint.
		if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo sed -i 's#advertiseAddress:.*#advertiseAddress: %s#g' /etc/kubernetes/kubeadm-config.yaml", node.InternalAddress), 0, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "Failed to modify kubeadm config")
		}

		for i := 0; i < 3; i++ {
			if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf(
				"sudo -E /bin/sh -c \"timeout -k 600s 600s /usr/local/bin/kubeadm upgrade apply -y %s --config=/etc/kubernetes/kubeadm-config.yaml "+
					"--ignore-preflight-errors=all --allow-experimental-upgrades --allow-release-candidate-upgrades --etcd-upgrade=false --certificate-renewal=true --force\"",
				mgr.Cluster.Kubernetes.Version),
				0, false); err != nil {
				if i == 1 {
					return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to upgrade master: %s", node.Name))
				}

				if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl restart kubelet\"", 2, true); err != nil {
					return err
				}
				time.Sleep(30 * time.Second)

			} else {
				break
			}
		}

		if err := kubernetes.GetKubeConfigForControlPlane(mgr); err != nil {
			return err
		}

		if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl stop kubelet\"", 2, true); err != nil {
			return err
		}

		if err := kubernetes.SetKubelet(mgr, node); err != nil {
			return err
		}

		if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl restart kubelet\"", 2, true); err != nil {
			return err
		}

		kubeCfgBase64Cmd := "cat /etc/kubernetes/admin.conf | base64 --wrap=0"
		output, err2 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", kubeCfgBase64Cmd), 1, false)
		if err2 != nil {
			return errors.Wrap(errors.WithStack(err2), "Failed to get new kubeconfig")
		}
		s.kubeConfig = output
	}

	time.Sleep(30 * time.Second)
	return nil
}

func (s *UpgradeStatus) upgradeKubeWorkers(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	kubeletVersion, err := mgr.Runner.ExecuteCmd("/usr/local/bin/kubelet --version", 3, false)
	if err != nil {
		return errors.Wrap(err, "Failed to get current kubelet version")
	}
	if strings.Split(kubeletVersion, " ")[1] != mgr.Cluster.Kubernetes.Version {
		mgr.Logger.Infof("Upgrading %s [%s]\n", node.Name, node.InternalAddress)
		if err := kubernetes.SyncKubeBinaries(mgr, node); err != nil {
			return err
		}

		_, _ = mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubeadm upgrade node\"", 2, true)

		if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl stop kubelet\"", 2, true); err != nil {
			return err
		}

		if err := kubernetes.SetKubelet(mgr, node); err != nil {
			return err
		}

		if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl restart kubelet\"", 2, true); err != nil {
			return err
		}
	}

	createConfigDirCmd := "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	chownKubeConfig := "chown $(id -u):$(id -g) $HOME/.kube/config"
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", createConfigDirCmd), 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to create kube dir")
	}
	syncKubeconfigCmd := fmt.Sprintf("echo %s | base64 -d > %s && echo %s | base64 -d > %s && %s", s.kubeConfig, "/root/.kube/config", s.kubeConfig, "$HOME/.kube/config", chownKubeConfig)
	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", syncKubeconfigCmd), 1, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to sync kube config")
	}

	return nil
}

func (s *UpgradeStatus) UpgradeCurrentKubeCluster(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Upgrading kube cluster")
	targetVersionStr := mgr.Cluster.Kubernetes.Version
	cmp, err := versionutil.MustParseSemantic(s.currentVersionStr).Compare(mgr.Cluster.Kubernetes.Version)
	if err != nil {
		return err
	}
	if cmp == 1 {
		mgr.Logger.Warningln(fmt.Sprintf("The current version (%s) is greater than the target version (%s)", s.currentVersionStr, targetVersionStr))
		os.Exit(0)
	}
Loop:
	for {
		if s.currentVersionStr != targetVersionStr {
			currentVersion := versionutil.MustParseSemantic(s.currentVersionStr)
			targetVersion := versionutil.MustParseSemantic(targetVersionStr)
			var nextVersionMinor uint
			if targetVersion.Minor() == currentVersion.Minor() {
				nextVersionMinor = currentVersion.Minor()
			} else {
				nextVersionMinor = currentVersion.Minor() + 1
			}

			if nextVersionMinor == versionutil.MustParseSemantic(targetVersionStr).Minor() {
				s.nextVersionStr = targetVersionStr
			} else {
				nextVersionPatchList := []int{}
				for supportVersionStr := range files.FileSha256["kubeadm"]["amd64"] {
					supportVersion, err := versionutil.ParseSemantic(supportVersionStr)
					if err != nil {
						return err
					}
					if supportVersion.Minor() == nextVersionMinor {
						nextVersionPatchList = append(nextVersionPatchList, int(supportVersion.Patch()))
					}
				}
				sort.Ints(nextVersionPatchList)

				nextVersion := currentVersion.WithMinor(nextVersionMinor)
				nextVersion = nextVersion.WithPatch(uint(nextVersionPatchList[len(nextVersionPatchList)-1]))

				s.nextVersionStr = fmt.Sprintf("v%s", nextVersion.String())
			}

			mgr.Cluster.Kubernetes.Version = s.nextVersionStr

			mgr.Logger.Infoln(fmt.Sprintf("Start Upgrade: %s -> %s", s.currentVersionStr, s.nextVersionStr))

			if err := preinstall.Prepare(mgr); err != nil {
				return err
			}

			if err := mgr.RunTaskOnK8sNodes(preinstall.PullImages, true); err != nil {
				return err
			}

			if err := mgr.RunTaskOnK8sNodes(s.upgradeNodes, false); err != nil {
				return err
			}

			if err := mgr.RunTaskOnMasterNodes(reconfigDns, false); err != nil {
				return err
			}
			s.currentVersionStr = s.nextVersionStr
		} else {
			break Loop
		}
	}

	return nil

}

func (s *UpgradeStatus) upgradeNodes(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {

	if node.IsMaster {
		if err := s.upgradeKubeMasters(mgr, node); err != nil {
			return err
		}
	} else {
		if err := s.upgradeKubeWorkers(mgr, node); err != nil {
			return nil
		}
	}

	return nil
}

func reconfigDns(mgr *manager.Manager, _ *kubekeyapiv1alpha1.HostCfg) error {
	if mgr.Runner.Index == 0 {
		patchCorednsCmd := `sudo -E /bin/sh -c "/usr/local/bin/kubectl patch deploy -n kube-system coredns -p \" 
spec:
    template:
       spec:
           volumes:
           - name: config-volume
             configMap:
                 name: coredns
                 items:
                 - key: Corefile
                   path: Corefile\""`

		_, _ = mgr.Runner.ExecuteCmd(patchCorednsCmd, 2, true)

		if err := dns.OverrideCorednsService(mgr); err != nil {
			return err
		}
		if err := dns.CreateClusterDns(mgr); err != nil {
			return err
		}
	}
	return nil
}
