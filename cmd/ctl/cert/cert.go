package cert

import (
	"github.com/kubesphere/kubekey/cmd/ctl/options"
	"github.com/spf13/cobra"
)

type CertOptions struct {
	CommonOptions *options.CommonOptions
}

func NewCertsOptions() *CertOptions {
	return &CertOptions{
		CommonOptions: options.NewCommonOptions(),
	}
}

// NewCmdCerts creates a new cert command
func NewCmdCerts() *cobra.Command {
	o := NewCertsOptions()
	cmd := &cobra.Command{
		Use:   "certs",
		Short: "cluster certs",
	}

	o.CommonOptions.AddCommonFlag(cmd)

	cmd.AddCommand(NewCmdCertList())
	cmd.AddCommand(NewCmdCertRenew())
	return cmd
}
