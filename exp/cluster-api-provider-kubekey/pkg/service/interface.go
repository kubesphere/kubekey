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

// Package service implements various services.
package service

import (
	"time"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/provisioning/commands"
)

// Bootstrap is the interface for bootstrap provision.
type Bootstrap interface {
	AddUsers() error
	SetHostname() error
	CreateDirectory() error
	ResetTmpDirectory() error
	ExecInitScript() error
	Repository() error
	ResetNetwork() error
	RemoveFiles() error
	DaemonReload() error
}

// BinaryService is the interface for binary provision.
type BinaryService interface {
	DownloadAll(timeout time.Duration) error
	ConfigureKubelet() error
	ConfigureKubeadm() error
}

// ContainerManager is the interface for container manager provision.
type ContainerManager interface {
	Type() string
	IsExist() bool
	Get(timeout time.Duration) error
	Install() error
}

// Provisioning is the interface for bootstrap generate by CABPK provision.
type Provisioning interface {
	RawBootstrapDataToProvisioningCommands(config []byte) ([]commands.Cmd, error)
}
