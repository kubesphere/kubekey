package runner

import (
	"errors"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/experiment/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/core/connector"
	"github.com/kubesphere/kubekey/experiment/core/logger"
	"os"
)

type Runner struct {
	Conn  connector.Connection
	Debug bool
	Host  *kubekeyapiv1alpha1.HostCfg
	Index int
}

// todo: return value may be too much
func (r *Runner) Cmd(cmd string, printOutput bool) (string, string, int, error) {
	if r.Conn == nil {
		return "", "", 1, errors.New("no ssh connection available")
	}

	stdout := NewTee(os.Stdout)
	defer stdout.Close()

	stderr := NewTee(os.Stderr)
	defer stderr.Close()

	code, err := r.Conn.PExec(cmd, nil, stdout, stderr)
	if printOutput {
		if stdout.String() != "" {
			logger.Log.Infof("[stdout]: %s", stdout.String())
		}
		if stderr.String() != "" {
			logger.Log.Infof("[stderr]: %s", stderr.String())
		}
	}
	if err != nil {
		return "", err.Error(), code, err
	}

	return stdout.String(), stderr.String(), code, nil
}

func (r *Runner) Fetch(local, remote string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Fetch(local, remote); err != nil {
		logger.Log.Errorf("fetch remote file %s to local %s failed: %v", remote, local, err)
		return err
	}
	logger.Log.Debugf("fetch remote file %s to local %s success", remote, local)
	return nil
}

func (r *Runner) Scp(remote, local string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Scp(remote, local); err != nil {
		logger.Log.Errorf("scp local file %s to remote %s failed: %v", remote, local, err)
		return err
	}
	logger.Log.Debugf("scp local file %s to remote %s success", remote, local)
	return nil
}

func (r *Runner) FileExist(remote string) (bool, error) {
	if r.Conn == nil {
		return false, errors.New("no ssh connection available")
	}

	ok := r.Conn.RemoteFileExist(remote)
	logger.Log.Debugf("check remote file exist: %v", ok)
	return ok, nil
}

func (r *Runner) DirExist(remote string) (bool, error) {
	if r.Conn == nil {
		return false, errors.New("no ssh connection available")
	}

	ok, err := r.Conn.RemoteDirExist(remote)
	if err != nil {
		logger.Log.Errorf("check remote dir exist failed: %v", err)
		return false, err
	}
	logger.Log.Debugf("check remote dir exist: %v", ok)
	return ok, nil
}
