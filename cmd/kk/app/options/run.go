/*
Copyright 2023 The KubeSphere Authors.

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

package options

import (
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
)

type KubekeyRunOptions struct {
	CommonOptions
	// ProjectAddr is the storage for executable packages (in Ansible format).
	// When starting with http or https, it will be obtained from a Git repository.
	// When starting with file path, it will be obtained from the local path.
	ProjectAddr string
	// ProjectName is the name of project. it will store to project dir use this name.
	// If empty generate from ProjectAddr
	ProjectName string
	// ProjectBranch is the git branch of the git Addr.
	ProjectBranch string
	// ProjectTag if the git tag of the git Addr.
	ProjectTag string
	// ProjectInsecureSkipTLS skip tls or not when git addr is https.
	ProjectInsecureSkipTLS bool
	// ProjectToken to clone and pull git project
	ProjectToken string
	// Tags is the tags of playbook which to execute
	Tags []string
	// SkipTags is the tags of playbook which skip execute
	SkipTags []string
}

func NewKubeKeyRunOptions() *KubekeyRunOptions {
	// add default values
	o := &KubekeyRunOptions{
		CommonOptions: newCommonOptions(),
	}
	return o
}

func (o *KubekeyRunOptions) Flags() cliflag.NamedFlagSets {
	fss := o.CommonOptions.Flags()
	gitfs := fss.FlagSet("project")
	gitfs.StringVar(&o.ProjectAddr, "project-addr", o.ProjectAddr, "the storage for executable packages (in Ansible format)."+
		" When starting with http or https, it will be obtained from a Git repository."+
		"When starting with file path, it will be obtained from the local path.")
	gitfs.StringVar(&o.ProjectBranch, "project-branch", o.ProjectBranch, "the git branch of the remote Addr")
	gitfs.StringVar(&o.ProjectTag, "project-tag", o.ProjectTag, "the git tag of the remote Addr")
	gitfs.BoolVar(&o.ProjectInsecureSkipTLS, "project-insecure-skip-tls", o.ProjectInsecureSkipTLS, "skip tls or not when git addr is https.")
	gitfs.StringVar(&o.ProjectToken, "project-token", o.ProjectToken, "the token for private project.")

	tfs := fss.FlagSet("tags")
	tfs.StringArrayVar(&o.Tags, "tags", o.Tags, "the tags of playbook which to execute")
	tfs.StringArrayVar(&o.SkipTags, "skip_tags", o.SkipTags, "the tags of playbook which skip execute")

	return fss
}

func (o *KubekeyRunOptions) Complete(cmd *cobra.Command, args []string) (*kubekeyv1.Pipeline, *kubekeyv1.Config, *kubekeyv1.Inventory, error) {
	pipeline := &kubekeyv1.Pipeline{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "run-",
			Namespace:    metav1.NamespaceDefault,
			Annotations:  map[string]string{},
		},
	}
	// complete playbook. now only support one playbook
	if len(args) == 1 {
		o.Playbook = args[0]
	} else {
		return nil, nil, nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}

	pipeline.Spec = kubekeyv1.PipelineSpec{
		Project: kubekeyv1.PipelineProject{
			Addr:            o.ProjectAddr,
			Name:            o.ProjectName,
			Branch:          o.ProjectBranch,
			Tag:             o.ProjectTag,
			InsecureSkipTLS: o.ProjectInsecureSkipTLS,
			Token:           o.ProjectToken,
		},
		Playbook: o.Playbook,
		Tags:     o.Tags,
		SkipTags: o.SkipTags,
		Debug:    o.Debug,
	}
	config, inventory, err := o.completeRef(pipeline)
	if err != nil {
		return nil, nil, nil, err
	}

	return pipeline, config, inventory, nil
}
