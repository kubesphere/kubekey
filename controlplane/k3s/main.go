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

// Package main
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/pflag"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	infrabootstrapv1 "github.com/kubesphere/kubekey/v3/bootstrap/k3s/api/v1beta1"
	infracontrolplanev1 "github.com/kubesphere/kubekey/v3/controlplane/k3s/api/v1beta1"
	"github.com/kubesphere/kubekey/v3/controlplane/k3s/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
	utilruntime.Must(infrabootstrapv1.AddToScheme(scheme))
	utilruntime.Must(infracontrolplanev1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

var (
	metricsBindAddr             string
	enableLeaderElection        bool
	leaderElectionLeaseDuration time.Duration
	leaderElectionRenewDeadline time.Duration
	leaderElectionRetryPeriod   time.Duration
	watchFilterValue            string
	watchNamespace              string
	k3sControlPlaneConcurrency  int
	syncPeriod                  time.Duration
	webhookPort                 int
	webhookCertDir              string
	healthAddr                  string
)

// InitFlags initializes the flags.
func InitFlags(fs *pflag.FlagSet) {
	fs.StringVar(&metricsBindAddr, "metrics-bind-addr", "localhost:8080",
		"The address the metric endpoint binds to.")

	fs.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")

	fs.DurationVar(&leaderElectionLeaseDuration, "leader-elect-lease-duration", 1*time.Minute,
		"Interval at which non-leader candidates will wait to force acquire leadership (duration string)")

	fs.DurationVar(&leaderElectionRenewDeadline, "leader-elect-renew-deadline", 40*time.Second,
		"Duration that the leading controller manager will retry refreshing leadership before giving up (duration string)")

	fs.DurationVar(&leaderElectionRetryPeriod, "leader-elect-retry-period", 5*time.Second,
		"Duration the LeaderElector clients should wait between tries of actions (duration string)")

	fs.StringVar(&watchNamespace, "namespace", "",
		"Namespace that the controller watches to reconcile cluster-api objects. If unspecified, the controller watches for cluster-api objects across all namespaces.")

	fs.IntVar(&k3sControlPlaneConcurrency, "k3scontrolplane-concurrency", 10,
		"Number of kubeadm control planes to process simultaneously")

	fs.DurationVar(&syncPeriod, "sync-period", 10*time.Minute,
		"The minimum interval at which watched resources are reconciled (e.g. 15m)")

	fs.StringVar(&watchFilterValue, "watch-filter", "",
		fmt.Sprintf("Label value that the controller watches to reconcile cluster-api objects. Label key is always %s. If unspecified, the controller watches for all cluster-api objects.", clusterv1.WatchLabel))

	fs.IntVar(&webhookPort, "webhook-port", 9443,
		"Webhook Server port")

	fs.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/k8s-webhook-server/serving-certs/",
		"Webhook cert dir, only used when webhook-port is specified.")

	fs.StringVar(&healthAddr, "health-addr", ":9440",
		"The address the health endpoint binds to.")
}

func main() {
	klog.InitFlags(nil)
	rand.Seed(time.Now().UnixNano())

	InitFlags(pflag.CommandLine)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	ctrl.SetLogger(klogr.New())

	ctx := ctrl.SetupSignalHandler()

	restConfig := ctrl.GetConfigOrDie()
	restConfig.UserAgent = remote.DefaultClusterAPIUserAgent("cluster-api-k3s-control-plane-manager")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsBindAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "k3s-control-plane-manager-leader-election-capkk",
		LeaseDuration:      &leaderElectionLeaseDuration,
		RenewDeadline:      &leaderElectionRenewDeadline,
		RetryPeriod:        &leaderElectionRetryPeriod,
		Namespace:          watchNamespace,
		SyncPeriod:         &syncPeriod,
		ClientDisableCacheFor: []client.Object{
			&corev1.ConfigMap{},
			&corev1.Secret{},
		},
		Port:                   webhookPort,
		HealthProbeBindAddress: healthAddr,
		CertDir:                webhookCertDir,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	log := ctrl.Log.WithName("remote").WithName("ClusterCacheTracker")
	tracker, err := remote.NewClusterCacheTracker(mgr, remote.ClusterCacheTrackerOptions{
		Log:     &log,
		Indexes: remote.DefaultIndexes,
		ClientUncachedObjects: []client.Object{
			&corev1.ConfigMap{},
			&corev1.Secret{},
			&corev1.Pod{},
			&appsv1.Deployment{},
			&appsv1.DaemonSet{},
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to create cluster cache tracker")
		os.Exit(1)
	}
	if err := (&remote.ClusterCacheReconciler{
		Client:           mgr.GetClient(),
		Tracker:          tracker,
		WatchFilterValue: watchFilterValue,
	}).SetupWithManager(ctx, mgr, concurrency(k3sControlPlaneConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterCacheReconciler")
		os.Exit(1)
	}

	if err = (&controllers.K3sControlPlaneReconciler{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		APIReader:        mgr.GetAPIReader(),
		Tracker:          tracker,
		WatchFilterValue: watchFilterValue,
	}).SetupWithManager(ctx, mgr, concurrency(k3sControlPlaneConcurrency)); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "K3sControlPlane")
		os.Exit(1)
	}
	if err = (&infracontrolplanev1.K3sControlPlane{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "K3sControlPlane")
		os.Exit(1)
	}
	if err = (&infracontrolplanev1.K3sControlPlaneTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "K3sControlPlaneTemplate")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func concurrency(c int) controller.Options {
	return controller.Options{MaxConcurrentReconciles: c}
}
