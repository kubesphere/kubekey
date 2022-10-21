/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/blang/semver"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/cluster-api/util/version"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	infrabootstrapv1 "github.com/kubesphere/kubekey/bootstrap/k3s/api/v1beta1"
)

func (in *K3sControlPlane) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/mutate-controlplane-cluster-x-k8s-io-v1beta1-k3scontrolplane,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=controlplane.cluster.x-k8s.io,resources=k3scontrolplanes,versions=v1beta1,name=default.k3scontrolplane.controlplane.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1
// +kubebuilder:webhook:verbs=create;update,path=/validate-controlplane-cluster-x-k8s-io-v1beta1-k3scontrolplane,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=controlplane.cluster.x-k8s.io,resources=k3scontrolplanes,versions=v1beta1,name=validation.k3scontrolplane.controlplane.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

var _ webhook.Defaulter = &K3sControlPlane{}
var _ webhook.Validator = &K3sControlPlane{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *K3sControlPlane) Default() {
	defaultK3sControlPlaneSpec(&in.Spec, in.Namespace)
}

func defaultK3sControlPlaneSpec(s *K3sControlPlaneSpec, namespace string) {
	if s.Replicas == nil {
		replicas := int32(1)
		s.Replicas = &replicas
	}

	if s.MachineTemplate.InfrastructureRef.Namespace == "" {
		s.MachineTemplate.InfrastructureRef.Namespace = namespace
	}

	if !strings.HasPrefix(s.Version, "v") {
		s.Version = "v" + s.Version
	}

	if s.K3sConfigSpec.ServerConfiguration.Database.DataStoreEndPoint == "" && s.K3sConfigSpec.ServerConfiguration.Database.ClusterInit {
		s.K3sConfigSpec.ServerConfiguration.Database.ClusterInit = true
	}

	infrabootstrapv1.DefaultK3sConfigSpec(&s.K3sConfigSpec)

	s.RolloutStrategy = defaultRolloutStrategy(s.RolloutStrategy)
}

func defaultRolloutStrategy(rolloutStrategy *RolloutStrategy) *RolloutStrategy {
	ios1 := intstr.FromInt(1)

	if rolloutStrategy == nil {
		rolloutStrategy = &RolloutStrategy{}
	}

	// Enforce RollingUpdate strategy and default MaxSurge if not set.
	if rolloutStrategy != nil {
		if len(rolloutStrategy.Type) == 0 {
			rolloutStrategy.Type = RollingUpdateStrategyType
		}
		if rolloutStrategy.Type == RollingUpdateStrategyType {
			if rolloutStrategy.RollingUpdate == nil {
				rolloutStrategy.RollingUpdate = &RollingUpdate{}
			}
			rolloutStrategy.RollingUpdate.MaxSurge = intstr.ValueOrDefault(rolloutStrategy.RollingUpdate.MaxSurge, ios1)
		}
	}

	return rolloutStrategy
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *K3sControlPlane) ValidateCreate() error {
	spec := in.Spec
	allErrs := validateK3sControlPlaneSpec(spec, in.Namespace, field.NewPath("spec"))
	allErrs = append(allErrs, validateServerConfiguration(spec.K3sConfigSpec.ServerConfiguration, nil, field.NewPath("spec", "k3sConfigSpec", "serverConfiguration"))...)
	allErrs = append(allErrs, spec.K3sConfigSpec.Validate(field.NewPath("spec", "k3sConfigSpec"))...)
	if len(allErrs) > 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind("KubeadmControlPlane").GroupKind(), in.Name, allErrs)
	}
	return nil
}

