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

package file

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/rootfs"
)

type Data map[string]interface{}

type Template struct {
	*File
	template *template.Template
	data     Data
	dst      string
}

func NewTemplate(sshClient ssh.Interface, rootFs rootfs.Interface, template *template.Template, data Data, dst string) (*Template, error) {
	file, err := NewFile(FileParams{
		SSHClient:      sshClient,
		RootFs:         rootFs,
		Name:           template.Name(),
		Type:           FileTemplate,
		LocalFullPath:  filepath.Join(rootFs.HostRootFsDir(sshClient.Host()), template.Name()),
		RemoteFullPath: dst,
	})
	if err != nil {
		return nil, err
	}
	return &Template{
		file,
		template,
		data,
		dst,
	}, nil
}

func (t *Template) RenderToLocal() error {
	dir := filepath.Dir(t.localFullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(t.localFullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	if err := t.template.Execute(f, t.data); err != nil {
		return err
	}
	return nil
}
