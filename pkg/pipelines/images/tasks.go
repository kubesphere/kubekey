package images

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"strings"
)

type PullImage struct {
	common.KubeAction
}

func (p *PullImage) Execute(runtime connector.Runtime) error {
	i := Images{}
	i.Images = []Image{
		GetImage(runtime, p.KubeConf, "etcd"),
		GetImage(runtime, p.KubeConf, "pause"),
		GetImage(runtime, p.KubeConf, "kube-apiserver"),
		GetImage(runtime, p.KubeConf, "kube-controller-manager"),
		GetImage(runtime, p.KubeConf, "kube-scheduler"),
		GetImage(runtime, p.KubeConf, "kube-proxy"),
		GetImage(runtime, p.KubeConf, "coredns"),
		GetImage(runtime, p.KubeConf, "k8s-dns-node-cache"),
		GetImage(runtime, p.KubeConf, "calico-kube-controllers"),
		GetImage(runtime, p.KubeConf, "calico-cni"),
		GetImage(runtime, p.KubeConf, "calico-node"),
		GetImage(runtime, p.KubeConf, "calico-flexvol"),
		GetImage(runtime, p.KubeConf, "cilium"),
		GetImage(runtime, p.KubeConf, "operator-generic"),
		GetImage(runtime, p.KubeConf, "flannel"),
		GetImage(runtime, p.KubeConf, "kubeovn"),
		GetImage(runtime, p.KubeConf, "haproxy"),
	}
	if err := i.PullImages(runtime, p.KubeConf); err != nil {
		return err
	}
	return nil
}

// GetImage defines the list of all images and gets image object by name.
func GetImage(runtime connector.ModuleRuntime, kubeConf *common.KubeConf, name string) Image {
	var image Image
	var pauseTag string

	cmp, err := versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.21.0")
	if err != nil {
		logger.Log.Fatal("Failed to compare version: %v", err)
	}
	if (cmp == 0 || cmp == 1) || (kubeConf.Cluster.Kubernetes.ContainerManager != "" && kubeConf.Cluster.Kubernetes.ContainerManager != "docker") {
		cmp, err := versionutil.MustParseSemantic(kubeConf.Cluster.Kubernetes.Version).Compare("v1.22.0")
		if err != nil {
			logger.Log.Fatal("Failed to compare version: %v", err)
		}
		if cmp == 0 || cmp == 1 {
			pauseTag = "3.5"
		} else {
			pauseTag = "3.4.1"
		}
	} else {
		pauseTag = "3.2"
	}

	ImageList := map[string]Image{
		"pause":                   {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyapiv1alpha1.DefaultKubeImageNamespace, Repo: "pause", Tag: pauseTag, Group: kubekeyapiv1alpha1.K8s, Enable: true},
		"kube-apiserver":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyapiv1alpha1.DefaultKubeImageNamespace, Repo: "kube-apiserver", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyapiv1alpha1.Master, Enable: true},
		"kube-controller-manager": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyapiv1alpha1.DefaultKubeImageNamespace, Repo: "kube-controller-manager", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyapiv1alpha1.Master, Enable: true},
		"kube-scheduler":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyapiv1alpha1.DefaultKubeImageNamespace, Repo: "kube-scheduler", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyapiv1alpha1.Master, Enable: true},
		"kube-proxy":              {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyapiv1alpha1.DefaultKubeImageNamespace, Repo: "kube-proxy", Tag: kubeConf.Cluster.Kubernetes.Version, Group: kubekeyapiv1alpha1.K8s, Enable: true},

		// network
		"coredns":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "coredns", Repo: "coredns", Tag: "1.8.4", Group: kubekeyapiv1alpha1.K8s, Enable: true},
		"k8s-dns-node-cache":      {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyapiv1alpha1.DefaultKubeImageNamespace, Repo: "k8s-dns-node-cache", Tag: "1.15.12", Group: kubekeyapiv1alpha1.K8s, Enable: kubeConf.Cluster.Kubernetes.EnableNodelocaldns()},
		"calico-kube-controllers": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "kube-controllers", Tag: kubekeyapiv1alpha1.DefaultCalicoVersion, Group: kubekeyapiv1alpha1.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-cni":              {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "cni", Tag: kubekeyapiv1alpha1.DefaultCalicoVersion, Group: kubekeyapiv1alpha1.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-node":             {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "node", Tag: kubekeyapiv1alpha1.DefaultCalicoVersion, Group: kubekeyapiv1alpha1.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-flexvol":          {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "pod2daemon-flexvol", Tag: kubekeyapiv1alpha1.DefaultCalicoVersion, Group: kubekeyapiv1alpha1.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico")},
		"calico-typha":            {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "calico", Repo: "typha", Tag: kubekeyapiv1alpha1.DefaultCalicoVersion, Group: kubekeyapiv1alpha1.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "calico") && len(runtime.GetHostsByRole(common.K8s)) > 50},
		"flannel":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: kubekeyapiv1alpha1.DefaultKubeImageNamespace, Repo: "flannel", Tag: kubekeyapiv1alpha1.DefaultFlannelVersion, Group: kubekeyapiv1alpha1.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "flannel")},
		"cilium":                  {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "cilium", Repo: "cilium", Tag: kubekeyapiv1alpha1.DefaultCiliumVersion, Group: kubekeyapiv1alpha1.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "cilium")},
		"operator-generic":        {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "cilium", Repo: "operator-generic", Tag: kubekeyapiv1alpha1.DefaultCiliumVersion, Group: kubekeyapiv1alpha1.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "cilium")},
		"kubeovn":                 {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "kubeovn", Repo: "kube-ovn", Tag: kubekeyapiv1alpha1.DefaultKubeovnVersion, Group: kubekeyapiv1alpha1.K8s, Enable: strings.EqualFold(kubeConf.Cluster.Network.Plugin, "kubeovn")},
		// storage
		"provisioner-localpv": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "openebs", Repo: "provisioner-localpv", Tag: "2.10.1", Group: kubekeyapiv1alpha1.Worker, Enable: false},
		"linux-utils":         {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "openebs", Repo: "linux-utils", Tag: "2.10.0", Group: kubekeyapiv1alpha1.Worker, Enable: false},

		// load balancer
		"haproxy": {RepoAddr: kubeConf.Cluster.Registry.PrivateRegistry, Namespace: "library", Repo: "haproxy", Tag: "2.3", Group: kubekeyapiv1alpha1.Worker, Enable: kubeConf.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
	}

	image = ImageList[name]
	return image
}
