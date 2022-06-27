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
)

func (s *Service) Make() error {
	_, err := s.SSHClient.SudoCmd(fmt.Sprintf("mkdir -p -m %s %s", s.Mode.String(), s.Path))
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Chown(user string) error {
	_, err := s.SSHClient.SudoCmd(fmt.Sprintf("chown %s -R %s", user, s.Path))
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Remove() error {
	_, err := s.SSHClient.SudoCmd(fmt.Sprintf("rm -rf %s", s.Path))
	if err != nil {
		return err
	}
	return nil
}
