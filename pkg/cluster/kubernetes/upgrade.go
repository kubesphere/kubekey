package kubernetes

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/kubernetes/tmpl"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
)

var (
	currentVersions   = make(map[string]string)
	currentVersionStr string
	nextVersionStr    string
	mu                sync.Mutex
)

func GetCurrentVersions(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Get current version")
	return mgr.RunTaskOnK8sNodes(getCurrentVersion, true)
}

func getCurrentVersion(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {
	kubeletVersionInfo, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubelet --version\"", 3, false)
	if err != nil {
		return errors.Wrap(err, "Failed to get current kubelet version")
	}
	kubeletVersionStr := strings.Split(kubeletVersionInfo, " ")[1]
	mu.Lock()
	currentVersions[kubeletVersionStr] = kubeletVersionStr
	if minVersion, err := getMinVersion(currentVersions); err != nil {
		return err
	} else {
		currentVersions = make(map[string]string)
		currentVersions[minVersion] = minVersion
		currentVersionStr = fmt.Sprintf("v%s", minVersion)
	}
	mu.Unlock()

	if node.IsMaster {
		apiserverVersionStr, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | cut -d ':' -f 3\"", 3, false)
		if err != nil {
			return errors.Wrap(err, "Failed to get current kube-apiserver version")
		}
		mu.Lock()
		currentVersions[apiserverVersionStr] = apiserverVersionStr
		if minVersion, err := getMinVersion(currentVersions); err != nil {
			return err
		} else {
			currentVersions = make(map[string]string)
			currentVersions[minVersion] = minVersion
			currentVersionStr = fmt.Sprintf("v%s", minVersion)
		}
		mu.Unlock()
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

func upgradeKubeMasters(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {
	kubeletVersion, err := mgr.Runner.ExecuteCmd("/usr/local/bin/kubelet --version", 3, false)
	if err != nil {
		return errors.Wrap(err, "Failed to get current kubelet version")
	}
	if strings.Split(kubeletVersion, " ")[1] != mgr.Cluster.Kubernetes.Version {
		if err := SyncKubeBinaries(mgr, node); err != nil {
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
			kubeadmCfg, err := tmpl.GenerateKubeadmCfg(mgr)
			if err != nil {
				return err
			}
			kubeadmCfgBase64 = base64.StdEncoding.EncodeToString([]byte(kubeadmCfg))
		}

		_, err1 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/kubernetes && echo %s | base64 -d > /etc/kubernetes/kubeadm-config.yaml\"", kubeadmCfgBase64), 1, false)
		if err1 != nil {
			return errors.Wrap(errors.WithStack(err1), "Failed to generate kubeadm config")
		}

		_, err2 := mgr.Runner.ExecuteCmd(fmt.Sprintf(
			"sudo -E /bin/sh -c \"/usr/local/bin/kubeadm upgrade apply -y %s --config=/etc/kubernetes/kubeadm-config.yaml "+
				"--ignore-preflight-errors=all --allow-experimental-upgrades --allow-release-candidate-upgrades --etcd-upgrade=false --certificate-renewal=true --force\"",
			mgr.Cluster.Kubernetes.Version),
			3, false)
		if err2 != nil {
			return errors.Wrap(errors.WithStack(err2), fmt.Sprintf("Failed to upgrade master: %s", node.Name))
		}

		if err := SetKubelet(mgr, node); err != nil {
			return err
		}

		if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl restart kubelet\"", 2, true); err != nil {
			return err
		}
	}

	return nil
}

func upgradeKubeWorkers(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {
	kubeletVersion, err := mgr.Runner.ExecuteCmd("/usr/local/bin/kubelet --version", 3, false)
	if err != nil {
		return errors.Wrap(err, "Failed to get current kubelet version")
	}
	if strings.Split(kubeletVersion, " ")[1] != mgr.Cluster.Kubernetes.Version {

		if err := SyncKubeBinaries(mgr, node); err != nil {
			return err
		}

		if err := SetKubelet(mgr, node); err != nil {
			return err
		}

		if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl restart kubelet\"", 2, true); err != nil {
			return err
		}
	}

	return nil
}

func UpgradeKubeCluster(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Upgrading kube cluster")
	targetVersionStr := mgr.Cluster.Kubernetes.Version
	cmp, err := versionutil.MustParseSemantic(currentVersionStr).Compare(mgr.Cluster.Kubernetes.Version)
	if err != nil {
		return err
	}
	if cmp == 1 {
		mgr.Logger.Warningln(fmt.Sprintf("The current version (%s) is greater than the target version (%s)", currentVersionStr, targetVersionStr))
		os.Exit(0)
	}
Loop:
	for {
		if currentVersionStr != targetVersionStr {
			currentVersion := versionutil.MustParseSemantic(currentVersionStr)
			targetVersion := versionutil.MustParseSemantic(targetVersionStr)

			var nextVersionMinor uint
			if targetVersion.Minor() == currentVersion.Minor() {
				nextVersionMinor = currentVersion.Minor()
			} else {
				nextVersionMinor = currentVersion.Minor() + 1
			}

			if nextVersionMinor == versionutil.MustParseSemantic(targetVersionStr).Minor() {
				nextVersionStr = targetVersionStr
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

				nextVersionStr = fmt.Sprintf("v%s", nextVersion.String())
			}

			mgr.Cluster.Kubernetes.Version = nextVersionStr

			mgr.Logger.Infoln(fmt.Sprintf("Start Upgrade: %s -> %s", currentVersionStr, nextVersionStr))

			preinstall.Prepare(mgr)

			if err := mgr.RunTaskOnK8sNodes(preinstall.PullImages, true); err != nil {
				return err
			}

			if err := mgr.RunTaskOnK8sNodes(upgradeCluster, false); err != nil {
				return err
			}

			currentVersionStr = nextVersionStr
		} else {
			break Loop
		}
	}

	return nil

}

func upgradeCluster(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {
	if node.IsMaster {
		if err := upgradeKubeMasters(mgr, node); err != nil {
			return err
		}
	} else {
		if err := upgradeKubeWorkers(mgr, node); err != nil {
			return nil
		}
	}

	return nil
}
