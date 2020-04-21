package create

import (
	"fmt"
	"github.com/pixiake/kubekey/install"
	"github.com/pixiake/kubekey/pkg/cluster/preinstall"
	"github.com/pixiake/kubekey/pkg/config"
	"github.com/pixiake/kubekey/pkg/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCmdCreateCluster() *cobra.Command {
	var (
		clusterCfgFile string
		addons         string
		pkgDir         string
	)
	var clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "Create Kubernetes Cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := util.InitLogger(true)
			return CreateCluster(clusterCfgFile, logger, addons, pkgDir)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "cluster-info", "", "", "")
	clusterCmd.Flags().StringVarP(&addons, "add", "", "", "")
	clusterCmd.Flags().StringVarP(&pkgDir, "pkg", "", "", "")
	return clusterCmd
}

func CreateCluster(clusterCfgFile string, logger *log.Logger, addons string, pkg string) error {
	cfg, err := config.ParseK2ClusterObj(clusterCfgFile, logger)
	if err != nil {
		return errors.Wrap(err, "failed to download cluster config")
	}
	fmt.Println(cfg)
	config := config.SetDefaultClusterCfg(&cfg.Spec)
	if err := preinstall.Prepare(config, logger); err != nil {
		return errors.Wrap(err, "failed to load kube binarys")
	}
	return install.NewExecutor(config, logger).Execute()
}
