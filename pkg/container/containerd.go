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
	"path/filepath"
	"strings"

	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/container/templates"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/registry"
	"github.com/kubesphere/kubekey/pkg/utils"
	"github.com/pkg/errors"
)

type SyncContainerd struct {
	common.KubeAction
}

func (s *SyncContainerd) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := s.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	containerd, ok := binariesMap[common.Conatinerd]
	if !ok {
		return errors.New("get KubeBinary key containerd by pipeline cache failed")
	}

	dst := filepath.Join(common.TmpDir, containerd.FileName)
	if err := runtime.GetRunner().Scp(containerd.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync containerd binaries failed"))
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("mkdir -p /usr/bin && tar -zxf %s && mv bin/* /usr/bin && rm -rf bin", dst),
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install containerd binaries failed"))
	}
	return nil
}

type SyncCrictlBinaries struct {
	common.KubeAction
}

func (s *SyncCrictlBinaries) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := s.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	crictl, ok := binariesMap[common.Crictl]
	if !ok {
		return errors.New("get KubeBinary key crictl by pipeline cache failed")
	}

	dst := filepath.Join(common.TmpDir, crictl.FileName)

	if err := runtime.GetRunner().Scp(crictl.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync crictl binaries failed"))
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("mkdir -p /usr/bin && tar -zxf %s -C /usr/bin ", dst),
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install crictl binaries failed"))
	}
	return nil
}

type EnableContainerd struct {
	common.KubeAction
}

func (e *EnableContainerd) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"systemctl daemon-reload && systemctl enable containerd && systemctl start containerd",
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("enable and start containerd failed"))
	}

	// install runc
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := e.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)

	containerd, ok := binariesMap[common.Runc]
	if !ok {
		return errors.New("get KubeBinary key runc by pipeline cache failed")
	}

	dst := filepath.Join(common.TmpDir, containerd.FileName)
	if err := runtime.GetRunner().Scp(containerd.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync runc binaries failed"))
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("install -m 755 %s /usr/local/sbin/runc", dst),
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install runc binaries failed"))
	}
	return nil
}

type DisableContainerd struct {
	common.KubeAction
}

func (d *DisableContainerd) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"systemctl disable containerd && systemctl stop containerd", true); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("disable and stop containerd failed"))
	}

	// remove containerd related files
	files := []string{
		"/usr/local/sbin/runc",
		"/usr/bin/crictl",
		"/usr/bin/containerd*",
		"/usr/bin/ctr",
		filepath.Join("/etc/systemd/system", templates.ContainerdService.Name()),
		filepath.Join("/etc/containerd", templates.ContainerdConfig.Name()),
		filepath.Join("/etc", templates.CrictlConfig.Name()),
	}
	if d.KubeConf.Cluster.Registry.DataRoot != "" {
		files = append(files, d.KubeConf.Cluster.Registry.DataRoot)
	} else {
		files = append(files, "/var/lib/containerd")
	}

	for _, file := range files {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), true)
	}
	return nil
}

type CordonNode struct {
	common.KubeAction
}

func (d *CordonNode) Execute(runtime connector.Runtime) error {
	nodeName := runtime.RemoteHost().GetName()
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("/usr/local/bin/kubectl cordon %s ", nodeName), true); err != nil {
		return errors.Wrap(err, fmt.Sprintf("cordon the node: %s failed", nodeName))
	}
	return nil
}

type UnCordonNode struct {
	common.KubeAction
}

func (d *UnCordonNode) Execute(runtime connector.Runtime) error {
	nodeName := runtime.RemoteHost().GetName()
	f := true
	for f {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("/usr/local/bin/kubectl uncordon %s", nodeName), true); err == nil {
			break
		}

	}
	return nil
}

type DrainNode struct {
	common.KubeAction
}

func (d *DrainNode) Execute(runtime connector.Runtime) error {
	nodeName := runtime.RemoteHost().GetName()
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("/usr/local/bin/kubectl drain %s --delete-emptydir-data --ignore-daemonsets --timeout=2m --force", nodeName), true); err != nil {
		return errors.Wrap(err, fmt.Sprintf("drain the node: %s failed", nodeName))
	}
	return nil
}

type RestartCri struct {
	common.KubeAction
}

func (i *RestartCri) Execute(runtime connector.Runtime) error {
	switch i.KubeConf.Arg.Type {
	case common.Docker:
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("systemctl daemon-reload && systemctl restart docker "), true); err != nil {
			return errors.Wrap(err, "restart docker")
		}
	case common.Conatinerd:
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("systemctl daemon-reload && systemctl restart containerd"), true); err != nil {
			return errors.Wrap(err, "restart containerd")
		}

	default:
		logger.Log.Fatalf("Unsupported container runtime: %s", strings.TrimSpace(i.KubeConf.Arg.Type))
	}
	return nil
}

