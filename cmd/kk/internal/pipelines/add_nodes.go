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

package pipelines

import (
	"fmt"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/cmd/kk/internal/artifact"
	"github.com/kubesphere/kubekey/cmd/kk/internal/binaries"
	"github.com/kubesphere/kubekey/cmd/kk/internal/bootstrap/confirm"
	"github.com/kubesphere/kubekey/cmd/kk/internal/bootstrap/os"
	"github.com/kubesphere/kubekey/cmd/kk/internal/bootstrap/precheck"
	"github.com/kubesphere/kubekey/cmd/kk/internal/bootstrap/registry"
	"github.com/kubesphere/kubekey/cmd/kk/internal/certs"
	"github.com/kubesphere/kubekey/cmd/kk/internal/common"
	"github.com/kubesphere/kubekey/cmd/kk/internal/container"
	"github.com/kubesphere/kubekey/cmd/kk/internal/etcd"
	"github.com/kubesphere/kubekey/cmd/kk/internal/filesystem"
	"github.com/kubesphere/kubekey/cmd/kk/internal/hooks"
	"github.com/kubesphere/kubekey/cmd/kk/internal/images"
	"github.com/kubesphere/kubekey/cmd/kk/internal/k3s"
	"github.com/kubesphere/kubekey/cmd/kk/internal/kubernetes"
	"github.com/kubesphere/kubekey/cmd/kk/internal/loadbalancer"
	kubekeycontroller "github.com/kubesphere/kubekey/controllers/kubekey"

	"github.com/kubesphere/kubekey/util/workflow/module"
	"github.com/kubesphere/kubekey/util/workflow/pipeline"
)

func NewAddNodesPipeline(runtime *common.KubeRuntime) error {
	noArtifact := runtime.Arg.Artifact == ""

	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.NodePreCheckModule{},
		&confirm.InstallConfirmModule{Skip: runtime.Arg.SkipConfirmCheck},
		&artifact.UnArchiveModule{Skip: noArtifact},
		&os.RepositoryModule{Skip: noArtifact || !runtime.Arg.InstallPackages},
		&binaries.NodeBinariesModule{},
		&os.ConfigureOSModule{},
		&registry.RegistryCertsModule{Skip: len(runtime.GetHostsByRole(common.Registry)) == 0},
		&kubernetes.StatusModule{},
		&container.InstallContainerModule{},
		&images.PullModule{Skip: runtime.Arg.SkipPullImages},
		&etcd.PreCheckModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.CertsModule{},
		&etcd.InstallETCDBinaryModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.ConfigureModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.BackupModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&kubernetes.InstallKubeBinariesModule{},
		&kubernetes.JoinNodesModule{},
		&loadbalancer.HaproxyModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
		&kubernetes.ConfigureKubernetesModule{},
		&filesystem.ChownModule{},
		&certs.AutoRenewCertsModule{Skip: !runtime.Cluster.Kubernetes.EnableAutoRenewCerts()},
	}

	p := pipeline.Pipeline{
		Name:            "AddNodesPipeline",
		Modules:         m,
		Runtime:         runtime,
		ModulePostHooks: []module.PostHookInterface{&hooks.UpdateCRStatusHook{}},
	}
	if err := p.Start(); err != nil {
		if runtime.Arg.InCluster {
			if err := kubekeycontroller.PatchNodeImportStatus(runtime, kubekeycontroller.Failed); err != nil {
				return err
			}
			if err := kubekeycontroller.UpdateStatus(runtime); err != nil {
				return err
			}
		}
		return err
	}

	if runtime.Arg.InCluster {
		if err := kubekeycontroller.PatchNodeImportStatus(runtime, kubekeycontroller.Success); err != nil {
			return err
		}
		if err := kubekeycontroller.UpdateStatus(runtime); err != nil {
			return err
		}
	}

	return nil
}

func NewK3sAddNodesPipeline(runtime *common.KubeRuntime) error {
	noArtifact := runtime.Arg.Artifact == ""

	m := []module.Module{
		&precheck.GreetingsModule{},
		&artifact.UnArchiveModule{Skip: noArtifact},
		&os.RepositoryModule{Skip: noArtifact || !runtime.Arg.InstallPackages},
		&binaries.K3sNodeBinariesModule{},
		&os.ConfigureOSModule{},
		&k3s.StatusModule{},
		&etcd.PreCheckModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.CertsModule{},
		&etcd.InstallETCDBinaryModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.ConfigureModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.BackupModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&k3s.InstallKubeBinariesModule{},
		&k3s.JoinNodesModule{},
		&loadbalancer.K3sHaproxyModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
		&kubernetes.ConfigureKubernetesModule{},
		&filesystem.ChownModule{},
		&certs.AutoRenewCertsModule{Skip: !runtime.Cluster.Kubernetes.EnableAutoRenewCerts()},
	}

	p := pipeline.Pipeline{
		Name:            "AddNodesPipeline",
		Modules:         m,
		Runtime:         runtime,
		ModulePostHooks: []module.PostHookInterface{&hooks.UpdateCRStatusHook{}},
	}
	if err := p.Start(); err != nil {
		if runtime.Arg.InCluster {
			if err := kubekeycontroller.PatchNodeImportStatus(runtime, kubekeycontroller.Failed); err != nil {
				return err
			}
			if err := kubekeycontroller.UpdateStatus(runtime); err != nil {
				return err
			}
		}
		return err
	}

	if runtime.Arg.InCluster {
		if err := kubekeycontroller.PatchNodeImportStatus(runtime, kubekeycontroller.Success); err != nil {
			return err
		}
		if err := kubekeycontroller.UpdateStatus(runtime); err != nil {
			return err
		}
	}

	return nil
}

func AddNodes(args common.Argument, downloadCmd string) error {
	args.DownloadCommand = func(path, url string) string {
		// this is an extension point for downloading tools, for example users can set the timeout, proxy or retry under
		// some poor network environment. Or users even can choose another cli, it might be wget.
		// perhaps we should have a build-in download function instead of totally rely on the external one
		return fmt.Sprintf(downloadCmd, path, url)
	}

	var loaderType string
	if args.FilePath != "" {
		loaderType = common.File
	} else {
		loaderType = common.AllInOne
	}

	runtime, err := common.NewKubeRuntime(loaderType, args)
	if err != nil {
		return err
	}
	if args.InCluster {
		c, err := kubekeycontroller.NewKubekeyClient()
		if err != nil {
			return err
		}
		runtime.ClientSet = c
	}

	if runtime.Arg.InCluster {
		if err := kubekeycontroller.CreateNodeForCluster(runtime); err != nil {
			return err
		}
		if err := kubekeycontroller.ClearConditions(runtime); err != nil {
			return err
		}
	}

	switch runtime.Cluster.Kubernetes.Type {
	case common.K3s:
		if err := NewK3sAddNodesPipeline(runtime); err != nil {
			return err
		}
	case common.Kubernetes:
		if err := NewAddNodesPipeline(runtime); err != nil {
			return err
		}
	default:
		if err := NewAddNodesPipeline(runtime); err != nil {
			return err
		}
	}
	return nil
}
