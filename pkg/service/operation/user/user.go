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

package user

import (
	"fmt"

	"github.com/pkg/errors"
)

// Add adds a new Linux user to the remote instance.
func (s *Service) Add() error {
	_, err := s.SSHClient.SudoCmd(fmt.Sprintf("useradd -M -c '%s' -s /sbin/nologin -r %s || :", s.Desc, s.Name))
	if err != nil {
		return errors.Wrapf(err, "failed to add user %s", s.Name)
	}
	return nil
}
