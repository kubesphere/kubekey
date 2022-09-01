package rook

import (
	"path/filepath"

	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	cluster "github.com/kubesphere/kubekey/pkg/rook/templates/cluster"
	filesystem "github.com/kubesphere/kubekey/pkg/rook/templates/filesystem"
	prereqs "github.com/kubesphere/kubekey/pkg/rook/templates/prereqs"
	"github.com/pkg/errors"
)

type GenerateRookCrdsManifest struct {
	common.KubeAction
}

func (g *GenerateRookCrdsManifest) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: prereqs.RookCrds,
		Dst:      filepath.Join(common.KubeManifestDir, prereqs.RookCrds.Name()),
		Data:     util.Data{},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeployRookCrds struct {
	common.KubeAction
}

func (d *DeployRookCrds) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/manifests/crds.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy rook crds failed")
	}
	return nil
}

type GenerateRookCommonManifest struct {
	common.KubeAction
}

func (g *GenerateRookCommonManifest) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: prereqs.RookCommon,
		Dst:      filepath.Join(common.KubeManifestDir, prereqs.RookCommon.Name()),
		Data:     util.Data{},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeployRookCommon struct {
	common.KubeAction
}

func (d *DeployRookCommon) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/manifests/common.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy rook common failed")
	}
	return nil
}

type GenerateRookOperatorManifest struct {
	common.KubeAction
}

func (g *GenerateRookOperatorManifest) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: prereqs.RookOperator,
		Dst:      filepath.Join(common.KubeManifestDir, prereqs.RookOperator.Name()),
		Data:     util.Data{},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeployRookOperator struct {
	common.KubeAction
}

func (d *DeployRookOperator) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/manifests/operator.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy rook operator failed")
	}
	return nil
}

type GenerateRookClusterManifest struct {
	common.KubeAction
}

func (g *GenerateRookClusterManifest) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: cluster.RookCluster,
		Dst:      filepath.Join(common.KubeManifestDir, cluster.RookCluster.Name()),
		Data:     util.Data{},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeployRookCluster struct {
	common.KubeAction
}

func (d *DeployRookCluster) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/manifests/multi-node.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy rook cluster failed")
	}
	return nil
}

type GenerateRookFilesysteManifest struct {
	common.KubeAction
}

func (g *GenerateRookFilesysteManifest) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: filesystem.RookFilesystem,
		Dst:      filepath.Join(common.KubeManifestDir, filesystem.RookFilesystem.Name()),
		Data:     util.Data{},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeployRookFilesystem struct {
	common.KubeAction
}

func (d *DeployRookFilesystem) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/manifests/multi-osd.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy rook filesystem failed")
	}
	return nil
}
