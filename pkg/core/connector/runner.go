/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package connector

import (
	"errors"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/common"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"os"
	"path/filepath"
)

type Runner struct {
	Conn  Connection
	Debug bool
	Host  Host
	Index int
}

func (r *Runner) Exec(cmd string, printOutput bool) (string, int, error) {
	if r.Conn == nil {
		return "", 1, errors.New("no ssh connection available")
	}

	//stdout := NewTee(os.Stdout)
	//defer stdout.Close()
	//
	//stderr := NewTee(os.Stderr)
	//defer stderr.Close()
	//
	//code, err := r.Conn.PExec(cmd, nil, stdout, stderr)

	//if printOutput {
	//	if stdout.String() != "" {
	//		logger.Log.Infof("[stdout]: %s", stdout.String())
	//	}
	//	if stderr.String() != "" {
	//		logger.Log.Infof("[stderr]: %s", stderr.String())
	//	}
	//}
	//if err != nil {
	//	return "", err.Error(), code, err
	//}
	//
	//return stdout.String(), stderr.String(), code, nil
	stdout, code, err := r.Conn.Exec(cmd, r.Host)
	if printOutput {
		if stdout != "" {
			logger.Log.Infof("stdout: [%s]\n%s", r.Host.GetName(), stdout)
		}
		//if stderr != "" {
		//	logger.Log.Infof("stderr: [%s]\n%s", r.Host.GetName(), stderr)
		//}
	}
	return stdout, code, err
}

func (r *Runner) Cmd(cmd string, printOutput bool) (string, error) {
	stdout, _, err := r.Exec(cmd, printOutput)
	if err != nil {
		return stdout, err
	}
	return stdout, nil
}

func (r *Runner) SudoExec(cmd string, printOutput bool) (string, int, error) {
	return r.Exec(SudoPrefix(cmd), printOutput)
}

func (r *Runner) SudoCmd(cmd string, printOutput bool) (string, error) {
	return r.Cmd(SudoPrefix(cmd), printOutput)
}

func (r *Runner) Fetch(local, remote string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Fetch(local, remote, r.Host); err != nil {
		logger.Log.Debugf("fetch remote file %s to local %s failed: %v", remote, local, err)
		return err
	}
	logger.Log.Debugf("fetch remote file %s to local %s success", remote, local)
	return nil
}

func (r *Runner) Scp(local, remote string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Scp(local, remote, r.Host); err != nil {
		logger.Log.Debugf("scp local file %s to remote %s failed: %v", local, remote, err)
		return err
	}
	logger.Log.Debugf("scp local file %s to remote %s success", local, remote)
	return nil
}

func (r *Runner) SudoScp(local, remote string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	// scp to tmp dir
	remoteTmp := filepath.Join(common.TmpDir, remote)
	//remoteTmp := remote
	if err := r.Scp(local, remoteTmp); err != nil {
		return err
	}

	baseRemotePath := remote
	if !util.IsDir(local) {
		baseRemotePath = filepath.Dir(remote)
	}
	if err := r.Conn.MkDirAll(baseRemotePath, "", r.Host); err != nil {
		return err
	}

	if _, err := r.SudoCmd(fmt.Sprintf(common.MoveCmd, remoteTmp, remote), false); err != nil {
		return err
	}

	if _, err := r.SudoCmd(fmt.Sprintf("rm -rf %s", filepath.Join(common.TmpDir, "*")), false); err != nil {
		return err
	}
	return nil
}

func (r *Runner) FileExist(remote string) (bool, error) {
	if r.Conn == nil {
		return false, errors.New("no ssh connection available")
	}

	ok := r.Conn.RemoteFileExist(remote, r.Host)
	logger.Log.Debugf("check remote file exist: %v", ok)
	return ok, nil
}

func (r *Runner) DirExist(remote string) (bool, error) {
	if r.Conn == nil {
		return false, errors.New("no ssh connection available")
	}

	ok, err := r.Conn.RemoteDirExist(remote, r.Host)
	if err != nil {
		logger.Log.Debugf("check remote dir exist failed: %v", err)
		return false, err
	}
	logger.Log.Debugf("check remote dir exist: %v", ok)
	return ok, nil
}

func (r *Runner) MkDir(path string) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.MkDirAll(path, "", r.Host); err != nil {
		logger.Log.Errorf("make remote dir %s failed: %v", path, err)
		return err
	}
	return nil
}

func (r *Runner) Chmod(path string, mode os.FileMode) error {
	if r.Conn == nil {
		return errors.New("no ssh connection available")
	}

	if err := r.Conn.Chmod(path, mode); err != nil {
		logger.Log.Errorf("chmod remote path %s failed: %v", path, err)
		return err
	}
	return nil
}

func (r *Runner) FileMd5(path string) (string, error) {
	if r.Conn == nil {
		return "", errors.New("no ssh connection available")
	}

	cmd := fmt.Sprintf("md5sum %s | cut -d\" \" -f1", path)
	out, _, err := r.Conn.Exec(cmd, r.Host)
	if err != nil {
		logger.Log.Errorf("count remote %s md5 failed: %v", path, err)
		return "", err
	}
	return out, nil
}
