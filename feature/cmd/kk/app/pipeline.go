package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

func newPipelineCommand() *cobra.Command {
	o := options.NewPipelineOptions()

	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Executor a pipeline in kubernetes",
		RunE: func(*cobra.Command, []string) error {
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
			// get pipeline
			var pipeline = new(kkcorev1.Pipeline)
			if err := client.Get(ctx, ctrlclient.ObjectKey{
				Name:      o.Name,
				Namespace: o.Namespace,
			}, pipeline); err != nil {
				return err
			}
			// get config
			var config = new(kkcorev1.Config)
			if err := client.Get(ctx, ctrlclient.ObjectKey{
				Name:      pipeline.Spec.ConfigRef.Name,
				Namespace: pipeline.Spec.ConfigRef.Namespace,
			}, config); err != nil {
				return err
			}
			// get inventory
			var inventory = new(kkcorev1.Inventory)
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