type EditKubeletCri struct {
	common.KubeAction
}

func (i *EditKubeletCri) Execute(runtime connector.Runtime) error {
	switch i.KubeConf.Arg.Type {
	case common.Docker:
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
			"sed -i 's#--container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --pod#--pod#' /var/lib/kubelet/kubeadm-flags.env"),
			true); err != nil {
			return errors.Wrap(err, "Change KubeletTo Containerd failed")
		}
	case common.Conatinerd:
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
			"sed -i 's#--network-plugin=cni --pod#--network-plugin=cni --container-runtime=remote --container-runtime-endpoint=unix:///run/containerd/containerd.sock --pod#' /var/lib/kubelet/kubeadm-flags.env"),
			true); err != nil {
			return errors.Wrap(err, "Change KubeletTo Containerd failed")
		}

	default:
		logger.Log.Fatalf("Unsupported container runtime: %s", strings.TrimSpace(i.KubeConf.Arg.Type))
	}
	return nil
}

type RestartKubeletNode struct {
	common.KubeAction
}

func (d *RestartKubeletNode) Execute(runtime connector.Runtime) error {

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("systemctl restart kubelet"), true); err != nil {
		return errors.Wrap(err, "RestartNode Kube failed")
	}
	return nil
}

func MigrateSelfNodeCriTasks(runtime connector.Runtime, kubeAction common.KubeAction) error {
	host := runtime.RemoteHost()
	tasks := []task.Interface{}
	CordonNode := &task.RemoteTask{
		Name:  "CordonNode",
		Desc:  "Cordon Node",
		Hosts: []connector.Host{host},

		Action:   new(CordonNode),
		Parallel: false,
	}
	DrainNode := &task.RemoteTask{
		Name:     "DrainNode",
		Desc:     "Drain Node",
		Hosts:    []connector.Host{host},
		Action:   new(DrainNode),
		Parallel: false,
	}
	RestartCri := &task.RemoteTask{
		Name:     "RestartCri",
		Desc:     "Restart Cri",
		Hosts:    []connector.Host{host},
		Action:   new(RestartCri),
		Parallel: false,
	}
	EditKubeletCri := &task.RemoteTask{
		Name:     "EditKubeletCri",
		Desc:     "Edit Kubelet Cri",
		Hosts:    []connector.Host{host},
		Action:   new(EditKubeletCri),
		Parallel: false,
	}
	RestartKubeletNode := &task.RemoteTask{
		Name:     "RestartKubeletNode",
		Desc:     "Restart Kubelet Node",
		Hosts:    []connector.Host{host},
		Action:   new(RestartKubeletNode),
		Parallel: false,
	}
	UnCordonNode := &task.RemoteTask{
		Name:     "UnCordonNode",
		Desc:     "UnCordon Node",
		Hosts:    []connector.Host{host},
		Action:   new(UnCordonNode),
		Parallel: false,
	}
	switch kubeAction.KubeConf.Cluster.Kubernetes.ContainerManager {
	case common.Docker:
		Uninstall := &task.RemoteTask{
			Name:  "DisableDocker",
			Desc:  "Disable docker",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				&DockerExist{Not: false},
			},
			Action:   new(DisableDocker),
			Parallel: false,
		}
		tasks = append(tasks, CordonNode, DrainNode, Uninstall)
	case common.Conatinerd:
		Uninstall := &task.RemoteTask{
			Name:  "UninstallContainerd",
			Desc:  "Uninstall containerd",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				&ContainerdExist{Not: false},
			},
			Action:   new(DisableContainerd),
			Parallel: false,
		}
		tasks = append(tasks, CordonNode, DrainNode, Uninstall)
	}
	if kubeAction.KubeConf.Arg.Type == common.Docker {
		syncBinaries := &task.RemoteTask{
			Name:  "SyncDockerBinaries",
			Desc:  "Sync docker binaries",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				// &kubernetes.NodeInCluster{Not: true},
				&DockerExist{Not: true},
			},
			Action:   new(SyncDockerBinaries),
			Parallel: false,
		}
		generateDockerService := &task.RemoteTask{
			Name:  "GenerateDockerService",
			Desc:  "Generate docker service",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				// &kubernetes.NodeInCluster{Not: true},
				&DockerExist{Not: true},
			},
			Action: &action.Template{
				Template: templates.DockerService,
				Dst:      filepath.Join("/etc/systemd/system", templates.DockerService.Name()),
			},
			Parallel: false,
		}
		generateDockerConfig := &task.RemoteTask{
			Name:  "GenerateDockerConfig",
			Desc:  "Generate docker config",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				// &kubernetes.NodeInCluster{Not: true},
				&DockerExist{Not: true},
			},
			Action: &action.Template{
				Template: templates.DockerConfig,
				Dst:      filepath.Join("/etc/docker/", templates.DockerConfig.Name()),
				Data: util.Data{
					"Mirrors":            templates.Mirrors(kubeAction.KubeConf),
					"InsecureRegistries": templates.InsecureRegistries(kubeAction.KubeConf),
					"DataRoot":           templates.DataRoot(kubeAction.KubeConf),
				},
			},
			Parallel: false,
		}
		enableDocker := &task.RemoteTask{
			Name:  "EnableDocker",
			Desc:  "Enable docker",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				// &kubernetes.NodeInCluster{Not: true},
				&DockerExist{Not: true},
			},
			Action:   new(EnableDocker),
			Parallel: false,
		}
		dockerLoginRegistry := &task.RemoteTask{
			Name:  "Login PrivateRegistry",
			Desc:  "Add auths to container runtime",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				// &kubernetes.NodeInCluster{Not: true},
				&DockerExist{},
				&PrivateRegistryAuth{},
			},
			Action:   new(DockerLoginRegistry),
			Parallel: false,
		}

		tasks = append(tasks, syncBinaries, generateDockerService, generateDockerConfig, enableDocker, dockerLoginRegistry,
			RestartCri, EditKubeletCri, RestartKubeletNode, UnCordonNode)
	}
	if kubeAction.KubeConf.Arg.Type == common.Conatinerd {
		syncContainerd := &task.RemoteTask{
			Name:  "SyncContainerd",
			Desc:  "Sync containerd binaries",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				&ContainerdExist{Not: true},
			},
			Action:   new(SyncContainerd),
			Parallel: false,
		}

		syncCrictlBinaries := &task.RemoteTask{
			Name:  "SyncCrictlBinaries",
			Desc:  "Sync crictl binaries",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				&CrictlExist{Not: true},
			},
			Action:   new(SyncCrictlBinaries),
			Parallel: false,
		}

		generateContainerdService := &task.RemoteTask{
			Name:  "GenerateContainerdService",
			Desc:  "Generate containerd service",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				&ContainerdExist{Not: true},
			},
			Action: &action.Template{
				Template: templates.ContainerdService,
				Dst:      filepath.Join("/etc/systemd/system", templates.ContainerdService.Name()),
			},
			Parallel: false,
		}

		generateContainerdConfig := &task.RemoteTask{
			Name:  "GenerateContainerdConfig",
			Desc:  "Generate containerd config",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				&ContainerdExist{Not: true},
			},
			Action: &action.Template{
				Template: templates.ContainerdConfig,
				Dst:      filepath.Join("/etc/containerd/", templates.ContainerdConfig.Name()),
				Data: util.Data{
					"Mirrors":            templates.Mirrors(kubeAction.KubeConf),
					"InsecureRegistries": kubeAction.KubeConf.Cluster.Registry.InsecureRegistries,
					"SandBoxImage":       images.GetImage(runtime, kubeAction.KubeConf, "pause").ImageName(),
					"Auths":              registry.DockerRegistryAuthEntries(kubeAction.KubeConf.Cluster.Registry.Auths),
					"DataRoot":           templates.DataRoot(kubeAction.KubeConf),
				},
			},
			Parallel: false,
		}

		generateCrictlConfig := &task.RemoteTask{
			Name:  "GenerateCrictlConfig",
			Desc:  "Generate crictl config",
			Hosts: []connector.Host{host},
			Prepare: &prepare.PrepareCollection{
				&ContainerdExist{Not: true},
			},
			Action: &action.Template{
				Template: templates.CrictlConfig,
				Dst:      filepath.Join("/etc/", templates.CrictlConfig.Name()),
				Data: util.Data{
					"Endpoint": kubeAction.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint,
				},
			},
			Parallel: false,
		}

		enableContainerd := &task.RemoteTask{
			Name:  "EnableContainerd",
			Desc:  "Enable containerd",
			Hosts: []connector.Host{host},
			// Prepare: &prepare.PrepareCollection{
			// 	&ContainerdExist{Not: true},
			// },
			Action:   new(EnableContainerd),
			Parallel: false,
		}
		tasks = append(tasks, syncContainerd, syncCrictlBinaries, generateContainerdService, generateContainerdConfig,
			generateCrictlConfig, enableContainerd, RestartCri, EditKubeletCri, RestartKubeletNode, UnCordonNode)
	}

	for i := range tasks {
		t := tasks[i]
		t.Init(runtime, kubeAction.ModuleCache, kubeAction.PipelineCache)
		if res := t.Execute(); res.IsFailed() {
			return res.CombineErr()
		}
	}
	return nil
}

type MigrateSelfNodeCri struct {
	common.KubeAction
}

func (d *MigrateSelfNodeCri) Execute(runtime connector.Runtime) error {

	if err := MigrateSelfNodeCriTasks(runtime, d.KubeAction); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("MigrateSelfNodeCriTasks failed:"))
	}
	return nil
}
