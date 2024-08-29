/*
 Copyright 2022 The KubeSphere Authors.

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

package registry

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/bootstrap/registry/templates"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/files"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/utils"
)

//go:embed templates/harbor.yml.tmpl

var f embed.FS

type SyncCertsFile struct {
	common.KubeAction
}

func (s *SyncCertsFile) Execute(runtime connector.Runtime) error {
	localCertsDir, ok := s.ModuleCache.Get(LocalCertsDir)
	if !ok {
		return errors.New("get registry local certs dir by module cache failed")
	}
	files, ok := s.ModuleCache.Get(CertsFileList)
	if !ok {
		return errors.New("get registry certs file list by module cache failed")
	}
	dir := localCertsDir.(string)
	fileList := files.([]string)

	for _, fileName := range fileList {
		if err := runtime.GetRunner().SudoScp(filepath.Join(dir, fileName), filepath.Join(common.RegistryCertDir, fileName)); err != nil {
			return errors.Wrap(errors.WithStack(err), "scp registry certs file failed")
		}
	}

	return nil
}

type SyncCertsToAllNodes struct {
	common.KubeAction
}

func (s *SyncCertsToAllNodes) Execute(runtime connector.Runtime) error {
	localCertsDir, ok := s.ModuleCache.Get(LocalCertsDir)
	if !ok {
		return errors.New("get registry local certs dir by module cache failed")
	}
	files, ok := s.ModuleCache.Get(CertsFileList)
	if !ok {
		return errors.New("get registry certs file list by module cache failed")
	}
	dir := localCertsDir.(string)
	fileList := files.([]string)

	for _, fileName := range fileList {
		var dstFileName string
		switch fileName {
		case "ca.pem":
			dstFileName = "ca.crt"
		case "ca-key.pem":
			continue
		default:
			if strings.HasSuffix(fileName, "-key.pem") {
				dstFileName = strings.Replace(fileName, "-key.pem", ".key", -1)
			} else {
				dstFileName = strings.Replace(fileName, ".pem", ".cert", -1)
			}
		}

		if err := runtime.GetRunner().SudoScp(filepath.Join(dir, fileName), filepath.Join(filepath.Join("/etc/docker/certs.d", s.KubeConf.Cluster.Registry.GetHost()), dstFileName)); err != nil {
			return errors.Wrap(errors.WithStack(err), "scp registry certs file to /etc/docker/certs.d/ failed")
		}

		if err := runtime.GetRunner().SudoScp(filepath.Join(dir, fileName), filepath.Join(common.RegistryCertDir, dstFileName)); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp registry certs file to %s failed", common.RegistryCertDir))
		}
	}

	return nil
}

type InstallRegistryBinary struct {
	common.KubeAction
}

func (g *InstallRegistryBinary) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := g.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	registry, ok := binariesMap[common.Registry]
	if !ok {
		return errors.New("get KubeBinary key registry by pipeline cache failed")
	}

	dst := filepath.Join(common.TmpDir, registry.FileName)
	if err := runtime.GetRunner().Scp(registry.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync etcd tar.gz failed")
	}

	installCmd := fmt.Sprintf("tar -zxf %s && mv -f registry /usr/local/bin/ && chmod +x /usr/local/bin/registry", dst)
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "install etcd binaries failed")
	}
	return nil
}

type StartRegistryService struct {
	common.KubeAction
}

func (g *StartRegistryService) Execute(runtime connector.Runtime) error {
	installCmd := "systemctl daemon-reload && systemctl enable registry && systemctl restart registry"
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "start registry service failed")
	}

	fmt.Println()
	fmt.Println(fmt.Sprintf("Local image registry created successfully. Address: %s", g.KubeConf.Cluster.Registry.GetHost()))
	fmt.Println()

	return nil
}

type InstallDockerCompose struct {
	common.KubeAction
}

func (g *InstallDockerCompose) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := g.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	compose, ok := binariesMap[common.DockerCompose]
	if !ok {
		return errors.New("get KubeBinary key docker-compose by pipeline cache failed")
	}

	dst := filepath.Join(common.TmpDir, compose.FileName)
	if err := runtime.GetRunner().Scp(compose.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync docker-compose failed")
	}

	installCmd := fmt.Sprintf("mv -f %s /usr/local/bin/docker-compose && chmod +x /usr/local/bin/docker-compose", dst)
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "install dokcer-compose failed")
	}

	return nil
}

type SyncHarborPackage struct {
	common.KubeAction
}

func (g *SyncHarborPackage) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := g.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	harbor, ok := binariesMap[common.Harbor]
	if !ok {
		return errors.New("get KubeBinary key harbor by pipeline cache failed")
	}

	dst := filepath.Join(common.TmpDir, harbor.FileName)
	if err := runtime.GetRunner().Scp(harbor.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync harbor package failed")
	}

	installCmd := fmt.Sprintf("tar -zxvf %s -C /opt", dst)
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "unzip harbor package failed")
	}

	return nil
}

type GenerateHarborConfig struct {
	common.KubeAction
}

func (g *GenerateHarborConfig) Execute(runtime connector.Runtime) error {
	harborContent, err := f.ReadFile("templates/harbor.yml.tmpl")
	if err != nil {
		return err
	}
	harbor := template.Must(template.New("harbor.yml").Parse(string(harborContent)))

	registryDomain := g.KubeConf.Cluster.Registry.GetHost()

	if g.KubeConf.Cluster.Registry.Type == "harbor-ha" {
		host := runtime.RemoteHost()
		registryDomain = host.GetName()
	}

	templateAction := action.Template{
		Template: harbor,
		Dst:      "/opt/harbor/harbor.yml",
		Data: util.Data{
			"Domain":      registryDomain,
			"Certificate": fmt.Sprintf("%s.pem", g.KubeConf.Cluster.Registry.GetHost()),
			"Key":         fmt.Sprintf("%s-key.pem", g.KubeConf.Cluster.Registry.GetHost()),
			"Password":    templates.Password(g.KubeConf, g.KubeConf.Cluster.Registry.GetHost()),
		},
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type StartHarbor struct {
	common.KubeAction
}

func (g *StartHarbor) Execute(runtime connector.Runtime) error {
	startCmd := "cd /opt/harbor && chmod +x install.sh && export PATH=$PATH:/usr/local/bin; ./install.sh --with-trivy && systemctl daemon-reload && systemctl enable harbor && systemctl restart harbor"
	//harbor 2.8.0开始移除--with-chartmuseum
	if VersionCompare(g.KubeConf.Cluster.Registry.Version, "v2.8.0") {
		startCmd = "cd /opt/harbor && chmod +x install.sh && export PATH=$PATH:/usr/local/bin; ./install.sh --with-notary --with-trivy --with-chartmuseum && systemctl daemon-reload && systemctl enable harbor && systemctl restart harbor"
	}
	if _, err := runtime.GetRunner().SudoCmd(startCmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "start harbor failed")
	}

	fmt.Println()
	fmt.Println(fmt.Sprintf("Local image registry created successfully. Address: %s", g.KubeConf.Cluster.Registry.GetHost()))
	fmt.Println()

	return nil
}

// VersionCompare
// @description: 版本比较，v2版本是否大于v1版本
// @param: v1 string 当前版本
// @param: v2 string 最新版本
// @author: GJing
// @email: gjing1st@gmail.com
// @date: 2024/8/29 14:10
// @success: v2>v1返回true
func VersionCompare(v1, v2 string) (res bool) {
	if len(v1) == 0 {
		return true
	}
	if len(v2) == 0 {
		return false
	}
	if v1[0] == 'v' || v1[0] == 'V' {
		v1 = v1[1:]
	}
	if v2[0] == 'v' || v2[0] == 'V' {
		v2 = v2[1:]
	}
	v1Arr := strings.Split(v1, ".")
	v2Arr := strings.Split(v2, ".")
	if len(v1Arr) == len(v2Arr) {
		//版本位数相同
		for i, v := range v2Arr {
			if v > v1Arr[i] {
				res = true
				return
			} else if v < v1Arr[i] {
				return
			}
		}
	} else if len(v1Arr) < len(v2Arr) {
		//新版本位数更多
		for i, v := range v1Arr {
			if v2Arr[i] > v {
				res = true
				return
			} else if v2Arr[i] < v {
				return
			}
			if i == len(v1Arr)-1 {
				//v1的版本位数跟v2的一样，v2的位数更多，返回true
				res = true
				return
			}
		}

	} else if len(v1Arr) > len(v2Arr) {
		//旧版本位数多，新版本位数少
		for i, v := range v2Arr {
			if v > v1Arr[i] {
				res = true
				return
			} else if v < v1Arr[i] {
				return
			}
		}
	}
	return
}
