package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

func newPipelineCommand() *cobra.Command {
	o := options.NewPipelineOption()

	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Executor a pipeline in kubernetes",
		RunE: func(cmd *cobra.Command, args []string) error {
			_const.SetWorkDir(o.WorkDir)
			restconfig, err := ctrl.GetConfig()
			if err != nil {
				klog.Infof("kubeconfig in empty, store resources local")
				restconfig = &rest.Config{}
			}
			restconfig, err = proxy.NewConfig(restconfig)
			if err != nil {
				return fmt.Errorf("could not get rest config: %w", err)
			}
			client, err := ctrlclient.New(restconfig, ctrlclient.Options{
				Scheme: _const.Scheme,
			})
			if err != nil {
				return fmt.Errorf("could not create client: %w", err)
			}
			ctx := signals.SetupSignalHandler()
			var pipeline = new(kubekeyv1.Pipeline)
			var config = new(kubekeyv1.Config)
			var inventory = new(kubekeyv1.Inventory)
			if err := client.Get(ctx, ctrlclient.ObjectKey{
				Name:      o.Name,
				Namespace: o.Namespace,
			}, pipeline); err != nil {
				return err
			}
			if err := client.Get(ctx, ctrlclient.ObjectKey{
				Name:      pipeline.Spec.ConfigRef.Name,
				Namespace: pipeline.Spec.ConfigRef.Namespace,
			}, config); err != nil {
				return err
			}
			if err := client.Get(ctx, ctrlclient.ObjectKey{
				Name:      pipeline.Spec.InventoryRef.Name,
				Namespace: pipeline.Spec.InventoryRef.Namespace,
			}, inventory); err != nil {
				return err
			}

			return manager.NewCommandManager(manager.CommandManagerOptions{
				Pipeline:  pipeline,
				Config:    config,
				Inventory: inventory,
				Client:    client,
			}).Run(ctx)
		},
	}

	fs := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		fs.AddFlagSet(f)
	}
	return cmd
}
