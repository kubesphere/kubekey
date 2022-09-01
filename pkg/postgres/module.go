package postgres

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
)

type PostgresModule struct {
	common.KubeModule
	Skip bool
}

func (k *PostgresModule) IsSkip() bool {
	return k.Skip
}

func (k *PostgresModule) Init() {
	k.Name = "PostgresPrereqsModule"
	k.Desc = "Install Postgres Prereqs"

	postgresPrereqsGenerateManifest := &task.RemoteTask{
		Name:     "GeneratePostgresPrereqsManifest",
		Desc:     "Generate Postgres prereqs manifest",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GeneratePostgresPrereqsManifest),
		Parallel: false,
	}

	applyPostgresPrereqs := &task.RemoteTask{
		Name:     "DeployPostgresPrereqs",
		Desc:     "Deploy Postgres prereqs",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeployPostgresPrereqs),
		Parallel: false,
		Retry:    5,
	}

	postgresClusterGenerateManifest := &task.RemoteTask{
		Name:     "GeneratePostgresClusterManifest",
		Desc:     "Generate Postgres cluster manifest",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GeneratePostgresClusterManifest),
		Parallel: false,
	}

	applyPostgresCluster := &task.RemoteTask{
		Name:     "DeployPostgresCluster",
		Desc:     "Deploy Postgres cluster",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeployPostgresCluster),
		Parallel: false,
		Retry:    5,
	}

	k.Tasks = []task.Interface{
		postgresPrereqsGenerateManifest,
		applyPostgresPrereqs,
		postgresClusterGenerateManifest,
		applyPostgresCluster,
	}

}
