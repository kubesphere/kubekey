package apiserver

import (
	"context"
	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
	"golang.org/x/sync/singleflight"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type KKClusterController struct {
	ctrlclient.Client
	record.EventRecorder
	Workdir string
	group   singleflight.Group
}

func (k *KKClusterController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	var needRequeue bool = true
	var requeueAfter time.Duration = 0
	_, err, shared := k.group.Do(request.NamespacedName.String(), func() (interface{}, error) {
		needRequeue = false
		klog.Infof("do reconcile %s", request.NamespacedName.String())
		var item capkkinfrav1beta1.KKCluster
		err := k.Get(ctx, request.NamespacedName, &item)
		if err != nil {
			klog.Infof("get kk cluster %s failed with error: %s", request.NamespacedName.String(), err.Error())
			return reconcile.Result{}, err
		}
		return nil, k.handleKKCluster(ctx, item)
	})

	if needRequeue && shared {
		requeueAfter = time.Second
	}
	return reconcile.Result{
		RequeueAfter: requeueAfter,
	}, err

}

func (k *KKClusterController) handleKKCluster(ctx context.Context, item capkkinfrav1beta1.KKCluster) error {

	updateStatus := false
	var err error
	switch item.Status.Status {
	case "":
		item.Status.Status = capkkinfrav1beta1.StatusNotInstall
		item.Status.Conditions = make([]metav1.Condition, 0)
		updateStatus = true
	case capkkinfrav1beta1.StatusNotInstall:
		// 需等待手动执行安装
	case capkkinfrav1beta1.StatusInitializing:
		// 安装中，
	case capkkinfrav1beta1.StatusRunning:
	case capkkinfrav1beta1.StatusUpgrading:
	case capkkinfrav1beta1.StatusScaling:
	case capkkinfrav1beta1.StatusNodeError:
	case capkkinfrav1beta1.StatusUpgradeError:
	case capkkinfrav1beta1.StatusFailed:

	}

	if updateStatus {
		err = k.Client.Status().Update(ctx, &item)
	}

	return err
}

func (k *KKClusterController) SetupWithManager(mgr ctrl.Manager, o options.ControllerManagerServerOptions) error {
	k.Client = mgr.GetClient()
	k.EventRecorder = mgr.GetEventRecorderFor(k.Name())

	return ctrl.NewControllerManagedBy(mgr).
		For(&capkkinfrav1beta1.KKCluster{}).
		Complete(k)
}

func (k *KKClusterController) Name() string {
	return "kkclusters-apiserver"
}

var _ options.Controller = &KKClusterController{}
var _ reconcile.Reconciler = &KKClusterController{}
