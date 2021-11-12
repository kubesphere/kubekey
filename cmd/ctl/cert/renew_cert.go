package cert

import (
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/kubesphere/kubekey/cmd/ctl/util"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/pipelines"
	"github.com/spf13/cobra"
)

type CertRenewOptions struct {
	CommonOptions  *options.CommonOptions
	ClusterCfgFile string
}

func NewCertRenewOptions() *CertRenewOptions {
	return &CertRenewOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCertRenew creates a new cert renew command
func NewCmdCertRenew() *cobra.Command {
	o := NewCertRenewOptions()
	cmd := &cobra.Command{
		Use:   "renew",
		Short: "renew a cluster certs",
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run())
		},
	}

	o.CommonOptions.AddCommonFlag(cmd)
	o.AddFlags(cmd)
	return cmd
}

func (o *CertRenewOptions) Run() error {
	arg := common.Argument{
		FilePath: o.ClusterCfgFile,
		Debug:    o.CommonOptions.Verbose,
	}
	return pipelines.RenewCerts(arg)
}

func (o *CertRenewOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.ClusterCfgFile, "filename", "f", "", "Path to a configuration file")
}
