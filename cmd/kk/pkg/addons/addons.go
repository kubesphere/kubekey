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

package addons

import (
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"

	"github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
)

func InstallAddons(kubeConf *common.KubeConf, addon *kubekeyapiv1alpha2.Addon, kubeConfig string) error {
	// install chart
	if addon.Sources.Chart.Name != "" {
		_ = os.Setenv("HELM_NAMESPACE", strings.TrimSpace(addon.Namespace))
		if err := InstallChart(kubeConf, addon, kubeConfig); err != nil {
			return err
		}
	}

	// install yaml
	if len(addon.Sources.Yaml.Path) != 0 {
		if err := render(addon.Sources.Yaml); err != nil {
			return errors.Wrapf(err, "tamplate addon %s yaml failed", addon.Name)
		}
		var settings = cli.New()
		p := getter.All(settings)
		for _, yaml := range addon.Sources.Yaml.Path {
			u, _ := url.Parse(yaml)
			_, err := p.ByScheme(u.Scheme)
			if err != nil {
				fp, err := filepath.Abs(yaml)
				if err != nil {
					return errors.Wrap(err, "Failed to look up current directory")
				}
				yamlPaths := []string{fp}
				if err := InstallYaml(yamlPaths, addon.Namespace, kubeConfig, kubeConf.Cluster.Kubernetes.Version); err != nil {
					return err
				}
			} else {
				yamlPaths := []string{yaml}
				if err := InstallYaml(yamlPaths, addon.Namespace, kubeConfig, kubeConf.Cluster.Kubernetes.Version); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func render(in v1alpha2.Yaml) error {
	if len(in.Values) == 0 {
		return nil
	}
	for _, yamlFile := range in.Path {
		// backup
		yamlTpl := yamlFile + ".tpl"
		if _, err := os.Stat(yamlTpl); err != nil {
			logger.Log.Infof("backup yaml: %s", yamlFile)
			if err := copyFile(yamlFile, yamlTpl); err != nil {
				return errors.Wrapf(err, "backup: %s failed", yamlFile)
			}
		}
		// render
		tpl, err := template.New(path.Base(yamlTpl)).ParseFiles(yamlTpl)
		if err != nil {
			return errors.Wrapf(err, "parse %s failed", yamlTpl)
		}

		f, err := os.OpenFile(yamlFile, os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := tpl.Execute(f, in.Values); err != nil {
			return errors.Wrapf(err, "render yaml %s failed", yamlFile)
		}
		logger.Log.Infof("render yaml: %s", yamlFile)
	}
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
