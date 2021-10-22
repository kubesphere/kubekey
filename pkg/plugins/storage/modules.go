package storage

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/plugins/storage/templates"
	"path/filepath"
)

type DeployLocalVolumeModule struct {
	common.KubeModule
	Skip bool
}

func (d *DeployLocalVolumeModule) IsSkip() bool {
	return d.Skip
}

func (d *DeployLocalVolumeModule) Init() {
	d.Name = "DeployStorageClassModule"

	generate := &module.RemoteTask{
		Name:  "GenerateOpenEBSManifest",
		Desc:  "Generate OpenEBS manifest",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CheckDefaultStorageClass),
		},
		Action: &action.Template{
			Template: templates.OpenEBS,
			Dst:      filepath.Join(common.KubeAddonsDir, templates.OpenEBS.Name()),
			Data: util.Data{
				"ProvisionerLocalPVImage": images.GetImage(d.Runtime, d.KubeConf, "provisioner-localpv").ImageName(),
				"LinuxUtilsImage":         images.GetImage(d.Runtime, d.KubeConf, "linux-utils").ImageName(),
			},
		},
		Parallel: true,
	}

	deploy := &module.RemoteTask{
		Name:  "DeployOpenEBS",
		Desc:  "Deploy OpenEBS as cluster default StorageClass",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CheckDefaultStorageClass),
		},
		Action:   new(DeployLocalVolume),
		Parallel: true,
	}

	d.Tasks = []module.Task{
		generate,
		deploy,
	}
}
