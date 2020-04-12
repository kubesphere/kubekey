package runner

import (
	"fmt"
	"github.com/pixiake/kubekey/util/dialer/ssh"
	"github.com/pkg/errors"
	"os"
	"strings"
	"time"
)

// Runner bundles a connection to a host with the verbosity and
// other options for running commands via SSH.
type Runner struct {
	Conn    ssh.Connection
	Prefix  string
	OS      string
	Verbose bool
}

// TemplateVariables is a render context for templates
type TemplateVariables map[string]interface{}

func (r *Runner) RunRaw(cmd string) (string, string, error) {
	if r.Conn == nil {
		return "", "", errors.New("runner is not tied to an opened SSH connection")
	}

	if !r.Verbose {
		stdout, stderr, _, err := r.Conn.Exec(cmd)
		if err != nil {
			err = errors.Wrap(err, stderr)
		}

		return stdout, stderr, err
	}

	stdout := NewTee(New(os.Stdout, r.Prefix))
	defer stdout.Close()

	stderr := NewTee(New(os.Stderr, r.Prefix))
	defer stderr.Close()

	// run the command
	_, err := r.Conn.Stream(cmd, stdout, stderr)

	return stdout.String(), stderr.String(), err
}

// Run executes a given command/script, optionally printing its output to
// stdout/stderr.
func (r *Runner) Run(cmd string, variables TemplateVariables) (string, string, error) {
	cmd, err := Render(cmd, variables)
	if err != nil {
		return "", "", err
	}

	return r.RunRaw(cmd)
}

// WaitForPod waits for the availability of the given Kubernetes element.
func (r *Runner) WaitForPod(namespace string, name string, timeout time.Duration) error {
	cmd := fmt.Sprintf(`sudo kubectl --kubeconfig=/etc/kubernetes/admin.conf -n "%s" get pod "%s" -o jsonpath='{.status.phase}' --ignore-not-found`, namespace, name)
	if !r.WaitForCondition(cmd, timeout, IsRunning) {
		return errors.Errorf("timed out while waiting for %s/%s to come up for %v", namespace, name, timeout)
	}

	return nil
}

type validatorFunc func(stdout string) bool

// IsRunning checks if the given output represents the "Running" status of a Kubernetes pod.
func IsRunning(stdout string) bool {
	return strings.ToLower(stdout) == "running"
}

// WaitForCondition waits for something to be true.
func (r *Runner) WaitForCondition(cmd string, timeout time.Duration, validator validatorFunc) bool {
	cutoff := time.Now().Add(timeout)

	for time.Now().Before(cutoff) {
		stdout, _, _ := r.Run(cmd, nil)
		if validator(stdout) {
			return true
		}

		time.Sleep(1 * time.Second)
	}

	return false
}
