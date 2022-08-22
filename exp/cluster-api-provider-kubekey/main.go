/*
Copyright 2022.

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

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/cluster-api/util/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
	utilruntime.Must(infrav1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

var (
	metricsAddr             string
	enableLeaderElection    bool
	leaderElectionNamespace string
	probeAddr               string
	watchFilterValue        string
	kkClusterConcurrency    int
	kkInstanceConcurrency   int
	kkMachineConcurrency    int
	syncPeriod              time.Duration
	watchNamespace          string
	dataDir                 string
)

func main() {
	klog.InitFlags(nil)

	rand.Seed(time.Now().UnixNano())
	initFlags(pflag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	ctrl.SetLogger(klogr.New())

	ctx := ctrl.SetupSignalHandler()

	restConfig := ctrl.GetConfigOrDie()
	restConfig.UserAgent = "cluster-api-provider-kk-controller"
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:                     scheme,
		MetricsBindAddress:         metricsAddr,
		LeaderElection:             enableLeaderElection,
		LeaderElectionID:           "controller-leader-election-capkk",
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
		LeaderElectionNamespace:    leaderElectionNamespace,
		SyncPeriod:                 &syncPeriod,
		Namespace:                  watchNamespace,
		Port:                       9443,
		HealthProbeBindAddress:     probeAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	log := ctrl.Log.WithName("remote").WithName("ClusterCacheTracker")
	tracker, err := remote.NewClusterCacheTracker(
		mgr,
		remote.ClusterCacheTrackerOptions{
			Log:     &log,
			Indexes: remote.DefaultIndexes,
		},
	)
	if err != nil {
		setupLog.Error(err, "unable to create cluster cache tracker")
		os.Exit(1)
	}

	// Initialize event recorder.
	record.InitFromRecorder(mgr.GetEventRecorderFor("kk-controller"))

	if err = (&controllers.KKClusterReconciler{
		Client:           mgr.GetClient(),
		Recorder:         mgr.GetEventRecorderFor("kkcluster-controller"),
		Scheme:           mgr.GetScheme(),
		WatchFilterValue: watchFilterValue,
		DataDir:          dataDir,
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: kkClusterConcurrency, RecoverPanic: true}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KKCluster")
		os.Exit(1)
	}
	if err = (&controllers.KKMachineReconciler{
		Client:           mgr.GetClient(),
		Recorder:         mgr.GetEventRecorderFor("kkmachine-controller"),
		Scheme:           mgr.GetScheme(),
		Tracker:          tracker,
		WatchFilterValue: watchFilterValue,
		DataDir:          dataDir,
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: kkMachineConcurrency, RecoverPanic: true}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KKMachine")
		os.Exit(1)
	}
	if err = (&controllers.KKInstanceReconciler{
		Client:           mgr.GetClient(),
		Recorder:         mgr.GetEventRecorderFor("kkinstance-controller"),
		Scheme:           mgr.GetScheme(),
		WatchFilterValue: watchFilterValue,
		DataDir:          dataDir,
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: kkInstanceConcurrency, RecoverPanic: true}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KKInstance")
		os.Exit(1)
	}

	if err = (&infrav1.KKCluster{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KKCluster")
		os.Exit(1)
	}
	if err = (&infrav1.KKClusterTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KKClusterTemplate")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("webhook", mgr.GetWebhookServer().StartedChecker()); err != nil {
		setupLog.Error(err, "unable to create health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("webhook", mgr.GetWebhookServer().StartedChecker()); err != nil {
		setupLog.Error(err, "unable to create ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func initFlags(fs *pflag.FlagSet) {
	fs.StringVar(
		&metricsAddr,
		"metrics-bind-address",
		":8080",
		"The address the metric endpoint binds to.",
	)

	fs.BoolVar(
		&enableLeaderElection,
		"leader-elect",
		false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.",
	)

	fs.StringVar(
		&watchNamespace,
		"namespace",
		"",
		"Namespace that the controller watches to reconcile cluster-api objects. If unspecified, the controller watches for cluster-api objects across all namespaces.",
	)

	fs.StringVar(
		&leaderElectionNamespace,
		"leader-elect-namespace",
		"",
		"Namespace that the controller performs leader election in. If unspecified, the controller will discover which namespace it is running in.",
	)

	fs.DurationVar(&syncPeriod,
		"sync-period",
		10*time.Minute,
		"The minimum interval at which watched resources are reconciled.",
	)

	fs.IntVar(&kkClusterConcurrency,
		"kkcluster-concurrency",
		5,
		"Number of KKClusters to process simultaneously.",
	)

	fs.IntVar(&kkInstanceConcurrency,
		"kkinstance-concurrency",
		10,
		"Number of KKInstance to process simultaneously.",
	)

	fs.IntVar(&kkMachineConcurrency,
		"kkmachine-concurrency",
		10,
		"Number of KKMachines to process simultaneously.",
	)

	fs.StringVar(
		&probeAddr,
		"health-probe-bind-address",
		":8081",
		"The address the probe endpoint binds to.",
	)

	fs.StringVar(
		&watchFilterValue,
		"watch-filter",
		"",
		fmt.Sprintf("Label value that the controller watches to reconcile cluster-api objects. Label key is always %s. If unspecified, the controller watches for all cluster-api objects.", clusterv1.WatchLabel),
	)

	fs.StringVar(
		&dataDir,
		"data-dir",
		"",
		"The KubeKey data dir.",
	)
}
