/*
 Copyright 2021 The KubeSphere Authors.

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

package kubesphere

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/prepare"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/task"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/version/kubesphere"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/version/kubesphere/templates"
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
	d.Desc = "Deploy KubeSphere"

	generateManifests := &task.RemoteTask{
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
				"Repo": MirrorRepo(d.KubeConf),
				"Tag":  d.KubeConf.Cluster.KubeSphere.Version,
			},
		},
		Parallel: true,
	}

	addConfig := &task.RemoteTask{
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

	createNamespace := &task.RemoteTask{
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

	setup := &task.RemoteTask{
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

	apply := &task.RemoteTask{
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

	d.Tasks = []task.Interface{
		generateManifests,
		// apply crd installer.kubesphere.io/v1alpha1
		apply,
		addConfig,
		createNamespace,
		setup,
		apply,
	}
}

func MirrorRepo(kubeConf *common.KubeConf) string {
	repo := kubeConf.Cluster.Registry.PrivateRegistry
	namespaceOverride := kubeConf.Cluster.Registry.NamespaceOverride
	version := kubeConf.Cluster.KubeSphere.Version

	_, ok := kubesphere.CNSource[version]
	if ok && os.Getenv("KKZONE") == "cn" {
		if repo == "" {
			repo = "registry.cn-beijing.aliyuncs.com/kubesphereio"
		} else if len(namespaceOverride) != 0 {
			repo = fmt.Sprintf("%s/%s", repo, namespaceOverride)
		} else {
			repo = fmt.Sprintf("%s/kubesphere", repo)
		}
	} else {
		if repo == "" {
			_, latest := kubesphere.LatestRelease(version)
			_, dev := kubesphere.DevRelease(version)
			_, stable := kubesphere.StabledVersionSupport(version)
			switch {
			case stable:
				repo = "kubesphere"
			case dev:
				repo = "kubespheredev"
			case latest:
				repo = "kubespheredev"
			default:
				repo = "kubesphere"
			}
		} else if len(namespaceOverride) != 0 {
			repo = fmt.Sprintf("%s/%s", repo, namespaceOverride)
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
	c.Desc = "Check deploy KubeSphere result"

	check := &task.RemoteTask{
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

	c.Tasks = []task.Interface{
		check,
	}
}

type CleanClusterConfigurationModule struct {
	common.KubeModule
	Skip bool
}

func (c *CleanClusterConfigurationModule) IsSkip() bool {
	return c.Skip
}

func (c *CleanClusterConfigurationModule) Init() {
	c.Name = "CleanClusterConfigurationModule"
	c.Desc = "Clean redundant ClusterConfiguration config"

	// ensure there is no cc config, and prevent to reset cc config when upgrade the cluster
	clean := &task.LocalTask{
		Name:   "CleanClusterConfiguration",
		Desc:   "Clean redundant ClusterConfiguration config",
		Action: new(CleanCC),
	}

	c.Tasks = []task.Interface{
		clean,
	}
}

type ConvertModule struct {
	common.KubeModule
	Skip bool
}

func (c *ConvertModule) IsSkip() bool {
	return c.Skip
}

func (c *ConvertModule) Init() {
	c.Name = "ConvertModule"
	c.Desc = "Convert ks-installer config v2 to v3"

	convert := &task.RemoteTask{
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

	c.Tasks = []task.Interface{
		convert,
	}
}
