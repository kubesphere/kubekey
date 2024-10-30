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
	cliflag "k8s.io/component-base/cli/flag"
)

// ControllerManagerServerOptions for NewControllerManagerServerOptions
type ControllerManagerServerOptions struct {
	// WorkDir is the baseDir which command find any resource (project etc.)
	WorkDir string
	// Debug mode, after a successful execution of Pipeline, will retain runtime data, which includes task execution status and parameters.
	Debug bool

	MaxConcurrentReconciles int
	LeaderElection          bool
}

// NewControllerManagerServerOptions for NewControllerManagerCommand
func NewControllerManagerServerOptions() *ControllerManagerServerOptions {
	return &ControllerManagerServerOptions{
		WorkDir:                 "/kubekey",
		MaxConcurrentReconciles: 1,
	}
}

// Flags add to NewControllerManagerCommand
func (o *ControllerManagerServerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.StringVar(&o.WorkDir, "work-dir", o.WorkDir, "the base Dir for kubekey. Default current dir. ")
	gfs.BoolVar(&o.Debug, "debug", o.Debug, "Debug mode, after a successful execution of Pipeline, "+"will retain runtime data, which includes task execution status and parameters.")
	cfs := fss.FlagSet("controller-manager")
	cfs.IntVar(&o.MaxConcurrentReconciles, "max-concurrent-reconciles", o.MaxConcurrentReconciles, "The number of maximum concurrent reconciles for controller.")
	cfs.BoolVar(&o.LeaderElection, "leader-election", o.LeaderElection, "Whether to enable leader election for controller-manager.")

	return fss
}

// Complete for ControllerManagerServerOptions
func (o *ControllerManagerServerOptions) Complete() {
	// do nothing
	if o.MaxConcurrentReconciles == 0 {
		o.MaxConcurrentReconciles = 1
	}
}
