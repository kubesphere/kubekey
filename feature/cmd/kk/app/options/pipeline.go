package options

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"
)

// PipelineOptions for NewPipelineOptions
type PipelineOptions struct {
	Name      string
	Namespace string
	WorkDir   string
}

// NewPipelineOptions for newPipelineCommand
func NewPipelineOptions() *PipelineOptions {
	return &PipelineOptions{
		Namespace: metav1.NamespaceDefault,
		WorkDir:   "/kubekey",
	}
}

// Flags add to newPipelineCommand
func (o *PipelineOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	pfs := fss.FlagSet("pipeline flags")
	pfs.StringVar(&o.Name, "name", o.Name, "name of pipeline")
	pfs.StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "namespace of pipeline")
	pfs.StringVar(&o.WorkDir, "work-dir", o.WorkDir, "the base Dir for kubekey. Default current dir. ")

	return fss
}
