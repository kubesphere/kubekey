package app

import (
	"fmt"
	"path/filepath"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/manager"
	"github.com/kubesphere/kubekey/v4/pkg/proxy"
)

func newPipelineCommand() *cobra.Command {
	o := options.NewPipelineOptions()

	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Executor a pipeline in kubernetes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			restconfig, err := ctrl.GetConfig()
			if err != nil {
				return fmt.Errorf("cannot get restconfig in kubernetes. error is %w", err)
			}
			kubeclient, err := ctrlclient.New(restconfig, ctrlclient.Options{
				Scheme: _const.Scheme,
			})
			if err != nil {
				return fmt.Errorf("could not create client: %w", err)
			}
			// get pipeline
			pipeline := &kkcorev1.Pipeline{}
			if err := kubeclient.Get(cmd.Context(), ctrlclient.ObjectKey{
				Name:      o.Name,
				Namespace: o.Namespace,
			}, pipeline); err != nil {
				return err
			}
			if pipeline.Status.Phase != kkcorev1.PipelinePhaseRunning {
				klog.InfoS("pipeline is not running, skip", "pipeline", ctrlclient.ObjectKeyFromObject(pipeline))

				return nil
			}
			// get inventory
			inventory := new(kkcorev1.Inventory)
			if err := kubeclient.Get(cmd.Context(), ctrlclient.ObjectKey{
				Name:      pipeline.Spec.InventoryRef.Name,
				Namespace: pipeline.Spec.InventoryRef.Namespace,
			}, inventory); err != nil {
				return err
			}
			if err := proxy.RestConfig(filepath.Join(_const.GetWorkdirFromConfig(pipeline.Spec.Config), _const.RuntimeDir), restconfig); err != nil {
				return fmt.Errorf("could not get rest config: %w", err)
			}
			// use proxy client to store task.
			proxyclient, err := ctrlclient.New(restconfig, ctrlclient.Options{
				Scheme: _const.Scheme,
			})
			if err != nil {
				return fmt.Errorf("could not create client: %w", err)
			}

			return manager.NewCommandManager(manager.CommandManagerOptions{
				Pipeline:  pipeline,
				Inventory: inventory,
				Client:    proxyclient,
			}).Run(cmd.Context())
		},
	}
	fs := cmd.Flags()
	for _, f := range o.Flags().FlagSets {
		fs.AddFlagSet(f)
	}

	return cmd
}