const (
	spec            = "spec"
	k3sConfigSpec   = "k3sConfigSpec"
	preK3sCommands  = "preK3sCommands"
	postK3sCommands = "postK3sCommands"
	files           = "files"
)

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *K3sControlPlane) ValidateUpdate(old runtime.Object) error {
	// add a * to indicate everything beneath is ok.
	// For example, {"spec", "*"} will allow any path under "spec" to change.
	allowedPaths := [][]string{
		{"metadata", "*"},
		{spec, k3sConfigSpec, preK3sCommands},
		{spec, k3sConfigSpec, postK3sCommands},
		{spec, k3sConfigSpec, files},
		{spec, "machineTemplate", "metadata", "*"},
		{spec, "machineTemplate", "infrastructureRef", "apiVersion"},
		{spec, "machineTemplate", "infrastructureRef", "name"},
		{spec, "machineTemplate", "infrastructureRef", "kind"},
		{spec, "machineTemplate", "nodeDrainTimeout"},
		{spec, "machineTemplate", "nodeDeletionTimeout"},
		{spec, "replicas"},
		{spec, "version"},
		{spec, "rolloutAfter"},
		{spec, "rolloutStrategy", "*"},
	}

	allErrs := validateK3sControlPlaneSpec(in.Spec, in.Namespace, field.NewPath("spec"))

	prev, ok := old.(*K3sControlPlane)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expecting K3sControlPlane but got a %T", old))
	}

	originalJSON, err := json.Marshal(prev)
	if err != nil {
		return apierrors.NewInternalError(err)
	}
	modifiedJSON, err := json.Marshal(in)
	if err != nil {
		return apierrors.NewInternalError(err)
	}

	diff, err := jsonpatch.CreateMergePatch(originalJSON, modifiedJSON)
	if err != nil {
		return apierrors.NewInternalError(err)
	}
	jsonPatch := map[string]interface{}{}
	if err := json.Unmarshal(diff, &jsonPatch); err != nil {
		return apierrors.NewInternalError(err)
	}
	// Build a list of all paths that are trying to change
	diffpaths := paths([]string{}, jsonPatch)
	// Every path in the diff must be valid for the update function to work.
	for _, path := range diffpaths {
		// Ignore paths that are empty
		if len(path) == 0 {
			continue
		}
		if !allowed(allowedPaths, path) {
			if len(path) == 1 {
				allErrs = append(allErrs, field.Forbidden(field.NewPath(path[0]), "cannot be modified"))
				continue
			}
			allErrs = append(allErrs, field.Forbidden(field.NewPath(path[0], path[1:]...), "cannot be modified"))
		}
	}

	allErrs = append(allErrs, in.validateVersion(prev.Spec.Version)...)
	allErrs = append(allErrs, validateServerConfiguration(in.Spec.K3sConfigSpec.ServerConfiguration, prev.Spec.K3sConfigSpec.ServerConfiguration, field.NewPath("spec", "k3sConfigSpec", "serverConfiguration"))...)
	allErrs = append(allErrs, in.Spec.K3sConfigSpec.Validate(field.NewPath("spec", "K3sConfigSpec"))...)

	if len(allErrs) > 0 {
		return apierrors.NewInvalid(GroupVersion.WithKind("K3sControlPlane").GroupKind(), in.Name, allErrs)
	}

	return nil
}

func validateK3sControlPlaneSpec(s K3sControlPlaneSpec, namespace string, pathPrefix *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if s.Replicas == nil {
		allErrs = append(
			allErrs,
			field.Required(
				pathPrefix.Child("replicas"),
				"is required",
			),
		)
	} else if *s.Replicas <= 0 {
		// The use of the scale subresource should provide a guarantee that negative values
		// should not be accepted for this field, but since we have to validate that Replicas != 0
		// it doesn't hurt to also additionally validate for negative numbers here as well.
		allErrs = append(
			allErrs,
			field.Forbidden(
				pathPrefix.Child("replicas"),
				"cannot be less than or equal to 0",
			),
		)
	}

	if s.MachineTemplate.InfrastructureRef.APIVersion == "" {
		allErrs = append(
			allErrs,
			field.Invalid(
				pathPrefix.Child("machineTemplate", "infrastructure", "apiVersion"),
				s.MachineTemplate.InfrastructureRef.APIVersion,
				"cannot be empty",
			),
		)
	}
	if s.MachineTemplate.InfrastructureRef.Kind == "" {
		allErrs = append(
			allErrs,
			field.Invalid(
				pathPrefix.Child("machineTemplate", "infrastructure", "kind"),
				s.MachineTemplate.InfrastructureRef.Kind,
				"cannot be empty",
			),
		)
	}
	if s.MachineTemplate.InfrastructureRef.Name == "" {
		allErrs = append(
			allErrs,
			field.Invalid(
				pathPrefix.Child("machineTemplate", "infrastructure", "name"),
				s.MachineTemplate.InfrastructureRef.Name,
				"cannot be empty",
			),
		)
	}
	if s.MachineTemplate.InfrastructureRef.Namespace != namespace {
		allErrs = append(
			allErrs,
			field.Invalid(
				pathPrefix.Child("machineTemplate", "infrastructure", "namespace"),
				s.MachineTemplate.InfrastructureRef.Namespace,
				"must match metadata.namespace",
			),
		)
	}

	if !version.KubeSemver.MatchString(s.Version) {
		allErrs = append(allErrs, field.Invalid(pathPrefix.Child("version"), s.Version, "must be a valid semantic version"))
	}

	allErrs = append(allErrs, validateRolloutStrategy(s.RolloutStrategy, s.Replicas, pathPrefix.Child("rolloutStrategy"))...)

	return allErrs
}

func validateRolloutStrategy(rolloutStrategy *RolloutStrategy, replicas *int32, pathPrefix *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if rolloutStrategy == nil {
		return allErrs
	}

	if rolloutStrategy.Type != RollingUpdateStrategyType {
		allErrs = append(
			allErrs,
			field.Required(
				pathPrefix.Child("type"),
				"only RollingUpdateStrategyType is supported",
			),
		)
	}

	ios1 := intstr.FromInt(1)
	ios0 := intstr.FromInt(0)

	if *rolloutStrategy.RollingUpdate.MaxSurge == ios0 && (replicas != nil && *replicas < int32(3)) {
		allErrs = append(
			allErrs,
			field.Required(
				pathPrefix.Child("rollingUpdate"),
				"when KubeadmControlPlane is configured to scale-in, replica count needs to be at least 3",
			),
		)
	}

	if *rolloutStrategy.RollingUpdate.MaxSurge != ios1 && *rolloutStrategy.RollingUpdate.MaxSurge != ios0 {
		allErrs = append(
			allErrs,
			field.Required(
				pathPrefix.Child("rollingUpdate", "maxSurge"),
				"value must be 1 or 0",
			),
		)
	}

	return allErrs
}

