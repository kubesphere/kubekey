/*
Copyright 2020 The KubeSphere Authors.

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

package manager

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	kubekeyclientset "github.com/kubesphere/kubekey/clients/clientset/versioned"
	"github.com/kubesphere/kubekey/pkg/util/runner"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	log "github.com/sirupsen/logrus"
)

type Manager struct {
	ObjName        string
	Cluster        *kubekeyapiv1alpha1.ClusterSpec
	Logger         log.FieldLogger
	Connector      *ssh.Dialer
	Runner         *runner.Runner
	AllNodes       []kubekeyapiv1alpha1.HostCfg
	EtcdNodes      []kubekeyapiv1alpha1.HostCfg
	MasterNodes    []kubekeyapiv1alpha1.HostCfg
	WorkerNodes    []kubekeyapiv1alpha1.HostCfg
	K8sNodes       []kubekeyapiv1alpha1.HostCfg
	EtcdContainer  bool
	ClusterHosts   []string
	WorkDir        string
	KsEnable       bool
	KsVersion      string
	Debug          bool
	SkipCheck      bool
	SkipPullImages bool
	SourcesDir     string
	AddImagesRepo  bool
	InCluster      bool
	Kubeconfig     string
	Conditions     []kubekeyapiv1alpha1.Condition
	ClientSet      *kubekeyclientset.Clientset
}

func (mgr *Manager) Copy() *Manager {
	newManager := *mgr
	return &newManager
}
