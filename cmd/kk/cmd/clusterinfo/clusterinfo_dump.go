package clusterinfo

import (
	"fmt"
	"github.com/kubesphere/kubekey/v3/cmd/kk/cmd/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/clusterinfo"
	"github.com/spf13/cobra"
)

type ClusterInfoDumpOptions struct {
	Options clusterinfo.DumpOption
}

func NewClusterInfoDumpOptions() *ClusterInfoDumpOptions {
	return &ClusterInfoDumpOptions{}
}

func NewCmdClusterInfoDump() *cobra.Command {
	o := NewClusterInfoDumpOptions()
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "Dumping key cluster configurations and files",
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckErr(o.Validate())
			util.CheckErr(o.Run())

		},
	}

	o.addFlag(cmd)
	return cmd
}

func (o *ClusterInfoDumpOptions) Validate() error {
	switch o.Options.Type {
	case "yaml", "YAML", "json", "JSON":
	default:
		return fmt.Errorf("unsupport output content  format [%s]", o.Options.Type)
	}
	return nil
}

func (o *ClusterInfoDumpOptions) addFlag(cmd *cobra.Command) {
	DefaultDumpNamespaces := []string{"kubesphere-system", "kubesphere-logging-system", "kubesphere-monitoring-system", "openpitrix-system", "kube-system", "istio-system", "kubesphere-devops-system", "porter-system"}
	cmd.Flags().StringArrayVar(&o.Options.Namespace, "namespaces", DefaultDumpNamespaces, "Namespaces to be dumped, separated by commas.")
	cmd.Flags().StringVar(&o.Options.KubeConfig, "kube-config", "", "Path to the kube-config file")
	cmd.Flags().BoolVarP(&o.Options.AllNamespaces, "all-namespaces", "A", false, "dump all namespaces.")
	cmd.Flags().StringVar(&o.Options.OutputDir, "output-dir", "", "output the dump result to the specified directory directory.")
	cmd.Flags().StringVarP(&o.Options.Type, "output", "o", "json", "output file content format. support in json,yaml")
	cmd.Flags().BoolVarP(&o.Options.Tar, "tar", "t", false, "build the dump result into a tar")
	cmd.Flags().IntVar(&o.Options.Queue, "queue", 5, "dump queue size")
	cmd.Flags().BoolVar(&o.Options.Logger, "log", false, "output the dump result to the log console")
}

func (o *ClusterInfoDumpOptions) Run() error {
	fmt.Println("dumping cluster info...")
	return clusterinfo.Dump(o.Options)

}
