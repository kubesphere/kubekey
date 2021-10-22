package dns

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/plugins/dns/templates"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
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
			"DndDomain":     g.KubeConf.Cluster.Kubernetes.ClusterName,
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
