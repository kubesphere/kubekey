package binaries

import (
	"crypto/sha256"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// K8sFilesDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func K8sFilesDownloadHTTP(kubeConf *common.KubeConf, filepath, version, arch string) error {
	kkzone := os.Getenv("KKZONE")
	etcd := files.KubeBinary{Name: "etcd", Arch: arch, Version: kubekeyapiv1alpha1.DefaultEtcdVersion}
	kubeadm := files.KubeBinary{Name: "kubeadm", Arch: arch, Version: version}
	kubelet := files.KubeBinary{Name: "kubelet", Arch: arch, Version: version}
	kubectl := files.KubeBinary{Name: "kubectl", Arch: arch, Version: version}
	kubecni := files.KubeBinary{Name: "kubecni", Arch: arch, Version: kubekeyapiv1alpha1.DefaultCniVersion}
	helm := files.KubeBinary{Name: "helm", Arch: arch, Version: kubekeyapiv1alpha1.DefaultHelmVersion}

	etcd.Path = fmt.Sprintf("%s/etcd-%s-linux-%s.tar.gz", filepath, kubekeyapiv1alpha1.DefaultEtcdVersion, arch)
	kubeadm.Path = fmt.Sprintf("%s/kubeadm", filepath)
	kubelet.Path = fmt.Sprintf("%s/kubelet", filepath)
	kubectl.Path = fmt.Sprintf("%s/kubectl", filepath)
	kubecni.Path = fmt.Sprintf("%s/cni-plugins-linux-%s-%s.tgz", filepath, arch, kubekeyapiv1alpha1.DefaultCniVersion)
	helm.Path = fmt.Sprintf("%s/helm", filepath)

	if kkzone == "cn" {
		etcd.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/etcd/release/download/%s/etcd-%s-linux-%s.tar.gz", etcd.Version, etcd.Version, etcd.Arch)
		kubeadm.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubeadm", kubeadm.Version, kubeadm.Arch)
		kubelet.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubelet", kubelet.Version, kubelet.Arch)
		kubectl.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubectl", kubectl.Version, kubectl.Arch)
		kubecni.Url = fmt.Sprintf("https://containernetworking.pek3b.qingstor.com/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", kubecni.Version, kubecni.Arch, kubecni.Version)
		helm.Url = fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm", helm.Arch, helm.Version)
		helm.GetCmd = kubeConf.Arg.DownloadCommand(helm.Path, helm.Url)
	} else {
		etcd.Url = fmt.Sprintf("https://github.com/coreos/etcd/releases/download/%s/etcd-%s-linux-%s.tar.gz", etcd.Version, etcd.Version, etcd.Arch)
		kubeadm.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubeadm", kubeadm.Version, kubeadm.Arch)
		kubelet.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubelet", kubelet.Version, kubelet.Arch)
		kubectl.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubectl", kubectl.Version, kubectl.Arch)
		kubecni.Url = fmt.Sprintf("https://github.com/containernetworking/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", kubecni.Version, kubecni.Arch, kubecni.Version)
		helm.Url = fmt.Sprintf("https://get.helm.sh/helm-%s-linux-%s.tar.gz", helm.Version, helm.Arch)
		getCmd := kubeConf.Arg.DownloadCommand(fmt.Sprintf("%s/helm-%s-linux-%s.tar.gz", filepath, helm.Version, helm.Arch), helm.Url)
		helm.GetCmd = fmt.Sprintf("%s && cd %s && tar -zxf helm-%s-linux-%s.tar.gz && mv linux-%s/helm . && rm -rf *linux-%s*", getCmd, filepath, helm.Version, helm.Arch, helm.Arch, helm.Arch)
	}

	kubeadm.GetCmd = kubeConf.Arg.DownloadCommand(kubeadm.Path, kubeadm.Url)
	kubelet.GetCmd = kubeConf.Arg.DownloadCommand(kubelet.Path, kubelet.Url)
	kubectl.GetCmd = kubeConf.Arg.DownloadCommand(kubectl.Path, kubectl.Url)
	kubecni.GetCmd = kubeConf.Arg.DownloadCommand(kubecni.Path, kubecni.Url)
	etcd.GetCmd = kubeConf.Arg.DownloadCommand(etcd.Path, etcd.Url)

	binaries := []files.KubeBinary{kubeadm, kubelet, kubectl, helm, kubecni, etcd}

	for _, binary := range binaries {
		logger.Log.Messagef(common.LocalHost, "downloading %s ...", binary.Name)

		if util.IsExist(binary.Path) {
			// download it again if it's incorrect
			if err := K8sSHA256Check(binary, version); err != nil {
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", binary.Path)).Run()
			} else {
				continue
			}
		}

		for i := 5; i > 0; i-- {
			if output, err := exec.Command("/bin/sh", "-c", binary.GetCmd).CombinedOutput(); err != nil {
				fmt.Println(string(output))

				if kkzone != "cn" {
					logger.Log.Warningln("Having a problem with accessing https://storage.googleapis.com? You can try again after setting environment 'export KKZONE=cn'")
				}
				return errors.New(fmt.Sprintf("Failed to download %s binary: %s", binary.Name, binary.GetCmd))
			}

			if err := K8sSHA256Check(binary, version); err != nil {
				if i == 1 {
					return err
				}
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", binary.Path)).Run()
				continue
			}
			break
		}
	}

	if kubeConf.Cluster.KubeSphere.Version == "v2.1.1" {
		logger.Log.Infoln(fmt.Sprintf("Downloading %s ...", "helm2"))
		if util.IsExist(fmt.Sprintf("%s/helm2", filepath)) == false {
			cmd := kubeConf.Arg.DownloadCommand(fmt.Sprintf("%s/helm2", filepath), fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm", helm.Arch, "v2.16.9"))
			if output, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput(); err != nil {
				fmt.Println(string(output))
				return errors.Wrap(err, "Failed to download helm2 binary")
			}
		}
	}

	return nil
}

// K8sSHA256Check is used to hash checks on downloaded binary. (sha256)
func K8sSHA256Check(binary files.KubeBinary, version string) error {
	output, err := sha256sum(binary.Path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to check SHA256 of %s", binary.Path))
	}

	if strings.TrimSpace(binary.GetSha256()) == "" {
		return errors.New(fmt.Sprintf("No SHA256 found for %s. %s is not supported.", version, version))
	}
	if output != binary.GetSha256() {
		return errors.New(fmt.Sprintf("SHA256 no match. %s not equal %s", binary.GetSha256(), output))
	}
	return nil
}

func sha256sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}
