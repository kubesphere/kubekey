package v1

import (
	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

const PlaybookFieldPlaybook = "spec.playbook"

// AddConversionFuncs adds the conversion functions to the given scheme.
// NOTE: ownerReferences:playbook is valid in proxy client.
func AddConversionFuncs(scheme *runtime.Scheme) error {
	return scheme.AddFieldLabelConversionFunc(
		SchemeGroupVersion.WithKind("Playbook"),
		func(label, value string) (string, string, error) {
			if label == PlaybookFieldPlaybook {
				return label, value, nil
			}

			return "", "", errors.Errorf("field label %q not supported for Playbook", label)
		},
	)
}
