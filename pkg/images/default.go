package images

import (
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"strings"
)

const (
	Etcd                  = "etcd"
	Master                = "master"
	Worker                = "worker"
	K8s                   = "k8s"
	Pause                 = "pause"
	KubeApiserver         = "kube-apiserver"
	KubeControllerManager = "kube-controller-manager"
	KubeScheduler         = "kube-scheduler"
	KubeProxy             = "kube-proxy"
)

func GetImage(mgr *manager.Manager, name string) *Image {
	var image Image
	ImageList := map[string]Image{
		"etcd":                    {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "etcd", Tag: kubekeyapi.DefaultEtcdVersion, Group: Etcd, Enable: true},
		"pause":                   {RepoAddr: "", Namespace: kubekeyapi.DefaultKubeImageRepo, Repo: "pause", Tag: "3.1", Group: K8s, Enable: true},
		"kube-apiserver":          {RepoAddr: "", Namespace: kubekeyapi.DefaultKubeImageRepo, Repo: "kube-apiserver", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true},
		"kube-controller-manager": {RepoAddr: "", Namespace: kubekeyapi.DefaultKubeImageRepo, Repo: "kube-controller-manager", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true},
		"kube-scheduler":          {RepoAddr: "", Namespace: kubekeyapi.DefaultKubeImageRepo, Repo: "kube-scheduler", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true},
		"kube-proxy":              {RepoAddr: "", Namespace: kubekeyapi.DefaultKubeImageRepo, Repo: "kube-proxy", Tag: mgr.Cluster.Kubernetes.Version, Group: K8s, Enable: true},
		"coredns":                 {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "coredns", Repo: "coredns", Tag: "1.6.0", Group: K8s, Enable: true},
		"k8s-dns-node-cache":      {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "k8s-dns-node-cache", Tag: "1.15.12", Group: K8s, Enable: true},
		"calico-kube-controllers": {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "kube-controllers", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-cni":              {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "cni", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-node":             {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "node", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-flexvol":          {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "pod2daemon-flexvol", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"flannel":                 {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "flannel", Tag: kubekeyapi.DefaultFlannelVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "flannel")},
	}
	image = ImageList[name]
	return &image
}
