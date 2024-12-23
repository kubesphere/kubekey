//go:build builtin
// +build builtin

package core

import (
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/controllers"
)

func init() {
	utilruntime.Must(controllers.Register(&PipelineReconciler{}))
}
