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

package repository

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type Interface interface {
	Backup() error
	IsAlreadyBackUp() bool
	Add(path string) error
	Update() error
	Install(pkg ...string) error
	Reset() error
}

func New(os string, runtime connector.Runtime) (Interface, error) {
	switch os {
	case "ubuntu":
		return NewDeb(runtime), nil
	case "centos":
		return NewRPM(runtime), nil
	default:
		return nil, fmt.Errorf("unsupported operation system %s", os)
	}
}
