package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/executor"
)

var playbookManager = &manager{
	manager: make(map[string]context.CancelFunc),
}

// playbookManager is responsible for managing playbook execution contexts and their cancellation.
// It uses a mutex to ensure thread-safe access to the manager map.
type manager struct {
	sync.Mutex
	manager map[string]context.CancelFunc // Map of playbook key to its cancel function
}

func (m *manager) executor(playbook *kkcorev1.Playbook, client ctrlclient.Client, promise string) error {
	f := func() error {
		// Build the log file path for the playbook execution
		filename := filepath.Join(
			_const.GetWorkdirFromConfig(playbook.Spec.Config),
			_const.RuntimeDir,
			kkcorev1.SchemeGroupVersion.Group,
			kkcorev1.SchemeGroupVersion.Version,
			"playbooks",
			playbook.Namespace,
			playbook.Name,
			playbook.Name+".log",
		)
		// Ensure the directory for the log file exists
		if _, err := os.Stat(filepath.Dir(filename)); err != nil {
			if !os.IsNotExist(err) {
				return errors.Wrapf(err, "failed to stat playbook dir %q", filepath.Dir(filename))
			}
			// If directory does not exist, create it
			if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
				return errors.Wrapf(err, "failed to create playbook dir %q", filepath.Dir(filename))
			}
		}
		// Open the log file for writing
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return errors.Wrapf(err, "failed to open log file", "file", filename)
		}
		defer file.Close()

		// Create a cancellable context for playbook execution
		ctx, cancel := context.WithCancel(context.Background())
		// Register the playbook and its cancel function in the playbookManager
		m.addPlaybook(playbook, cancel)
		// Execute the playbook and write output to the log file
		if err := executor.NewPlaybookExecutor(ctx, client, playbook, file).Exec(ctx); err != nil {
			// recode to log file
			fmt.Fprintf(file, "%s [Playbook %s] ERROR: %v\n", time.Now().Format(time.TimeOnly+" MST"), ctrlclient.ObjectKeyFromObject(playbook), err)
		}
		// Remove the playbook from the playbookManager after execution
		m.deletePlaybook(playbook)
		return nil
	}
	if promise == "true" {
		go func() {
			if err := f(); err != nil {
				klog.ErrorS(err, "failed to execute playbook", "playbook", ctrlclient.ObjectKeyFromObject(playbook))
			}
		}()
		return nil
	}
	return f()
}

// addPlaybook adds a playbook and its cancel function to the manager map.
func (m *manager) addPlaybook(playbook *kkcorev1.Playbook, cancel context.CancelFunc) {
	m.Lock()
	defer m.Unlock()

	m.manager[ctrlclient.ObjectKeyFromObject(playbook).String()] = cancel
}

// deletePlaybook removes a playbook from the manager map.
func (m *manager) deletePlaybook(playbook *kkcorev1.Playbook) {
	m.Lock()
	defer m.Unlock()

	delete(m.manager, ctrlclient.ObjectKeyFromObject(playbook).String())
}

// stopPlaybook cancels the context for a running playbook, if it exists.
func (m *manager) stopPlaybook(playbook *kkcorev1.Playbook) {
	// Attempt to cancel the playbook's context if it exists in the manager
	if cancel, ok := m.manager[ctrlclient.ObjectKeyFromObject(playbook).String()]; ok {
		cancel()
	}
}
