package app

import (
	"path/filepath"

	"github.com/cockroachdb/errors"
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

func newPlaybookCommand() *cobra.Command {
	o := options.NewPlaybookOptions()

	cmd := &cobra.Command{
		Use:   "playbook",
		Short: "Executor a playbook in kubernetes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			restconfig, err := ctrl.GetConfig()
			if err != nil {
				return errors.Wrap(err, "failed to get restconfig")
			}
			kubeclient, err := ctrlclient.New(restconfig, ctrlclient.Options{
				Scheme: _const.Scheme,
			})
			if err != nil {
				return errors.Wrap(err, "failed to create client")
			}
			// get playbook
			playbook := &kkcorev1.Playbook{}
			if err := kubeclient.Get(cmd.Context(), ctrlclient.ObjectKey{
				Name:      o.Name,
				Namespace: o.Namespace,
			}, playbook); err != nil {
				return errors.Wrap(err, "failed to get playbook")
			}
			if playbook.Status.Phase != kkcorev1.PlaybookPhaseRunning {
				klog.InfoS("playbook is not running, skip", "playbook", ctrlclient.ObjectKeyFromObject(playbook))

				return nil
			}
			// get inventory
			inventory := new(kkcorev1.Inventory)
			if err := kubeclient.Get(cmd.Context(), ctrlclient.ObjectKey{
				Name:      playbook.Spec.InventoryRef.Name,
				Namespace: playbook.Spec.InventoryRef.Namespace,
			}, inventory); err != nil {
				return errors.Wrap(err, "failed to get inventory")
			}
			if err := proxy.RestConfig(filepath.Join(_const.GetWorkdirFromConfig(playbook.Spec.Config), _const.RuntimeDir), restconfig); err != nil {
				return errors.Wrap(err, "failed to get rest config")
			}
			// use proxy client to store task.
			proxyclient, err := ctrlclient.New(restconfig, ctrlclient.Options{
				Scheme: _const.Scheme,
			})
			if err != nil {
				return errors.Wrap(err, "failed to create client")
			}

			return manager.NewCommandManager(manager.CommandManagerOptions{
				Playbook:  playbook,
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
