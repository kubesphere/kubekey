package common

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	kubekeyclientset "github.com/kubesphere/kubekey/clients/clientset/versioned"
	"github.com/kubesphere/kubekey/pkg/core/modules"
)

type KubeConf struct {
	ClusterHosts []string
	Cluster      *kubekeyapiv1alpha1.ClusterSpec
	Kubeconfig   string
	Conditions   []kubekeyapiv1alpha1.Condition
	ClientSet    *kubekeyclientset.Clientset
	Arg          Argument
}

type KubeModule struct {
	modules.BaseTaskModule
	KubeConf *KubeConf
}

func (k *KubeModule) AutoAssert() {
	kubeRuntime := k.Runtime.(*KubeRuntime)
	conf := &KubeConf{
		ClusterHosts: kubeRuntime.ClusterHosts,
		Cluster:      kubeRuntime.Cluster,
		Kubeconfig:   kubeRuntime.Kubeconfig,
		Conditions:   kubeRuntime.Conditions,
		ClientSet:    kubeRuntime.ClientSet,
		Arg:          kubeRuntime.Arg,
	}

	k.KubeConf = conf
}

type KubeCustomModule struct {
	modules.CustomModule
	KubeConf *KubeConf
}

func (k *KubeCustomModule) AutoAssert() {
	kubeRuntime := k.Runtime.(*KubeRuntime)
	conf := &KubeConf{
		ClusterHosts: kubeRuntime.ClusterHosts,
		Cluster:      kubeRuntime.Cluster,
		Kubeconfig:   kubeRuntime.Kubeconfig,
		Conditions:   kubeRuntime.Conditions,
		ClientSet:    kubeRuntime.ClientSet,
		Arg:          kubeRuntime.Arg,
	}

	k.KubeConf = conf
}
