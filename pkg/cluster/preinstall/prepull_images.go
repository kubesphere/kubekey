package preinstall

import (
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"strings"
)

func PrePullImages(mgr *manager.Manager) error {

	if !mgr.SkipPullImages {
		mgr.Logger.Infoln("Start to download images on all nodes")
		if err := mgr.RunTaskOnAllNodes(PullImages, true); err != nil {
			return err
		}
	}

	return nil
}

func PullImages(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {
	i := images.Images{}
	i.Images = []images.Image{
		GetImage(mgr, "etcd"),
		GetImage(mgr, "pause"),
		GetImage(mgr, "kube-apiserver"),
		GetImage(mgr, "kube-controller-manager"),
		GetImage(mgr, "kube-scheduler"),
		GetImage(mgr, "kube-proxy"),
		GetImage(mgr, "coredns"),
		GetImage(mgr, "k8s-dns-node-cache"),
		GetImage(mgr, "calico-kube-controllers"),
		GetImage(mgr, "calico-cni"),
		GetImage(mgr, "calico-node"),
		GetImage(mgr, "calico-flexvol"),
		GetImage(mgr, "flannel"),
	}
	if err := i.PullImages(mgr, node); err != nil {
		return err
	}
	return nil
}

func GetImage(mgr *manager.Manager, name string) images.Image {
	var image images.Image
	var pauseTag string

	result := util.CompareVersion(strings.TrimSpace(mgr.Cluster.Kubernetes.Version), "v1.18.0")
	if result == 0 || result == 1 {
		pauseTag = "3.2"
	} else {
		pauseTag = "3.1"
	}

	ImageList := map[string]images.Image{
		"pause":                   {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "pause", Tag: pauseTag, Group: kubekeyapi.K8s, Enable: true},
		"kube-apiserver":          {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-apiserver", Tag: mgr.Cluster.Kubernetes.Version, Group: kubekeyapi.Master, Enable: true},
		"kube-controller-manager": {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-controller-manager", Tag: mgr.Cluster.Kubernetes.Version, Group: kubekeyapi.Master, Enable: true},
		"kube-scheduler":          {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-scheduler", Tag: mgr.Cluster.Kubernetes.Version, Group: kubekeyapi.Master, Enable: true},
		"kube-proxy":              {RepoAddr: "", Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-proxy", Tag: mgr.Cluster.Kubernetes.Version, Group: kubekeyapi.K8s, Enable: true},
		"etcd":                    {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "etcd", Tag: kubekeyapi.DefaultEtcdVersion, Group: kubekeyapi.Etcd, Enable: true},
		// network
		"coredns":                 {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "coredns", Repo: "coredns", Tag: "1.6.0", Group: kubekeyapi.K8s, Enable: true},
		"k8s-dns-node-cache":      {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "k8s-dns-node-cache", Tag: "1.15.12", Group: kubekeyapi.K8s, Enable: true},
		"calico-kube-controllers": {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "kube-controllers", Tag: kubekeyapi.DefaultCalicoVersion, Group: kubekeyapi.K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-cni":              {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "cni", Tag: kubekeyapi.DefaultCalicoVersion, Group: kubekeyapi.K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-node":             {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "node", Tag: kubekeyapi.DefaultCalicoVersion, Group: kubekeyapi.K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-flexvol":          {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "pod2daemon-flexvol", Tag: kubekeyapi.DefaultCalicoVersion, Group: kubekeyapi.K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		"calico-typha":            {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "typha", Tag: kubekeyapi.DefaultCalicoVersion, Group: kubekeyapi.K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico") && len(mgr.K8sNodes) > 50},
		"flannel":                 {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "flannel", Tag: kubekeyapi.DefaultFlannelVersion, Group: kubekeyapi.K8s, Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "flannel")},
		// storage
		"provisioner-localpv":    {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "provisioner-localpv", Tag: "1.10.0", Group: kubekeyapi.Worker, Enable: false},
		"openebs-tools":          {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "openebs-tools", Tag: "3.8", Group: kubekeyapi.Worker, Enable: false},
		"node-disk-manager":      {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "node-disk-manager", Tag: "0.5.0", Group: kubekeyapi.Worker, Enable: false},
		"node-disk-operator":     {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "node-disk-operator", Tag: "0.5.0", Group: kubekeyapi.Worker, Enable: false},
		"linux-utils":            {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "linux-utils", Tag: "1.10.0", Group: kubekeyapi.Worker, Enable: false},
		"rbd-provisioner":        {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "rbd-provisioner", Tag: "v2.1.1-k8s1.11", Group: kubekeyapi.Worker, Enable: false},
		"nfs-client-provisioner": {RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: "kubesphere", Repo: "nfs-client-provisioner", Tag: "v3.1.0-k8s1.11", Group: kubekeyapi.Worker, Enable: false},
	}

	if mgr.Cluster.Registry.PrivateRegistry != "" {
		ImageList["pause"] = images.Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "pause", Tag: pauseTag, Group: kubekeyapi.K8s, Enable: true}
		ImageList["kube-apiserver"] = images.Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-apiserver", Tag: mgr.Cluster.Kubernetes.Version, Group: kubekeyapi.Master, Enable: true}
		ImageList["kube-controller-manager"] = images.Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-controller-manager", Tag: mgr.Cluster.Kubernetes.Version, Group: kubekeyapi.Master, Enable: true}
		ImageList["kube-scheduler"] = images.Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-scheduler", Tag: mgr.Cluster.Kubernetes.Version, Group: kubekeyapi.Master, Enable: true}
		ImageList["kube-proxy"] = images.Image{RepoAddr: mgr.Cluster.Registry.PrivateRegistry, Namespace: mgr.Cluster.Kubernetes.ImageRepo, Repo: "kube-proxy", Tag: mgr.Cluster.Kubernetes.Version, Group: kubekeyapi.Master, Enable: true}
	}

	image = ImageList[name]
	return image
}
