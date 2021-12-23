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

package images

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/pkg/errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
)

const (
	cnRegistry          = "registry.cn-beijing.aliyuncs.com"
	cnNamespaceOverride = "kubesphereio"
)

// Image defines image's info.
type Image struct {
	RepoAddr          string
	Namespace         string
	NamespaceOverride string
	Repo              string
	Tag               string
	Group             string
	Enable            bool
}

// Images contains a list of Image
type Images struct {
	Images []Image
}

// ImageName is used to generate image's full name.
func (image Image) ImageName() string {
	return fmt.Sprintf("%s:%s", image.ImageRepo(), image.Tag)
}

// ImageRepo is used to generate image's repo address.
func (image Image) ImageRepo() string {
	var prefix string

	if os.Getenv("KKZONE") == "cn" {
		if image.RepoAddr == "" || image.RepoAddr == cnRegistry {
			image.RepoAddr = cnRegistry
			image.NamespaceOverride = cnNamespaceOverride
		}
	}

	if image.RepoAddr == "" {
		if image.Namespace == "" {
			prefix = ""
		} else {
			prefix = fmt.Sprintf("%s/", image.Namespace)
		}
	} else {
		if image.NamespaceOverride == "" {
			if image.Namespace == "" {
				prefix = fmt.Sprintf("%s/library/", image.RepoAddr)
			} else {
				prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.Namespace)
			}
		} else {
			prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.NamespaceOverride)
		}
	}

	return fmt.Sprintf("%s%s", prefix, image.Repo)
}

// PullImages is used to pull images in the list of Image.
func (images *Images) PullImages(runtime connector.Runtime, kubeConf *common.KubeConf) error {
	pullCmd := "docker"
	switch kubeConf.Cluster.Kubernetes.ContainerManager {
	case "crio":
		pullCmd = "crictl"
	case "containerd":
		pullCmd = "crictl"
	case "isula":
		pullCmd = "isula"
	default:
		pullCmd = "docker"
	}

	host := runtime.RemoteHost()

	for _, image := range images.Images {
		switch {
		case host.IsRole(common.Master) && image.Group == kubekeyapiv1alpha2.Master && image.Enable,
			host.IsRole(common.Worker) && image.Group == kubekeyapiv1alpha2.Worker && image.Enable,
			(host.IsRole(common.Master) || host.IsRole(common.Worker)) && image.Group == kubekeyapiv1alpha2.K8s && image.Enable,
			host.IsRole(common.ETCD) && image.Group == kubekeyapiv1alpha2.Etcd && image.Enable:

			logger.Log.Messagef(host.GetName(), "downloading image: %s", image.ImageName())
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("env PATH=$PATH %s pull %s", pullCmd, image.ImageName()), false); err != nil {
				return errors.Wrap(err, "pull image failed")
			}
		default:
			continue
		}

	}
	return nil
}

func Push(fileName string, prePath string, runtime connector.Runtime, kubeConf *common.KubeConf) (Image, error) {
	// just like: docker.io-calico-cni-v3.20.0.tar, docker.io-kubesphere-kube-apiserver-v1.21.5.tar .e.g.
	nameArr := strings.Split(fileName, "-")

	// docker.io
	registry := nameArr[0]

	// calico or kubesphere .e.g
	namespace := nameArr[1]

	// cni or kube-apiserver
	imageName := strings.Join(nameArr[2:len(nameArr)-1], "-")

	// v3.20.0.tar
	tag := nameArr[len(nameArr)-1]
	// .tar
	tagExt := path.Ext(tag)
	// v3.20.0
	tag = strings.TrimSuffix(tag, tagExt)

	privateRegistry := kubeConf.Cluster.Registry.PrivateRegistry
	image := Image{
		RepoAddr:  privateRegistry,
		Namespace: namespace,
		Repo:      imageName,
		Tag:       tag,
	}

	fullPath := filepath.Join(prePath, fileName)
	oldName := fmt.Sprintf("%s/%s/%s:%s", registry, namespace, imageName, tag)
	switch kubeConf.Cluster.Kubernetes.ContainerManager {
	case common.Docker:
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("env PATH=$PATH docker load -i %s", fullPath), false); err != nil {
			return image, errors.Wrap(err, "pull image failed")
		}

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("env PATH=$PATH docker tag %s %s", oldName, image.ImageName()), false); err != nil {
			return image, errors.Wrap(err, "pull image failed")
		}
	case common.Conatinerd:
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("env PATH=$PATH ctr images import --base-name %s %s", image.ImageName(), fullPath), false); err != nil {
			return image, errors.Wrapf(err, "load %s tar onto image %s failed", imageName, fullPath)
		}

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("env PATH=$PATH ctr images push %s %s", image.ImageName(), oldName), false); err != nil {
			return image, errors.Wrap(err, "pull image failed")
		}
	case common.Isula:
	case common.Crio:

	default:
		return image, fmt.Errorf("unsupport container manager [%s]", kubeConf.Cluster.Kubernetes.ContainerManager)
	}
	return image, nil
}
