/*
 Copyright 2021 The KubeSphere Authors.

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

package artifact

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/files"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/images"

	mapset "github.com/deckarep/golang-set"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	versionutil "k8s.io/apimachinery/pkg/util/version"

	kubekeyv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/artifact/templates"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/client/kubernetes"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
)

func CreateManifest(arg common.Argument, name string, registry bool) error {
	checkFileExists(arg.FilePath)

	client, err := kubernetes.NewClient(arg.KubeConfig)
	if err != nil {
		return errors.Wrap(err, "get kubernetes client failed")
	}

	nodes, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	archSet := mapset.NewThreadUnsafeSet()
	containerSet := mapset.NewThreadUnsafeSet()
	imagesSet := mapset.NewThreadUnsafeSet()
	osSet := mapset.NewThreadUnsafeSet()

	maxKubeletVersion := versionutil.MustParseGeneric("v0.0.0")
	kubernetesDistribution := kubekeyv1alpha2.KubernetesDistribution{}
	for _, node := range nodes.Items {
		containerStrArr := strings.Split(node.Status.NodeInfo.ContainerRuntimeVersion, "://")
		containerRuntime := kubekeyv1alpha2.ContainerRuntime{
			Type:    containerStrArr[0],
			Version: containerStrArr[1],
		}
		if containerRuntime.Type == "containerd" &&
			versionutil.MustParseSemantic(containerRuntime.Version).LessThan(versionutil.MustParseSemantic("1.6.2")) {
			containerRuntime.Version = "1.6.2"
		}
		containerSet.Add(containerRuntime)

		archSet.Add(node.Status.NodeInfo.Architecture)
		for _, image := range node.Status.Images {
			for _, name := range image.Names {
				if !strings.Contains(name, "@sha256") {
					if containerRuntime.Type == kubekeyv1alpha2.Docker {
						arr := strings.Split(name, "/")
						switch len(arr) {
						case 1:
							name = fmt.Sprintf("docker.io/library/%s", name)
						case 2:
							name = fmt.Sprintf("docker.io/%s", name)
						}
					}
					imagesSet.Add(name)
				}
			}
		}

		// todo: for now, the cases only have ubuntu, centos. Ant it need to check all linux distribution.
		var (
			id, version string
		)
		osImageArr := strings.Split(node.Status.NodeInfo.OSImage, " ")
		switch strings.ToLower(osImageArr[0]) {
		case "ubuntu":
			id = "ubuntu"
			v := strings.Split(osImageArr[1], ".")
			if len(v) >= 2 {
				version = fmt.Sprintf("%s.%s", v[0], v[1])
			}
		case "centos":
			id = "centos"
			version = osImageArr[2]
		case "rhel":
			id = "rhel"
			version = osImageArr[2]
		default:
			id = strings.ToLower(osImageArr[0])
			version = "Can't get the os version. Please edit it manually."
		}

		osObj := kubekeyv1alpha2.OperatingSystem{
			Arch:    node.Status.NodeInfo.Architecture,
			Type:    node.Status.NodeInfo.OperatingSystem,
			Id:      id,
			Version: version,
			OsImage: node.Status.NodeInfo.OSImage,
		}
		osSet.Add(osObj)

		kubeletStrArr := strings.Split(node.Status.NodeInfo.KubeletVersion, "+")
		kubeletVersion := kubeletStrArr[0]
		distribution := "kubernetes"
		if len(kubeletStrArr) == 2 {
			distribution = "k3s"
		}
		if maxKubeletVersion.LessThan(versionutil.MustParseGeneric(kubeletVersion)) {
			maxKubeletVersion = versionutil.MustParseGeneric(kubeletVersion)
			kubernetesDistribution.Version = fmt.Sprintf("v%s", maxKubeletVersion.String())
			kubernetesDistribution.Type = distribution
		}

	}

	archArr := make([]string, 0, archSet.Cardinality())
	for _, v := range archSet.ToSlice() {
		arch := v.(string)
		archArr = append(archArr, arch)
	}
	imageArr := make([]string, 0, imagesSet.Cardinality())
	for _, v := range imagesSet.ToSlice() {
		image := v.(string)
		imageArr = append(imageArr, image)
	}
	osArr := make([]kubekeyv1alpha2.OperatingSystem, 0, osSet.Cardinality())
	for _, v := range osSet.ToSlice() {
		osObj := v.(kubekeyv1alpha2.OperatingSystem)
		osArr = append(osArr, osObj)
	}
	containerArr := make([]kubekeyv1alpha2.ContainerRuntime, 0, containerSet.Cardinality())
	for _, v := range containerSet.ToSlice() {
		container := v.(kubekeyv1alpha2.ContainerRuntime)
		containerArr = append(containerArr, container)
	}

	// todo: Whether it need to detect components version
	sort.Strings(imageArr)
	options := &templates.Options{
		Name:                    name,
		Arches:                  archArr,
		OperatingSystems:        osArr,
		KubernetesDistributions: []kubekeyv1alpha2.KubernetesDistribution{kubernetesDistribution},
		Components: kubekeyv1alpha2.Components{
			Helm:              kubekeyv1alpha2.Helm{Version: kubekeyv1alpha2.DefaultHelmVersion},
			CNI:               kubekeyv1alpha2.CNI{Version: kubekeyv1alpha2.DefaultCniVersion},
			ETCD:              kubekeyv1alpha2.ETCD{Version: kubekeyv1alpha2.DefaultEtcdVersion},
			Crictl:            kubekeyv1alpha2.Crictl{Version: kubekeyv1alpha2.DefaultCrictlVersion},
			Calicoctl:         kubekeyv1alpha2.Calicoctl{Version: kubekeyv1alpha2.DefaultCalicoVersion},
			ContainerRuntimes: containerArr,
		},
		Images: imageArr,
	}

	if registry {
		options.Components.DockerRegistry.Version = kubekeyv1alpha2.DefaultRegistryVersion
		options.Components.DockerCompose.Version = kubekeyv1alpha2.DefaultDockerComposeVersion
		options.Components.Harbor.Version = kubekeyv1alpha2.DefaultHarborVersion
	}

	manifestStr, err := templates.RenderManifest(options)

	if err := os.WriteFile(arg.FilePath, []byte(manifestStr), 0644); err != nil {
		return errors.Wrap(err, fmt.Sprintf("write file %s failed", arg.FilePath))
	}

	fmt.Println("Generate KubeKey manifest file successfully")
	return nil
}

func checkFileExists(fileName string) {
	if util.IsExist(fileName) {
		reader := bufio.NewReader(os.Stdin)
		stop := false
		for {
			if stop {
				break
			}
			fmt.Printf("%s already exists. Are you sure you want to overwrite this file? [yes/no]: ", fileName)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input != "" {
				switch input {
				case "yes":
					stop = true
				case "no":
					os.Exit(0)
				}
			}
		}
	}
}

func CreateManifestSpecifyVersion(arg common.Argument, name, version string, registry bool, arch []string) error {
	checkFileExists(arg.FilePath)

	var kubernetesDistribution []kubekeyv1alpha2.KubernetesDistribution

	k8sVersion := strings.Split(version, ",")

	imageNames := []string{
		"pause",
		"kube-apiserver",
		"kube-controller-manager",
		"kube-scheduler",
		"kube-proxy",

		// network
		"coredns",
		"k8s-dns-node-cache",
		"calico-kube-controllers",
		"calico-cni",
		"calico-node",
		"calico-flexvol",
		"calico-typha",
		"flannel",
		"flannel-cni-plugin",
		"cilium",
		"cilium-operator-generic",
		"hybridnet",
		"kubeovn",
		"multus",
		// storage
		"provisioner-localpv",
		"linux-utils",
		// load balancer
		"haproxy",
		"kubevip",
		// kata-deploy
		"kata-deploy",
		// node-feature-discovery
		"node-feature-discovery",
	}
	var imageArr []string
	for _, v := range k8sVersion {
		versionutil.MustParseGeneric(v)

		for _, a := range arch {
			_, ok := files.FileSha256["kubeadm"][a][v]
			if !ok {
				return fmt.Errorf("invalid kubernetes version %s", v)
			}
		}

		kubernetesDistribution = append(kubernetesDistribution, kubekeyv1alpha2.KubernetesDistribution{
			Type:    "kubernetes",
			Version: v,
		})

		for _, imageName := range imageNames {
			repo := images.GetImage(&common.KubeRuntime{
				BaseRuntime: connector.NewBaseRuntime("image list", connector.NewDialer(), arg.Debug, arg.IgnoreErr),
			}, &common.KubeConf{
				Cluster: &kubekeyv1alpha2.ClusterSpec{
					Kubernetes: kubekeyv1alpha2.Kubernetes{Version: v},
					Registry:   kubekeyv1alpha2.RegistryConfig{PrivateRegistry: images.DefaultRegistry()},
				},
			}, imageName).ImageName()
			if !imageIsExist(repo, imageArr) {
				imageArr = append(imageArr, repo)
			}

		}
	}

	options := &templates.Options{
		Name:                    name,
		Arches:                  arch,
		OperatingSystems:        []kubekeyv1alpha2.OperatingSystem{},
		KubernetesDistributions: kubernetesDistribution,
		Components: kubekeyv1alpha2.Components{
			Helm:      kubekeyv1alpha2.Helm{Version: kubekeyv1alpha2.DefaultHelmVersion},
			CNI:       kubekeyv1alpha2.CNI{Version: kubekeyv1alpha2.DefaultCniVersion},
			ETCD:      kubekeyv1alpha2.ETCD{Version: kubekeyv1alpha2.DefaultEtcdVersion},
			Crictl:    kubekeyv1alpha2.Crictl{Version: kubekeyv1alpha2.DefaultCrictlVersion},
			Calicoctl: kubekeyv1alpha2.Calicoctl{Version: kubekeyv1alpha2.DefaultCalicoVersion},
			ContainerRuntimes: []kubekeyv1alpha2.ContainerRuntime{
				{
					Type:    "docker",
					Version: kubekeyv1alpha2.DefaultDockerVersion,
				},
				{
					Type:    "containerd",
					Version: kubekeyv1alpha2.DefaultContainerdVersion,
				},
			},
		},
		Images: imageArr,
	}

	if registry {
		options.Components.DockerRegistry.Version = kubekeyv1alpha2.DefaultRegistryVersion
		options.Components.DockerCompose.Version = kubekeyv1alpha2.DefaultDockerComposeVersion
		options.Components.Harbor.Version = kubekeyv1alpha2.DefaultHarborVersion
	}

	manifestStr, err := templates.RenderManifest(options)

	if err = os.WriteFile(arg.FilePath, []byte(manifestStr), 0644); err != nil {
		return errors.Wrap(err, fmt.Sprintf("write file %s failed", arg.FilePath))
	}

	fmt.Println("Generate KubeKey manifest file successfully")
	return nil
}

func imageIsExist(s string, arr []string) bool {
	for _, s2 := range arr {
		if s2 == s {
			return true
		}
	}
	return false
}
