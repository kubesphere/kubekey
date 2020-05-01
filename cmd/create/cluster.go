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
		Verbose        bool
	)
	var clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "Create Kubernetes Cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := util.InitLogger(Verbose)
			return install.CreateCluster(clusterCfgFile, logger, addons, pkgDir, Verbose)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "config", "f", "", "cluster info config")
	clusterCmd.Flags().StringVarP(&addons, "add", "", "", "add plugins")
	clusterCmd.Flags().StringVarP(&pkgDir, "pkg", "", "", "release package (offline)")
	clusterCmd.Flags().BoolVarP(&Verbose, "debug", "", true, "debug info")
	return clusterCmd
}
