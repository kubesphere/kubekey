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

package upgrade

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/pipelines"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/version/kubernetes"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/version/kubesphere"

	"github.com/kubesphere/kubekey/cmd/kk/cmd/options"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/upgrade/phase"
	"github.com/kubesphere/kubekey/cmd/kk/cmd/util"
)

type UpgradeOptions struct {
	CommonOptions    *options.CommonOptions
	ClusterCfgFile   string
	Kubernetes       string
	EnableKubeSphere bool
	KubeSphere       string
	SkipPullImages   bool
	DownloadCmd      string
	Artifact         string
}

func NewUpgradeOptions() *UpgradeOptions {
	return &UpgradeOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdUpgrade creates a new upgrade command
func NewCmdUpgrade() *cobra.Command {
	o := NewUpgradeOptions()
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade your cluster smoothly to a newer version with this command",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(cmd, args))
			util.CheckErr(o.Run())
		},
	}
	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	cmd.AddCommand(phase.NewPhaseCommand())

	if err := completionSetting(cmd); err != nil {
		panic(fmt.Sprintf("Got error with the completion setting"))
	}
	return cmd
}

func (o *UpgradeOptions) Complete(cmd *cobra.Command, args []string) error {
	var ksVersion string
	if o.EnableKubeSphere && len(args) > 0 {
		ksVersion = args[0]
	} else {
		ksVersion = kubesphere.Latest().Version
	}
	o.KubeSphere = ksVersion
	return nil
}

func (o *UpgradeOptions) Run() error {
	arg := common.Argument{
		FilePath:          o.ClusterCfgFile,
		KubernetesVersion: o.Kubernetes,
		KsEnable:          o.EnableKubeSphere,
		KsVersion:         o.KubeSphere,
		SkipPullImages:    o.SkipPullImages,
		Debug:             o.CommonOptions.Verbose,
		SkipConfirmCheck:  o.CommonOptions.SkipConfirmCheck,
		Artifact:          o.Artifact,
	}
	return pipelines.UpgradeCluster(arg, o.DownloadCmd)
}

func (o *UpgradeOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
	cmd.Flags().StringVarP(&o.Kubernetes, "with-kubernetes", "", "", "Specify a supported version of kubernetes")
	cmd.Flags().BoolVarP(&o.EnableKubeSphere, "with-kubesphere", "", false, fmt.Sprintf("Deploy a specific version of kubesphere (default %s)", kubesphere.Latest().Version))
	cmd.Flags().BoolVarP(&o.SkipPullImages, "skip-pull-images", "", false, "Skip pre pull images")
	cmd.Flags().StringVarP(&o.DownloadCmd, "download-cmd", "", "curl -L -o %s %s",
		`The user defined command to download the necessary binary files. The first param '%s' is output path, the second param '%s', is the URL`)
	cmd.Flags().StringVarP(&o.Artifact, "artifact", "a", "", "Path to a KubeKey artifact")
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
