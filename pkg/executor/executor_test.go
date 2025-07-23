package executor

import (
	"context"
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

// newTestOption creates a new test option struct for the given hosts.
// It sets up a fake client and playbook using the provided hosts, and initializes the variable subsystem
// with a memory-backed source. This is intended for use in unit tests.
func newTestOption(hosts []string) (*option, error) {
	var err error

	// Convert the slice of hostnames to an InventoryHost map (not strictly needed here, but kept for clarity).
	inventoryHost := make(kkcorev1.InventoryHost)
	for _, h := range hosts {
		inventoryHost[h] = runtime.RawExtension{}
	}

	// Create a fake client and playbook for testing.
	client, playbook, err := _const.NewTestPlaybook(hosts)
	if err != nil {
		return nil, err
	}

	// Initialize the option struct with the fake client, playbook, and standard output for logging.
	o := &option{
		client:    client,
		playbook:  playbook,
		logOutput: os.Stdout,
	}

	// Initialize the variable subsystem using the memory-backed source.
	o.variable, err = variable.New(context.TODO(), o.client, *o.playbook, source.MemorySource)
	if err != nil {
		return nil, err
	}

	return o, nil
}
