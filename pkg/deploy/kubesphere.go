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

package deploy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/pkg/errors"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"strings"
)

func DeployKubeSphere(version, repo, kubeconfig string) error {
	restCfg, err := util.NewDynamicClient(kubeconfig)
	if err != nil {
		return err
	}

	var kubesphereConfig, installerYaml string

	switch version {
	case "":
		kubesphereConfig = kubesphere.V3_0_0
		str, err := kubesphere.GenerateKubeSphereYaml(repo, "latest")
		if err != nil {
			return err
		}
		installerYaml = str
	case "v3.0.0":
		kubesphereConfig = kubesphere.V3_0_0
		str, err := kubesphere.GenerateKubeSphereYaml(repo, "latest")
		if err != nil {
			return err
		}
		installerYaml = str
	case "v2.1.1":
		kubesphereConfig = kubesphere.V2_1_1
		str, err := kubesphere.GenerateKubeSphereYaml(repo, "v2.1.1")
		if err != nil {
			return err
		}
		installerYaml = str
	default:
		return errors.New(fmt.Sprintf("Unsupported version: %s", strings.TrimSpace(version)))
	}

	b1 := bufio.NewReader(bytes.NewReader([]byte(installerYaml)))
	for {
		content, err := k8syaml.NewYAMLReader(b1).Read()
		if len(content) == 0 {
			break
		}
		if err != nil {
			return errors.Wrap(err, "Unable to read the manifests")
		}

		if len(strings.TrimSpace(string(content))) == 0 {
			continue
		}

		if err := DoServerSideApply(context.TODO(), restCfg, content); err != nil {
			return err
		}
	}

	if err := DoServerSideApply(context.TODO(), restCfg, []byte(kubesphereConfig)); err != nil {
		return err
	}

	return nil
}
