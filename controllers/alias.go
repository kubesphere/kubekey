/*
 Copyright 2022 The KubeSphere Authors.

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

package controllers

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/controllers/remote"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	kkclustercontroller "github.com/kubesphere/kubekey/controllers/kkcluster"
	kkinstancecontroller "github.com/kubesphere/kubekey/controllers/kkinstance"
	kkmachinecontroller "github.com/kubesphere/kubekey/controllers/kkmachine"
)

// KKClusterReconciler reconciles a KKCluster object
type KKClusterReconciler struct {
	client.Client
	Recorder         record.EventRecorder
	Scheme           *runtime.Scheme
	WatchFilterValue string
	DataDir          string
}

// SetupWithManager sets up the controller with the Manager.
func (r *KKClusterReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return (&kkclustercontroller.Reconciler{
		Client:           r.Client,
		Recorder:         r.Recorder,
		Scheme:           r.Scheme,
		WatchFilterValue: r.WatchFilterValue,
		DataDir:          r.DataDir,
	}).SetupWithManager(ctx, mgr, options)
}

// KKMachineReconciler reconciles a KKMachine object
type KKMachineReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	Recorder         record.EventRecorder
	Tracker          *remote.ClusterCacheTracker
	WatchFilterValue string
	DataDir          string
}

// SetupWithManager sets up the controller with the Manager.
func (r *KKMachineReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return (&kkmachinecontroller.Reconciler{
		Client:           r.Client,
		Recorder:         r.Recorder,
		Scheme:           r.Scheme,
		Tracker:          r.Tracker,
		WatchFilterValue: r.WatchFilterValue,
		DataDir:          r.DataDir,
	}).SetupWithManager(ctx, mgr, options)
}

// KKInstanceReconciler reconciles a KKInstance object
type KKInstanceReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	Tracker          *remote.ClusterCacheTracker
	Recorder         record.EventRecorder
	WatchFilterValue string
	DataDir          string

	WaitKKInstanceInterval time.Duration
	WaitKKInstanceTimeout  time.Duration
}

// SetupWithManager sets up the controller with the Manager.
func (r *KKInstanceReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return (&kkinstancecontroller.Reconciler{
		Client:                 r.Client,
		Recorder:               r.Recorder,
		Tracker:                r.Tracker,
		Scheme:                 r.Scheme,
		WatchFilterValue:       r.WatchFilterValue,
		DataDir:                r.DataDir,
		WaitKKInstanceInterval: r.WaitKKInstanceInterval,
		WaitKKInstanceTimeout:  r.WaitKKInstanceTimeout,
	}).SetupWithManager(ctx, mgr, options)
}
