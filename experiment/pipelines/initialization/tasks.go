package initialization

import (
	"fmt"
	"github.com/kubesphere/kubekey/experiment/core/action"
	"github.com/kubesphere/kubekey/experiment/core/logger"
	"github.com/kubesphere/kubekey/experiment/core/vars"
	"strings"
)

type NodePreCheck struct {
	action.BaseAction
}

func (n *NodePreCheck) Execute(vars vars.Vars) error {
	var results = make(map[string]string)
	results["name"] = n.Runtime.Runner.Host.Name
	for _, software := range baseSoftware {
		_, err := n.Runtime.Runner.SudoCmd(fmt.Sprintf("which %s", software), false)
		switch software {
		case showmount:
			software = nfs
		case rbd:
			software = ceph
		case glusterfs:
			software = glusterfs
		}
		if err != nil {
			results[software] = ""
			logger.Log.Warnf("exec cmd 'which %s' got err return: %v", software, err)
		} else {
			results[software] = "y"
			if software == docker {
				dockerVersion, err := n.Runtime.Runner.SudoCmd("docker version --format '{{.Server.Version}}'", false)
				if err != nil {
					results[software] = UnknownVersion
				} else {
					results[software] = dockerVersion
				}
			}
		}
	}

	output, err := n.Runtime.Runner.Cmd("date +\"%Z %H:%M:%S\"", false)
	if err != nil {
		results["time"] = ""
	} else {
		results["time"] = strings.TrimSpace(output)
	}

	checkResults := make(map[string]interface{})
	checkResults[n.Runtime.Runner.Host.Name] = results
	if res, ok := n.RootCache.GetOrSet("nodePreCheck", checkResults); ok {
		m := res.(map[string]interface{})
		m[n.Runtime.Runner.Host.Name] = results
		n.RootCache.Set("nodePreCheck", m)
	} else {
		n.RootCache.Set("nodePreCheck", checkResults)
	}
	//n.RootCache.Set(n.Runtime.Runner.Host.Name, results)
	return nil
}

type NodePreCheckConfirm struct {
	action.BaseAction
}
