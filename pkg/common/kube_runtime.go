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
	"fmt"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	kubekeyclientset "github.com/kubesphere/kubekey/clients/clientset/versioned"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type KubeRuntime struct {
	connector.BaseRuntime
	ClusterHosts []string
	ClusterName  string
	Cluster      *kubekeyapiv1alpha2.ClusterSpec
	Kubeconfig   string
	ClientSet    *kubekeyclientset.Clientset
	Arg          Argument
}

type Argument struct {
	NodeName            string
	FilePath            string
	KubernetesVersion   string
	KsEnable            bool
	KsVersion           string
	Debug               bool
	IgnoreErr           bool
	SkipPullImages      bool
	SKipPushImages      bool
	AddImagesRepo       bool
	DeployLocalStorage  bool
	SourcesDir          string
	DownloadCommand     func(path, url string) string
	SkipConfirmCheck    bool
	InCluster           bool
	ContainerManager    string
	FromCluster         bool
	KubeConfig          string
	Artifact            string
	SkipInstallPackages bool
}

func NewKubeRuntime(flag string, arg Argument) (*KubeRuntime, error) {
	loader := NewLoader(flag, arg)
	cluster, err := loader.Load()
	if err != nil {
		return nil, err
	}
	clusterSpec := &cluster.Spec

	defaultCluster, hostGroups, err := clusterSpec.SetDefaultClusterSpec(arg.InCluster)
	if err != nil {
		return nil, err
	}

	base := connector.NewBaseRuntime(cluster.Name, connector.NewDialer(), arg.Debug, arg.IgnoreErr)
	for _, v := range hostGroups.All {
		host := ToHosts(v)
		if v.IsMaster {
			host.SetRole(Master)
		}
		if v.IsWorker {
			host.SetRole(Worker)
		}
		if v.IsEtcd {
			host.SetRole(ETCD)
		}
		if v.IsMaster || v.IsWorker {
			host.SetRole(K8s)
		}
		base.AppendHost(host)
		base.AppendRoleMap(host)
	}

	arg.KsEnable = defaultCluster.KubeSphere.Enabled
	arg.KsVersion = defaultCluster.KubeSphere.Version
	r := &KubeRuntime{
		ClusterHosts: generateHosts(hostGroups, defaultCluster),
		Cluster:      defaultCluster,
		ClusterName:  cluster.Name,
		Arg:          arg,
	}
	r.BaseRuntime = base

	return r, nil
}

// Copy is used to create a copy for Runtime.
func (k *KubeRuntime) Copy() connector.Runtime {
	runtime := *k
	return &runtime
}

func ToHosts(cfg kubekeyapiv1alpha2.HostCfg) *connector.BaseHost {
	host := connector.NewHost()
	host.Name = cfg.Name
	host.Address = cfg.Address
	host.InternalAddress = cfg.InternalAddress
	host.Port = cfg.Port
	host.User = cfg.User
	host.Password = cfg.Password
	host.PrivateKey = cfg.PrivateKey
	host.PrivateKeyPath = cfg.PrivateKeyPath
	host.Arch = cfg.Arch
	return host
}

func generateHosts(hostGroups *kubekeyapiv1alpha2.HostGroups, cfg *kubekeyapiv1alpha2.ClusterSpec) []string {
	var lbHost string
	var hostsList []string

	if cfg.ControlPlaneEndpoint.Address != "" {
		lbHost = fmt.Sprintf("%s  %s", cfg.ControlPlaneEndpoint.Address, cfg.ControlPlaneEndpoint.Domain)
	} else {
		lbHost = fmt.Sprintf("%s  %s", hostGroups.Master[0].InternalAddress, cfg.ControlPlaneEndpoint.Domain)
	}

	for _, host := range cfg.Hosts {
		if host.Name != "" {
			hostsList = append(hostsList, fmt.Sprintf("%s  %s.%s %s", host.InternalAddress, host.Name, cfg.Kubernetes.ClusterName, host.Name))
		}
	}

	hostsList = append(hostsList, lbHost)
	return hostsList
}
