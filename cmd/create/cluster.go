package create

import (
	"github.com/kubesphere/kubekey/pkg/install"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

func NewCmdCreateCluster() *cobra.Command {
	var (
		clusterCfgFile string
		addons         string
		pkgDir         string
		verbose        bool
	)
	var clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "Create Kubernetes Cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := util.InitLogger(verbose)
			return install.CreateCluster(clusterCfgFile, logger, addons, pkgDir, verbose)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "config", "f", "", "cluster info config")
	clusterCmd.Flags().StringVarP(&addons, "add", "", "", "add plugins")
	clusterCmd.Flags().StringVarP(&pkgDir, "pkg", "", "", "release package (offline)")
	clusterCmd.Flags().BoolVarP(&verbose, "debug", "", true, "debug info")
	return clusterCmd
}
