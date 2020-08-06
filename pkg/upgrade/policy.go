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

package upgrade

import (
	"bufio"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"os"
	"strings"
)

func GetClusterInfo(mgr *manager.Manager) error {
	return mgr.RunTaskOnMasterNodes(getClusterInfo, true)
}

func getClusterInfo(mgr *manager.Manager, node *kubekeyapi.HostCfg) error {
	if mgr.Runner.Index == 0 {
		componentstatus, err := mgr.Runner.ExecuteCmd("sudo -E /bin/bash -c \"/usr/local/bin/kubectl get componentstatus\"", 2, false)
		if err != nil {
			return err
		}
		fmt.Println("Cluster components status:")
		fmt.Println(componentstatus + "\n")
		nodestatus, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"/usr/local/bin/kubectl get node -o wide\"", 2, false)
		if err != nil {
			return err
		}
		fmt.Println("Cluster nodes status:")
		fmt.Println(nodestatus + "\n\n")

		reader := bufio.NewReader(os.Stdin)
	Loop:
		for {
			fmt.Printf("Continue upgrading cluster? [yes/no]: ")
			input, err := reader.ReadString('\n')
			if err != nil {
				mgr.Logger.Fatal(err)
			}
			input = strings.TrimSpace(input)

			switch input {
			case "yes":
				break Loop
			case "no":
				os.Exit(0)
			default:
				continue
			}
		}
	}
	return nil
}
