package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/executor"
	"golang.org/x/sync/singleflight"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type KKMachineController struct {
	ctrlclient.Client
	record.EventRecorder
	Workdir string
	group   singleflight.Group
}

func (k *KKMachineController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	var needRequeue bool = true
	var requeueAfter time.Duration = 0
	_, err, shared := k.group.Do(request.NamespacedName.String(), func() (interface{}, error) {
		needRequeue = false
		klog.Infof("do reconcile %s", request.NamespacedName.String())
		var kkmachine capkkinfrav1beta1.KKMachine
		err := k.Get(ctx, request.NamespacedName, &kkmachine)
		if err != nil {
			klog.Infof("get kkmachine %s failed with error: %s", request.NamespacedName.String(), err.Error())
			return reconcile.Result{}, err
		}
		return nil, k.handleKKMachine(ctx, kkmachine)
	})

	if needRequeue && shared {
		requeueAfter = time.Second
	}
	return reconcile.Result{
		RequeueAfter: requeueAfter,
	}, err
}

func (k *KKMachineController) handleKKMachine(ctx context.Context, machine capkkinfrav1beta1.KKMachine) error {
	updateStatus := false
	var err error
	switch machine.Status.Status {
	case "":
		machine.Status = capkkinfrav1beta1.KKMachineStatus{
			Ready:      false,
			Status:     capkkinfrav1beta1.KKMachineStatusCreating,
			Conditions: make(clusterv1beta1.Conditions, 0),
		}
		updateStatus = true
	case capkkinfrav1beta1.KKMachineStatusCreating:
		updateStatus, err = k.handleNodeCreate(ctx, &machine)
	case capkkinfrav1beta1.KKMachineStatusWarning:
	case capkkinfrav1beta1.KKMachineStatusReady:
	case capkkinfrav1beta1.KKMachineStatusRunning:
	case capkkinfrav1beta1.KKMachineStatusFault:
	case capkkinfrav1beta1.KKMachineStatusUnschedulable:

	}

	if err != nil {
		fmt.Println("err:", err)
		return err
	}

	if updateStatus {
		err = k.Client.Status().Update(ctx, &machine)
	}

	return err
}

func (k *KKMachineController) handleNodeCreate(ctx context.Context, machine *capkkinfrav1beta1.KKMachine) (bool, error) {

	hostCheckPlaybook := &kkcorev1.Playbook{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "host-check-",
			Namespace:    machine.GetNamespace(),
		},
		Spec: kkcorev1.PlaybookSpec{
			InventoryRef: &corev1.ObjectReference{
				Kind:      "Inventory",
				Namespace: machine.GetNamespace(),
				Name:      machine.GetNamespace(),
			},
			Playbook: "host_check.yaml",
			Config: kkcorev1.Config{
				Spec: machine.Spec.Config,
			},
		},
		Status: kkcorev1.PlaybookStatus{
			Phase: kkcorev1.PlaybookPhasePending,
		},
	}
	// Set the workdir in the playbook's config
	if err := unstructured.SetNestedField(hostCheckPlaybook.Spec.Config.Value(), k.Workdir, _const.Workdir); err != nil {
		return false, err
	}
	// Create the playbook resource in the cluster
	if err := k.Client.Create(ctx, hostCheckPlaybook); err != nil {
		return false, err
	}

	// Execute the playbook asynchronously if "promise" is true (default)
	bs, _ := json.Marshal(hostCheckPlaybook)
	fmt.Println("start exec playbook:\n", string(bs))
	if err := executor.PlaybookManager.Executor(hostCheckPlaybook, k.Client, "false"); err != nil {
		fmt.Println("exec playbook with error:", err.Error())
		machine.Status.Status = capkkinfrav1beta1.KKMachineStatusWarning
	} else {
		machine.Status.Status = capkkinfrav1beta1.KKMachineStatusReady
	}

	machine.Status.Conditions = append(machine.Status.Conditions, clusterv1beta1.Condition{
		LastTransitionTime: metav1.Now(),
		Type:               "PlayBook",
		Reason:             hostCheckPlaybook.Status.Result.String(),
		Message:            fmt.Sprintf("%s/%s", hostCheckPlaybook.GetNamespace(), hostCheckPlaybook.GetName()),
	})

	fmt.Println("playbook exec success")
	bs, _ = json.Marshal(hostCheckPlaybook)
	fmt.Println("after exec playbook data:\n", string(bs))

	return true, nil
}

func (k *KKMachineController) SetupWithManager(mgr ctrl.Manager, o options.ControllerManagerServerOptions) error {
	k.Client = mgr.GetClient()
	k.EventRecorder = mgr.GetEventRecorderFor(k.Name())

	return ctrl.NewControllerManagedBy(mgr).
		For(&capkkinfrav1beta1.KKMachine{}).
		Complete(k)
}

func (k *KKMachineController) Name() string {
	return "kkmachine-apiserver"
}

var _ options.Controller = &KKMachineController{}
var _ reconcile.Reconciler = &KKMachineController{}
