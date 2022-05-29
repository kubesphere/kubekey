/*
 Copyright 2020 The KubeSphere Authors.

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

package files

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/util/workflow/logger"
	"github.com/kubesphere/kubekey/util/workflow/util"
)

const (
	kubeadm    = "kubeadm"
	kubelet    = "kubelet"
	kubectl    = "kubectl"
	kubecni    = "kubecni"
	etcd       = "etcd"
	helm       = "helm"
	amd64      = "amd64"
	arm64      = "arm64"
	k3s        = "k3s"
	docker     = "docker"
	crictl     = "crictl"
	registry   = "registry"
	harbor     = "harbor"
	compose    = "compose"
	containerd = "containerd"
	runc       = "runc"
)

// KubeBinary Type field const
const (
	CNI        = "cni"
	CRICTL     = "crictl"
	DOCKER     = "docker"
	ETCD       = "etcd"
	HELM       = "helm"
	KUBE       = "kube"
	REGISTRY   = "registry"
	CONTAINERD = "containerd"
	RUNC       = "runc"
)

type KubeBinary struct {
	Type     string
	ID       string
	FileName string
	Arch     string
	Version  string
	Url      string
	BaseDir  string
	Zone     string
	getCmd   func(path, url string) string
}

func NewKubeBinary(name, arch, version, prePath string, getCmd func(path, url string) string) *KubeBinary {
	component := new(KubeBinary)
	component.ID = name
	component.Arch = arch
	component.Version = version
	component.Zone = os.Getenv("KKZONE")
	component.getCmd = getCmd

	switch name {
	case etcd:
		component.Type = ETCD
		component.FileName = fmt.Sprintf("etcd-%s-linux-%s.tar.gz", version, arch)
		component.Url = fmt.Sprintf("https://github.com/coreos/etcd/releases/download/%s/etcd-%s-linux-%s.tar.gz", version, version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf(
				"https://kubernetes-release.pek3b.qingstor.com/etcd/release/download/%s/etcd-%s-linux-%s.tar.gz",
				component.Version, component.Version, component.Arch)
		}
	case kubeadm:
		component.Type = KUBE
		component.FileName = kubeadm
		component.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubeadm", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubeadm", version, arch)
		}
	case kubelet:
		component.Type = KUBE
		component.FileName = kubelet
		component.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubelet", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubelet", version, arch)
		}
	case kubectl:
		component.Type = KUBE
		component.FileName = kubectl
		component.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubectl", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubectl", version, arch)
		}
	case kubecni:
		component.Type = CNI
		component.FileName = fmt.Sprintf("cni-plugins-linux-%s-%s.tgz", arch, version)
		component.Url = fmt.Sprintf("https://github.com/containernetworking/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", version, arch, version)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://containernetworking.pek3b.qingstor.com/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", version, arch, version)
		}
	case helm:
		component.Type = HELM
		component.FileName = helm
		component.Url = fmt.Sprintf("https://get.helm.sh/helm-%s-linux-%s.tar.gz", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm", arch, version)
		}
	case docker:
		component.Type = DOCKER
		component.FileName = fmt.Sprintf("docker-%s.tgz", version)
		component.Url = fmt.Sprintf("https://download.docker.com/linux/static/stable/%s/docker-%s.tgz", util.ArchAlias(arch), version)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://mirrors.aliyun.com/docker-ce/linux/static/stable/%s/docker-%s.tgz", util.ArchAlias(arch), version)
		}
	case crictl:
		component.Type = CRICTL
		component.FileName = fmt.Sprintf("crictl-%s-linux-%s.tar.gz", version, arch)
		component.Url = fmt.Sprintf("https://github.com/kubernetes-sigs/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz", version, version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz", version, version, arch)
		}
	case k3s:
		component.Type = KUBE
		component.FileName = k3s
		component.Url = fmt.Sprintf("https://github.com/k3s-io/k3s/releases/download/%s+k3s1/k3s", version)
		if arch == arm64 {
			component.Url = fmt.Sprintf("https://github.com/k3s-io/k3s/releases/download/%s+k3s1/k3s-%s", version, arch)
		}
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/k3s/releases/download/%s+k3s1/linux/%s/k3s", version, arch)
		}
	case registry:
		component.Type = REGISTRY
		component.FileName = fmt.Sprintf("registry-%s-linux-%s.tar.gz", version, arch)
		component.Url = fmt.Sprintf("https://github.com/kubesphere/kubekey/releases/download/v2.0.0-alpha.1/registry-%s-linux-%s.tar.gz", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/registry/%s/registry-%s-linux-%s.tar.gz", version, version, arch)
		}
		component.BaseDir = filepath.Join(prePath, component.Type, component.ID, component.Version, component.Arch)
	case harbor:
		component.Type = REGISTRY
		component.FileName = fmt.Sprintf("harbor-offline-installer-%s.tgz", version)
		component.Url = fmt.Sprintf("https://github.com/goharbor/harbor/releases/download/%s/harbor-offline-installer-%s.tgz", version, version)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/harbor/releases/download/%s/harbor-offline-installer-%s.tgz", version, version)
		}
		component.BaseDir = filepath.Join(prePath, component.Type, component.ID, component.Version, component.Arch)
	case compose:
		component.Type = REGISTRY
		component.FileName = "docker-compose-linux-x86_64"
		component.Url = fmt.Sprintf("https://github.com/docker/compose/releases/download/%s/docker-compose-linux-x86_64", version)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/docker/compose/releases/download/%s/docker-compose-linux-x86_64", version)
		}
		component.BaseDir = filepath.Join(prePath, component.Type, component.ID, component.Version, component.Arch)
	case containerd:
		component.Type = CONTAINERD
		component.FileName = fmt.Sprintf("containerd-%s-linux-%s.tar.gz", version, arch)
		component.Url = fmt.Sprintf("https://github.com/containerd/containerd/releases/download/v%s/containerd-%s-linux-%s.tar.gz", version, version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/containerd/containerd/releases/download/v%s/containerd-%s-linux-%s.tar.gz", version, version, arch)
		}
	case runc:
		component.Type = RUNC
		component.FileName = fmt.Sprintf("runc.%s", arch)
		component.Url = fmt.Sprintf("https://github.com/opencontainers/runc/releases/download/%s/runc.%s", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/opencontainers/runc/releases/download/%s/runc.%s", version, arch)
		}
	default:
		logger.Log.Fatalf("unsupported kube binaries %s", name)
	}

	if component.BaseDir == "" {
		component.BaseDir = filepath.Join(prePath, component.Type, component.Version, component.Arch)
	}

	return component
}

func (b *KubeBinary) CreateBaseDir() error {
	if err := util.CreateDir(b.BaseDir); err != nil {
		return err
	}
	return nil
}

func (b *KubeBinary) Path() string {
	return filepath.Join(b.BaseDir, b.FileName)
}

func (b *KubeBinary) GetCmd() string {
	cmd := b.getCmd(b.Path(), b.Url)

	if b.ID == helm && b.Zone != "cn" {
		get := b.getCmd(filepath.Join(b.BaseDir, fmt.Sprintf("helm-%s-linux-%s.tar.gz", b.Version, b.Arch)), b.Url)
		cmd = fmt.Sprintf("%s && cd %s && tar -zxf helm-%s-linux-%s.tar.gz && mv linux-%s/helm . && rm -rf *linux-%s*",
			get, b.BaseDir, b.Version, b.Arch, b.Arch, b.Arch)
	}
	return cmd
}

func (b *KubeBinary) GetSha256() string {
	s := FileSha256[b.ID][b.Arch][b.Version]
	return s
}

func (b *KubeBinary) Download() error {
	for i := 5; i > 0; i-- {
		cmd := exec.Command("/bin/sh", "-c", b.GetCmd())
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		cmd.Stderr = cmd.Stdout

		if err = cmd.Start(); err != nil {
			return err
		}
		for {
			tmp := make([]byte, 1024)
			_, err := stdout.Read(tmp)
			fmt.Print(string(tmp)) // Get the output from the pipeline in real time and print it to the terminal
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				logger.Log.Errorln(err)
				break
			}
		}
		if err = cmd.Wait(); err != nil {
			if os.Getenv("KKZONE") != "cn" {
				logger.Log.Warningln("Having a problem with accessing https://storage.googleapis.com? You can try again after setting environment 'export KKZONE=cn'")
			}
			return err
		}

		if err := b.SHA256Check(); err != nil {
			if i == 1 {
				return err
			}
			path := b.Path()
			_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", path)).Run()
			continue
		}
		break
	}
	return nil
}

// SHA256Check is used to hash checks on downloaded binary. (sha256)
func (b *KubeBinary) SHA256Check() error {
	output, err := sha256sum(b.Path())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to check SHA256 of %s", b.Path()))
	}

	if strings.TrimSpace(b.GetSha256()) == "" {
		return errors.New(fmt.Sprintf("No SHA256 found for %s. %s is not supported.", b.ID, b.Version))
	}
	if output != b.GetSha256() {
		return errors.New(fmt.Sprintf("SHA256 no match. %s not equal %s", b.GetSha256(), output))
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

var (
	FileSha256 = map[string]map[string]map[string]string{
		kubeadm: {
			amd64: {
				"v1.15.12": "e052bae41e731921a9197b4d078c30d33fac5861716dc275bfee4670addbac9b",
				"v1.16.8":  "58a74986af13b969abc8b471822f36f3fda71f95ed1c006f48c8d2ab88f8edf1",
				"v1.16.10": "726d42c569f25078d03b758477f17f543c845aef2ff48acd9d4269705ca1aa9d",
				"v1.16.12": "bb4d0f045600b883745016416c14533f823d582f4f20df691b7f79a6545b6480",
				"v1.16.13": "3ddce3fb919f1e8b0a3e0a1ae1d20c9af0fd4a7d731be1e818597b3ecdb49023",
				"v1.17.0":  "0d8443f50fb7caab2e5e7e53f9dc56d5ffe55f021ec061f2e2bcba0481df5a48",
				"v1.17.4":  "3cdcffcf8a1660241a045cfdfed3ebbf7f7c6a0840f008e2b049b533bca5bb8c",
				"v1.17.5":  "9bd2fd1118b3d07d12e2a806c04bf34d99e79886c5318ddc003ba38f30da390c",
				"v1.17.6":  "d4cfc9a0a734ba015594974ee4253b8965b95cdb6e83d8a6a946675aad418b40",
				"v1.17.7":  "9d4b97e93ddb204798b91fec063743e218c92b42798779b5248a49e1476226e2",
				"v1.17.8":  "c59b85696c4cbabe896ba71f4bbc99e4ad2444fcea851e3ee740705584420aad",
				"v1.17.9":  "5ef1660d3d56e93e3d87d6a7028aa64745984be0b0678c45c32f66043b4d69b4",
				"v1.18.3":  "a60974e9840e006076d204fd4ddcba96213beba10fb89ff01882095546c9684d",
				"v1.18.5":  "e428fc9d1cf860090346a83eb66082c3be6b6032f0db9e4f8e6d52492d46231f",
				"v1.18.6":  "11b4180b9f82a8b6bb30250e3d7341b104521f3b654076b8569853ec9451b2a9",
				"v1.18.8":  "27c8f4d4398d57762998b157d35802a36a7ea9b2b6f9a363c397a9d65b4f3c89",
				"v1.19.0":  "88ce7dc5302d8847f6e679aab9e4fa642a819e8a33d70731fb7bc8e110d8659f",
				"v1.19.8":  "9c6646cdf03efc3194afc178647205195da4a43f58d0b70954953f566fa15c76",
				"v1.19.9":  "917712bbd38b625aca456ffa78bf134d64f0efb186cc5772c9844ba6d74fd920",
				"v1.20.4":  "dcc5629da2c31a000b9b50db077b1cd51a6840e08233fd64b67e37f3f098c392",
				"v1.20.6":  "ff6fca46edeccd8a4dbf162079d0b3d27841b04885b3f47f80377b3a93ab1533",
				"v1.20.10": "da5864968a38e0bf2317965e87b5425e1b9101a49dd5178f2e967c0a46547270",
				"v1.21.4":  "286794aed41148e82a77087d79111052ea894796c6ae81fc463275dcd848f98d",
				"v1.21.5":  "e384171fcb3c0de924904007bfd7babb0f970997b93223ed7ffee14d29019353",
				"v1.21.6":  "fef4b40acd982da99294be07932eabedd476113ce5dc38bb9149522e32dada6d",
				"v1.22.1":  "50a5f0d186d7aefae309539e9cc7d530ef1a9b45ce690801655c2bee722d978c",
				"v1.22.9":  "e3061f3a9c52bff82ae740c928fe389a256964a5756d691758bf3611904d7183",
				"v1.23.0":  "e21269a058d4ad421cf5818d4c7825991b8ba51cd06286932a33b21293b071b0",
				"v1.23.6":  "9213c7d738e86c9a562874021df832735236fcfd5599fd4474bab3283d34bfd7",
				"v1.24.0":  "5e58a29eaaf69ea80e90d9780d2a2d5f189fd74f94ec3bec9e3823d472277318",
			},
			arm64: {
				"v1.15.12": "dfc1af35cccac89099a7e9a48dcc4b0d956a8b1d4dfcdd2be12191b6f6c384a3",
				"v1.16.8":  "2300e2a7dc16512595c7aebc486799239039d33f33db2d085550d1f2d5f3129b",
				"v1.16.12": "67f675f8fb1ff3af56ca0a976323a65cabc35efa53b7896146684b8f53990741",
				"v1.16.13": "bb4d0f045600b883745016416c14533f823d582f4f20df691b7f79a6545b6480",
				"v1.17.0":  "0b94d1ace240a8f9995358ca2b66ac92072e3f3cd0543275b315dcd317798546",
				"v1.17.7":  "6c8622adf5a7a2dfc66ebe15058353b2e2660b01f1e8990bab7a9c7fca76bccb",
				"v1.17.8":  "5a52e7d0306890e68ed66fc47ecd70bf14628c70527442fd0cd2973dbde7064c",
				"v1.17.9":  "b56dc03177636fdafb4f8ab329d087b804cb7395c142f76e8246e86083c6d750",
				"v1.18.5":  "0e2a9de622177015c2514498382b0d821ac8f71c7ed5f02e5684d456ff3c0e4d",
				"v1.18.6":  "df5a3d7c70c3f8221d57093c5cb17558aad6e65725d7a096c6620302fbf64730",
				"v1.18.8":  "71f6d95f165a9e8066c6f299217af779829ab3d798f6130caf6daa4784dc0464",
				"v1.19.0":  "db1c432646e6e6484989b6f7191f3610996ac593409f12574290bfc008ea11f5",
				"v1.19.8":  "dfb838ffb88d79e4d881326f611ae5e5999accb54cdd666c75664da264b5d58e",
				"v1.19.9":  "403c767bef0d681aebc45d5643787fc8c0b9344866cbd339368637a05ea1d11c",
				"v1.20.4":  "c3ff7f944826889a23a002c85e8f9f9d9a8bc95e9083fbdda59831e3e34245a7",
				"v1.20.6":  "33837e290bd76fcb16af27db0e814ec023c25e6c41f25a0907b48756d4a2ffc2",
				"v1.20.10": "ec1f8df0f57b8aa6bddce2d6bb8d0503e016b022ba8a5f113ddf412d9a99c03c",
				"v1.21.4":  "30645f57296281d214a9dd787a90bd16207df4b1fca7ac320913c616818a92cd",
				"v1.21.5":  "5a273b023eaa60d7820436b0f0062c4bd467274d6f2b86a9e13270c91d663618",
				"v1.21.6":  "498325da2521ce67b27902967daf4087153c5797070e03bf0bdd7c846f4d61a8",
				"v1.22.1":  "85df7978b2e5bb78064ed0bcce14a39d105a1a3968bb92ee5d2f96a1fa09ed12",
				"v1.22.9":  "0168c60d1997435b006b17c95a1d42e55743048cc50ee16c8774498aa203a202",
				"v1.23.0":  "989d117128dcaa923b2c7a917a03f4836c1b023fe1ee723541e0e39b068b93a6",
				"v1.23.6":  "a4db7458e224c3a2a7b468fc2704b31fec437614914b26a9e3d9efb6eecf61ee",
				"v1.24.0":  "3e0fa21b8ebce04ca919fdfea7cc756e5f645166b95d6e4b5d9912d7721f9004",
			},
		},
		kubelet: {
			amd64: {
				"v1.15.12": "dff48393a3116b8f7dea206b81678e52f7fad298f1aff976f18f1bfa4e9ccdde",
				"v1.16.8":  "4573da19fed14c84f4434ab7cbedf5ded4bf89710c078d58c0703cf2332df198",
				"v1.16.10": "82b38f444d11c2436040165b1addf46d0909a6daec9133cc979678835ef8e14b",
				"v1.16.12": "fbc8c16b148dbb3234a3e13f80e6c6736557c10f8c046edfb1dc5337fe2dd40f",
				"v1.16.13": "a88c0e9f8c4b5a2e91c2c4a8d772cc65ca3a0eb5d477cbce06fbf82d3e50c158",
				"v1.17.0":  "c2af77f501c3164e80171903028d35c632366f53dec0c8419828d4e55d86146f",
				"v1.17.4":  "f3a427ddf610b568db60c8d47565041901220e1bbe257614b61bb4c76801d765",
				"v1.17.5":  "c5fbfa83444bdeefb51934c29f0b4b7ffc43ce5a98d7f957d8a11e3440055383",
				"v1.17.6":  "4b7fd5123bfafe2249bf91ed83469c2655a8d3295966e5fbd952f89b64b75f57",
				"v1.17.7":  "a6b66c94a37dd6ae830a9af5b9200884a2c0af868096a3c2553b2e876723c2a2",
				"v1.17.8":  "b39081fb40332ae12d262b04dc81630e5c6550fb196f09b60f3d726283dff17f",
				"v1.17.9":  "3b6cdfcd38a646c7b553821ef9bb67e93541da658305c00705e6ab2ba15e73af",
				"v1.18.3":  "6aac8853028a4f185de5ccb5b41b3fbd87726161445dee56f351e3e51442d669",
				"v1.18.5":  "8c328f65d30f0edd0fd4f529b09d6fc588cfb7b524d5c9f181e36de6e494e19c",
				"v1.18.6":  "2eb9baf5a65a7b94c653dbd7af03a768a520961eb27ef369e43ef12711e22d4a",
				"v1.18.8":  "a4116675ac52bf80e224fba8ff6db6f2d7aed192bf6fffd5f8e4d5efb4368f31",
				"v1.19.0":  "3f03e5c160a8b658d30b34824a1c00abadbac96e62c4d01bf5c9271a2debc3ab",
				"v1.19.8":  "f5cad5260c29584dd370ec13e525c945866957b1aaa719f1b871c31dc30bcb3f",
				"v1.19.9":  "296e72c395f030209e712167fc5f6d2fdfe3530ca4c01bcd9bfb8c5e727c3d8d",
				"v1.20.4":  "a9f28ac492b3cbf75dee284576b2e1681e67170cd36f3f5cdc31495f1bdbf809",
				"v1.20.6":  "7688a663dd06222d337c8fdb5b05e1d9377e6d64aa048c6acf484bc3f2a596a8",
				"v1.20.10": "de1b24f33d47cc4dc14a10f051d7d6fbbcf3800d3a07ddb45fc83660183c3a73",
				"v1.21.4":  "cdd46617d1a501531c62421de3754d65f30ad24d75beae2693688993a12bb557",
				"v1.21.5":  "600f70fe0e69151b9d8ac65ec195bcc840687f86ba397fce27be1faae3538a6f",
				"v1.21.6":  "422c29a1ba3bfeb2fc26ebd1c3596847fbbeeeef0ce2694515504513dc907813",
				"v1.22.1":  "2079780ad2ff993affc9b8e1a378bf5ee759bf87fdc446e6a892a0bbd7353683",
				"v1.22.9":  "61530a9e6a5cb1f971295de860a8ade29db65d0dff50d1ffff3de1155dfd0c02",
				"v1.23.0":  "4756ff345dd80704b749d87efb8eb294a143a1f4a251ec586197d26ad20ea518",
				"v1.23.6":  "fbb83e35f6b9f7cae19c50694240291805ca9c4028676af868306553b3e9266c",
				"v1.24.0":  "3d98ac8b4fb8dc99f9952226f2565951cc366c442656a889facc5b1b2ec2ba52",
			},
			arm64: {
				"v1.15.12": "c7f586a77acdb3c3e27a6b3bd749760538b830414575f8718f03f7ce53b138d8",
				"v1.16.8":  "a6889c9957d8ec3ba15676b1e2eff021c9d120284f185d367626763dd15a245b",
				"v1.16.12": "0ef9d42e27bf85e9ff276f2181e17e2912941c3a7ae9086de722ac3c9cea997f",
				"v1.16.13": "bb4d0f045600b883745016416c14533f823d582f4f20df691b7f79a6545b6480",
				"v1.17.0":  "b1a4a2325383854a69ec768e7dc00f69378d3ccbc554859d910bf5b582264ea2",
				"v1.17.7":  "eb1715a745281f6aee34644653f73787acdd9f3904e3d58e1319ded4a16be013",
				"v1.17.8":  "673355f62aa422915682ae595e4e53813e4656f2c272eb032f97492211cfced5",
				"v1.17.9":  "d57c25a3d67c937a9d6778de07295478185f73938937868525030a01d15c372f",
				"v1.18.5":  "c3815bc740755aa9fd3ec240ad808a13628a4deb6ec2b4338e772fd0cf77e1a2",
				"v1.18.6":  "257fd42be375025fb93724bda9bef23b73eb40531f22bab9e19f6d6ff1ca57cf",
				"v1.18.8":  "d36e2d656bad232e8b48b19c948164ee3966669f4566cf5ea43ca22f6eed1aa5",
				"v1.19.0":  "d8fa5a9739ecc387dfcc55afa91ac6f4b0ccd01f1423c423dbd312d787bbb6bf",
				"v1.19.8":  "a00146c16266d54f961c40fc67f92c21967596c2d730fa3dc95868d4efb44559",
				"v1.19.9":  "796f080c53ec50b11152558b4a744432349b800e37b80516bcdc459152766a4f",
				"v1.20.4":  "66bcdc7521e226e4acaa93c08e5ea7b2f57829e1a5b9decfd2b91d237e216e1d",
				"v1.20.6":  "6e7b44d1ca65f970b0646f7d093dcf0cfefc44d4a67f29d542fe1b7ca6dcf715",
				"v1.20.10": "5107a4b2eb017039dda900cf263ec19484eee8bec070fc88803d3d9d4cc9fb18",
				"v1.21.4":  "12c849ccc627e9404187adf432a922b895c8bdecfd7ca901e1928396558eb043",
				"v1.21.5":  "746a535956db55807ef71772d2a4afec5cc438233da23952167ec0aec6fe937b",
				"v1.21.6":  "041441623c31bc6b0295342b8a2a5930d87545473e7c761ea79f3ff186c0ff52",
				"v1.22.1":  "d5ffd67d8285fb224a1c49622fd739131f7b941e3d68f233dec96e72c9ebee63",
				"v1.22.9":  "d7a692ee4f5f5929a15c61947ae2deecb71b0945461f6064ced83d13094028e8",
				"v1.23.0":  "a546fb7ccce69c4163e4a0b19a31f30ea039b4e4560c23fd6e3016e2b2dfd0d9",
				"v1.23.6":  "11a0310e8e7af5a11539ac26d6c14cf1b77d35bce4ca74e4bbd053ed1afc8650",
				"v1.24.0":  "8f066c9a048dd1704bf22ccf6e994e2fa2ea1175c9768a786f6cb6608765025e",
			},
		},
		kubectl: {
			amd64: {
				"v1.15.12": "a32b762279c33cb8d8f4198f3facdae402248c3164e9b9b664c3afbd5a27472e",
				"v1.16.8":  "1d8602496ca4b843824a9746206509991eb8d30b5bb8436b36a02718729934ed",
				"v1.16.10": "246d36e4ce67e74e95ff2ba578b9189f58e5def0e8830a24cd30fa3cf279742f",
				"v1.16.12": "db72e5c90de59e1bf287bef55eaf0b603c8d74b3dc552f356ccc02b08c2eb348",
				"v1.16.13": "ab861ec3ec347062bd1b87f8d78d15cd1ce251e74c5fe662e434056962d2a2c9",
				"v1.17.0":  "6e0aaaffe5507a44ec6b1b8a0fb585285813b78cc045f8804e70a6aac9d1cb4c",
				"v1.17.4":  "465b2d2bd7512b173860c6907d8127ee76a19a385aa7865608e57a5eebe23597",
				"v1.17.5":  "03cd1fa19f90d38005148793efdb17a9b58d01dedea641a8496b9cf228db3ab4",
				"v1.17.6":  "5e245f6af6fb761fbe4b3ac06b753f33b361ce0486c48c85b45731a7ee5e4cca",
				"v1.17.7":  "7124a296518edda2ae326e754aec9be6d0ac86131e6f61b52f5ecaa413b66ae4",
				"v1.17.8":  "01283cbc2b09555cbf2a71c162097552a62a4fd48a0a4c06e34e9b853b815486",
				"v1.17.9":  "2ca83eecd221bedf3eceb0ccfcf45bb2e27950c382c2326211303adb0a9c4232",
				"v1.18.3":  "6fcf70aae5bc64870c358fac153cdfdc93f55d8bae010741ecce06bb14c083ea",
				"v1.18.5":  "69d9b044ffaf544a4d1d4b40272f05d56aaf75d7e3c526d5418d1d3c78249e45",
				"v1.18.6":  "62fcb9922164725c7cba5747562f2ad2f4d834ad0a458c1e4c794cc203dcdfb3",
				"v1.18.8":  "a076f5eff0710de94d1eb77bee458ea43b8f4d9572bbb3a3aec1edf0dde0a3e7",
				"v1.19.0":  "79bb0d2f05487ff533999a639c075043c70a0a1ba25c1629eb1eef6ebe3ba70f",
				"v1.19.8":  "a0737d3a15ca177816b6fb1fd59bdd5a3751bfdc66de4e08dffddba84e38bf3f",
				"v1.19.9":  "7128c9e38ab9c445a3b02d3d0b3f0f15fe7fbca56fd87b84e575d7b29e999ad9",
				"v1.20.4":  "98e8aea149b00f653beeb53d4bd27edda9e73b48fed156c4a0aa1dabe4b1794c",
				"v1.20.6":  "89ae000df6bbdf38ae4307cc4ecc0347d5c871476862912c0a765db9bf05284e",
				"v1.20.10": "1e87edb99b7a92a142b458976ae75412d3ee22421793968b03213ddd007c0530",
				"v1.21.4":  "9410572396fb31e49d088f9816beaebad7420c7686697578691be1651d3bf85a",
				"v1.21.5":  "060ede75550c63bdc84e14fcc4c8ab3017f7ffc032fc4cac3bf20d274fab1be4",
				"v1.21.6":  "810eadc2673e0fab7044f88904853e8f3f58a4134867370bf0ccd62c19889eaa",
				"v1.22.1":  "78178a8337fc6c76780f60541fca7199f0f1a2e9c41806bded280a4a5ef665c9",
				"v1.22.9":  "ae6a9b585f9a366d24bb71f508bfb9e2bb90822136138109d3a91cd28e6563bb",
				"v1.23.0":  "2d0f5ba6faa787878b642c151ccb2c3390ce4c1e6c8e2b59568b3869ba407c4f",
				"v1.23.6":  "703a06354bab9f45c80102abff89f1a62cbc2c6d80678fd3973a014acc7c500a",
				"v1.24.0":  "94d686bb6772f6fb59e3a32beff908ab406b79acdfb2427abdc4ac3ce1bb98d7",
			},
			arm64: {
				"v1.15.12": "ef9a4272d556851c645d6788631a2993823260a7e1176a281620284b4c3406da",
				"v1.16.8":  "d08aab5f02db63690672e5d9052659589301323c010d90734788d5332ac99daa",
				"v1.16.12": "7f493dcf9d4edfeea68284c4cd7c74383be23f24e9aefd59c08dc37bc20b46db",
				"v1.16.13": "bb4d0f045600b883745016416c14533f823d582f4f20df691b7f79a6545b6480",
				"v1.17.0":  "cba12bfe0ee447b06f00813d7d4ba3fbdbf5116eccc4d3291987044f2d6f93c2",
				"v1.17.7":  "00c71ceffa9b50af081d2838b102be49ca224a8aa928f5c948b804af84c58818",
				"v1.17.8":  "4dfd36dbd637b8dca9a7c4e789fb3fe4ca420062c90d3a872ae751dfb9777cb6",
				"v1.17.9":  "4d818e97073113eb1e62bf97d63876757be0f273c47807c09f34511155e25afd",
				"v1.18.5":  "28c1edb2d76f80e70e10fa8cd2a30b9fccc5f003d8b3e853535d8317db7f424a",
				"v1.18.6":  "7b3d6cc019747a7ee5f6cc2b187423daaac4e153140cb290e60d316c3f456430",
				"v1.18.8":  "9046c4086528427462544e1a6dcbe709de4d7ae44d1a155375de330fecd067b1",
				"v1.19.0":  "d4adf1b6b97252025cb2f7febf55daa3f42dc305822e3da133f77fd33071ec2f",
				"v1.19.8":  "8f037ab2aa798bbc66ebd1d52653f607f223b07813bcf98d9c1d0c0e136910ec",
				"v1.19.9":  "628627d01c9eaf624ffe3cf1195947a256ea5f842851e42682057e4233a9e283",
				"v1.20.4":  "0fd64b3e5d3fda4637c174a5aea0119b46d6cbede591a4dc9130a81481fc952f",
				"v1.20.6":  "1d0a29420c4488b15adb44044b193588989b95515cd6c8c03907dafe9b3d53f3",
				"v1.20.10": "e559bcf16c824a2337125f20a2d64bfbf3959c713aa4f711871a694e2f58d4d8",
				"v1.21.4":  "8ac78de847118c94e2d87844e9b974556dfb30aff0e0d15fd03b82681df3ac98",
				"v1.21.5":  "fca8de7e55b55cceab9902aae03837fb2f1e72b97aa09b2ac9626bdbfd0466e4",
				"v1.21.6":  "a193997181cdfa00be0420ac6e7f4cfbf6cedd6967259c5fda1d558fa9f4efe0",
				"v1.22.1":  "5c7ef1e505c35a8dc0b708f6b6ecdad6723875bb85554e9f9c3fe591e030ae5c",
				"v1.22.9":  "33724bed4dddf4d8ecd6ae75667552d121e2fb575ff2db427ce66516e048edac",
				"v1.23.0":  "1d77d6027fc8dfed772609ad9bd68f611b7e4ce73afa949f27084ad3a92b15fe",
				"v1.23.6":  "4be771c8e6a082ba61f0367077f480237f9858ef5efe14b1dbbfc05cd42fc360",
				"v1.24.0":  "449278789de283648e4076ade46816da249714f96e71567e035e9d17e1fff06d",
			},
		},
		etcd: {
			amd64: {
				"v3.4.13": "2ac029e47bab752dacdb7b30032f230f49e2f457cbc32e8f555c2210bb5ff107",
			},
			arm64: {
				"v3.4.13": "1934ebb9f9f6501f706111b78e5e321a7ff8d7792d3d96a76e2d01874e42a300",
			},
		},
		helm: {
			amd64: {
				"v3.2.1": "98c57f2b86493dd36ebaab98990e6d5117510f5efbf21c3344c3bdc91a4f947c",
				"v3.6.3": "6e5498e0fa82ba7b60423b1632dba8681d629e5a4818251478cb53f0b71b3c82",
			},
			arm64: {
				"v3.2.1": "20bb9d66e74f618cd104ca07e4525a8f2f760dd6d5611f7d59b6ac574624d672",
				"v3.6.3": "fce1f94dd973379147bb63d8b6190983ad63f3a1b774aad22e54d2a27049414f",
			},
		},
		kubecni: {
			amd64: {
				"v0.8.2": "21283754ffb953329388b5a3c52cef7d656d535292bda2d86fcdda604b482f85",
				"v0.8.6": "994fbfcdbb2eedcfa87e48d8edb9bb365f4e2747a7e47658482556c12fd9b2f5",
				"v0.9.1": "962100bbc4baeaaa5748cdbfce941f756b1531c2eadb290129401498bfac21e7",
			},
			arm64: {
				"v0.8.6": "43fbf750c5eccb10accffeeb092693c32b236fb25d919cf058c91a677822c999",
				"v0.9.1": "ef17764ffd6cdcb16d76401bac1db6acc050c9b088f1be5efa0e094ea3b01df0",
			},
		},
		k3s: {
			amd64: {
				"v1.20.2": "ce3055783cf115ee68fc00bb8d25421d068579ece2fafa4ee1d09f3415aaeabf",
				"v1.20.4": "1c7b68b0b7d54f21a9c1727545a7db181668115f161a3986bc137261dd817e98",
				"v1.21.4": "47e686ad5390670da79a467ba94399d72e472364bc064a20fecd3937a8d928b5",
				"v1.21.6": "89eb5f3d12524d0a9d5b56ba3e2707b106e1731dd0e6d2e7b898ac585f4959df",
			},
			arm64: {
				"v1.21.4": "b7f8c026c5346b3e894d731f1dc2490cd7281687549f34c28a849f58c62e3e48",
				"v1.21.6": "1f06a2da0e1e8596220a5504291ce69237979ebf520e2458c2d72573945a9c1d",
			},
		},
		docker: {
			amd64: {
				"20.10.8": "7ea11ecb100fdc085dbfd9ab1ff380e7f99733c890ed815510a5952e5d6dd7e0",
			},
			arm64: {
				"20.10.8": "4eb9d5e2adf718cd7ee59f6951715f3113c9c4ee49c75c9efb9747f2c3457b2b",
			},
		},
		containerd: {
			amd64: {
				"1.6.2": "3d94f887de5f284b0d6ee61fa17ba413a7d60b4bb27d756a402b713a53685c6a",
				"1.6.4": "f23c8ac914d748f85df94d3e82d11ca89ca9fe19a220ce61b99a05b070044de0",
			},
			arm64: {
				"1.6.2": "a4b24b3c38a67852daa80f03ec2bc94e31a0f4393477cd7dc1c1a7c2d3eb2a95",
				"1.6.4": "0205bd1907154388dc85b1afeeb550cbb44c470ef4a290cb1daf91501c85cae6",
			},
		},
		runc: {
			amd64: {
				"v1.1.1": "5798c85d2c8b6942247ab8d6830ef362924cd72a8e236e77430c3ab1be15f080",
			},
			arm64: {
				"v1.1.1": "20c436a736547309371c7ac2a335f5fe5a42b450120e497d09c8dc3902c28444",
			},
		},
		crictl: {
			amd64: {
				"v1.22.0": "45e0556c42616af60ebe93bf4691056338b3ea0001c0201a6a8ff8b1dbc0652a",
				"v1.23.0": "b754f83c80acdc75f93aba191ff269da6be45d0fc2d3f4079704e7d1424f1ca8",
				"v1.24.0": "3df4a4306e0554aea4fdc26ecef9eea29a58c8460bebfaca3405799787609880",
			},
			arm64: {
				"v1.22.0": "a713c37fade0d96a989bc15ebe906e08ef5c8fe5e107c2161b0665e9963b770e",
				"v1.23.0": "91094253e77094435027998a99b9b6a67b0baad3327975365f7715a1a3bd9595",
				"v1.24.0": "b6fe172738dfa68ca4c71ade53574e859bf61a3e34d21b305587b1ad4ab28d24",
			},
		},
		registry: {
			amd64: {
				"2": "7706e46674fa2cf20f734dfb7e4dd7f1390710e9c0a2c520563e3c55f3e4b5c5",
			},
			arm64: {
				"2": "a6e98123b850da5f6476c08b357e504de352a00f656279ec2636625d352abd5a",
			},
		},
		compose: {
			amd64: {
				"v2.2.2": "92551cd3d22b41536ce8345fe06795ad0d08cb3c17b693ecbfe41176e501bfd4",
			},
		},
		harbor: {
			amd64: {
				"v2.4.1": "cfd799c150b59353aefb34835f3a2e859763cb2e91966cd3ffeb1b6ceaa19841",
			},
		},
	}
)
