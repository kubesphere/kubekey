package v1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

const PipelineFieldPlaybook = "spec.playbook"

// AddConversionFuncs adds the conversion functions to the given scheme.
// NOTE: ownerReferences:pipeline is valid in proxy client.
func AddConversionFuncs(scheme *runtime.Scheme) error {
	return scheme.AddFieldLabelConversionFunc(
		SchemeGroupVersion.WithKind("Pipeline"),
		func(label, value string) (string, string, error) {
			if label == PipelineFieldPlaybook {
				return label, value, nil
			}

			return "", "", fmt.Errorf("field label %q not supported for Pipeline", label)
		},
	)
}