func validateServerConfiguration(newServerConfiguration, oldServerConfiguration *infrabootstrapv1.ServerConfiguration, pathPrefix *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if newServerConfiguration == nil {
		return allErrs
	}

	if newServerConfiguration.Database.ClusterInit && newServerConfiguration.Database.DataStoreEndPoint != "" {
		allErrs = append(
			allErrs,
			field.Forbidden(
				pathPrefix.Child("database", "clusterInit"),
				"cannot have both external and local etcd",
			),
		)
	}

	// update validations
	if oldServerConfiguration != nil {
		if newServerConfiguration.Database.ClusterInit && oldServerConfiguration.Database.DataStoreEndPoint != "" {
			allErrs = append(
				allErrs,
				field.Forbidden(
					pathPrefix.Child("database", "clusterInit"),
					"cannot change between external and local etcd",
				),
			)
		}

		if newServerConfiguration.Database.DataStoreEndPoint != "" && oldServerConfiguration.Database.ClusterInit {
			allErrs = append(
				allErrs,
				field.Forbidden(
					pathPrefix.Child("database", "dataStoreEndPoint"),
					"cannot change between external and local etcd",
				),
			)
		}
	}

	return allErrs
}

func allowed(allowList [][]string, path []string) bool {
	for _, allowed := range allowList {
		if pathsMatch(allowed, path) {
			return true
		}
	}
	return false
}

func pathsMatch(allowed, path []string) bool {
	// if either are empty then no match can be made
	if len(allowed) == 0 || len(path) == 0 {
		return false
	}
	i := 0
	for i = range path {
		// reached the end of the allowed path and no match was found
		if i > len(allowed)-1 {
			return false
		}
		if allowed[i] == "*" {
			return true
		}
		if path[i] != allowed[i] {
			return false
		}
	}
	// path has been completely iterated and has not matched the end of the path.
	// e.g. allowed: []string{"a","b","c"}, path: []string{"a"}
	return i >= len(allowed)-1
}

// paths builds a slice of paths that are being modified.
func paths(path []string, diff map[string]interface{}) [][]string {
	allPaths := [][]string{}
	for key, m := range diff {
		nested, ok := m.(map[string]interface{})
		if !ok {
			// We have to use a copy of path, because otherwise the slice we append to
			// allPaths would be overwritten in another iteration.
			tmp := make([]string, len(path))
			copy(tmp, path)
			allPaths = append(allPaths, append(tmp, key))
			continue
		}
		allPaths = append(allPaths, paths(append(path, key), nested)...)
	}
	return allPaths
}

func (in *K3sControlPlane) validateVersion(previousVersion string) (allErrs field.ErrorList) {
	fromVersion, err := version.ParseMajorMinorPatch(previousVersion)
	if err != nil {
		allErrs = append(allErrs,
			field.InternalError(
				field.NewPath("spec", "version"),
				errors.Wrapf(err, "failed to parse current k3scontrolplane version: %s", previousVersion),
			),
		)
		return allErrs
	}

	toVersion, err := version.ParseMajorMinorPatch(in.Spec.Version)
	if err != nil {
		allErrs = append(allErrs,
			field.InternalError(
				field.NewPath("spec", "version"),
				errors.Wrapf(err, "failed to parse updated k3scontrolplane version: %s", in.Spec.Version),
			),
		)
		return allErrs
	}

	// Check if we're trying to upgrade to Kubernetes v1.19.0, which is not supported.
	//
	// See https://github.com/kubernetes-sigs/cluster-api/issues/3564
	if fromVersion.NE(toVersion) && toVersion.Equals(semver.MustParse("1.19.0")) {
		allErrs = append(allErrs,
			field.Forbidden(
				field.NewPath("spec", "version"),
				"cannot update Kubernetes version to v1.19.0, for more information see https://github.com/kubernetes-sigs/cluster-api/issues/3564",
			),
		)
		return allErrs
	}

	// Since upgrades to the next minor version are allowed, irrespective of the patch version.
	ceilVersion := semver.Version{
		Major: fromVersion.Major,
		Minor: fromVersion.Minor + 2,
		Patch: 0,
	}
	if toVersion.GTE(ceilVersion) {
		allErrs = append(allErrs,
			field.Forbidden(
				field.NewPath("spec", "version"),
				fmt.Sprintf("cannot update Kubernetes version from %s to %s", previousVersion, in.Spec.Version),
			),
		)
	}

	return allErrs
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *K3sControlPlane) ValidateDelete() error {
	return nil
}
