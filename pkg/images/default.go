package images

import (
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"strings"
)

const (
	Etcd   = "etcd"
	Master = "master"
	Worker = "worker"
	K8s    = "k8s"
)

func GetImage(mgr *manager.Manager, name string) *Image {
	var image Image
	ImageList := map[string]Image{
		"pause":                   {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "pause", Tag: "3.1", Group: K8s, Enable: true},
		"kube-apiserver":          {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-apiserver", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true},
		"kube-controller-manager": {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-controller-manager", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true},
		"kube-scheduler":          {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-scheduler", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true},
		"kube-proxy":              {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-proxy", Tag: mgr.Cluster.Kubernetes.Version, Group: K8s, Enable: true},
		"etcd":                    {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "etcd", Tag: kubekeyapi.DefaultEtcdVersion, Group: Etcd, Enable: true},
		"coredns":                 {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "coredns", Repo: "coredns", Tag: "1.6.0", Group: K8s, Enable: true},
		"k8s-dns-node-cache":      {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "k8s-dns-node-cache", Tag: "1.15.12", Group: K8s, Enable: true},
		"calico-kube-controllers": {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "kube-controllers", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-cni":              {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "cni", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-node":             {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "node", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-flexvol":          {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "pod2daemon-flexvol", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"flannel":                 {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "flannel", Tag: kubekeyapi.DefaultFlannelVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "flannel")},
	}

	if mgr.Cluster.Registry.PrivateRegistry != "" {
		ImageList["pause"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "pause", Tag: "3.1", Group: K8s, Enable: true}
		ImageList["kube-apiserver"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-apiserver", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true}
		ImageList["kube-controller-manager"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-controller-manager", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true}
		ImageList["kube-scheduler"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-scheduler", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true}
		ImageList["kube-proxy"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-proxy", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true}
	}

	image = ImageList[name]
	return &image
}
