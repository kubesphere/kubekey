package options

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliflag "k8s.io/component-base/cli/flag"
)

// PlaybookOptions for NewPlaybookOptions
type PlaybookOptions struct {
	Name      string
	Namespace string
}

// NewPlaybookOptions for newPlaybookCommand
func NewPlaybookOptions() *PlaybookOptions {
	return &PlaybookOptions{
		Namespace: metav1.NamespaceDefault,
	}
}

// Flags add to newPlaybookCommand
func (o *PlaybookOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}
	pfs := fss.FlagSet("playbook flags")
	pfs.StringVar(&o.Name, "name", o.Name, "name of playbook")
	pfs.StringVarP(&o.Namespace, "namespace", "n", o.Namespace, "namespace of playbook")

	return fss
}
