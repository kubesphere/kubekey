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

	"github.com/kubesphere/kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/pkg/rootfs"
	"github.com/kubesphere/kubekey/pkg/util/filesystem"
)

// Data is the data that will be passed to the template.
type Data map[string]interface{}

// Template is an implementation of the Template interface.
type Template struct {
	*File
	template *template.Template
	data     Data
	dst      string
}

// NewTemplate returns a new Template.
func NewTemplate(sshClient ssh.Interface, rootFs rootfs.Interface, template *template.Template, data Data, dst string) (*Template, error) {
	file, err := NewFile(Params{
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

// RenderToLocal renders the template to the local filesystem.
func (t *Template) RenderToLocal() error {
	dir := filepath.Dir(t.localFullPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, filesystem.FileMode0755); err != nil {
			return err
		}
	}

	f, err := os.Create(t.localFullPath)
	if err != nil {
		return err
	}

	if err := t.template.Execute(f, t.data); err != nil {
		return err
	}
	return nil
}
