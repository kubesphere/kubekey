package storage

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	local_volume "github.com/kubesphere/kubekey/pkg/plugins/storage/local-volume"
	nfs_client "github.com/kubesphere/kubekey/pkg/plugins/storage/nfs-client"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
)

func DeployStoragePlugins(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Deploying storage plugin ...")

	return mgr.RunTaskOnMasterNodes(deployStoragePlugins, true)
}

func deployStoragePlugins(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	if mgr.Runner.Index == 0 {
		mgr.Runner.RunCmd("sudo -E /bin/sh -c \"mkdir -p /etc/kubernetes/addons\" && /usr/local/bin/helm repo add kubesphere https://charts.kubesphere.io/qingcloud")
		if mgr.Cluster.Storage.LocalVolume.StorageClassName != "" {
			if err := DeployLocalVolume(mgr); err != nil {
				return err
			}
		}
		if mgr.Cluster.Storage.NfsClient.StorageClassName != "" {
			if err := DeployNfsClient(mgr); err != nil {
				return err
			}
		}
	}
	return nil
}

func DeployLocalVolume(mgr *manager.Manager) error {
	localVolumeFile, err := local_volume.GenerateOpenebsManifests(mgr)
	if err != nil {
		return err
	}
	localVolumeFileBase64 := base64.StdEncoding.EncodeToString([]byte(localVolumeFile))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/addons/local-volume.yaml\"", localVolumeFileBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate local-volume file")
	}

	_, err2 := mgr.Runner.RunCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/local-volume.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to deploy local-volume.yaml")
	}
	return nil
}

func DeployNfsClient(mgr *manager.Manager) error {

	_, err1 := mgr.Runner.RunCmd("sudo -E /bin/sh -c \"rm -rf /etc/kubernetes/addons/nfs-client-provisioner && /usr/local/bin/helm fetch kubesphere/nfs-client-provisioner -d /etc/kubernetes/addons --untar\"")
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to fetch nfs-client-provisioner chart")
	}

	nfsClientValuesFile, err := nfs_client.GenerateNfsClientValuesFile(mgr)
	if err != nil {
		return err
	}
	nfsClientValuesFileBase64 := base64.StdEncoding.EncodeToString([]byte(nfsClientValuesFile))
	_, err2 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/addons/custom-values-nfs-client.yaml\"", nfsClientValuesFileBase64))
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to generate nfs-client values file")
	}

	_, err3 := mgr.Runner.RunCmd("sudo -E /bin/sh -c \"/usr/local/bin/helm upgrade -i nfs-client /etc/kubernetes/addons/nfs-client-provisioner -f /etc/kubernetes/addons/custom-values-nfs-client.yaml -n kube-system\"")
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "Failed to deploy nfs-client-provisioner")
	}
	return nil
}
