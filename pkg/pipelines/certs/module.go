package certs

import (
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
)

type CheckCertsModule struct {
	common.KubeModule
}

func (c *CheckCertsModule) Init() {
	c.Name = "CheckCertsModule"

	//check := &modules.Task{
	//	Name: "CheckClusterCerts",
	//	Desc: "check cluster certs",
	//	Hosts: c.Runtime.GetHostsByRole(common.Master),
	//	Action: ,
	//	Parallel: true,
	//}
}
