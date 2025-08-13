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

// newWebCommand creates a new cobra command for starting the KubeKey web server
// It initializes the web server with the provided configuration options and starts
// the HTTP server with web UI interface
func newWebCommand() *cobra.Command {
	o := options.NewKubeKeyWebOptions()

	cmd := &cobra.Command{
		Use:   "web",
		Short: "start a http server with web UI.",
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

			// Initialize and run the web manager with provided options
			return manager.NewWebManager(manager.WebManagerOptions{
				Workdir:    o.Workdir,
				Port:       o.Port,
				SchemaPath: o.SchemaPath,
				UIPath:     o.UIPath,
				Client:     client,
				Config:     restconfig,
			}).Run(cmd.Context())
		},
	}
	for _, f := range o.Flags().FlagSets {
		cmd.Flags().AddFlagSet(f)
	}

	return cmd
}
