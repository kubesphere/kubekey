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

package manager

import (
	"context"

	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/controllers"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

type controllerManager struct {
	ControllerGates         []string
	MaxConcurrentReconciles int
	LeaderElection          bool
}

func (c controllerManager) Run(ctx context.Context) error {
	ctrl.SetLogger(klog.NewKlogr())

	restconfig := config.GetConfigOrDie()
	proxyTransport, err := proxy.NewProxyTransport(false)
	if err != nil {
		klog.ErrorS(err, "Create proxy transport error")
		return err
	}
	restconfig.Transport = proxyTransport
	mgr, err := ctrl.NewManager(restconfig, ctrl.Options{
		Scheme:           _const.Scheme,
		LeaderElection:   c.LeaderElection,
		LeaderElectionID: "controller-leader-election-kk",
	})
	if err != nil {
		klog.ErrorS(err, "Create manager error")
		return err
	}

	if err := (&controllers.PipelineReconciler{
		Client:        mgr.GetClient(),
		EventRecorder: mgr.GetEventRecorderFor("pipeline"),
		Scheme:        mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controllers.Options{
		ControllerGates: c.ControllerGates,
		Options: ctrlcontroller.Options{
			MaxConcurrentReconciles: c.MaxConcurrentReconciles,
		},
	}); err != nil {
		klog.ErrorS(err, "create pipeline controller error")
		return err
	}

	return mgr.Start(ctx)
}
