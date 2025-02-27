package executor

import (
	"context"
	"io"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// Executor all task in pipeline
type Executor interface {
	Exec(ctx context.Context) error
}

// option for pipelineExecutor, blockExecutor, taskExecutor
type option struct {
	client ctrlclient.Client

	pipeline *kkcorev1.Pipeline
	variable variable.Variable
	// commandLine log output. default os.stdout
	logOutput io.Writer
}
