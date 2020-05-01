package images

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"strings"
)

type Image struct {
	Prefix string
	Repo   string
	Tag    string
	Group  string
	Enable bool
}

func (image *Image) NewImage() string {
	return fmt.Sprintf("%s%s:%s", image.Prefix, image.Repo, image.Tag)
}

func GetImagePrefix(privateRegistry, ns string) string {
	var prefix string
	if privateRegistry == "" {
		if ns == "" {
			prefix = ""
		} else {
			prefix = fmt.Sprintf("%s/", ns)
		}
	} else {
		if ns == "" {
			prefix = fmt.Sprintf("%s/library/", privateRegistry)
		} else {
			prefix = fmt.Sprintf("%s/%s/", privateRegistry, ns)
		}
	}
	return prefix
}

func PreDownloadImages(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Pre-download images")

	return mgr.RunTaskOnAllNodes(preDownloadImages, true)
}

func preDownloadImages(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	imagesList := []Image{
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, kubekeyapi.DefaultKubeImageRepo), Repo: "etcd", Tag: kubekeyapi.DefaultEtcdVersion, Group: "etcd", Enable: true},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, kubekeyapi.DefaultKubeImageRepo), Repo: "pause", Tag: "3.1", Group: "k8s", Enable: true},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, kubekeyapi.DefaultKubeImageRepo), Repo: "kube-apiserver", Tag: mgr.Cluster.KubeCluster.Version, Group: "master", Enable: true},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, kubekeyapi.DefaultKubeImageRepo), Repo: "kube-controller-manager", Tag: mgr.Cluster.KubeCluster.Version, Group: "master", Enable: true},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, kubekeyapi.DefaultKubeImageRepo), Repo: "kube-scheduler", Tag: mgr.Cluster.KubeCluster.Version, Group: "master", Enable: true},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, kubekeyapi.DefaultKubeImageRepo), Repo: "kube-proxy", Tag: mgr.Cluster.KubeCluster.Version, Group: "k8s", Enable: true},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, "coredns"), Repo: "coredns", Tag: "1.6.0", Group: "k8s", Enable: true},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, "calico"), Repo: "kube-controllers", Tag: kubekeyapi.DefaultCalicoVersion, Group: "k8s", Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, "calico"), Repo: "cni", Tag: kubekeyapi.DefaultCalicoVersion, Group: "k8s", Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, "calico"), Repo: "node", Tag: kubekeyapi.DefaultCalicoVersion, Group: "k8s", Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, "calico"), Repo: "pod2daemon-flexvol", Tag: kubekeyapi.DefaultCalicoVersion, Group: "k8s", Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "calico")},
		{Prefix: GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, "kubesphere"), Repo: "flannel", Tag: kubekeyapi.DefaultFlannelVersion, Group: "k8s", Enable: strings.EqualFold(mgr.Cluster.Network.Plugin, "flannel")},
	}

	for _, image := range imagesList {

		if node.IsMaster && image.Group == "master" && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.HostName, image.NewImage())
			_, err := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E docker pull %s", image.NewImage()))
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to download image: %s", image.NewImage()))
			}
		}
		if (node.IsMaster || node.IsWorker) && image.Group == "k8s" && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.HostName, image.NewImage())
			_, err := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E docker pull %s", image.NewImage()))
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to download image: %s", image.NewImage()))
			}
		}
		if node.IsEtcd && image.Group == "etcd" && image.Enable {
			fmt.Printf("[%s] Downloading image: %s\n", node.HostName, image.NewImage())
			_, err := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E docker pull %s", image.NewImage()))
			if err != nil {
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to download image: %s", image.NewImage()))
			}
		}
	}

	return nil
}
