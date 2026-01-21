package app

import (
	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"path/filepath"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ApiServerCommand creates a new cobra command for starting the KubeKey api server
// It initializes the web server with the provided configuration options and starts
// the HTTP server with web UI interface
func ApiServerCommand() *cobra.Command {
	o := options.NewKubeKeyWebOptions()

	cmd := &cobra.Command{
		Use:   "api-server",
		Short: "start a http api server for web installer.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize REST config for Kubernetes client
			restconfig := &rest.Config{}
			if err := proxy.RestConfig(filepath.Join(o.Workdir, _const.RuntimeDir), restconfig); err != nil {
				return err
			}

			// Create Kubernetes client with the REST config
			client, err := ctrlclient.New(restconfig, ctrlclient.Options{
				Scheme: _const.Scheme,
			})
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			mgr := manager.NewApiManager(manager.ApiManagerOptions{
				Workdir:    o.Workdir,
				SchemaPath: o.SchemaPath,
				Port:       o.Port,
				Client:     client,
				Config:     restconfig,
			})

			if err = mgr.PrepareRun(ctx.Done()); err != nil {
				return err
			}

			// Initialize and run the web manager with provided options
			return mgr.Run(ctx)
		},
	}
	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}

	return cmd
}
