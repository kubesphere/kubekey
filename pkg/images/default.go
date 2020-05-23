package images

import (
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
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
	var pauseTag string

	result := util.CompareVersion(strings.TrimSpace(mgr.Cluster.Kubernetes.Version), "v1.18.0")
	if result == 0 || result == 1 {
		pauseTag = "3.2"
	} else {
		pauseTag = "3.1"
	}

	ImageList := map[string]Image{
		"pause":                   {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "pause", Tag: pauseTag, Group: K8s, Enable: true},
		"kube-apiserver":          {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-apiserver", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true},
		"kube-controller-manager": {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-controller-manager", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true},
		"kube-scheduler":          {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-scheduler", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true},
		"kube-proxy":              {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-proxy", Tag: mgr.Cluster.Kubernetes.Version, Group: K8s, Enable: true},
		"etcd":                    {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "etcd", Tag: kubekeyapi.DefaultEtcdVersion, Group: Etcd, Enable: true},
		// network
		"coredns":                 {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "coredns", Repo: "coredns", Tag: "1.6.0", Group: K8s, Enable: true},
		"k8s-dns-node-cache":      {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "k8s-dns-node-cache", Tag: "1.15.12", Group: K8s, Enable: true},
		"calico-kube-controllers": {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "kube-controllers", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-cni":              {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "cni", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-node":             {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "node", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-flexvol":          {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "pod2daemon-flexvol", Tag: kubekeyapi.DefaultCalicoVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"flannel":                 {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "flannel", Tag: kubekeyapi.DefaultFlannelVersion, Group: K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "flannel")},
		// storage
		"provisioner-localpv":      {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "provisioner-localpv", Tag: "1.1.0", Group: K8s, Enable: mgr.Cluster.Storage.LocalVolume.Enabled},
		"openebs-tools":            {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "openebs-tools", Tag: "3.8", Group: K8s, Enable: mgr.Cluster.Storage.LocalVolume.Enabled},
		"node-disk-manager-amd64":  {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "node-disk-manager-amd64", Tag: "v0.4.1", Group: Worker, Enable: mgr.Cluster.Storage.LocalVolume.Enabled},
		"node-disk-operator-amd64": {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "node-disk-operator-amd64", Tag: "v0.4.1", Group: Worker, Enable: mgr.Cluster.Storage.LocalVolume.Enabled},
		"linux-utils":              {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "linux-utils", Tag: "3.9", Group: Worker, Enable: mgr.Cluster.Storage.LocalVolume.Enabled},
		"rbd-provisioner":          {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "rbd-provisioner", Tag: "v2.1.1-k8s1.11", Group: Worker, Enable: mgr.Cluster.Storage.CephRBD.Enabled},
		"nfs-client-provisioner":   {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "nfs-client-provisioner", Tag: "v3.1.0-k8s1.11", Group: Worker, Enable: mgr.Cluster.Storage.NfsClient.Enabled},
	}

	if mgr.Cluster.Registry.PrivateRegistry != "" {
		ImageList["pause"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "pause", Tag: pauseTag, Group: K8s, Enable: true}
		ImageList["kube-apiserver"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-apiserver", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true}
		ImageList["kube-controller-manager"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-controller-manager", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true}
		ImageList["kube-scheduler"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-scheduler", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true}
		ImageList["kube-proxy"] = Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-proxy", Tag: mgr.Cluster.Kubernetes.Version, Group: Master, Enable: true}
	}

	image = ImageList[name]
	return &image
}
