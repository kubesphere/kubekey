package preinstall

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
)

func FilesDownloadHttp(cfg *kubekeyapi.K2ClusterSpec, filepath string, logger *log.Logger) error {
	var kubeVersion string
	if cfg.Kubernetes.Version == "" {
		kubeVersion = kubekeyapi.DefaultKubeVersion
	} else {
		kubeVersion = cfg.Kubernetes.Version
	}

	kubeadmUrl := fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubeadm", kubeVersion, kubekeyapi.DefaultArch)
	kubeletUrl := fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubelet", kubeVersion, kubekeyapi.DefaultArch)
	kubectlUrl := fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubectl", kubeVersion, kubekeyapi.DefaultArch)
	kubeCniUrl := fmt.Sprintf("https://containernetworking.pek3b.qingstor.com/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", kubekeyapi.DefaultCniVersion, kubekeyapi.DefaultArch, kubekeyapi.DefaultCniVersion)
	HelmUrl := fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-amd64/%s/helm", kubekeyapi.DefaultHelmVersion)

	kubeadm := fmt.Sprintf("%s/kubeadm-%s", filepath, kubeVersion)
	kubelet := fmt.Sprintf("%s/kubelet-%s", filepath, kubeVersion)
	kubectl := fmt.Sprintf("%s/kubectl-%s", filepath, kubeVersion)
	kubeCni := fmt.Sprintf("%s/cni-plugins-linux-%s-%s.tgz", filepath, kubekeyapi.DefaultArch, kubekeyapi.DefaultCniVersion)
	helm := fmt.Sprintf("%s/helm-%s", filepath, kubekeyapi.DefaultHelmVersion)

	getKubeadmCmd := fmt.Sprintf("curl -o %s  %s", kubeadm, kubeadmUrl)
	getKubeletCmd := fmt.Sprintf("curl -o %s  %s", kubelet, kubeletUrl)
	getKubectlCmd := fmt.Sprintf("curl -o %s  %s", kubectl, kubectlUrl)
	getKubeCniCmd := fmt.Sprintf("curl -o %s  %s", kubeCni, kubeCniUrl)
	getHelmCmd := fmt.Sprintf("curl -o %s  %s", helm, HelmUrl)

	logger.Info("Downloading Kubeadm ...")
	if util.IsExist(kubeadm) == false {
		if out, err := exec.Command("/bin/sh", "-c", getKubeadmCmd).CombinedOutput(); err != nil {
			fmt.Println(string(out))
			return errors.Wrap(err, "Failed to download kubeadm binary")
		}
	}

	logger.Info("Downloading Kubelet ...")
	if util.IsExist(kubelet) == false {
		if out, err := exec.Command("/bin/sh", "-c", getKubeletCmd).CombinedOutput(); err != nil {
			fmt.Println(string(out))
			return errors.Wrap(err, "Failed to download kubelet binary")
		}
	}

	logger.Info("Downloading Kubectl ...")
	if util.IsExist(kubectl) == false {
		if out, err := exec.Command("/bin/sh", "-c", getKubectlCmd).CombinedOutput(); err != nil {
			fmt.Println(string(out))
			return errors.Wrap(err, "Failed to download kubectl binary")
		}
	}

	logger.Info("Downloading KubeCni ...")
	if util.IsExist(kubeCni) == false {
		if out, err := exec.Command("/bin/sh", "-c", getKubeCniCmd).CombinedOutput(); err != nil {
			fmt.Println(string(out))
			return errors.Wrap(err, "Faild to download kubecni")
		}
	}

	logger.Info("Downloading Helm ...")
	if util.IsExist(helm) == false {
		if out, err := exec.Command("/bin/sh", "-c", getHelmCmd).CombinedOutput(); err != nil {
			fmt.Println(string(out))
			return errors.Wrap(err, "Failed to download helm binary")
		}
	}

	return nil
}

func Prepare(cfg *kubekeyapi.K2ClusterSpec, logger *log.Logger) error {
	logger.Info("Downloading Installation Files")

	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "Faild to get current dir")
	}

	filepath := fmt.Sprintf("%s/%s", currentDir, kubekeyapi.DefaultPreDir)
	if err := util.CreateDir(filepath); err != nil {
		return errors.Wrap(err, "Failed to create download target dir")
	}

	if err := FilesDownloadHttp(cfg, filepath, logger); err != nil {
		return err
	}
	return nil
}
