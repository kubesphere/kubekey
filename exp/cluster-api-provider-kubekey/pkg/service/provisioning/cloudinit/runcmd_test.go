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

package cloudinit

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/provisioning/commands"
)

func TestRunCmdUnmarshal(t *testing.T) {
	g := NewWithT(t)

	cloudData := `
runcmd:
- [ ls, -l, / ]
- "ls -l /"`
	r := runCmd{}
	err := r.Unmarshal([]byte(cloudData))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(r.Cmds).To(HaveLen(2))

	expected0 := commands.Cmd{Cmd: "ls", Args: []string{"-l", "/"}}
	g.Expect(r.Cmds[0]).To(Equal(expected0))

	expected1 := commands.Cmd{Cmd: "/bin/sh", Args: []string{"-c", "ls -l /"}}
	g.Expect(r.Cmds[1]).To(Equal(expected1))
}

func TestHackKubeadmIgnoreErrors(t *testing.T) {
	g := NewWithT(t)

	cloudData := `
runcmd:
- kubeadm init --config=/run/kubeadm/kubeadm.yaml
- [ kubeadm, join, --config=/run/kubeadm/kubeadm-controlplane-join-config.yaml ]
- kubeadm init --ignore-preflight-errors=all --config=/run/kubeadm/kubeadm.yaml
- [ kubeadm, join, --ignore-preflight-errors=all, --config=/run/kubeadm/kubeadm-controlplane-join-config.yaml ]`

	r := runCmd{}
	err := r.Unmarshal([]byte(cloudData))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(r.Cmds).To(HaveLen(4))

	r.Cmds[0] = hackKubeadmIgnoreErrors(r.Cmds[0])
	expected0 := commands.Cmd{Cmd: "/bin/sh", Args: []string{"-c", "kubeadm init --ignore-preflight-errors=all --config=/run/kubeadm/kubeadm.yaml"}}
	g.Expect(r.Cmds[0]).To(Equal(expected0))

	r.Cmds[1] = hackKubeadmIgnoreErrors(r.Cmds[1])
	expected1 := commands.Cmd{Cmd: "kubeadm", Args: []string{"join", "--ignore-preflight-errors=all", "--config=/run/kubeadm/kubeadm-controlplane-join-config.yaml"}}
	g.Expect(r.Cmds[1]).To(Equal(expected1))

	r.Cmds[2] = hackKubeadmIgnoreErrors(r.Cmds[2])
	expected2 := commands.Cmd{Cmd: "/bin/sh", Args: []string{"-c", "kubeadm init --ignore-preflight-errors=all --config=/run/kubeadm/kubeadm.yaml"}}
	g.Expect(r.Cmds[0]).To(Equal(expected2))

	r.Cmds[3] = hackKubeadmIgnoreErrors(r.Cmds[3])
	expected3 := commands.Cmd{Cmd: "kubeadm", Args: []string{"join", "--ignore-preflight-errors=all", "--config=/run/kubeadm/kubeadm-controlplane-join-config.yaml"}}
	g.Expect(r.Cmds[1]).To(Equal(expected3))
}
