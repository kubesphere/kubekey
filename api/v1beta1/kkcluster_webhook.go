/*
Copyright 2022 The KubeSphere Authors.

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
	"fmt"
	"net"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/cluster-api/util/version"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	defaultSSHUser             = "root"
	defaultSSHPort             = 22
	defaultSSHEstablishTimeout = 30 * time.Second
)

// log is for logging in this package.
var kkclusterlog = logf.Log.WithName("kkcluster-resource")

func (k *KKCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(k).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-kkcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkclusters,verbs=create;update,versions=v1beta1,name=default.kkcluster.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KKCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *KKCluster) Default() {
	kkclusterlog.Info("default", "name", k.Name)

	defaultDistribution(&k.Spec)
	defaultAuth(&k.Spec.Nodes.Auth)
	defaultInstance(&k.Spec)
	defaultInPlaceUpgradeAnnotation(k.GetAnnotations())
}

func defaultDistribution(spec *KKClusterSpec) {
	if spec.Distribution == "" {
		spec.Distribution = "kubernetes"
	}
	if spec.Distribution == "k8s" {
		spec.Distribution = "kubernetes"
	}
}

func defaultAuth(auth *Auth) {
	if auth.User == "" {
		auth.User = defaultSSHUser
	}
	if auth.Port == nil {
		p := defaultSSHPort
		auth.Port = &p
	}
	if auth.Timeout == nil {
		t := defaultSSHEstablishTimeout
		auth.Timeout = &t
	}
}

func defaultInstance(spec *KKClusterSpec) {
	for i := range spec.Nodes.Instances {
		instance := &spec.Nodes.Instances[i]
		if instance.Name == "" {
			instance.Name = names.SimpleNameGenerator.GenerateName("kk-instance-")
		}
		if instance.InternalAddress == "" {
			instance.InternalAddress = instance.Address
		}
		if instance.Arch == "" {
			instance.Arch = "amd64"
		}
		if len(instance.Roles) == 0 {
			instance.Roles = []Role{ControlPlane, Worker}
		}
	}
}

func defaultInPlaceUpgradeAnnotation(annotation map[string]string) {
	upgradeVersion, ok := annotation[InPlaceUpgradeVersionAnnotation]
	if !ok {
		return
	}

	if !strings.HasPrefix(upgradeVersion, "v") {
		annotation[InPlaceUpgradeVersionAnnotation] = "v" + upgradeVersion
	}
}

//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-kkcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkclusters,verbs=create;update,versions=v1beta1,name=validation.kkcluster.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Validator = &KKCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *KKCluster) ValidateCreate() error {
	kkclusterlog.Info("validate create", "name", k.Name)

	var allErrs field.ErrorList
	allErrs = append(allErrs, validateDistribution(k.Spec)...)
	allErrs = append(allErrs, validateClusterNodes(k.Spec.Nodes)...)
	allErrs = append(allErrs, validateLoadBalancer(k.Spec.ControlPlaneLoadBalancer)...)

	return aggregateObjErrors(k.GroupVersionKind().GroupKind(), k.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *KKCluster) ValidateUpdate(old runtime.Object) error {
	kkclusterlog.Info("validate update", "name", k.Name)

	var allErrs field.ErrorList
	oldC, ok := old.(*KKCluster)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected an KKCluster but got a %T", old))
	}

	newLoadBalancer := &KKLoadBalancerSpec{}

	if k.Spec.ControlPlaneLoadBalancer != nil {
		newLoadBalancer = k.Spec.ControlPlaneLoadBalancer.DeepCopy()
	}

	if oldC.Spec.ControlPlaneLoadBalancer != nil {
		// If old scheme was not nil, the new scheme should be the same.
		existingLoadBalancer := oldC.Spec.ControlPlaneLoadBalancer.DeepCopy()
		if !cmp.Equal(newLoadBalancer, existingLoadBalancer) {
			allErrs = append(allErrs,
				field.Invalid(field.NewPath("spec", "controlPlaneLoadBalancer"),
					k.Spec.ControlPlaneLoadBalancer, "field is immutable"),
			)
		}
	}

	allErrs = append(allErrs, validateClusterNodes(k.Spec.Nodes)...)
	allErrs = append(allErrs, validateInPlaceUpgrade(k.GetAnnotations())...)
	return aggregateObjErrors(k.GroupVersionKind().GroupKind(), k.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *KKCluster) ValidateDelete() error {
	return nil
}

func validateDistribution(spec KKClusterSpec) []*field.Error {
	var errs field.ErrorList
	path := field.NewPath("spec", "distribution")
	switch spec.Distribution {
	case K3S:
		return errs
	case KUBERNETES:
		return errs
	default:
		errs = append(errs, field.NotSupported(path, spec.Distribution, []string{K3S, KUBERNETES}))
	}
	return errs
}

func validateLoadBalancer(loadBalancer *KKLoadBalancerSpec) []*field.Error {
	var errs field.ErrorList
	path := field.NewPath("spec", "controlPlaneLoadBalancer")
	if loadBalancer.Host == "" {
		errs = append(errs, field.Required(path.Child("host"), "can't be empty"))
	}
	return errs
}

func validateClusterNodes(nodes Nodes) []*field.Error {
	var errs field.ErrorList

	if nodes.Auth.Password == "" && nodes.Auth.PrivateKey == "" && nodes.Auth.PrivateKeyPath == "" {
		errs = append(errs, field.Required(field.NewPath("spec", "nodes", "auth"), "password and privateKey can't both be empty"))
	}

	nameSet := mapset.NewThreadUnsafeSet()
	addrSet := mapset.NewThreadUnsafeSet()
	internalAddrSet := mapset.NewThreadUnsafeSet()
	for i := range nodes.Instances {
		instance := nodes.Instances[i]
		path := field.NewPath("spec", "nodes", fmt.Sprintf("instances[%d]", i))
		if strings.ToLower(instance.Name) != instance.Name {
			errs = append(errs,
				field.Forbidden(path.Child("name"), "instance name must be the lower case"))
		}
		if !nameSet.Add(instance.Name) {
			errs = append(errs, field.Duplicate(path.Child("name"), instance.Name))
		}

		if net.ParseIP(instance.Address) == nil {
			errs = append(errs, field.Invalid(path.Child("address"), instance.Address, "instance address is invalid"))
		}
		if !addrSet.Add(instance.Address) {
			errs = append(errs, field.Duplicate(path.Child("address"), instance.Address))
		}

		if net.ParseIP(instance.InternalAddress) == nil {
			errs = append(errs, field.Invalid(path.Child("internalAddress"), instance.InternalAddress, "instance internalAddress is invalid"))
		}
		if !internalAddrSet.Add(instance.InternalAddress) {
			errs = append(errs, field.Duplicate(path.Child("internalAddress"), instance.InternalAddress))
		}
	}
	return errs
}

func validateInPlaceUpgrade(newAnnotation map[string]string) []*field.Error {
	var allErrs field.ErrorList

	if v, ok := newAnnotation[InPlaceUpgradeVersionAnnotation]; ok {
		_, err := version.ParseMajorMinorPatch(v)
		if err != nil {
			allErrs = append(allErrs,
				field.InternalError(
					field.NewPath("metadata", "annotations"),
					errors.Wrapf(err, "failed to parse in-place upgrade version: %s", v)),
			)
		}
	}
	return allErrs
}
