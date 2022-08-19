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

package directory

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/util/filesystem"
)

// Make wraps the Linux command "mkdir -p -m <mode> <path>".
func (s *Service) Make() error {
	_, err := s.sshClient.SudoCmdf("mkdir -p -m %o %s", filesystem.ToChmodPerm(s.mode), s.path)
	if err != nil {
		return errors.Wrapf(err, "failed to mkdir -p -m %o %s", filesystem.ToChmodPerm(s.mode), s.path)
	}
	return nil
}

// Chown wraps the linux command "chown <user> -R <path>".
func (s *Service) Chown(user string) error {
	_, err := s.sshClient.SudoCmd(fmt.Sprintf("chown %s -R %s", user, s.path))
	if err != nil {
		return errors.Wrapf(err, "failed to chown %s -R %s", user, s.path)
	}
	return nil
}

// Remove wraps the linux command "rm -rf <path>".
func (s *Service) Remove() error {
	_, err := s.sshClient.SudoCmd(fmt.Sprintf("rm -rf %s", s.path))
	if err != nil {
		return errors.Wrapf(err, "failed to rm -rf %s", s.path)
	}
	return nil
}
