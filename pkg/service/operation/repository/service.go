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
	"github.com/kubesphere/kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/util/osrelease"
)

// Service holds a collection of interfaces.
// The interfaces are broken down like this to group functions together.
type Service interface {
	Add(path string) error
	Update() error
	Install(pkg ...string) error
}

// NewService returns a new service given the remote instance package manager client.
func NewService(sshClient ssh.Interface, os *osrelease.Data) Service {
	if os == nil {
		return nil
	}

	switch {
	case os.IsLikeDebian():
		return NewDeb(sshClient)
	case os.IsLikeFedora():
		return NewRPM(sshClient)
	default:
		return nil
	}
}
