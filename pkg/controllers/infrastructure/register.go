//go:build clusterapi
// +build clusterapi

package infrastructure

import (
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/controllers"
)

func init() {
	utilruntime.Must(controllers.Register(&KKClusterReconciler{}))
	utilruntime.Must(controllers.Register(&InventoryReconciler{}))
	utilruntime.Must(controllers.Register(&KKMachineReconciler{}))
}
