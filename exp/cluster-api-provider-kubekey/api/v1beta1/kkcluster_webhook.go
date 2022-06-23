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
	"fmt"
	"net"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	"github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/storage/names"
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

func (r *KKCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-kkcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkclusters,verbs=create;update,versions=v1beta1,name=default.kkcluster.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KKCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KKCluster) Default() {
	kkclusterlog.Info("default", "name", r.Name)

	defaultAuth(&r.Spec.Nodes.Auth)
	defaultContainerManager(&r.Spec)
	defaultInstance(&r.Spec)
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

func defaultContainerManager(spec *KKClusterSpec) {
	// Direct connection to the user-provided CRI socket
	if spec.Nodes.ContainerManager.CRISocket != "" {
		return
	}

	if spec.Nodes.ContainerManager.Type == "" {
		spec.Nodes.ContainerManager.Type = ContainerdType
	}

	switch spec.Nodes.ContainerManager.Type {
	case ContainerdType:
		if spec.Nodes.ContainerManager.Version == "" {
			spec.Nodes.ContainerManager.Version = DefaultContainerdVersion
		}
		spec.Nodes.ContainerManager.CRISocket = DefaultContainerdCRISocket
	case DockerType:
		if spec.Nodes.ContainerManager.Version == "" {
			spec.Nodes.ContainerManager.Version = DefaultDockerVersion
		}
		spec.Nodes.ContainerManager.CRISocket = DefaultDockerCRISocket
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

//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-kkcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkclusters,verbs=create;update,versions=v1beta1,name=validation.kkcluster.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Validator = &KKCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KKCluster) ValidateCreate() error {
	kkclusterlog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList
	allErrs = append(allErrs, validateClusterNodes(r.Spec.Nodes)...)
	allErrs = append(allErrs, validateLoadBalancer(r.Spec.ControlPlaneLoadBalancer)...)

	return aggregateObjErrors(r.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KKCluster) ValidateUpdate(old runtime.Object) error {
	kkclusterlog.Info("validate update", "name", r.Name)

	var allErrs field.ErrorList
	oldC, ok := old.(*KKCluster)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected an KKCluster but got a %T", old))
	}

	newLoadBalancer := &KKLoadBalancerSpec{}

	if r.Spec.ControlPlaneLoadBalancer != nil {
		newLoadBalancer = r.Spec.ControlPlaneLoadBalancer.DeepCopy()
	}

	if oldC.Spec.ControlPlaneLoadBalancer != nil {
		// If old scheme was not nil, the new scheme should be the same.
		existingLoadBalancer := oldC.Spec.ControlPlaneLoadBalancer.DeepCopy()
		if !cmp.Equal(newLoadBalancer, existingLoadBalancer) {
			allErrs = append(allErrs,
				field.Invalid(field.NewPath("spec", "controlPlaneLoadBalancer"),
					r.Spec.ControlPlaneLoadBalancer, "field is immutable"),
			)
		}
	}

	allErrs = append(allErrs, validateClusterNodes(r.Spec.Nodes)...)
	return aggregateObjErrors(r.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KKCluster) ValidateDelete() error {
	return nil
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
