/*
 Copyright 2022 The KubeSphere Authors.

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

package k8e

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v2/pkg/common"
	"github.com/kubesphere/kubekey/v2/pkg/core/connector"
)

type K8eStatus struct {
	Version     string
	ClusterInfo string
	NodeToken   string
	KubeConfig  string
	NodesInfo   map[string]string
}

func NewK8eStatus() *K8eStatus {
	return &K8eStatus{NodesInfo: make(map[string]string)}
}

func (k *K8eStatus) SearchVersion(runtime connector.Runtime) error {
	cmd := "/usr/local/bin/k8e --version | grep 'k8e' | awk '{print $3}'"
	if output, err := runtime.GetRunner().Cmd(cmd, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "search current version failed")
	} else {
		k.Version = output
	}
	return nil
}

func (k *K8eStatus) SearchKubeConfig(runtime connector.Runtime) error {
	kubeCfgCmd := "cat /etc/k8e/k8e.yaml"
	if kubeConfigStr, err := runtime.GetRunner().SudoCmd(kubeCfgCmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "search cluster kubeconfig failed")
	} else {
		k.KubeConfig = kubeConfigStr
	}
	return nil
}

func (k *K8eStatus) SearchNodeToken(runtime connector.Runtime) error {
	nodeTokenBase64Cmd := "cat /var/lib/k8e/server/node-token"
	output, err := runtime.GetRunner().SudoCmd(nodeTokenBase64Cmd, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get cluster node token failed")
	}
	k.NodeToken = output
	return nil
}

func (k *K8eStatus) SearchInfo(runtime connector.Runtime) error {
	output, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl --no-headers=true get nodes -o custom-columns=:metadata.name,:status.nodeInfo.kubeletVersion,:status.addresses",
		true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get K8e cluster info failed")
	}
	k.ClusterInfo = output
	return nil
}

func (k *K8eStatus) SearchNodesInfo(_ connector.Runtime) error {
	ipv4Regexp, err := regexp.Compile(common.IPv4Regexp)
	if err != nil {
		return err
	}
	ipv6Regexp, err := regexp.Compile(common.IPv6Regexp)
	if err != nil {
		return err
	}
	tmp := strings.Split(k.ClusterInfo, "\r\n")
	if len(tmp) >= 1 {
		for i := 0; i < len(tmp); i++ {
			if ipv4 := ipv4Regexp.FindStringSubmatch(tmp[i]); len(ipv4) != 0 {
				k.NodesInfo[ipv4[0]] = ipv4[0]
			}
			if ipv6 := ipv6Regexp.FindStringSubmatch(tmp[i]); len(ipv6) != 0 {
				k.NodesInfo[ipv6[0]] = ipv6[0]
			}
			if len(strings.Fields(tmp[i])) > 3 {
				k.NodesInfo[strings.Fields(tmp[i])[0]] = strings.Fields(tmp[i])[1]
			} else {
				k.NodesInfo[strings.Fields(tmp[i])[0]] = ""
			}
		}
	}
	return nil
}

func (k *K8eStatus) LoadKubeConfig(runtime connector.Runtime, kubeConf *common.KubeConf) error {
	kubeConfigPath := filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.GetObjName()))

	oldServer := "server: https://127.0.0.1:6443"
	newServer := fmt.Sprintf("server: https://%s:%d", kubeConf.Cluster.ControlPlaneEndpoint.Address, kubeConf.Cluster.ControlPlaneEndpoint.Port)
	newKubeConfigStr := strings.Replace(k.KubeConfig, oldServer, newServer, -1)

	if err := ioutil.WriteFile(kubeConfigPath, []byte(newKubeConfigStr), 0644); err != nil {
		return err
	}
	return nil
}
