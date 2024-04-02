/*
 Copyright 2022 The KubeSphere Authors.

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
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/images"
	"github.com/spf13/cobra"
)

type ArtifactImagesListOptions struct {
	CommonOptions *options.CommonOptions

	KubernetesVersion string
}

func NewArtifactImagesListOptions() *ArtifactImagesListOptions {
	return &ArtifactImagesListOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdArtifactImagesList  creates a new `kubekey artifacts images push` command
func NewCmdArtifactImagesList() *cobra.Command {
	o := NewArtifactImagesListOptions()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list images to a registry from an artifact",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *ArtifactImagesListOptions) Run() error {

	imageName := []string{
		"pause",
		"etcd",
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

	for _, name := range imageName {
		repo := images.GetImage(&common.KubeRuntime{
			BaseRuntime: connector.NewBaseRuntime("image list", connector.NewDialer(), o.CommonOptions.Verbose, o.CommonOptions.IgnoreErr),
		}, &common.KubeConf{
			Cluster: &kubekeyapiv1alpha2.ClusterSpec{
				Kubernetes: kubekeyapiv1alpha2.Kubernetes{Version: o.KubernetesVersion},
				Registry:   kubekeyapiv1alpha2.RegistryConfig{PrivateRegistry: "docker.io"},
			},
		}, name).ImageName()
		fmt.Println(repo)
	}

	return nil
}

func (o *ArtifactImagesListOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.KubernetesVersion, "k8s-version", "", "v1.23.15", "Path to a KubeKey artifact images directory")
}
