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

package repository

import (
	"strings"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
)

type Service interface {
	Update() error
	Install(pkg ...string) error
}

func NewService(sshClient ssh.Interface, os string) Service {

	switch strings.ToLower(os) {
	case "ubuntu", "debian":
		return NewDeb(sshClient)
	case "centos", "rhel", "fedora":
		return NewRPM(sshClient)
	default:
		return nil
	}
}
