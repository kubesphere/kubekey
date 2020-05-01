package runner

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	ssh2 "github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type Runner struct {
	Conn    ssh2.Connection
	Prefix  string
	OS      string
	Verbose bool
	Host    *kubekeyapi.HostCfg
	Index   int
}

type TemplateVariables map[string]interface{}

func (r *Runner) RunCmd(cmd string) (string, error) {
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

	if err != nil {
		return output, err
	}

	if output != "" {
		if strings.Contains(cmd, "base64") && strings.Contains(cmd, "--wrap=0") || strings.Contains(cmd, "make-ssl-etcd.sh") || strings.Contains(cmd, "docker-install.sh") || strings.Contains(cmd, "docker pull") {
		} else {
			fmt.Printf("[%s %s] MSG:\n", r.Host.HostName, r.Host.SSHAddress)
			fmt.Println(output)
		}
	}

	return output, nil
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
		stdout, _ := r.RunCmd(cmd)
		if validator(stdout) {
			return true
		}

		time.Sleep(1 * time.Second)
	}

	return false
}
