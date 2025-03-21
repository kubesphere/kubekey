package executor

import (
	"context"
	"io"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// Executor all task in playbook
type Executor interface {
	Exec(ctx context.Context) error
}

// option for playbookExecutor, blockExecutor, taskExecutor
type option struct {
	client ctrlclient.Client

	playbook *kkcorev1.Playbook
	variable variable.Variable
	// commandLine log output. default os.stdout
	logOutput io.Writer
}
