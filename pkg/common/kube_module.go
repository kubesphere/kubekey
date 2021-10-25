package common

import (
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	kubekeyclientset "github.com/kubesphere/kubekey/clients/clientset/versioned"
	"github.com/kubesphere/kubekey/pkg/core/module"
)

type KubeConf struct {
	ClusterHosts []string
	ClusterName  string
	Cluster      *kubekeyapiv1alpha2.ClusterSpec
	Kubeconfig   string
	Conditions   []kubekeyapiv1alpha2.Condition
	ClientSet    *kubekeyclientset.Clientset
	Arg          Argument
}

type KubeModule struct {
	module.BaseTaskModule
	KubeConf *KubeConf
}

func (k *KubeModule) IsSkip() bool {
	return k.Skip
}

func (k *KubeModule) AutoAssert() {
	kubeRuntime := k.Runtime.(*KubeRuntime)
	conf := &KubeConf{
		ClusterHosts: kubeRuntime.ClusterHosts,
		ClusterName:  kubeRuntime.ClusterName,
		Cluster:      kubeRuntime.Cluster,
		Kubeconfig:   kubeRuntime.Kubeconfig,
		Conditions:   kubeRuntime.Conditions,
		ClientSet:    kubeRuntime.ClientSet,
		Arg:          kubeRuntime.Arg,
	}

	k.KubeConf = conf
}

type KubeCustomModule struct {
	module.CustomModule
	KubeConf *KubeConf
}

func (k *KubeCustomModule) AutoAssert() {
	kubeRuntime := k.Runtime.(*KubeRuntime)
	conf := &KubeConf{
		ClusterHosts: kubeRuntime.ClusterHosts,
		ClusterName:  kubeRuntime.ClusterName,
		Cluster:      kubeRuntime.Cluster,
		Kubeconfig:   kubeRuntime.Kubeconfig,
		Conditions:   kubeRuntime.Conditions,
		ClientSet:    kubeRuntime.ClientSet,
		Arg:          kubeRuntime.Arg,
	}

	k.KubeConf = conf
}
