package kubesphere

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/version/kubesphere"
	"github.com/kubesphere/kubekey/pkg/pipelines/version/kubesphere/templates"
	"os"
	"path/filepath"
)

type DeployModule struct {
	common.KubeModule
	Skip bool
}

func (d *DeployModule) IsSkip() bool {
	return d.Skip
}

func (d *DeployModule) Init() {
	d.Name = "DeployKubeSphereModule"

	generateManifests := &modules.RemoteTask{
		Name:  "GenerateKsInstallerCRD",
		Desc:  "Generate KubeSphere ks-installer crd manifests",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action: &action.Template{
			Template: templates.KsInstaller,
			Dst:      filepath.Join(common.KubeAddonsDir, templates.KsInstaller.Name()),
			Data: util.Data{
				"Repo": MirrorRepo(d.KubeConf.Cluster.KubeSphere.Version),
				"Tag":  d.KubeConf.Cluster.KubeSphere.Version,
			},
		},
		Parallel: true,
	}

	addConfig := &modules.RemoteTask{
		Name:  "AddKsInstallerConfig",
		Desc:  "Add config to ks-installer manifests",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(AddInstallerConfig),
		Parallel: true,
	}

	createNamespace := &modules.RemoteTask{
		Name:  "CreateKubeSphereNamespace",
		Desc:  "Create the kubesphere namespace",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(CreateNamespace),
		Parallel: true,
	}

	setup := &modules.RemoteTask{
		Name:  "SetupKsInstallerConfig",
		Desc:  "Setup ks-installer config",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(Setup),
		Parallel: true,
	}

	apply := &modules.RemoteTask{
		Name:  "ApplyKsInstaller",
		Desc:  "Apply ks-installer",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(Apply),
		Parallel: true,
	}

	d.Tasks = []modules.Task{
		generateManifests,
		addConfig,
		createNamespace,
		setup,
		apply,
	}
}

func MirrorRepo(version string) string {
	var repo string

	_, ok := kubesphere.CNSource[version]
	if ok && os.Getenv("KKZONE") == "cn" {
		repo = "registry.cn-beijing.aliyuncs.com/kubesphereio"
	} else {
		if repo == "" {
			if kubesphere.PreRelease(version) {
				repo = "kubespheredev"
			} else {
				repo = "kubesphere"
			}
		} else {
			repo = fmt.Sprintf("%s/kubesphere", repo)
		}
	}
	return repo
}

type CheckResultModule struct {
	common.KubeModule
	Skip bool
}

func (c *CheckResultModule) IsSkip() bool {
	return c.Skip
}

func (c *CheckResultModule) Init() {
	c.Name = "CheckResultModule"

	check := &modules.RemoteTask{
		Name:  "CheckKsInstallerResult",
		Desc:  "Check ks-installer result",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(Check),
		Parallel: true,
	}

	c.Tasks = []modules.Task{
		check,
	}
}

type ConvertModule struct {
	common.KubeModule
}

func (c *ConvertModule) Init() {
	c.Name = "ConvertModule"

	convert := &modules.RemoteTask{
		Name:  "ConvertV2ToV3",
		Desc:  "Convert ks-installer config v2 to v3",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
			new(VersionBelowV3),
		},
		Action:   new(ConvertV2ToV3),
		Parallel: true,
	}

	c.Tasks = []modules.Task{
		convert,
	}
}
