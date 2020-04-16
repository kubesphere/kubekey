package runner

import (
	"fmt"
	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
	"github.com/pixiake/kubekey/util"
	ssh2 "github.com/pixiake/kubekey/util/ssh"
	"github.com/pkg/errors"
	"strings"
	"text/template"
	"time"
)

// Runner bundles a connection to a host with the verbosity and
// other options for running commands via SSH.
type Runner struct {
	Conn    ssh2.Connection
	Prefix  string
	OS      string
	Verbose bool
	Host    *kubekeyapi.HostCfg
	Result  chan string
}

// TemplateVariables is a render context for templates
type TemplateVariables map[string]interface{}

func (r *Runner) RunRaw(cmd string) (string, error) {
	if r.Conn == nil {
		return "", errors.New("runner is not tied to an opened SSH connection")
	}
	output, _, err := r.Conn.Exec(cmd, r.Host)
	if !r.Verbose {
		if err != nil {
			return "", err
		}
		return output, nil
	}

	if output != "" {
		fmt.Printf("[%s %s] MSG:\n", r.Host.HostName, r.Host.SSHAddress)
		fmt.Println(output)
	}

	return "", err
}

func (r *Runner) ScpFile(src, dst string) error {
	if r.Conn == nil {
		return errors.New("runner is not tied to an opened SSH connection")
	}

	err := r.Conn.Scp(src, dst)
	if err != nil {
		if r.Verbose {
			fmt.Printf("push %s to %s:%s   Failed\n", src, r.Host.SSHAddress, dst)
			return err
		}
	} else {
		if r.Verbose {
			fmt.Printf("push %s to %s:%s   Done\n", src, r.Host.SSHAddress, dst)
		}
	}
	return nil
}

// Run executes a given command/script, optionally printing its output to
// stdout/stderr.
func (r *Runner) Run(cmd string, variables TemplateVariables) (string, error) {
	tmpl, _ := template.New("base").Parse(cmd)
	cmd, err := util.Render(tmpl, variables)
	if err != nil {
		return "", err
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
		stdout, _ := r.Run(cmd, nil)
		if validator(stdout) {
			return true
		}

		time.Sleep(1 * time.Second)
	}

	return false
}
