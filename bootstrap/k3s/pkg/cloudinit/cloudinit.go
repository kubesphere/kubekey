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

package cloudinit

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/pkg/errors"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
)

const (
	// sentinelFileCommand writes a file to /run/cluster-api to signal successful Kubernetes bootstrapping in a way that
	// works both for Linux and Windows OS.
	sentinelFileCommand = "echo success > /run/cluster-api/bootstrap-success.complete"
	cloudConfigHeader   = `## template: jinja
#cloud-config
`
)

// BaseUserData is shared across all the various types of files written to disk.
type BaseUserData struct {
	Header              string
	PreK3sCommands      []string
	PostK3sCommands     []string
	AdditionalFiles     []bootstrapv1.File
	WriteFiles          []bootstrapv1.File
	ConfigFile          bootstrapv1.File
	SentinelFileCommand string
}

func (input *BaseUserData) prepare() error {
	input.Header = cloudConfigHeader
	input.WriteFiles = append(input.WriteFiles, input.AdditionalFiles...)
	k3sScriptFile, err := generateBootstrapScript(input)
	if err != nil {
		return errors.Wrap(err, "failed to generate user data for machine install k3s")
	}
	input.WriteFiles = append(input.WriteFiles, *k3sScriptFile)
	input.SentinelFileCommand = sentinelFileCommand
	return nil
}

func generate(kind string, tpl string, data interface{}) ([]byte, error) {
	tm := template.New(kind).Funcs(defaultTemplateFuncMap)
	if _, err := tm.Parse(filesTemplate); err != nil {
		return nil, errors.Wrap(err, "failed to parse files template")
	}

	if _, err := tm.Parse(commandsTemplate); err != nil {
		return nil, errors.Wrap(err, "failed to parse commands template")
	}

	t, err := tm.Parse(tpl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s template", kind)
	}

	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return nil, errors.Wrapf(err, "failed to generate %s template", kind)
	}

	return out.Bytes(), nil
}

var (
	//go:embed k3s-install.sh
	k3sBootstrapScript string
)

func generateBootstrapScript(input interface{}) (*bootstrapv1.File, error) {
	k3sScript, err := generate("K3sInstallScript", k3sBootstrapScript, input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to bootstrap script for machine joins")
	}
	return &bootstrapv1.File{
		Path:        "/usr/local/bin/k3s-install.sh",
		Owner:       "root",
		Permissions: "0755",
		Content:     string(k3sScript),
	}, nil
}
