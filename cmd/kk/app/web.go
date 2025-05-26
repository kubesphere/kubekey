package app

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

func newWebCommand() *cobra.Command {
	o := options.NewKubeKeyWebOptions()

	cmd := &cobra.Command{
		Use:   "web",
		Short: "start a http server with web UI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			restconfig := &rest.Config{}
			if err := proxy.RestConfig(filepath.Join(o.Workdir, _const.RuntimeDir), restconfig); err != nil {
				return err
			}

			client, err := ctrlclient.New(restconfig, ctrlclient.Options{
				Scheme: _const.Scheme,
			})
			if err != nil {
				return err
			}
			return manager.NewWebManager(manager.WebManagerOptions{
				Workdir: o.Workdir,
				Port:    o.Port,
				Client:  client,
				Config:  restconfig,
			}).Run(cmd.Context())
		},
	}
	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}

	return cmd
}
