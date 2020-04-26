package cmd

import (
	"github.com/pixiake/kubekey/pkg/util"
	"github.com/pixiake/kubekey/scale"
	"github.com/spf13/cobra"
)

func NewCmdScaleCluster() *cobra.Command {
	var (
		clusterCfgFile string
		pkgDir         string
	)
	var clusterCmd = &cobra.Command{
		Use:   "scale",
		Short: "Scale cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := util.InitLogger(true)
			return scale.ScaleCluster(clusterCfgFile, logger, pkgDir)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "config", "f", "", "")
	clusterCmd.Flags().StringVarP(&pkgDir, "pkg", "", "", "")
	return clusterCmd
}
