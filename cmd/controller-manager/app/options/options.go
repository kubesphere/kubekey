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
	"flag"
	"strings"

	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
)

type ControllerManagerServerOptions struct {
	// Enable gops or not.
	GOPSEnabled bool
	// WorkDir is the baseDir which command find any resource (project etc.)
	WorkDir string
	// Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.
	Debug bool
	// ControllerGates is the list of controller gates to enable or disable controller.
	// '*' means "all enabled by default controllers"
	// 'foo' means "enable 'foo'"
	// '-foo' means "disable 'foo'"
	// first item for a particular name wins.
	//     e.g. '-foo,foo' means "disable foo", 'foo,-foo' means "enable foo"
	// * has the lowest priority.
	//     e.g. *,-foo, means "disable 'foo'"
	ControllerGates         []string
	MaxConcurrentReconciles int
	LeaderElection          bool
}

func NewControllerManagerServerOptions() *ControllerManagerServerOptions {
	return &ControllerManagerServerOptions{
		WorkDir:                 "/var/lib/kubekey",
		ControllerGates:         []string{"*"},
		MaxConcurrentReconciles: 1,
	}
}

func (o *ControllerManagerServerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.BoolVar(&o.GOPSEnabled, "gops", o.GOPSEnabled, "Whether to enable gops or not.  When enabled this option, "+
		"controller-manager will listen on a random port on 127.0.0.1, then you can use the gops tool to list and diagnose the controller-manager currently running.")
	gfs.StringVar(&o.WorkDir, "work-dir", o.WorkDir, "the base Dir for kubekey. Default current dir. ")
	gfs.BoolVar(&o.Debug, "debug", o.Debug, "Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.")

	kfs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		kfs.AddGoFlag(fl)
	})

	cfs := fss.FlagSet("controller-manager")
	cfs.StringSliceVar(&o.ControllerGates, "controllers", o.ControllerGates, "The list of controller gates to enable or disable controller. "+
		"'*' means \"all enabled by default controllers\"")
	cfs.IntVar(&o.MaxConcurrentReconciles, "max-concurrent-reconciles", o.MaxConcurrentReconciles, "The number of maximum concurrent reconciles for controller.")
	cfs.BoolVar(&o.LeaderElection, "leader-election", o.LeaderElection, "Whether to enable leader election for controller-manager.")
	return fss
}

func (o *ControllerManagerServerOptions) Complete(cmd *cobra.Command, args []string) {
	// do nothing
}
