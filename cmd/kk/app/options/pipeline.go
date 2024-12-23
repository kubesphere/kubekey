package options

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"
)

// PipelineOptions for NewPipelineOptions
type PipelineOptions struct {
	Name      string
	Namespace string
}

// NewPipelineOptions for newPipelineCommand
func NewPipelineOptions() *PipelineOptions {
	return &PipelineOptions{
		Namespace: metav1.NamespaceDefault,
	}
}

// Flags add to newPipelineCommand
func (o *PipelineOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	pfs := fss.FlagSet("pipeline flags")
	pfs.StringVar(&o.Name, "name", o.Name, "name of pipeline")
	pfs.StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "namespace of pipeline")

	return fss
}
