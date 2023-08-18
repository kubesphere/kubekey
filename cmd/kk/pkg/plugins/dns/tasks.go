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
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/images"
	"github.com/pkg/errors"
	"path/filepath"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/plugins/dns/templates"
)

type GenerateCorednsmanifests struct {
	common.KubeAction
}

func (g *GenerateCorednsmanifests) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: templates.Coredns,
		Dst:      filepath.Join(common.KubeConfigDir, templates.Coredns.Name()),
		Data: util.Data{
			"ClusterIP":    g.KubeConf.Cluster.CorednsClusterIP(),
			"CorednsImage": images.GetImage(runtime, g.KubeConf, "coredns").ImageName(),
			"DNSEtcHosts":  g.KubeConf.Cluster.DNS.DNSEtcHosts,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeployCoreDNS struct {
	common.KubeAction
}

func (o *DeployCoreDNS) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl delete svc -n kube-system --field-selector metadata.name=kube-dns", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "remove old coredns svc")
	}
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/coredns.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "update coredns failed")
	}
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl -n kube-system rollout restart deploy coredns", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "restart coredns failed")
	}
	return nil
}

type ApplyCorednsConfigMap struct {
	common.KubeAction
}

func (o *ApplyCorednsConfigMap) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/coredns-configmap.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "create coredns configmap failed")
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
			"ExternalZones": g.KubeConf.Cluster.DNS.NodeLocalDNS.ExternalZones,
			"DNSEtcHosts":   g.KubeConf.Cluster.DNS.DNSEtcHosts,
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
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/nodelocaldns-configmap.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "apply nodelocaldns configmap failed")
	}
	return nil
}
