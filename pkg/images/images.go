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
	"github.com/kubesphere/kubekey/pkg/container/templates"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/pkg/errors"
	"os"
	"os/exec"
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

func CmdPush(fileName string, prePath string, kubeConf *common.KubeConf, arches []string) error {
	// just like: docker.io-calico-cni:v3.20.0.tar, docker.io-kubesphere-kube-apiserver:v1.21.5.tar
	// docker.io-fluent-fluentd:v1.4.2-2.0.tar .e.g.
	fullArr := strings.Split(fileName, ":")

	// v3.20.0.tar, v1.21.5.tar or v1.4.2-2.0.tar
	tag := fullArr[len(fullArr)-1]
	// .tar
	tagExt := path.Ext(tag)
	// v3.20.0, v1.21.5 or v1.4.2-2.0
	tag = strings.TrimSuffix(tag, tagExt)

	// docker.io-calico-cni, docker.io-kubesphere-kube-apiserver or docker.io-fluent-fluentd
	nameArr := strings.Split(strings.Join(fullArr[:len(fullArr)-1], ":"), "-")

	// docker.io
	registry := nameArr[0]
	// calico, kubesphere or fluent
	namespace := nameArr[1]
	// cni, kube-apiserver or fluentd
	imageName := strings.Join(nameArr[2:], "-")

	privateRegistry := kubeConf.Cluster.Registry.PrivateRegistry
	image := Image{
		RepoAddr:  privateRegistry,
		Namespace: namespace,
		Repo:      imageName,
		Tag:       tag,
	}

	auths := templates.Auths(kubeConf)

	fullPath := filepath.Join(prePath, fileName)
	oldName := fmt.Sprintf("%s/%s/%s:%s", registry, namespace, imageName, tag)

	if out, err := exec.Command("/bin/bash", "-c",
		fmt.Sprintf("sudo ctr images import --base-name %s %s", oldName, fullPath)).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "import image %s failed: %s", oldName, out)
	}

	pushCmd := fmt.Sprintf("sudo ctr images push  %s %s --platform %s -k",
		image.ImageName(), oldName, strings.Join(arches, " "))
	if kubeConf.Cluster.Registry.PlainHTTP {
		pushCmd = fmt.Sprintf("%s --plain-http", pushCmd)
	}
	if _, ok := auths[privateRegistry]; ok {
		auth := auths[privateRegistry]
		pushCmd = fmt.Sprintf("%s --user %s:%s", pushCmd, auth.Username, auth.Password)
	}

	logger.Log.Debugf("push command: %s", pushCmd)
	if out, err := exec.Command("/bin/bash", "-c", pushCmd).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "push image %s failed: %s", oldName, out)
	}
	fmt.Printf("push %s success \n", image.ImageName())
	return nil
}
