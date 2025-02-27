//go:build builtin
// +build builtin

package core

import (
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
)

func init() {
	utilruntime.Must(options.Register(&PipelineReconciler{}))
	utilruntime.Must(options.Register(&PipelineWebhook{}))
}
