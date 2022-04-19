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
	mapset "github.com/deckarep/golang-set"
	kubekeyv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/artifact/templates"
	"github.com/kubesphere/kubekey/pkg/client/kubernetes"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/pkg/errors"
	"io/ioutil"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"os"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateManifest(arg common.Argument, name string) error {
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
			version = fmt.Sprintf("%s.%s", v[0], v[1])
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
			ContainerRuntimes: containerArr,
		},
		Images: imageArr,
	}

	manifestStr, err := templates.RenderManifest(options)

	if err := ioutil.WriteFile(arg.FilePath, []byte(manifestStr), 0644); err != nil {
		return errors.Wrap(err, fmt.Sprintf("write file %s failed", arg.FilePath))
	}

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
