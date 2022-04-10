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

package common

import (
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	kubekeyclientset "github.com/kubesphere/kubekey/clients/clientset/versioned"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type KubeRuntime struct {
	connector.BaseRuntime
	ClusterName string
	Cluster     *kubekeyapiv1alpha2.ClusterSpec
	Kubeconfig  string
	ClientSet   *kubekeyclientset.Clientset
	Arg         Argument
}

type Argument struct {
	NodeName           string
	FilePath           string
	KubernetesVersion  string
	KsEnable           bool
	KsVersion          string
	Debug              bool
	IgnoreErr          bool
	SkipPullImages     bool
	SKipPushImages     bool
	AddImagesRepo      bool
	DeployLocalStorage *bool
	SourcesDir         string
	DownloadCommand    func(path, url string) string
	SkipConfirmCheck   bool
	InCluster          bool
	ContainerManager   string
	FromCluster        bool
	KubeConfig         string
	Artifact           string
	InstallPackages    bool
	ImagesDir          string
	Namespace          string
}

func NewKubeRuntime(flag string, arg Argument) (*KubeRuntime, error) {
	loader := NewLoader(flag, arg)
	cluster, err := loader.Load()
	if err != nil {
		return nil, err
	}

	base := connector.NewBaseRuntime(cluster.Name, connector.NewDialer(), arg.Debug, arg.IgnoreErr)

	clusterSpec := &cluster.Spec
	defaultCluster, roleGroups := clusterSpec.SetDefaultClusterSpec(arg.InCluster)

	hostSet := make(map[string]struct{})
	for _, role := range roleGroups {
		for _, host := range role {
			if host.IsRole(Master) || host.IsRole(Worker) {
				host.SetRole(K8s)
			}
			if _, ok := hostSet[host.GetName()]; !ok {
				hostSet[host.GetName()] = struct{}{}
				base.AppendHost(host)
				base.AppendRoleMap(host)
			}
		}
	}

	arg.KsEnable = defaultCluster.KubeSphere.Enabled
	arg.KsVersion = defaultCluster.KubeSphere.Version
	r := &KubeRuntime{
		Cluster:     defaultCluster,
		ClusterName: cluster.Name,
		Arg:         arg,
	}
	r.BaseRuntime = base

	return r, nil
}

// Copy is used to create a copy for Runtime.
func (k *KubeRuntime) Copy() connector.Runtime {
	runtime := *k
	return &runtime
}
