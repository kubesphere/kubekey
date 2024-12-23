//go:build clusterapi
// +build clusterapi

package infrastructure

import (
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
)

func init() {
	utilruntime.Must(options.Register(&KKClusterReconciler{}))
	utilruntime.Must(options.Register(&KKClusterWebhook{}))
	utilruntime.Must(options.Register(&InventoryReconciler{}))
	utilruntime.Must(options.Register(&KKMachineReconciler{}))
}
