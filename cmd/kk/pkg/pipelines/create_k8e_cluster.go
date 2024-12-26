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

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/addons"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/artifact"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/binaries"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/bootstrap/customscripts"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/bootstrap/os"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/bootstrap/precheck"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/certs"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/module"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/etcd"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/filesystem"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/images"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/k8e"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/kubernetes"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/kubesphere"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/loadbalancer"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/plugins/network"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/plugins/storage"
)

func NewK8eCreateClusterPipeline(runtime *common.KubeRuntime) error {
	noArtifact := runtime.Arg.Artifact == ""
	skipPushImages := runtime.Arg.SkipPushImages || noArtifact || (!noArtifact && runtime.Cluster.Registry.PrivateRegistry == "")
	skipLocalStorage := true
	if runtime.Arg.DeployLocalStorage != nil {
		skipLocalStorage = !*runtime.Arg.DeployLocalStorage
	} else if runtime.Cluster.KubeSphere.Enabled {
		skipLocalStorage = false
	}

	m := []module.Module{
		&precheck.GreetingsModule{},
		&artifact.UnArchiveModule{Skip: noArtifact},
		&os.RepositoryModule{Skip: noArtifact || !runtime.Arg.InstallPackages},
		&binaries.K8eNodeBinariesModule{},
		&os.ConfigureOSModule{Skip: runtime.Cluster.System.SkipConfigureOS},
		&customscripts.CustomScriptsModule{Phase: "PreInstall", Scripts: runtime.Cluster.System.PreInstall},
		&k8e.StatusModule{},
		&etcd.PreCheckModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.CertsModule{},
		&etcd.InstallETCDBinaryModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.ConfigureModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&etcd.BackupModule{Skip: runtime.Cluster.Etcd.Type != kubekeyapiv1alpha2.KubeKey},
		&loadbalancer.K3sKubevipModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
		&k8e.InstallKubeBinariesModule{},
		&k8e.InitClusterModule{},
		&k8e.StatusModule{},
		&k8e.JoinNodesModule{},
		&images.CopyImagesToRegistryModule{Skip: skipPushImages},
		&loadbalancer.K3sHaproxyModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabled()},
		&network.DeployNetworkPluginModule{},
		&kubernetes.ConfigureKubernetesModule{},
		&filesystem.ChownModule{},
		&certs.AutoRenewCertsModule{Skip: !runtime.Cluster.Kubernetes.EnableAutoRenewCerts()},
		&k8e.SaveKubeConfigModule{},
		&customscripts.CustomScriptsModule{Phase: "PostClusterInstall", Scripts: runtime.Cluster.System.PostClusterInstall},
		&addons.AddonsModule{Skip: runtime.Arg.SkipInstallAddons},
		&storage.DeployLocalVolumeModule{Skip: skipLocalStorage},
		&kubesphere.DeployModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
		&kubesphere.CheckResultModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
		&customscripts.CustomScriptsModule{Phase: "PostInstall", Scripts: runtime.Cluster.System.PostInstall},
	}

	p := pipeline.Pipeline{
		Name:    "K8eCreateClusterPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}

	if runtime.Cluster.KubeSphere.Enabled {

		fmt.Print(`Installation is complete.

Please check the result using the command:

	kubectl logs -n kubesphere-system $(kubectl get pod -n kubesphere-system -l 'app in (ks-install, ks-installer)' -o jsonpath='{.items[0].metadata.name}') -f   

`)
	} else {
		fmt.Print(`Installation is complete.

Please check the result using the command:
		
	kubectl get pod -A

`)

	}

	return nil
}
