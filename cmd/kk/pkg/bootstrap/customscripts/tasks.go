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

package customscripts

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
)

type CustomScriptTask struct {
	action.BaseAction
	taskDir string
	script  kubekeyapiv1alpha2.CustomScripts
}

func (t *CustomScriptTask) Execute(runtime connector.Runtime) error {

	if len(t.script.Bash) <= 0 {
		return errors.Errorf("custom script %s Bash is empty", t.script.Name)
	}

	remoteTaskHome := common.TmpDir + t.taskDir

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", remoteTaskHome), false); err != nil {
		return errors.Wrapf(err, "create remoteTaskHome: %s  err:%s", remoteTaskHome, err)
	}

	// dilver the dependency materials to the remotehost
	for idx, localPath := range t.script.Materials {

		if !util.IsExist(localPath) {
			return errors.Errorf("Not found Path: %s", localPath)
		}

		targetPath := filepath.Join(remoteTaskHome, filepath.Base(localPath))

		// first clean the target to makesure target path always is the lastest.
		cleanCmd := fmt.Sprintf("rm -fr %s", targetPath)
		if _, err := runtime.GetRunner().SudoCmd(cleanCmd, false); err != nil {
			return errors.Wrapf(err, "Can not remove target found Path: %s", targetPath)
		}

		start := time.Now()
		err := runtime.GetRunner().SudoScp(localPath, targetPath)
		if err != nil {
			return errors.Wrapf(err, "Can not Scp -fr %s root@%s:%s", localPath, runtime.RemoteHost().GetAddress(), targetPath)
		}

		fmt.Printf("Copy %d/%d materials: Scp -fr %s root@%s:%s done, take %s\n",
			idx, len(t.script.Materials), localPath, runtime.RemoteHost().GetAddress(), targetPath, time.Since(start))
	}

	// wrap use bash file if shell has many lines.
	RunBash := t.script.Bash
	if strings.Index(RunBash, "\n") > 0 {
		tmpFile, err := ioutil.TempFile(os.TempDir(), t.taskDir)
		if err != nil {
			return errors.Wrapf(err, "create tmp Bash: %s/%s in local node, err:%s", os.TempDir(), t.taskDir, err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(RunBash); err != nil {
			return errors.Wrapf(err, "write to tmp:%s in local node, err:%s", tmpFile.Name(), err)
		}

		targetPath := filepath.Join(remoteTaskHome, "task.sh")
		if err := runtime.GetRunner().SudoScp(tmpFile.Name(), targetPath); err != nil {
			return errors.Wrapf(err, "Can not Scp -fr %s root@%s:%s", tmpFile.Name(), runtime.RemoteHost().GetAddress(), targetPath)
		}

		RunBash = "/bin/bash " + targetPath
	}

	start := time.Now()
	out, err := runtime.GetRunner().SudoCmd(RunBash, false)
	if err != nil {
		return errors.Errorf("Exec Bash: %s err:%s", RunBash, err)
	}

	if !runtime.GetRunner().Debug {
		fmt.Printf("Exec Bash:%s done, take %s", RunBash, time.Since(start))
		cleanCmd := fmt.Sprintf("rm -fr %s", remoteTaskHome)
		if _, err := runtime.GetRunner().SudoCmd(cleanCmd, false); err != nil {
			return errors.Wrapf(err, "Exec cmd:%s err:%s", cleanCmd, err)
		}
	} else {
		// keep the Materials for debug
		fmt.Printf("Exec Bash:%s done, take %s, output:\n%s", RunBash, time.Since(start), out)
	}

	return nil
}
