package create

import (
	"github.com/kubesphere/kubekey/pkg/install"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

func NewCmdCreateCluster() *cobra.Command {
	var (
		clusterCfgFile string
		//addons         string
		//pkgDir         string
		verbose bool
		all     bool
	)
	var clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "Create a Kubernetes or KubeSphere cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := util.InitLogger(verbose)
			return install.CreateCluster(clusterCfgFile, logger, all, verbose)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "file", "f", "", "configuration file name")
	//clusterCmd.Flags().StringVarP(&pkgDir, "pkg", "", "", "release package (offline)")
	clusterCmd.Flags().BoolVarP(&verbose, "debug", "", true, "debug info")
	clusterCmd.Flags().BoolVarP(&all, "all", "", false, "deploy kubernetes and kubesphere")
	return clusterCmd
}
