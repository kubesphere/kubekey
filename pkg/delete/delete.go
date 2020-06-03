/*
Copyright 2020 The KubeSphere Authors.

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

package delete

import (
	"bufio"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/config"
	"github.com/kubesphere/kubekey/pkg/util/executor"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func ResetCluster(clusterCfgFile string, logger *log.Logger, verbose bool) error {
	cfg, err := config.ParseClusterCfg(clusterCfgFile, false, logger)
	if err != nil {
		return errors.Wrap(err, "Failed to download cluster config")
	}

	return Execute(executor.NewExecutor(&cfg.Spec, logger, verbose))
}

func Execute(executor *executor.Executor) error {
	mgr, err := executor.CreateManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
}

func ExecTasks(mgr *manager.Manager) error {
	resetTasks := []manager.Task{
		{Task: ResetKubeCluster, ErrMsg: "Failed to reset kube cluster"},
	}

	for _, step := range resetTasks {
		if err := step.Run(mgr); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}

	mgr.Logger.Infoln("Successful.")

	return nil
}

func ResetKubeCluster(mgr *manager.Manager) error {
	reader := bufio.NewReader(os.Stdin)
	input, err := Confirm(reader)
	if err != nil {
		return err
	}
	if input == "no" {
		os.Exit(0)
	}

	mgr.Logger.Infoln("Resetting kubernetes cluster ...")

	return mgr.RunTaskOnK8sNodes(resetKubeCluster, true)
}

var clusterFiles = []string{
	"/usr/local/bin/etcd",
	"/etc/ssl/etcd",
	"/var/lib/etcd",
	"/etc/etcd.env",
	"/etc/systemd/system/etcd.service",
	"/var/log/calico",
	"/etc/cni",
	"/var/log/pods/",
	"/var/lib/cni",
	"/var/lib/calico",
	"/run/calico",
	"/run/flannel",
	"/etc/flannel",
}

var cmdsList = []string{
	"iptables -F",
	"iptables -X",
	"iptables -F -t nat",
	"iptables -X -t nat",
	"ip link del kube-ipvs0",
	"ip link del nodelocaldns",
}

func resetKubeCluster(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	_, err := mgr.Runner.RunCmdOutput("sudo -E /bin/sh -c \"/usr/local/bin/kubeadm reset -f\"")
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to reset kube cluster")
	}
	fmt.Println(strings.Join(cmdsList, " && "))
	mgr.Runner.RunCmdOutput(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", strings.Join(cmdsList, " && ")))
	deleteFiles(mgr)
	return nil
}

func deleteFiles(mgr *manager.Manager) error {
	mgr.Runner.RunCmdOutput("sudo -E /bin/sh -c \"systemctl stop etcd && exit 0\"")
	for _, file := range clusterFiles {
		mgr.Runner.RunCmdOutput(fmt.Sprintf("sudo -E /bin/sh -c \"rm -rf %s\"", file))
	}
	return nil
}

func Confirm(reader *bufio.Reader) (string, error) {
	for {
		fmt.Printf("Are you sure to delete this cluster? [yes/no]: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)

		if input != "" && (input == "yes" || input == "no") {
			return input, nil
		}
	}
}
