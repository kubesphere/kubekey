package rook

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/task"
)

type RookModule struct {
	common.KubeModule
	Skip bool
}

func (k *RookModule) IsSkip() bool {
	return k.Skip
}

func (k *RookModule) Init() {
	k.Name = "RookCrdsModule"
	k.Desc = "Install Rook CRDs"

	rookCrdsGenerateManifest := &task.RemoteTask{
		Name:     "GenerateRookCrdsManifest",
		Desc:     "Generate Rook CRDs manifest",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GenerateRookCrdsManifest),
		Parallel: false,
	}

	applyRookCrds := &task.RemoteTask{
		Name:     "DeployRookCrds",
		Desc:     "Deploy Rook CRDs",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeployRookCrds),
		Parallel: false,
		Retry:    5,
	}

	rookCommonGenerateManifest := &task.RemoteTask{
		Name:     "GenerateRookCommonManifest",
		Desc:     "Generate Rook common manifest",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GenerateRookCommonManifest),
		Parallel: false,
	}

	applyRookCommon := &task.RemoteTask{
		Name:     "DeployRookCommon",
		Desc:     "Deploy Rook common",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeployRookCommon),
		Parallel: false,
		Retry:    5,
	}

	rookOperatorGenerateManifest := &task.RemoteTask{
		Name:     "GenerateRookOperatorManifest",
		Desc:     "Generate Rook Operator manifest",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GenerateRookOperatorManifest),
		Parallel: false,
	}

	applyRookOperator := &task.RemoteTask{
		Name:     "DeployRookOperator",
		Desc:     "Deploy Rook Operator",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeployRookOperator),
		Parallel: false,
		Retry:    5,
	}

	rookClusterGenerateManifest := &task.RemoteTask{
		Name:     "GenerateRookClusterManifest",
		Desc:     "Generate Rook Cluster manifest",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GenerateRookClusterManifest),
		Parallel: false,
	}

	applyRookCluster := &task.RemoteTask{
		Name:     "DeployRookCluster",
		Desc:     "Deploy Rook Cluster",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeployRookCluster),
		Parallel: false,
		Retry:    5,
	}

	rookFilesystemGenerateManifest := &task.RemoteTask{
		Name:     "GenerateRookFilesystemManifest",
		Desc:     "Generate Rook Filesystem manifest",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(GenerateRookFilesysteManifest),
		Parallel: false,
	}

	applyRookFilesystem := &task.RemoteTask{
		Name:     "DeployRookFilesystem",
		Desc:     "Deploy Rook Filesystem",
		Hosts:    k.Runtime.GetHostsByRole(common.Master),
		Action:   new(DeployRookFilesystem),
		Parallel: false,
		Retry:    5,
	}

	k.Tasks = []task.Interface{
		rookCrdsGenerateManifest,
		applyRookCrds,
		rookCommonGenerateManifest,
		applyRookCommon,
		rookOperatorGenerateManifest,
		applyRookOperator,
		rookClusterGenerateManifest,
		applyRookCluster,
		rookFilesystemGenerateManifest,
		applyRookFilesystem,
	}

}
