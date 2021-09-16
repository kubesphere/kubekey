package config

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"unicode"
)

type ModifyConfigModule struct {
	common.KubeModule
}

func (m *ModifyConfigModule) Init() {
	m.Name = "ModifyConfigModule"
	m.Desc = "Modify the KubeKey config file"
}

func (m *ModifyConfigModule) Run() error {
	fp, _ := filepath.Abs(m.KubeConf.Arg.FilePath)
	cmd0 := fmt.Sprintf("cat %s | grep %s | wc -l", fp, m.KubeConf.Arg.NodeName)
	nodeNameNum, err0 := exec.Command("/bin/sh", "-c", cmd0).CombinedOutput()
	if err0 != nil {
		return errors.Wrap(err0, "Failed to get node num")
	}

	host := &connector.BaseHost{Name: m.KubeConf.Arg.NodeName}
	if string(nodeNameNum) == "2\n" {
		cmd := fmt.Sprintf("sed -i /%s/d %s", m.KubeConf.Arg.NodeName, fp)
		_ = exec.Command("/bin/sh", "-c", cmd).Run()
		m.Runtime.DeleteHost(host)
	} else if string(nodeNameNum) == "1\n" {
		cmd := fmt.Sprintf("sed -i /%s/d %s", m.KubeConf.Arg.NodeName, fp)
		_ = exec.Command("/bin/sh", "-c", cmd).Run()
		m.Runtime.DeleteHost(host)

		var newNodeName []string
		for i := 0; i < len(m.Runtime.GetHostsByRole(common.Worker)); i++ {
			name := m.Runtime.GetHostsByRole(common.Worker)[i].GetName()
			if m.KubeConf.Arg.NodeName == name {
				continue
			} else {
				newNodeName = append(newNodeName, name)
			}
		}
		var connNodeName []string
		for j := 0; j < len(newNodeName); j++ {
			t := j
			nodename1 := newNodeName[t]
			for t+1 < len(newNodeName) && IsAdjoin(newNodeName[t], newNodeName[t+1]) {
				t++
			}
			if t == j {
				connNodeName = append(connNodeName, nodename1)
			} else {
				connNodeName = append(connNodeName, Merge(nodename1, newNodeName[t]))
				j = t
			}
		}
		cmd1 := fmt.Sprintf("sed -i -n '1,/worker/p;/controlPlaneEndpoint/,$p' %s", fp)
		_ = exec.Command("/bin/sh", "-c", cmd1).Run()
		for k := 0; k < len(connNodeName); k++ {
			workPar := connNodeName[k]
			workPar1 := fmt.Sprintf("%s", workPar)
			cmd2 := fmt.Sprintf("sed -i '/worker/a\\ \\ \\ \\ \\- %s' %s", workPar1, fp)
			_ = exec.Command("/bin/sh", "-c", cmd2).Run()
		}
	} else {
		fmt.Println("Please check the node name in the config-sample.yaml or only support to delete worker")
		os.Exit(0)
	}
	return nil
}

func Merge(name1, name2 string) (endName string) {
	par1, par2 := SplitNum(name1)
	_, par4 := SplitNum(name2)
	endName = fmt.Sprintf("%s[%s:%s]", par1, strconv.Itoa(par2), strconv.Itoa(par4))
	return endName
}
func IsAdjoin(name1, name2 string) bool {
	IsAd := false
	par1, par2 := SplitNum(name1)
	par3, par4 := SplitNum(name2)
	if par1 == par3 && par4 == par2+1 {
		IsAd = true
	}
	return IsAd
}

func SplitNum(nodeName string) (name string, num int) {
	nodeLen := len(nodeName)
	i := nodeLen - 1
	for ; nodeLen > 0; i-- {
		if !unicode.IsDigit(rune(nodeName[i])) {
			num, _ := strconv.Atoi(nodeName[i+1:])
			name := nodeName[:i+1]
			return name, num
		}
	}
	return "", 0
}
