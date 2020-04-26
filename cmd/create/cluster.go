package create

import (
	"github.com/pixiake/kubekey/pkg/install"
	"github.com/pixiake/kubekey/pkg/util"
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
			return install.CreateCluster(clusterCfgFile, logger, addons, pkgDir)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "config", "f", "", "")
	clusterCmd.Flags().StringVarP(&addons, "add", "", "", "")
	clusterCmd.Flags().StringVarP(&pkgDir, "pkg", "", "", "")
	return clusterCmd
}
