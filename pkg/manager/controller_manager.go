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
	"github.com/kubesphere/kubekey/v4/pkg/controllers"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
	"github.com/kubesphere/kubekey/v4/pkg/task"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
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
		klog.ErrorS(err, "Add default scheme error")
		return err
	}
	// add kubekey scheme,
	// exclude task resource,Because will manager in local
	if err := kubekeyv1.AddToScheme(scheme); err != nil {
		klog.ErrorS(err, "Add kk scheme error")
		return err
	}

	mgr, err := ctrl.NewManager(config.GetConfigOrDie(), ctrlmanager.Options{
		Scheme:           scheme,
		LeaderElection:   c.LeaderElection,
		LeaderElectionID: "controller-leader-election-kk",
	})
	if err != nil {
		klog.ErrorS(err, "Create manager error")
		return err
	}

	taskController, err := task.NewController(task.ControllerOptions{
		VariableCache: variable.Cache,
		MaxConcurrent: c.MaxConcurrentReconciles,
		Client:        proxy.NewDelegatingClient(mgr.GetClient()),
		TaskReconciler: &controllers.TaskReconciler{
			Client:        proxy.NewDelegatingClient(mgr.GetClient()),
			VariableCache: variable.Cache,
		},
	})
	if err != nil {
		klog.ErrorS(err, "Create task controller error")
		return err
	}

	// add task controller to manager
	if err := mgr.Add(taskController); err != nil {
		klog.ErrorS(err, "Add task controller error")
		return err
	}

	if err := (&controllers.PipelineReconciler{
		Client:         proxy.NewDelegatingClient(mgr.GetClient()),
		EventRecorder:  mgr.GetEventRecorderFor("pipeline"),
		TaskController: taskController,
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
