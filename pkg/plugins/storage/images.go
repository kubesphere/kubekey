package storage

import (
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
)

func prePullStorageImages(mgr *manager.Manager, node *kubekeyapi.HostCfg, _ ssh.Connection) error {
	i := images.Images{}
	i.Images = []images.Image{
		preinstall.GetImage(mgr, "provisioner-localpv"),
		preinstall.GetImage(mgr, "node-disk-manager"),
		preinstall.GetImage(mgr, "node-disk-operator"),
		preinstall.GetImage(mgr, "linux-utils"),
		preinstall.GetImage(mgr, "rbd-provisioner"),
		preinstall.GetImage(mgr, "nfs-client-provisioner"),
	}
	if err := i.PullImages(mgr, node); err != nil {
		return err
	}
	return nil
}
