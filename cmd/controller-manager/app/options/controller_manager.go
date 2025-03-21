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
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	cliflag "k8s.io/component-base/cli/flag"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ControllerManagerServerOptions for NewControllerManagerServerOptions
type ControllerManagerServerOptions struct {
	// Debug mode, after a successful execution of Playbook, will retain runtime data, which includes task execution status and parameters.
	Debug bool

	MaxConcurrentReconciles int
	LeaderElection          bool
	LeaderElectionID        string
	// LeaderElectionResourceLock determines which resource lock to use for leader election,
	// defaults to "leases". Change this value only if you know what you are doing.
	//
	// If you are using `configmaps`/`endpoints` resource lock and want to migrate to "leases",
	// you might do so by migrating to the respective multilock first ("configmapsleases" or "endpointsleases"),
	// which will acquire a leader lock on both resources.
	// After all your users have migrated to the multilock, you can go ahead and migrate to "leases".
	// Please also keep in mind, that users might skip versions of your controller.
	//
	// Note: before controller-runtime version v0.7, it was set to "configmaps".
	// And from v0.7 to v0.11, the default was "configmapsleases", which was
	// used to migrate from configmaps to leases.
	// Since the default was "configmapsleases" for over a year, spanning five minor releases,
	// any actively maintained operators are very likely to have a released version that uses
	// "configmapsleases". Therefore defaulting to "leases" should be safe since v0.12.
	//
	// So, what do you have to do when you are updating your controller-runtime dependency
	// from a lower version to v0.12 or newer?
	// - If your operator matches at least one of these conditions:
	//   - the LeaderElectionResourceLock in your operator has already been explicitly set to "leases"
	//   - the old controller-runtime version is between v0.7.0 and v0.11.x and the
	//     LeaderElectionResourceLock wasn't set or was set to "leases"/"configmapsleases"/"endpointsleases"
	//   feel free to update controller-runtime to v0.12 or newer.
	// - Otherwise, you may have to take these steps:
	//   1. update controller-runtime to v0.12 or newer in your go.mod
	//   2. set LeaderElectionResourceLock to "configmapsleases" (or "endpointsleases")
	//   3. package your operator and upgrade it in all your clusters
	//   4. only if you have finished 3, you can remove the LeaderElectionResourceLock to use the default "leases"
	// Otherwise, your operator might end up with multiple running instances that
	// each acquired leadership through different resource locks during upgrades and thus
	// act on the same resources concurrently.
	LeaderElectionResourceLock string
	// ControllerGates is the list of controller gates to enable or disable controller.
	// '*' means "all enabled by default controllers"
	// 'foo' means "enable 'foo'"
	// '-foo' means "disable 'foo'"
	// first item for a particular name wins.
	//     e.g. '-foo,foo' means "disable foo", 'foo,-foo' means "enable foo"
	// * has the lowest priority.
	//     e.g. *,-foo, means "disable 'foo'"
	ControllerGates []string
	Controllers     []Controller
}

// Keys returns a sorted list of controller names
func (o *ControllerManagerServerOptions) controllerKeys() string {
	var keys = make([]string, 0)
	for _, c := range o.Controllers {
		keys = append(keys, c.Name())
	}
	sort.Strings(keys)

	return strings.Join(keys, ",")
}

// NewControllerManagerServerOptions for NewControllerManagerCommand
func NewControllerManagerServerOptions() *ControllerManagerServerOptions {
	return &ControllerManagerServerOptions{
		MaxConcurrentReconciles: 1,
		Controllers:             controllers,
	}
}

// Flags add to NewControllerManagerCommand
func (o *ControllerManagerServerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	gfs := fss.FlagSet("generic")
	gfs.BoolVar(&o.Debug, "debug", o.Debug, "Debug mode, after a successful execution of Playbook, "+"will retain runtime data, which includes task execution status and parameters.")
	cfs := fss.FlagSet("controller-manager")
	cfs.IntVar(&o.MaxConcurrentReconciles, "max-concurrent-reconciles", o.MaxConcurrentReconciles, "The number of maximum concurrent reconciles for controller.")
	cfs.BoolVar(&o.LeaderElection, "leader-election", o.LeaderElection, "Whether to enable leader election for controller-manager.")
	cfs.StringVar(&o.LeaderElectionID, "leader-election-id", o.LeaderElectionID, "Whether to enable leader election for controller-manager.")
	cfs.StringVar(&o.LeaderElectionResourceLock, "leader-election-lock", o.LeaderElectionResourceLock, " which resource lock to use for leader election.")
	cfs.StringSliceVar(&o.ControllerGates, "controllers", []string{"*"}, fmt.Sprintf(""+
		"A list of controllers to enable. '*' enables all on-by-default controllers, 'foo' enables the controller "+
		"named 'foo', '-foo' disables the controller named 'foo'.\nAll controllers: %s",
		o.controllerKeys()))

	return fss
}

// Complete for ControllerManagerServerOptions
func (o *ControllerManagerServerOptions) Complete() {
	// do nothing
	if o.MaxConcurrentReconciles == 0 {
		o.MaxConcurrentReconciles = 1
	}
}

var controllers []Controller

// Register controller
func Register(reconciler Controller) error {
	for _, c := range controllers {
		if c.Name() == reconciler.Name() {
			return errors.Errorf("%s has register", reconciler.Name())
		}
	}
	controllers = append(controllers, reconciler)

	return nil
}

// Controller should add in ctrl.manager
type Controller interface {
	// setup reconcile with manager
	SetupWithManager(mgr ctrl.Manager, o ControllerManagerServerOptions) error
	// the Name of controller
	Name() string
}
