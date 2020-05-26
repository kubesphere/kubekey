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

package create

import (
	"github.com/kubesphere/kubekey/pkg/config"
	"github.com/kubesphere/kubekey/pkg/util"
	log "github.com/sirupsen/logrus"
	//"github.com/kubesphere/kubekey/cluster"
	//"github.com/kubesphere/kubekey/cluster/etcd"
	"github.com/spf13/cobra"
)

func NewCmdCreateEtcd() *cobra.Command {
	var (
		clusterCfgFile string
	)
	var clusterCmd = &cobra.Command{
		Use:   "etcd",
		Short: "Manage Etcd cluster",
		Run: func(cmd *cobra.Command, args []string) {
			logger := util.InitLogger(true)
			createEtcdCluster(clusterCfgFile, logger)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "cluster-info", "", "", "")
	return clusterCmd
}

func createEtcdCluster(clusterCfgFile string, logger *log.Logger) {
	config.ParseClusterCfg(clusterCfgFile, false, logger)
}
