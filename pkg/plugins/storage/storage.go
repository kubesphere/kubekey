/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package storage

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	ceph_rbd "github.com/kubesphere/kubekey/pkg/plugins/storage/ceph-rbd"
	csi_neonsan "github.com/kubesphere/kubekey/pkg/plugins/storage/csi-neonsan"
	"github.com/kubesphere/kubekey/pkg/plugins/storage/glusterfs"
	local_volume "github.com/kubesphere/kubekey/pkg/plugins/storage/local-volume"
	nfs_client "github.com/kubesphere/kubekey/pkg/plugins/storage/nfs-client"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
)

func DeployStoragePlugins(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Deploying storage plugin ...")

	err := mgr.RunTaskOnMasterNodes(deployStoragePlugins, true)
	if err != nil {
		return err
	}
	if mgr.Cluster.Storage.NeonsanCSI.Enable {
		return mgr.RunTaskOnK8sNodes(csi_neonsan.InstallNeonsanPlugin, true)
	}
	return nil
}

func deployStoragePlugins(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	_, _ = node, conn
	if mgr.Runner.Index == 0 {
		mgr.Runner.RunCmdOutput("sudo -E /bin/sh -c \"mkdir -p /etc/kubernetes/addons\" && /usr/local/bin/helm repo add kubesphere https://charts.kubesphere.io/qingcloud")
		if mgr.Cluster.Storage.LocalVolume.Enabled {
			if err := DeployLocalVolume(mgr); err != nil {
				return err
			}
		}
		if mgr.Cluster.Storage.NfsClient.Enabled {
			if err := DeployNfsClient(mgr); err != nil {
				return err
			}
		}
		if mgr.Cluster.Storage.CephRBD.Enabled {
			if err := DeployRBDProvisioner(mgr); err != nil {
				return err
			}
		}
		if mgr.Cluster.Storage.GlusterFS.Enabled {
			if err := DeployGlusterFS(mgr); err != nil {
				return err
			}
		}
		if mgr.Cluster.Storage.NeonsanCSI.Enable {
			if err := csi_neonsan.DeployNeonsanCSI(mgr); err != nil {
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
	_, err1 := mgr.Runner.RunCmdOutput(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/addons/local-volume.yaml\"", localVolumeFileBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate local-volume manifests")
	}

	_, err2 := mgr.Runner.RunCmdOutput("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/local-volume.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to deploy local-volume.yaml")
	}
	return nil
}

func DeployNfsClient(mgr *manager.Manager) error {

	_, err1 := mgr.Runner.RunCmdOutput("sudo -E /bin/sh -c \"rm -rf /etc/kubernetes/addons/nfs-client-provisioner && /usr/local/bin/helm fetch kubesphere/nfs-client-provisioner -d /etc/kubernetes/addons --untar\"")
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

	_, err3 := mgr.Runner.RunCmdOutput("sudo -E /bin/sh -c \"/usr/local/bin/helm upgrade -i nfs-client /etc/kubernetes/addons/nfs-client-provisioner -f /etc/kubernetes/addons/custom-values-nfs-client.yaml -n kube-system\"")
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "Failed to deploy nfs-client-provisioner")
	}
	return nil
}

func DeployRBDProvisioner(mgr *manager.Manager) error {
	RBDProvisionerFile, err := ceph_rbd.GenerateRBDProvisionerManifests(mgr)
	if err != nil {
		return err
	}
	RBDProvisionerFileBase64 := base64.StdEncoding.EncodeToString([]byte(RBDProvisionerFile))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/addons/rbd-provisioner.yaml\"", RBDProvisionerFileBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate rbd-provisioner manifests")
	}

	_, err2 := mgr.Runner.RunCmdOutput("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/rbd-provisioner.yaml -n kube-system")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to deploy rbd-provisioner.yaml")
	}
	return nil
}

func DeployGlusterFS(mgr *manager.Manager) error {
	glusterfsFile, err := glusterfs.GenerateGlusterFSManifests(mgr)
	if err != nil {
		return err
	}
	glusterfsFileBase64 := base64.StdEncoding.EncodeToString([]byte(glusterfsFile))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/addons/glusterfs.yaml\"", glusterfsFileBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate glusterfs manifests")
	}

	_, err2 := mgr.Runner.RunCmdOutput("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/glusterfs.yaml -n kube-system")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to deploy glusterfs.yaml")
	}
	return nil
}
