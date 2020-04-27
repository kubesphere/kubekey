package kubernetes

import (
	"fmt"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/images"
	"github.com/pixiake/kubekey/pkg/util/manager"
)

func PreDownloadNodeImages(mgr *manager.Manager, node *kubekeyapi.HostCfg) {
	imagesList := []images.Image{
		{Prefix: images.GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, kubekeyapi.DefaultKubeImageRepo), Repo: "pause", Tag: "3.1"},
		{Prefix: images.GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, "coredns"), Repo: "coredns", Tag: "1.6.0"},
		{Prefix: images.GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, "node"), Repo: "coredns", Tag: "1.6.0"},
	}

	for _, image := range imagesList {
		fmt.Printf("[%s] Download image: %s\n", node.HostName, image.NewImage())
		mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh \"docker pull %s\"", image.NewImage()))
	}
}
