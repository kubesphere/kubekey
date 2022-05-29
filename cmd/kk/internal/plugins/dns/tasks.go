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

package dns

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/cmd/kk/internal/common"
	"github.com/kubesphere/kubekey/cmd/kk/internal/plugins/dns/templates"
	"github.com/kubesphere/kubekey/util/workflow/action"
	"github.com/kubesphere/kubekey/util/workflow/connector"
	"github.com/kubesphere/kubekey/util/workflow/util"
)

type OverrideCoreDNS struct {
	common.KubeAction
}

func (o *OverrideCoreDNS) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl delete -n kube-system svc kube-dns", true); err != nil {
		if !strings.Contains(err.Error(), "NotFound") {
			return errors.Wrap(errors.WithStack(err), "delete kube-dns failed")
		}
	}

	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/coredns-svc.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "create coredns service failed")
	}
	return nil
}

type DeployNodeLocalDNS struct {
	common.KubeAction
}

func (d *DeployNodeLocalDNS) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/nodelocaldns.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy nodelocaldns failed")
	}
	return nil
}

type GenerateNodeLocalDNSConfigMap struct {
	common.KubeAction
}

func (g *GenerateNodeLocalDNSConfigMap) Execute(runtime connector.Runtime) error {
	clusterIP, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get svc -n kube-system coredns -o jsonpath='{.spec.clusterIP}'", false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get clusterIP failed")
	}

	if len(clusterIP) == 0 {
		clusterIP = g.KubeConf.Cluster.CorednsClusterIP()
	}

	templateAction := action.Template{
		Template: templates.NodeLocalDNSConfigMap,
		Dst:      filepath.Join(common.KubeConfigDir, templates.NodeLocalDNSConfigMap.Name()),
		Data: util.Data{
			"ForwardTarget": clusterIP,
			"DNSDomain":     g.KubeConf.Cluster.Kubernetes.DNSDomain,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type ApplyNodeLocalDNSConfigMap struct {
	common.KubeAction
}

func (a *ApplyNodeLocalDNSConfigMap) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/nodelocaldnsConfigmap.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "apply nodelocaldns configmap failed")
	}
	return nil
}
