package kubernetes

import (
	"encoding/base64"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

type KubernetesStatus struct {
	Version       string
	JoinMasterCmd string
	JoinWorkerCmd string
	Info          string
	KubeConfig    string
	NodesInfo     map[string]string
}

func NewKubernetesStatus() *KubernetesStatus {
	return &KubernetesStatus{NodesInfo: make(map[string]string)}
}

func (k *KubernetesStatus) SearchVersion(runtime connector.Runtime) error {
	cmd := "cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | awk -F '[:]' '{print $(NF-0)}'"
	if output, err := runtime.GetRunner().SudoCmd(cmd, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "search current version failed")
	} else {
		if !strings.Contains(output, "No such file or directory") {
			k.Version = output
		}
	}
	return nil
}

func (k *KubernetesStatus) SearchKubeConfig(runtime connector.Runtime) error {
	kubeCfgBase64Cmd := "cat /etc/kubernetes/admin.conf | base64 --wrap=0"
	if kubeConfigStr, err := runtime.GetRunner().SudoCmd(kubeCfgBase64Cmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "search cluster kubeconfig failed")
	} else {
		k.KubeConfig = kubeConfigStr
	}
	return nil
}

func (k *KubernetesStatus) SearchJoinCmd(runtime connector.Runtime) error {
	uploadCertsCmd := "/usr/local/bin/kubeadm init phase upload-certs --upload-certs"
	output, err := runtime.GetRunner().SudoCmd(uploadCertsCmd, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to upload kubeadm certs")
	}
	reg := regexp.MustCompile("[0-9|a-z]{64}")
	certificateKey := reg.FindAllString(output, -1)[0]

	if err := patchKubeadmSecret(runtime); err != nil {
		return err
	}

	tokenCreateMasterCmd := "/usr/local/bin/kubeadm token create --print-join-command"
	joinCmd, err := runtime.GetRunner().SudoCmd(tokenCreateMasterCmd, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to get join node cmd")
	}

	joinWorkerStrList := strings.Split(joinCmd, "kubeadm join")
	// if "127.0.0.1" in the join command, replace it with the cluster config file's first internal address in master group
	if strings.Contains(joinWorkerStrList[1], "127.0.0.1") {
		joinWorkerStrList[1] = strings.Replace(joinWorkerStrList[1], "127.0.0.1", runtime.GetHostsByRole(common.Master)[0].GetInternalAddress(), 1)
	}

	k.JoinWorkerCmd = fmt.Sprintf("/usr/local/bin/kubeadm join %s", joinWorkerStrList[1])
	k.JoinMasterCmd = fmt.Sprintf("%s --control-plane --certificate-key %s", k.JoinWorkerCmd, certificateKey)
	return nil
}

func (k *KubernetesStatus) SearchInfo(runtime connector.Runtime) error {
	output, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl --no-headers=true get nodes -o custom-columns=:metadata.name,:status.nodeInfo.kubeletVersion,:status.addresses",
		true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get kubernetes cluster info failed")
	}
	k.Info = output
	return nil
}

func (k *KubernetesStatus) SearchNodesInfo(_ connector.Runtime) error {
	ipv4Regexp, err := regexp.Compile(common.IPv4Regexp)
	if err != nil {
		return err
	}
	ipv6Regexp, err := regexp.Compile(common.IPv6Regexp)
	if err != nil {
		return err
	}
	tmp := strings.Split(k.Info, "\r\n")
	if len(tmp) >= 1 {
		for i := 0; i < len(tmp); i++ {
			if ipv4 := ipv4Regexp.FindStringSubmatch(tmp[i]); len(ipv4) != 0 {
				k.NodesInfo[ipv4[0]] = ipv4[0]
			}
			if ipv6 := ipv6Regexp.FindStringSubmatch(tmp[i]); len(ipv6) != 0 {
				k.NodesInfo[ipv6[0]] = ipv6[0]
			}
			if len(strings.Fields(tmp[i])) > 3 {
				k.NodesInfo[strings.Fields(tmp[i])[0]] = strings.Fields(tmp[i])[1]
			} else {
				k.NodesInfo[strings.Fields(tmp[i])[0]] = ""
			}
		}
	}
	return nil
}

func (k *KubernetesStatus) LoadKubeConfig(runtime connector.Runtime, kubeConf *common.KubeConf) error {
	kubeConfigPath := filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.GetObjName()))
	kubeConfigStr, err := base64.StdEncoding.DecodeString(k.KubeConfig)
	if err != nil {
		return err
	}

	oldServer := fmt.Sprintf("server: https://%s:%d", kubeConf.Cluster.ControlPlaneEndpoint.Domain, kubeConf.Cluster.ControlPlaneEndpoint.Port)
	newServer := fmt.Sprintf("server: https://%s:%d", kubeConf.Cluster.ControlPlaneEndpoint.Address, kubeConf.Cluster.ControlPlaneEndpoint.Port)
	newKubeConfigStr := strings.Replace(string(kubeConfigStr), oldServer, newServer, -1)

	if err := ioutil.WriteFile(kubeConfigPath, []byte(newKubeConfigStr), 0644); err != nil {
		return err
	}
	return nil
}

// PatchKubeadmSecret is used to patch etcd's certs for kubeadm-certs secret.
func patchKubeadmSecret(runtime connector.Runtime) error {
	externalEtcdCerts := []string{"external-etcd-ca.crt", "external-etcd.crt", "external-etcd.key"}
	for _, cert := range externalEtcdCerts {
		_, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("/usr/local/bin/kubectl patch -n kube-system secret kubeadm-certs -p '{\\\"data\\\": {\\\"%s\\\": \\\"\\\"}}'", cert),
			true)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "patch kubeadm secret failed")
		}
	}
	return nil
}
