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
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
				"v1.21.0":  "7bdaf0d58f0d286538376bc40b50d7e3ab60a3fe7a0709194f53f1605129550f",
				"v1.21.1":  "1553c07a6a777c4cf71d45d5892915f0ea6586b8a80f9fea39e7a659d6315d42",
				"v1.21.2":  "6a83e52e51f41d67658a13ce8ac9deb77a6d82a71ced2d106756f6d38756ec00",
				"v1.21.3":  "82fff4fc0cdb1110150596ab14a3ddcd3dbe53f40c404917d2e9703f8f04787a",
				"v1.21.4":  "286794aed41148e82a77087d79111052ea894796c6ae81fc463275dcd848f98d",
				"v1.21.5":  "e384171fcb3c0de924904007bfd7babb0f970997b93223ed7ffee14d29019353",
				"v1.21.6":  "fef4b40acd982da99294be07932eabedd476113ce5dc38bb9149522e32dada6d",
				"v1.21.7":  "c4480121b629a0f563f718aa11440ae26a569e37e0229c093a5785c90725a03c",
				"v1.21.8":  "51d266e91e2aec0e994c046b4d80901a1b1e7be05e30b83461f0563571f1224d",
				"v1.21.9":  "3333116f9f0d72e0598f52dcbef7ecab1ce88192fdcfd5384ca919fdc075e8d5",
				"v1.21.10": "61aaadd98806d979b65e031a144d9379390d26ccb5383d47bdd8b7c727e94a7b",
				"v1.21.11": "3514ea5acaae9c2779a341deb24832df17722cb612fa7a78d34f602f91e94d17",
				"v1.21.12": "f6ef1d2d19ba0aaaba4c57c4eda94e2725c3f7e9412feb5d6fe12c1827e7c1cb",
				"v1.21.13": "5d25cc16bd38e0aaf7010d115827f7d95a1dcf7343e6b096b3df1b700ce23f7e",
				"v1.22.0":  "90a48b92a57ff6aef63ff409e2feda0713ca926b2cd243fe7e88a84c483456cc",
				"v1.22.1":  "50a5f0d186d7aefae309539e9cc7d530ef1a9b45ce690801655c2bee722d978c",
				"v1.22.2":  "4ff09d3cd2118ee2670bc96ed034620a9a1ea6a69ef38804363d4710a2f90d8c",
				"v1.22.3":  "3964e6fd46052eb4a9672421d8e8ce133b83b45abb77481b688dc6375390e480",
				"v1.22.4":  "33b799df2941f12a53ffe995d86a385c35d3c543f9d2c00c0cdb47ec91a98c5c",
				"v1.22.5":  "a512be0fa429f43d3457472efd73529cd2ba2cd54ef714faf6b69486beea054f",
				"v1.22.6":  "0bf8e47ad91215cd8c5e0ded565645aeb1ad6f0a9223a2486eb913bff929d472",
				"v1.22.7":  "7e4be37fc5ddeeae732886bf83c374198813e76d84ed2f6590145e08ece1a8b2",
				"v1.22.8":  "fc10b4e5b66c9bfa6dc297bbb4a93f58051a6069c969905ef23c19680d8d49dc",
				"v1.22.9":  "e3061f3a9c52bff82ae740c928fe389a256964a5756d691758bf3611904d7183",
				"v1.22.10": "df5e090a3c0e24b92b26f22f1d7689b6ea860099ea89b97edf5d4c19fa6da0ca",
				"v1.23.0":  "e21269a058d4ad421cf5818d4c7825991b8ba51cd06286932a33b21293b071b0",
				"v1.23.1":  "4d5766cb90050ee84e15df5e09148072da2829492fdb324521c4fa6d74d3aa34",
				"v1.23.2":  "58487391ec37489bb32fe532e367995e9ecaeafdb65c2113ff3675e7a8407219",
				"v1.23.3":  "57ec7f2921568dcf4cda0699b877cc830d49ddd2709e035c339a5afc3b83586f",
				"v1.23.4":  "c91912c9fd34a50492f889e08ff94c447fdceff150b588016fecc9051a1e56b8",
				"v1.23.5":  "8eebded187ee84c97003074eaa347e34131fef3acdf3e589a9b0200f94687667",
				"v1.23.6":  "9213c7d738e86c9a562874021df832735236fcfd5599fd4474bab3283d34bfd7",
				"v1.23.7":  "d7d863213eeb4791cdbd7c5fd398cf0cc2ef1547b3a74de8285786040f75efd2",
				"v1.23.8":  "edbd60fd6a7e11c71f848b3a6e5d1b5a2bb8ebd703e5490caa8db267361a7b89",
				"v1.24.0":  "5e58a29eaaf69ea80e90d9780d2a2d5f189fd74f94ec3bec9e3823d472277318",
				"v1.24.1":  "15e3193eecbc69330ada3f340c5a47999959bc227c735fa95e4aa79470c085d0",
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
				"v1.21.0":  "50bb95d1827455346b5643dcf83a52520733c3a582b8b1ffb50f04a8e66f00e7",
				"v1.21.1":  "1c9a93ac74f2756c1eb40a9d18bb7e146eeab0b33177c0f66f5e617ed7261d1b",
				"v1.21.2":  "245125dc436f649466123a2d2c922d17f300cbc20d2b75edad5e42d734ead4a3",
				"v1.21.3":  "5bff1c6cd1d683ce191d271b968d7b776ae5ed7403bdab5fa88446100e74972c",
				"v1.21.4":  "30645f57296281d214a9dd787a90bd16207df4b1fca7ac320913c616818a92cd",
				"v1.21.5":  "5a273b023eaa60d7820436b0f0062c4bd467274d6f2b86a9e13270c91d663618",
				"v1.21.6":  "498325da2521ce67b27902967daf4087153c5797070e03bf0bdd7c846f4d61a8",
				"v1.21.7":  "d2d17f37f1e4de446cf75f60a2a6f7fba3cbc8e27a1d176cfa0fa48862fad4bc",
				"v1.21.8":  "abf2d57cb42e8dfbcb3632dd278991bcf422891cc91e3967e00f7f45183bb43e",
				"v1.21.9":  "8947309c985911a99fb0a6e30f9ca85d9b7adc1215149e45e5be150c7e5e5de9",
				"v1.21.10": "7607bfd40317a24a276e452b46a26a7298dde2988fce826f1ee0fe9355eae786",
				"v1.21.11": "97117a6d984ff88628654494181b62502cbf4c310af70d4de92dab35482900e5",
				"v1.21.12": "6b59aab97cabb8becdd0aa1260bc0553998c8e6511507c07b0fa231c0865211d",
				"v1.21.13": "dad351cf95224f7eea54d12c84141420a750a3f6289eb4f442b9c0488def8858",
				"v1.22.0":  "9fc14b993de2c275b54445255d7770bd1d6cdb49f4cf9c227c5b035f658a2351",
				"v1.22.1":  "85df7978b2e5bb78064ed0bcce14a39d105a1a3968bb92ee5d2f96a1fa09ed12",
				"v1.22.2":  "77b4c6a56ae0ec142f54a6f5044a7167cdd7193612b04b77bf433ffe1d1918ef",
				"v1.22.3":  "dcd1ecfb7f51fb3929b9c63a984b00cf6baa6136e1d58f943ee2c9a47af5875d",
				"v1.22.4":  "3dfb128e108a3f07c53cae777026f529784a057628c721062d8fdd94b6870b69",
				"v1.22.5":  "47aa54533289277ac13419c16ffd1a2c35c7af2d6a571261e3d728990bc5fc7d",
				"v1.22.6":  "bc10e4fb42a182515f4232205bea53f90270b8f80ec1a6c1cc3301bff05e86b7",
				"v1.22.7":  "2ae0287769a70f442757e49af0ecd9ca2c6e5748e8ba72cb822d669a7aeeb8fa",
				"v1.22.8":  "67f09853d10434347eb75dbb9c63d57011ba3e4f7e1b320a0c30612b8185be8c",
				"v1.22.9":  "0168c60d1997435b006b17c95a1d42e55743048cc50ee16c8774498aa203a202",
				"v1.22.10": "8ea22a05b428de70a430711e8f75553e1be2925977ab773b5be1c240bc5b9fcd",
				"v1.23.0":  "989d117128dcaa923b2c7a917a03f4836c1b023fe1ee723541e0e39b068b93a6",
				"v1.23.1":  "eb865da197f4595dec21e6fb1fa1751ef25ac66b64fa77fd4411bbee33352a40",
				"v1.23.2":  "a29fcde7f92e1abfe992e99f415d3aee0fa381478b4a3987e333438b5380ddff",
				"v1.23.3":  "5eceefa3ca737ff1532f91bdb9ef7162882029a2a0300b4348a0980249698398",
				"v1.23.4":  "90fd5101e321053cdb66d165879a9cde18f19ba9bb8eae152fd4f4fcbe497be1",
				"v1.23.5":  "22a8468abc5d45b3415d694ad52cc8099114248c3d1fcf4297ec2b336f5cc274",
				"v1.23.6":  "a4db7458e224c3a2a7b468fc2704b31fec437614914b26a9e3d9efb6eecf61ee",
				"v1.23.7":  "65fd71aa138166039b7f4f3695308064abe7f41d2f157175e6527e60fb461eae",
				"v1.23.8":  "9b3d8863ea4ab0438881ccfbe285568529462bc77ef4512b515397a002d81b22",
				"v1.24.0":  "3e0fa21b8ebce04ca919fdfea7cc756e5f645166b95d6e4b5d9912d7721f9004",
				"v1.24.1":  "04f18fe097351cd16dc91cd3bde979201916686c6f4e1b87bae69ab4479fda04",
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
				"v1.21.0":  "681c81b7934ae2bf38b9f12d891683972d1fbbf6d7d97e50940a47b139d41b35",
				"v1.21.1":  "e77ff3ea404b2e69519ea4dce41cbdf11ae2bcba75a86d409a76eecda1c76244",
				"v1.21.2":  "aaf144b19c0676e1fe34a93dc753fb38f4de057a0e2d7521b0bef4e82f8ccc28",
				"v1.21.3":  "5bd542d656caabd75e59757a3adbae3e13d63c7c7c113d2a72475574c3c640fe",
				"v1.21.4":  "cdd46617d1a501531c62421de3754d65f30ad24d75beae2693688993a12bb557",
				"v1.21.5":  "600f70fe0e69151b9d8ac65ec195bcc840687f86ba397fce27be1faae3538a6f",
				"v1.21.6":  "422c29a1ba3bfeb2fc26ebd1c3596847fbbeeeef0ce2694515504513dc907813",
				"v1.21.7":  "59f8d7da2e994f59a369ea1705e4933949fc142bf47693e0918f4811c2e1c7b5",
				"v1.21.8":  "32f7eb6af9f1fd4e8b944f4f59582d455572147745e9fc04d044c383bd995c98",
				"v1.21.9":  "1fa0c296df6af71fca1bdd94f9fb19c7051b4b3f8cf19c353192cb96b413fcf2",
				"v1.21.10": "8e0dab1cb93e61771fba594484a37a6079073ed2d707cf300c472e79b2f91bf0",
				"v1.21.11": "ea22e3683016643344c5839a317b5e7b0061fdded321339a6d545766765bb10a",
				"v1.21.12": "56246c4d0433a7cfd29e3e989fe3835a7545a781ff0123738713c8c78a99ec17",
				"v1.21.13": "4de3bf88be86e4661f55fa69a91c3414e8e23341038b1cf366914a0794f68efb",
				"v1.22.0":  "fec5c596f7f815f17f5d7d955e9707df1ef02a2ca5e788b223651f83376feb7f",
				"v1.22.1":  "2079780ad2ff993affc9b8e1a378bf5ee759bf87fdc446e6a892a0bbd7353683",
				"v1.22.2":  "0fd6572e24e3bebbfd6b2a7cb7adced41dad4a828ef324a83f04b46378a8cb24",
				"v1.22.3":  "3f00a5f98cec024abace5bcc3580b80afc78181caf52e100fc800e588774d6eb",
				"v1.22.4":  "8d014cfe511d8c0a127b4e65ae2a6e60db592f9b1b512bb822490ea35958b10d",
				"v1.22.5":  "2be340f236a25881969eaa7d58b2279a4e31dc393cab289a74c78c0c37ba2154",
				"v1.22.6":  "7b009835b0ab74aa16ebf57f5179893035e0cf5994e1bcf9b783275921a0393a",
				"v1.22.7":  "cfc96b5f781bfbfdcb05115f4e26a5a6afc9d74bb4a5647c057b2c13086fb24d",
				"v1.22.8":  "2e6d1774f18c4d4527c3b9197a64ea5705edcf1b547c77b3e683458d771f3ce7",
				"v1.22.9":  "61530a9e6a5cb1f971295de860a8ade29db65d0dff50d1ffff3de1155dfd0c02",
				"v1.22.10": "c1aa6e9f59cfc765d33b382f604140699ab97c9c4212a905d5e1bcd7ef9a5c8b",
				"v1.23.0":  "4756ff345dd80704b749d87efb8eb294a143a1f4a251ec586197d26ad20ea518",
				"v1.23.1":  "7ff47abf62096a41005d18c6d482cf73f26b613854173327fa9f2b98720804d4",
				"v1.23.2":  "c3c4be17910935d234b776288461baf7a9c6a7414d1f1ac2ef8d3a1af4e41ab6",
				"v1.23.3":  "8f9d2dd992af82855fbac2d82e030429b08ba7775e4fee7bf043eb857dfb0317",
				"v1.23.4":  "ec3db57edcce219c24ef37f4a6a2eef5a1543e4a9bd15e7ecc993b9f74950d91",
				"v1.23.5":  "253b9db2299b09b91e4c09781ce1d2db6bad2099cf16ba210245159f48d0d5e4",
				"v1.23.6":  "fbb83e35f6b9f7cae19c50694240291805ca9c4028676af868306553b3e9266c",
				"v1.23.7":  "518f67200e853253ed6424488d6148476144b6b796ec7c6160cff15769b3e12a",
				"v1.23.8":  "1ba15ad4d9d99cfc3cbef922b5101492ad74e812629837ac2e5705a68cb7af1e",
				"v1.24.0":  "3d98ac8b4fb8dc99f9952226f2565951cc366c442656a889facc5b1b2ec2ba52",
				"v1.24.1":  "fc352d5c983b0ccf47acd8816eb826d781f408d27263dd8f761dfb63e69abfde",
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
				"v1.21.0":  "17832b192be5ea314714f7e16efd5e5f65347974bbbf41def6b02f68931380c4",
				"v1.21.1":  "5b37d7fc2da65a25896447685166769333b5896488de21bc9667edb4e799905e",
				"v1.21.2":  "525cf5506595e70bffc4c1845b3c535c7121fa2ee3daac6ca3edc69d8d63b89f",
				"v1.21.3":  "5d21da1145c25181605b9ad0810401545262fc421bbaae683bdb599632e834c1",
				"v1.21.4":  "12c849ccc627e9404187adf432a922b895c8bdecfd7ca901e1928396558eb043",
				"v1.21.5":  "746a535956db55807ef71772d2a4afec5cc438233da23952167ec0aec6fe937b",
				"v1.21.6":  "041441623c31bc6b0295342b8a2a5930d87545473e7c761ea79f3ff186c0ff52",
				"v1.21.7":  "02adf21a8de206cf64c4bff5723adb08377ecdcc38ff1efbfefd3abe2e415bb8",
				"v1.21.8":  "1d880cd437457b6a52c95fa5cfb62f05bdcea8fc29b87aaa5535a67c89a279d4",
				"v1.21.9":  "8797c78961cb71a757f35714d2735bb8bdbea94fc13d567bc0f1cf4f8e49e880",
				"v1.21.10": "5278427751381b90299e4ef330f41ca6b691aab39c3100cd200344ce6a7481c9",
				"v1.21.11": "ec0df7cf90f3422d674f9881e33d6e329a12e0f5bb438b422999493fd4370edf",
				"v1.21.12": "cb523115a0aef43fc7f1de58c33d364185b3888af2083c303e6cc59335431ac2",
				"v1.21.13": "4ffef9ed33067858f96c5662f61753791191eebe208a75ae263ca96270448249",
				"v1.22.0":  "cea637a7da4f1097b16b0195005351c07032a820a3d64c3ff326b9097cfac930",
				"v1.22.1":  "d5ffd67d8285fb224a1c49622fd739131f7b941e3d68f233dec96e72c9ebee63",
				"v1.22.2":  "f5fe3d6f4b2df5a794ebf325dc17fcdfe905a188e25f7c7e47d9cd15f14f8c2d",
				"v1.22.3":  "d0570f09bd5137ff2f672a0b177a6b78fd294a42db21f094dc02c613436ce8d1",
				"v1.22.4":  "c0049ab240b27a9dd57be2bb98356c62582d975ba2f790a61b34f155b12ab7e6",
				"v1.22.5":  "e68536cff9172d1562edddd7194d20302472a064009bf7c0ed8d79d030cb61aa",
				"v1.22.6":  "fbb823fe82b16c6f37911e907d3e4921f4642d5d48eb60e56aba1d7be0665430",
				"v1.22.7":  "8291d304c0ba4faec4336336d4cdd5159f5c90652b8b0d6be0cb5ce8f8bf92e3",
				"v1.22.8":  "604c672908a3b3cbbcf9d109d8d5fef0879992ddcf0d3e0766079d3bb7d0ca3e",
				"v1.22.9":  "d7a692ee4f5f5929a15c61947ae2deecb71b0945461f6064ced83d13094028e8",
				"v1.22.10": "2376a7ecc044bc4b5cdae9a0a14d058ae5c1803450f3a8ffdce656785e9e251e",
				"v1.23.0":  "a546fb7ccce69c4163e4a0b19a31f30ea039b4e4560c23fd6e3016e2b2dfd0d9",
				"v1.23.1":  "c24e4ab211507a39141d227595610383f7c5686cae3795b7d75eebbce8606f3d",
				"v1.23.2":  "65372ad077a660dfb8a863432c8a22cd0b650122ca98ce2e11f51a536449339f",
				"v1.23.3":  "95c36d0d1e65f6167f8fa80df04b3a816bc803e6bb5554f04d6af849c729a77d",
				"v1.23.4":  "c4f09c9031a34549fbaa48231b115fee6e170ce6832dce26d4b50b040aad2311",
				"v1.23.5":  "61f7e3ae0eb00633d3b5163c046cfcae7e73b5f26d4ffcf343f3a45904323583",
				"v1.23.6":  "11a0310e8e7af5a11539ac26d6c14cf1b77d35bce4ca74e4bbd053ed1afc8650",
				"v1.23.7":  "e96b746a77b00c04f1926035899a583ce28f02e9a5dca26c1bfb8251ca6a43bb",
				"v1.23.8":  "1b4ec707e29e8136e3516a437cb541a79c52c69b1331a7add2b47e7ac7d032e6",
				"v1.24.0":  "8f066c9a048dd1704bf22ccf6e994e2fa2ea1175c9768a786f6cb6608765025e",
				"v1.24.1":  "c2189c6956afda0f6002839f9f14a9b48c89dcc0228701e84856be36a3aac6bf",
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
				"v1.21.0":  "9f74f2fa7ee32ad07e17211725992248470310ca1988214518806b39b1dad9f0",
				"v1.21.1":  "58785190e2b4fc6891e01108e41f9ba5db26e04cebb7c1ac639919a931ce9233",
				"v1.21.2":  "55b982527d76934c2f119e70bf0d69831d3af4985f72bb87cd4924b1c7d528da",
				"v1.21.3":  "631246194fc1931cb897d61e1d542ef2321ec97adcb859a405d3b285ad9dd3d6",
				"v1.21.4":  "9410572396fb31e49d088f9816beaebad7420c7686697578691be1651d3bf85a",
				"v1.21.5":  "060ede75550c63bdc84e14fcc4c8ab3017f7ffc032fc4cac3bf20d274fab1be4",
				"v1.21.6":  "810eadc2673e0fab7044f88904853e8f3f58a4134867370bf0ccd62c19889eaa",
				"v1.21.7":  "d25d6b6f67456cc059680e7443c424eb613d9e840850a7be5195cff73fed41b8",
				"v1.21.8":  "84eaef3da0b508666e58917ebe9a6b32dcc6367bddf6e4489b909451877e3e70",
				"v1.21.9":  "195d5387f2a6ca7b8ab5c2134b4b6cc27f29372f54b771947ba7c18ee983fbe6",
				"v1.21.10": "24ce60269b1ffe1ca151af8bfd3905c2427ebef620bc9286484121adf29131c0",
				"v1.21.11": "9c45ce24ad412701beeac8d9f0004787209d76dd66390915f38a8682358484cb",
				"v1.21.12": "5a8bde5198dc0e87dfa8ebc50c29f69becdc94c756254f6b2c3f37cdbfaf2e42",
				"v1.21.13": "24fc367b5add5a06713ea8103041f6fc0cf4560a17f2c17916e7930037adc84a",
				"v1.22.0":  "703e70d49b82271535bc66bc7bd469a58c11d47f188889bd37101c9772f14fa1",
				"v1.22.1":  "78178a8337fc6c76780f60541fca7199f0f1a2e9c41806bded280a4a5ef665c9",
				"v1.22.2":  "aeca0018958c1cae0bf2f36f566315e52f87bdab38b440df349cd091e9f13f36",
				"v1.22.3":  "0751808ca8d7daba56bf76b08848ef5df6b887e9d7e8a9030dd3711080e37b54",
				"v1.22.4":  "21f24aa723002353eba1cc2668d0be22651f9063f444fd01626dce2b6e1c568c",
				"v1.22.5":  "fcb54488199c5340ff1bc0e8641d0adacb27bb18d87d0899a45ddbcc45468611",
				"v1.22.6":  "1ab07643807a45e2917072f7ba5f11140b40f19675981b199b810552d6af5c53",
				"v1.22.7":  "4dd14c5b61f112b73a5c9c844011a7887c4ffd6b91167ca76b67197dee54d388",
				"v1.22.8":  "761bf1f648056eeef753f84c8365afe4305795c5f605cd9be6a715483fe7ca6b",
				"v1.22.9":  "ae6a9b585f9a366d24bb71f508bfb9e2bb90822136138109d3a91cd28e6563bb",
				"v1.22.10": "225bc8d4ac86e3a9e36b85d2d9cb90cd4b4afade29ba0292f47834ecf570abf2",
				"v1.23.0":  "2d0f5ba6faa787878b642c151ccb2c3390ce4c1e6c8e2b59568b3869ba407c4f",
				"v1.23.1":  "156fd5e7ebbedf3c482fd274089ad75a448b04cf42bc53f370e4e4ea628f705e",
				"v1.23.2":  "5b55b58205acbafa7f4e3fc69d9ce5a9257be63455db318e24db4ab5d651cbde",
				"v1.23.3":  "d7da739e4977657a3b3c84962df49493e36b09cc66381a5e36029206dd1e01d0",
				"v1.23.4":  "3f0398d4c8a5ff633e09abd0764ed3b9091fafbe3044970108794b02731c72d6",
				"v1.23.5":  "715da05c56aa4f8df09cb1f9d96a2aa2c33a1232f6fd195e3ffce6e98a50a879",
				"v1.23.6":  "703a06354bab9f45c80102abff89f1a62cbc2c6d80678fd3973a014acc7c500a",
				"v1.23.7":  "b4c27ad52812ebf3164db927af1a01e503be3fb9dc5ffa058c9281d67c76f66e",
				"v1.23.8":  "299803a347e2e50def7740c477f0dedc69fc9e18b26b2f10e9ff84a411edb894",
				"v1.24.0":  "94d686bb6772f6fb59e3a32beff908ab406b79acdfb2427abdc4ac3ce1bb98d7",
				"v1.24.1":  "0ec3c2dbafc6dd27fc8ad25fa27fc527b5d7356d1830c0efbb8adcf975d9e84a",
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
				"v1.21.0":  "a4dd7100f547a40d3e2f83850d0bab75c6ea5eb553f0a80adcf73155bef1fd0d",
				"v1.21.1":  "d7e1163f4127efd841e5f5db6eacced11c2a3b20384457341b19ca295d0c535f",
				"v1.21.2":  "5753051ed464d0f1af05a3ca351577ba5680a332d5b2fa7738f287c8a40d81cf",
				"v1.21.3":  "2be58b5266faeeb93f38fa72d36add13a950643d2ae16a131f48f5a21c66ef23",
				"v1.21.4":  "8ac78de847118c94e2d87844e9b974556dfb30aff0e0d15fd03b82681df3ac98",
				"v1.21.5":  "fca8de7e55b55cceab9902aae03837fb2f1e72b97aa09b2ac9626bdbfd0466e4",
				"v1.21.6":  "a193997181cdfa00be0420ac6e7f4cfbf6cedd6967259c5fda1d558fa9f4efe0",
				"v1.21.7":  "50e5d76831af7b83228a5191ae10313c33639d03fadd89ad3cd492d280be4f88",
				"v1.21.8":  "ec122a1c239798c8a233377113b71bed808191dd931137f0631faa2d91fddb2a",
				"v1.21.9":  "6e2893b5de590fd9587ba327c048e5318e9e12e2acdc5a83c995c57ae822e6e4",
				"v1.21.10": "d0a88f897824954ec104895eae5f9ff9a173b162d1c9245c274cfe8db323fb37",
				"v1.21.11": "2d51a37128d823520f5f2b70436f5e3ae426eeacd16d671ae7806d421e4f57d8",
				"v1.21.12": "3f3739cff2d1a4c28d2f89d06a2bd39388af95ce25f70b6d5cc0de0538d2ce4b",
				"v1.21.13": "ca40722f3a3cb1b7687e5cdbd3a374b64ba4566e979e8ee6cd6023fa09f82ffe",
				"v1.22.0":  "8d9cc92dcc942f5ea2b2fc93c4934875d9e0e8ddecbde24c7d4c4e092cfc7afc",
				"v1.22.1":  "5c7ef1e505c35a8dc0b708f6b6ecdad6723875bb85554e9f9c3fe591e030ae5c",
				"v1.22.2":  "c5bcc7e5321d34ac42c4635ad4f6fe8bd4698e9c879dc3367be542a0b301297b",
				"v1.22.3":  "ebeac516cc073cfe9550f114ca326f762d958cb91a33c8c9d03ede6ba94a6088",
				"v1.22.4":  "3fcec0284c0fdfc22e89a5b73ebd7f51120cc3505a11a4f6d6f819d46a40b26a",
				"v1.22.5":  "a122ef299d75c0bec1dc1e28670d358e13743144e68223c8178462ba5c436e1d",
				"v1.22.6":  "b43199fe66a58f292f2c685b922330819190eb22ac41cc5c10c33fdf9f2bbc29",
				"v1.22.7":  "44342131947bc61e6b03103e7e1302d16fa3e5b2e2cd67e27194f66223ecf798",
				"v1.22.8":  "48105735b74e941a84dec6bd53637c023ad53dc5fadd9bf616347cb339c76b47",
				"v1.22.9":  "33724bed4dddf4d8ecd6ae75667552d121e2fb575ff2db427ce66516e048edac",
				"v1.22.10": "6ce1a1315225d7d62f7d17083c9f87d4f3f5684c80da108799c99780ad520cb3",
				"v1.23.0":  "1d77d6027fc8dfed772609ad9bd68f611b7e4ce73afa949f27084ad3a92b15fe",
				"v1.23.1":  "c0c24c7f6a974390e15148a575c84878e925f32328ff96ae173ec762678e4524",
				"v1.23.2":  "6e7bb8ddc5fc8fa89a4c31aba02942718b092a5107585bd09a83c95039c7510b",
				"v1.23.3":  "6708d7a701b3d9ab3b359c6be27a3012b1c486fa1e81f79e5bdc71ffca2c38f9",
				"v1.23.4":  "aa45dba48791eeb78a994a2723c462d155af4e39fdcfbcb39ce9c96f604a967a",
				"v1.23.5":  "15cd560c04def7bbe5ee3f6f75e2cfd3913371c7e76354f4b2d5d6f536b70e39",
				"v1.23.6":  "4be771c8e6a082ba61f0367077f480237f9858ef5efe14b1dbbfc05cd42fc360",
				"v1.23.7":  "5d59447a5facd8623a79c2a296a68a573789d2b102b902aafb3a730fc4bb0d3b",
				"v1.23.8":  "b293fce0b3dec37d3f5b8875b8fddc64e02f0f54f54dd7742368973c52530890",
				"v1.24.0":  "449278789de283648e4076ade46816da249714f96e71567e035e9d17e1fff06d",
				"v1.24.1":  "b817b54183e089494f8b925096e9b65af3a356d87f94b73929bf5a6028a06271",
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
