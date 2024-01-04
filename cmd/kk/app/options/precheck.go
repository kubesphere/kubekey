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

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	cliflag "k8s.io/component-base/cli/flag"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
)

type PreCheckOptions struct {
	// Playbook which to execute.
	Playbook string
	// HostFile is the path of host file
	InventoryFile string
	// ConfigFile is the path of config file
	ConfigFile string
	// WorkDir is the baseDir which command find any resource (project etc.)
	WorkDir string
	// Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.
	Debug bool
}

func NewPreCheckOption() *PreCheckOptions {
	o := &PreCheckOptions{
		WorkDir: "/var/lib/kubekey",
	}
	return o
}

func (o *PreCheckOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.StringVar(&o.WorkDir, "work-dir", o.WorkDir, "the base Dir for kubekey. Default current dir. ")
	gfs.StringVar(&o.ConfigFile, "config", o.ConfigFile, "the config file path. support *.yaml ")
	gfs.StringVar(&o.InventoryFile, "inventory", o.InventoryFile, "the host list file path. support *.ini")
	gfs.BoolVar(&o.Debug, "debug", o.Debug, "Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.")

	return fss
}

func (o *PreCheckOptions) Complete(cmd *cobra.Command, args []string) (*kubekeyv1.Pipeline, error) {
	kk := &kubekeyv1.Pipeline{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pipeline",
			APIVersion: "kubekey.kubesphere.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              fmt.Sprintf("precheck-%s", rand.String(6)),
			Namespace:         metav1.NamespaceDefault,
			UID:               types.UID(uuid.NewString()),
			CreationTimestamp: metav1.Now(),
			Annotations: map[string]string{
				kubekeyv1.BuiltinsProjectAnnotation: "",
			},
		},
	}

	// complete playbook. now only support one playbook
	if len(args) == 1 {
		o.Playbook = args[0]
	} else {
		return nil, fmt.Errorf("%s\nSee '%s -h' for help and examples", cmd.Use, cmd.CommandPath())
	}

	kk.Spec = kubekeyv1.PipelineSpec{
		Playbook: o.Playbook,
		Debug:    o.Debug,
	}
	return kk, nil
}
