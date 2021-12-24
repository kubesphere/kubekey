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

package create

import (
	"fmt"
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/kubesphere/kubekey/pkg/version/kubernetes"
	"github.com/kubesphere/kubekey/pkg/version/kubesphere"
	"github.com/spf13/cobra"
	"time"
)

type CreateClusterOptions struct {
	CommonOptions *options.CommonOptions

	ClusterCfgFile      string
	Kubernetes          string
	EnableKubeSphere    bool
	KubeSphere          string
	LocalStorage        bool
	SkipPullImages      bool
	SkipPushImages      bool
	ContainerManager    string
	DownloadCmd         string
	Artifact            string
	SkipInstallPackages bool
}

func NewCreateClusterOptions() *CreateClusterOptions {
	return &CreateClusterOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCreateCluster creates a new create cluster command
func NewCmdCreateCluster() *cobra.Command {
	o := NewCreateClusterOptions()
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Create a Kubernetes or KubeSphere cluster",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(cmd, args))
			util.CheckErr(o.Validate(cmd, args))
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	if err := completionSetting(cmd); err != nil {
		panic(fmt.Sprintf("Got error with the completion setting"))
	}
	return cmd
}

func (o *CreateClusterOptions) Complete(cmd *cobra.Command, args []string) error {
	var ksVersion string
	if o.EnableKubeSphere && len(args) > 0 {
		ksVersion = args[0]
	} else {
		ksVersion = kubesphere.Latest().Version
	}
	o.KubeSphere = ksVersion

	if o.Artifact == "" {
		o.SkipInstallPackages = false
		o.SkipPushImages = false
	}
	return nil
}

func (o *CreateClusterOptions) Validate(cmd *cobra.Command, args []string) error {
	switch o.ContainerManager {
	case common.Docker, common.Conatinerd, common.Crio, common.Isula:
	default:
		return fmt.Errorf("unsupport container runtime [%s]", o.ContainerManager)
	}
	return nil
}

func (o *CreateClusterOptions) Run() error {
	arg := common.Argument{
		FilePath:            o.ClusterCfgFile,
		KubernetesVersion:   o.Kubernetes,
		KsEnable:            o.EnableKubeSphere,
		KsVersion:           o.KubeSphere,
		SkipPullImages:      o.SkipPullImages,
		SKipPushImages:      o.SkipPushImages,
		InCluster:           o.CommonOptions.InCluster,
		DeployLocalStorage:  o.LocalStorage,
		Debug:               o.CommonOptions.Verbose,
		IgnoreErr:           o.CommonOptions.IgnoreErr,
		SkipConfirmCheck:    o.CommonOptions.SkipConfirmCheck,
		ContainerManager:    o.ContainerManager,
		Artifact:            o.Artifact,
		SkipInstallPackages: o.SkipInstallPackages,
	}

	return pipelines.CreateCluster(arg, o.DownloadCmd)
}

func (o *CreateClusterOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
	cmd.Flags().BoolVarP(&o.LocalStorage, "with-local-storage", "", false, "Deploy a local PV provisioner")
	cmd.Flags().BoolVarP(&o.EnableKubeSphere, "with-kubesphere", "", false, "Deploy a specific version of kubesphere (default v3.2.0)")
	cmd.Flags().BoolVarP(&o.SkipPullImages, "skip-pull-images", "", false, "Skip pre pull images")
	cmd.Flags().BoolVarP(&o.SkipPushImages, "skip-push-images", "", false, "Skip pre push images")
	cmd.Flags().StringVarP(&o.ContainerManager, "container-manager", "", "docker", "Container runtime: docker, crio, containerd and isula.")
	cmd.Flags().StringVarP(&o.DownloadCmd, "download-cmd", "", "curl -L -o %s %s",
		`The user defined command to download the necessary binary files. The first param '%s' is output path, the second param '%s', is the URL`)
	cmd.Flags().StringVarP(&o.Artifact, "artifact", "a", "", "Path to a KubeKey artifact")
	cmd.Flags().BoolVarP(&o.SkipInstallPackages, "skip-install-packages", "", false, "Skip install packages by artifact")
}

func completionSetting(cmd *cobra.Command) (err error) {
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) (
		strings []string, directive cobra.ShellCompDirective) {
		versionArray := kubesphere.VersionsStringArr()
		versionArray = append(versionArray, time.Now().Add(-time.Hour*24).Format("nightly-20060102"))
		return versionArray, cobra.ShellCompDirectiveNoFileComp
	}

	err = cmd.RegisterFlagCompletionFunc("with-kubernetes", func(cmd *cobra.Command, args []string, toComplete string) (
		strings []string, directive cobra.ShellCompDirective) {
		return kubernetes.SupportedK8sVersionList(), cobra.ShellCompDirectiveNoFileComp
	})
	return
}
