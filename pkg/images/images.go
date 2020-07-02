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

package images

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
)

type Image struct {
	RepoAddr  string
	Namespace string
	Repo      string
	Tag       string
	Group     string
	Enable    bool
}

type Images struct {
	Images []Image
}

func (image Image) ImageName() string {
	var prefix string

	if image.RepoAddr == "" {
		if image.Namespace == "" {
			prefix = ""
		} else {
			prefix = fmt.Sprintf("%s/", image.Namespace)
		}
	} else {
		if image.Namespace == "" {
			prefix = fmt.Sprintf("%s/library/", image.RepoAddr)
		} else {
			prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.Namespace)
		}
	}

	return fmt.Sprintf("%s%s:%s", prefix, image.Repo, image.Tag)
}

func (image Image) ImageRepo() string {
	var prefix string

	if image.RepoAddr == "" {
		if image.Namespace == "" {
			prefix = ""
		} else {
			prefix = fmt.Sprintf("%s/", image.Namespace)
		}
	} else {
		if image.Namespace == "" {
			prefix = fmt.Sprintf("%s/library/", image.RepoAddr)
		} else {
			prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.Namespace)
		}
	}

	return fmt.Sprintf("%s%s", prefix, image.Repo)
}

func (images *Images) PullImages(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {
	for _, image := range images.Images {

		if node.IsMaster && image.Group == kubekeyapi.Master && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.Name, image.ImageName())
			_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E docker pull %s", image.ImageName()), 5, false)
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to download image: %s", image.ImageName()))
			}
		}
		if node.IsWorker && image.Group == kubekeyapi.Worker && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.Name, image.ImageName())
			_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E docker pull %s", image.ImageName()), 5, false)
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to download image: %s", image.ImageName()))
			}
		}
		if (node.IsMaster || node.IsWorker) && image.Group == kubekeyapi.K8s && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.Name, image.ImageName())
			_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E docker pull %s", image.ImageName()), 5, false)
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to download image: %s", image.ImageName()))
			}
		}
		if node.IsEtcd && image.Group == kubekeyapi.Etcd && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.Name, image.ImageName())
			_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E docker pull %s", image.ImageName()), 5, false)
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to download image: %s", image.ImageName()))
			}
		}
	}
	return nil
}
