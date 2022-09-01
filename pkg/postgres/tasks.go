package postgres

import (
	"path/filepath"

	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	cluster "github.com/kubesphere/kubekey/pkg/postgres/templates/cluster"
	prereqs "github.com/kubesphere/kubekey/pkg/postgres/templates/prereqs"
	"github.com/pkg/errors"
)

type GeneratePostgresPrereqsManifest struct {
	common.KubeAction
}

func (g *GeneratePostgresPrereqsManifest) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: prereqs.PostgresPrereqs,
		Dst:      filepath.Join(common.KubeManifestDir, prereqs.PostgresPrereqs.Name()),
		Data:     util.Data{},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeployPostgresPrereqs struct {
	common.KubeAction
}

func (d *DeployPostgresPrereqs) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/manifests/postgres-prereqs.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy postgres prereqs failed")
	}
	return nil
}

type GeneratePostgresClusterManifest struct {
	common.KubeAction
}

func (g *GeneratePostgresClusterManifest) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: cluster.PostgresCluster,
		Dst:      filepath.Join(common.KubeManifestDir, cluster.PostgresCluster.Name()),
		Data:     util.Data{},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeployPostgresCluster struct {
	common.KubeAction
}

func (d *DeployPostgresCluster) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/manifests/postgres-cluster.yaml", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy postgres cluster failed")
	}
	return nil
}
