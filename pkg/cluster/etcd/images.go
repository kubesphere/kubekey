package etcd

import (
	"fmt"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/images"
	"github.com/pixiake/kubekey/pkg/util/manager"
)

func PreDownloadEtcdImages(mgr *manager.Manager, node *kubekeyapi.HostCfg) {
	imagesList := []images.Image{
		{Prefix: images.GetImagePrefix(mgr.Cluster.Registry.PrivateRegistry, kubekeyapi.DefaultKubeImageRepo), Repo: "etcd", Tag: kubekeyapi.DefaultEtcdVersion},
	}

	for _, image := range imagesList {
		fmt.Printf("[%s] Download image: %s\n", node.HostName, image.NewImage())
		mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh \"docker pull %s\"", image.NewImage()))
	}
}
