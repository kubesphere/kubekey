/*
 Copyright 2022 The KubeSphere Authors.

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

package images

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kubesphere/kubekey/v2/cmd/ctl/options"
	"github.com/kubesphere/kubekey/v2/cmd/ctl/util"
	"github.com/kubesphere/kubekey/v2/pkg/artifact"
	"github.com/kubesphere/kubekey/v2/pkg/bootstrap/precheck"
	"github.com/kubesphere/kubekey/v2/pkg/common"
	"github.com/kubesphere/kubekey/v2/pkg/core/module"
	"github.com/kubesphere/kubekey/v2/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/v2/pkg/filesystem"
	"github.com/kubesphere/kubekey/v2/pkg/images"
)

type ArtifactImagesPushOptions struct {
	CommonOptions *options.CommonOptions

	ImageDirPath   string
	Artifact       string
	ClusterCfgFile string
}

func NewArtifactImagesPushOptions() *ArtifactImagesPushOptions {
	return &ArtifactImagesPushOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdArtifactImagesPush creates a new `kubekey artifacts images push` command
func NewCmdArtifactImagesPush() *cobra.Command {
	o := NewArtifactImagesPushOptions()
	cmd := &cobra.Command{
		Use:   "push",
		Short: "push images to a registry from an artifact",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(cmd, args))
			util.CheckErr(o.Validate(args))
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *ArtifactImagesPushOptions) Complete(_ *cobra.Command, _ []string) error {
	if o.ImageDirPath == "" && o.Artifact == "" {
		currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return errors.Wrap(err, "failed to get current directory")
		}
		o.ImageDirPath = filepath.Join(currentDir, "kubekey", "images")
	}

	return nil
}

func (o *ArtifactImagesPushOptions) Validate(_ []string) error {
	if o.ClusterCfgFile == "" {
		return errors.New("kubekey config file is required")
	}
	if o.ImageDirPath != "" && o.Artifact != "" {
		return errors.New("only one of --image-dir or --artifact can be specified")
	}
	return nil
}

func (o *ArtifactImagesPushOptions) Run() error {
	arg := common.Argument{
		ImagesDir: o.ImageDirPath,
		Artifact:  o.Artifact,
		FilePath:  o.ClusterCfgFile,
		Debug:     o.CommonOptions.Verbose,
		IgnoreErr: o.CommonOptions.IgnoreErr,
	}
	return runPush(arg)
}

func (o *ArtifactImagesPushOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ImageDirPath, "images-dir", "", "", "Path to a KubeKey artifact images directory")
	cmd.Flags().StringVarP(&o.Artifact, "artifact", "a", "", "Path to a KubeKey artifact")
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}

func runPush(arg common.Argument) error {
	runtime, err := common.NewKubeRuntime(common.File, arg)
	if err != nil {
		return err
	}

	if err := newImagesPushPipeline(runtime); err != nil {
		return err
	}

	return nil
}

func newImagesPushPipeline(runtime *common.KubeRuntime) error {
	noArtifact := runtime.Arg.Artifact == ""

	m := []module.Module{
		&precheck.GreetingsModule{},
		&artifact.UnArchiveModule{Skip: noArtifact},
		&images.CopyImagesToRegistryModule{ImagePath: runtime.Arg.ImagesDir},
		&filesystem.ChownWorkDirModule{},
	}

	p := pipeline.Pipeline{
		Name:    "ArtifactImagesPushPipeline",
		Modules: m,
		Runtime: runtime,
	}

	if err := p.Start(); err != nil {
		return err
	}

	return nil
}
