package kubeark

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/kubeark/templates"
	"github.com/pkg/errors"
	"path/filepath"
)

type GenerateKubearkManifest struct {
	common.KubeAction
}

func (g *GenerateKubearkManifest) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: templates.KubearkManifest,
		Dst:      filepath.Join(common.KubeManifestDir, templates.KubearkManifest.Name()),
		Data:     util.Data{},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeployKubeark struct {
	common.KubeAction
}

func (d *DeployKubeark) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/manifests/kubeark.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy Kubeark failed")
	}
	return nil
}
