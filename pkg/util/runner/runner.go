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

package runner

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	ssh2 "github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"strings"
)

type Runner struct {
	Conn    ssh2.Connection
	Prefix  string
	OS      string
	Verbose bool
	Host    *kubekeyapi.HostCfg
	Index   int
}

func (r *Runner) RunCmd(cmd string) (string, error) {
	if r.Conn == nil {
		return "", errors.New("Runner is not tied to an opened SSH connection")
	}
	output, _, err := r.Conn.Exec(cmd, r.Host)
	if !r.Verbose {
		if err != nil {
			return "", err
		}
		return output, nil
	}

	if err != nil {
		return output, err
	}

	if output != "" {
		if strings.Contains(cmd, "base64") && strings.Contains(cmd, "--wrap=0") || strings.Contains(cmd, "make-ssl-etcd.sh") || strings.Contains(cmd, "docker-install.sh") || strings.Contains(cmd, "docker pull") {
		} else {
			fmt.Printf("[%s %s] MSG:\n", r.Host.Name, r.Host.Address)
			fmt.Println(output)
		}
	}

	return output, nil
}

func (r *Runner) ScpFile(src, dst string) error {
	if r.Conn == nil {
		return errors.New("Runner is not tied to an opened SSH connection")
	}

	err := r.Conn.Scp(src, dst)
	if err != nil {
		if r.Verbose {
			fmt.Printf("Push %s to %s:%s   Failed\n", src, r.Host.Address, dst)
			return err
		}
	} else {
		if r.Verbose {
			fmt.Printf("Push %s to %s:%s   Done\n", src, r.Host.Address, dst)
		}
	}
	return nil
}
