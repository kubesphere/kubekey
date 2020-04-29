package cmd

import (
	"github.com/pixiake/kubekey/pkg/scale"
	"github.com/pixiake/kubekey/pkg/util"
	"github.com/spf13/cobra"
)

func NewCmdScaleCluster() *cobra.Command {
	var (
		clusterCfgFile string
		pkgDir         string
		Verbose        bool
	)
	var clusterCmd = &cobra.Command{
		Use:   "scale",
		Short: "Scale cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := util.InitLogger(Verbose)
			return scale.ScaleCluster(clusterCfgFile, logger, pkgDir, Verbose)
		},
	}

	clusterCmd.Flags().StringVarP(&clusterCfgFile, "config", "f", "", "")
	clusterCmd.Flags().StringVarP(&pkgDir, "pkg", "", "", "")
	clusterCmd.Flags().BoolVarP(&Verbose, "debug", "", true, "")
	return clusterCmd
}
