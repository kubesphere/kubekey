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

package task

import (
	"context"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	"github.com/kubesphere/kubekey/v4/pkg/cache"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// Controller is the interface for running tasks
type Controller interface {
	// Start the controller
	Start(ctx context.Context) error
	// AddTasks adds tasks to the controller
	AddTasks(ctx context.Context, o AddTaskOptions) error
}

type AddTaskOptions struct {
	*kubekeyv1.Pipeline
	// set by AddTask function
	variable variable.Variable
}

type ControllerOptions struct {
	MaxConcurrent int
	ctrlclient.Client
	TaskReconciler reconcile.Reconciler
}

func NewController(o ControllerOptions) (Controller, error) {
	if o.MaxConcurrent == 0 {
		o.MaxConcurrent = 1
	}
	if o.Client == nil {
		o.Client = cache.NewDelegatingClient(nil)
	}

	return &taskController{
		MaxConcurrent:  o.MaxConcurrent,
		wq:             workqueue.NewRateLimitingQueue(&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)}),
		client:         o.Client,
		taskReconciler: o.TaskReconciler,
	}, nil
}
