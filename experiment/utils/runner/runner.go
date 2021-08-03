package runner

import (
	"errors"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/connector"
	"os"
	"time"
)

type Runner struct {
	Conn  connector.Connection
	Debug bool
	Host  *kubekeyapiv1alpha1.HostCfg
	Index int
}

const retryTime = 5 * time.Second

func (r *Runner) Cmd(cmd string, printOutput bool) (string, string, int, error) {
	if r.Conn == nil {
		return "", "", 1, errors.New("no ssh connection available")
	}

	stdout := NewTee(os.Stdout)
	defer stdout.Close()

	stderr := NewTee(os.Stderr)
	defer stderr.Close()

	var exitCode int

	code, err := r.Conn.PExec(cmd, nil, stdout, stderr)
	exitCode = code
	if err != nil {
		return "", err.Error(), exitCode, err
	}

	if printOutput && stdout.String() != "" {
		fmt.Printf("[%s %s] MSG:\n", r.Host.Name, r.Host.Address)
		fmt.Println(stdout.String())
	}

	return stdout.String(), stderr.String(), exitCode, nil
}

func (r *Runner) ScpFile(src, dst string) error {
	if r.Conn == nil {
		return errors.New("runner is not tied to an opened SSH connection")
	}

	err := r.Conn.Scp(src, dst)
	if err != nil {
		if r.Debug {
			fmt.Printf("Push %s to %s:%s   Failed\n", src, r.Host.Address, dst)
			return err
		}
	} else {
		if r.Debug {
			fmt.Printf("Push %s to %s:%s   Done\n", src, r.Host.Address, dst)
		}
	}
	return nil
}
