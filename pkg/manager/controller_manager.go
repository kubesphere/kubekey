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
	"errors"
	"fmt"

	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/controllers"
	_ "github.com/kubesphere/kubekey/v4/pkg/controllers/core"
	_ "github.com/kubesphere/kubekey/v4/pkg/controllers/infrastructure"
)

type controllerManager struct {
	MaxConcurrentReconciles int
	LeaderElection          bool
	LeaderElectionID        string
}

// Run controllerManager, run controller in kubernetes
func (m controllerManager) Run(ctx context.Context) error {
	ctrl.SetLogger(klog.NewKlogr())
	restconfig, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("cannot get restconfig in kubernetes. error is %w", err)
	}

	mgr, err := ctrl.NewManager(restconfig, ctrl.Options{
		Scheme:                 _const.Scheme,
		LeaderElection:         m.LeaderElection,
		LeaderElectionID:       m.LeaderElectionID,
		HealthProbeBindAddress: ":9440",
	})
	if err != nil {
		return fmt.Errorf("failed to create controller manager. error: %w", err)
	}
	if err := mgr.AddHealthzCheck("default", healthz.Ping); err != nil {
		return fmt.Errorf("failed to add default healthcheck. error: %w", err)
	}
	if err := mgr.AddReadyzCheck("default", healthz.Ping); err != nil {
		return fmt.Errorf("failed to add default readycheck. error: %w", err)
	}

	if err := m.register(mgr); err != nil {
		return err
	}

	return mgr.Start(ctx)
}

func (m controllerManager) register(mgr ctrl.Manager) error {
	o := &ctrlcontroller.Options{
		MaxConcurrentReconciles: m.MaxConcurrentReconciles,
	}
	if len(controllers.Controllers) == 0 {
		return errors.New("register controllers is empty")
	}
	for _, c := range controllers.Controllers {
		if err := c.SetupWithManager(mgr, *o); err != nil {
			return fmt.Errorf("failed to register controller %q. error: %w", c.Name(), err)
		}
	}

	return nil
}
