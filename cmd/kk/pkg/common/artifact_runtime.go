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

package common

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"

	kubekeyv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
)

type ArtifactArgument struct {
	ManifestFile       string
	Output             string
	CriSocket          string
	Debug              bool
	IgnoreErr          bool
	DownloadCommand    func(path, url string) string
	ImageStartIndex    int
	SkipRemoveArtifact bool
}

type ArtifactRuntime struct {
	LocalRuntime
	Spec *kubekeyv1alpha2.ManifestSpec
	Arg  ArtifactArgument
}

func NewArtifactRuntime(arg ArtifactArgument) (*ArtifactRuntime, error) {
	localRuntime, err := NewLocalRuntime(arg.Debug, arg.IgnoreErr)
	if err != nil {
		return nil, err
	}

	fp, err := filepath.Abs(arg.ManifestFile)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to look up current directory")
	}

	fileByte, err := os.ReadFile(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read file %s", fp)
	}

	contentToJson, err := k8syaml.ToJSON(fileByte)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to convert configuration to json")
	}

	manifest := &kubekeyv1alpha2.Manifest{}
	if err := json.Unmarshal(contentToJson, manifest); err != nil {
		return nil, errors.Wrapf(err, "Failed to json unmarshal")
	}

	r := &ArtifactRuntime{
		Spec: &manifest.Spec,
		Arg:  arg,
	}
	r.LocalRuntime = localRuntime
	return r, nil
}

// Copy is used to create a copy for Runtime.
func (a *ArtifactRuntime) Copy() connector.Runtime {
	runtime := *a
	return &runtime
}
