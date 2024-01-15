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

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	ctrlmanager "sigs.k8s.io/controller-runtime/pkg/manager"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	"github.com/kubesphere/kubekey/v4/pkg/cache"
	"github.com/kubesphere/kubekey/v4/pkg/controllers"
	"github.com/kubesphere/kubekey/v4/pkg/task"
)

type controllerManager struct {
	ControllerGates         []string
	MaxConcurrentReconciles int
	LeaderElection          bool
}

func (c controllerManager) Run(ctx context.Context) error {
	ctrl.SetLogger(klog.NewKlogr())
	scheme := runtime.NewScheme()
	// add default scheme
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		klog.Errorf("add default scheme error: %v", err)
		return err
	}
	// add kubekey scheme
	if err := kubekeyv1.AddToScheme(scheme); err != nil {
		klog.Errorf("add kubekey scheme error: %v", err)
		return err
	}
	mgr, err := ctrl.NewManager(config.GetConfigOrDie(), ctrlmanager.Options{
		Scheme:           scheme,
		LeaderElection:   c.LeaderElection,
		LeaderElectionID: "controller-leader-election-kk",
	})
	if err != nil {
		klog.Errorf("create manager error: %v", err)
		return err
	}

	taskController, err := task.NewController(task.ControllerOptions{
		MaxConcurrent: c.MaxConcurrentReconciles,
		Client:        cache.NewDelegatingClient(mgr.GetClient()),
		TaskReconciler: &controllers.TaskReconciler{
			Client:        cache.NewDelegatingClient(mgr.GetClient()),
			VariableCache: cache.LocalVariable,
		},
	})
	if err != nil {
		klog.Errorf("create task controller error: %v", err)
		return err
	}
	if err := mgr.Add(taskController); err != nil {
		klog.Errorf("add task controller error: %v", err)
		return err
	}

	if err := (&controllers.PipelineReconciler{
		Client:         cache.NewDelegatingClient(mgr.GetClient()),
		EventRecorder:  mgr.GetEventRecorderFor("pipeline"),
		TaskController: taskController,
	}).SetupWithManager(ctx, mgr, controllers.Options{
		ControllerGates: c.ControllerGates,
		Options: ctrlcontroller.Options{
			MaxConcurrentReconciles: c.MaxConcurrentReconciles,
		},
	}); err != nil {
		klog.Errorf("create pipeline controller error: %v", err)
		return err
	}

	return mgr.Start(ctx)
}
