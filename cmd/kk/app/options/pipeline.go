package options

import (
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"
)

type PipelineOptions struct {
	Name      string
	Namespace string
	WorkDir   string
}

func NewPipelineOption() *PipelineOptions {
	return &PipelineOptions{
		Namespace: metav1.NamespaceDefault,
		WorkDir:   "/kubekey",
	}
}

func (o *PipelineOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	pfs := fss.FlagSet("pipeline flags")
	pfs.StringVar(&o.Name, "name", o.Name, "name of pipeline")
	pfs.StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "namespace of pipeline")
	pfs.StringVar(&o.WorkDir, "work-dir", o.WorkDir, "the base Dir for kubekey. Default current dir. ")
	return fss

}

func (o *PipelineOptions) Complete(cmd *cobra.Command, args []string) {
	// do nothing
}
