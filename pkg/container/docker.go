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

package container

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/container/templates"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/utils"
	"github.com/pkg/errors"
)

type SyncDockerBinaries struct {
	common.KubeAction
}

func (s *SyncDockerBinaries) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := s.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]files.KubeBinary)

	docker, ok := binariesMap[common.Docker]
	if !ok {
		return errors.New("get KubeBinary key docker by pipeline cache failed")
	}

	fileName := path.Base(docker.Path)
	dst := filepath.Join(common.TmpDir, fileName)

	if err := runtime.GetRunner().Scp(docker.Path, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync docker binaries failed"))
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("mkdir -p /usr/bin && tar -zxf %s && mv docker/* /usr/bin && rm -rf docker", dst),
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install container runtime docker binaries failed"))
	}
	return nil
}

type EnableDocker struct {
	common.KubeAction
}

func (e *EnableDocker) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"systemctl daemon-reload && systemctl enable docker && systemctl start docker",
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("enable and start docker failed"))
	}
	return nil
}

type DockerLoginRegistry struct {
	common.KubeAction
}

func (p *DockerLoginRegistry) Execute(runtime connector.Runtime) error {

	auths := templates.Auths(p.KubeConf)

	for repo, entry := range auths {

		cmd := fmt.Sprintf("docker login --username \"%s\" --password \"%s\" %s", entry.Username, entry.Password, repo)
		if _, err := runtime.GetRunner().SudoCmd(cmd, false); err != nil {
			return errors.Wrapf(err, "login registry failed, cmd: %v, err:%v", cmd, err)
		}
	}

	cmd := "mkdir -p /.docker && cp -f $HOME/.docker/config.json /.docker/ && chmod 0644 /.docker/config.json "
	if _, err := runtime.GetRunner().SudoCmd(cmd, false); err != nil {
		return errors.Wrapf(err, "copy docker auths failed cmd: %v, err:%v", cmd, err)
	}

	return nil
}
