package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

// supportedFields defines the set of field labels that are allowed for field selector queries
// on the Task resource. Only these fields can be used in field selectors for filtering.
var supportedFields = sets.NewString(
	"metadata.name",
	"metadata.namespace",
	"playbook.name",
	"playbook.uid",
)

// RegisterFieldLabelConversion registers a field label conversion function for the Task resource.
// This function ensures that only supported field labels can be used in field selectors.
// If an unsupported field label is used, an error is returned.
func RegisterFieldLabelConversion(scheme *runtime.Scheme) error {
	return scheme.AddFieldLabelConversionFunc(
		SchemeGroupVersion.WithKind("Task"),
		func(label, value string) (string, string, error) {
			if !supportedFields.Has(label) {
				return "", "", fmt.Errorf("field label %q is not supported", label)
			}
			return label, value, nil
		},
	)
}
