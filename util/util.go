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

// Package util contains utility functions
package util

import (
	"context"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/v3/api/v1beta1"
	"github.com/kubesphere/kubekey/v3/pkg/scope"
)

// GetInfraCluster returns the infrastructure cluster object corresponding to a Cluster.
func GetInfraCluster(ctx context.Context, c client.Client, log logr.Logger, cluster *clusterv1.Cluster, controllerName string,
	dataDir string) (*scope.ClusterScope, error) {
	kkCluster := &infrav1.KKCluster{}
	infraClusterName := client.ObjectKey{
		Namespace: cluster.Spec.InfrastructureRef.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}

	if err := c.Get(ctx, infraClusterName, kkCluster); err != nil {
		return nil, err
	}

	// Create the cluster scope
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Client:         c,
		Logger:         &log,
		Cluster:        cluster,
		KKCluster:      kkCluster,
		ControllerName: controllerName,
		RootFsBasePath: dataDir,
	})
	if err != nil {
		return nil, err
	}

	return clusterScope, nil
}

// GetOwnerKKMachine returns the Machine object owning the given object.
func GetOwnerKKMachine(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*infrav1.KKMachine, error) {
	for _, ref := range obj.OwnerReferences {
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, err
		}
		if ref.Kind == "KKMachine" && gv.Group == infrav1.GroupVersion.Group {
			return GetKKMachineByName(ctx, c, obj.Namespace, ref.Name)
		}
	}
	return nil, nil
}

// GetKKMachineByName finds and return a Machine object using the specified params.
func GetKKMachineByName(ctx context.Context, c client.Client, namespace, name string) (*infrav1.KKMachine, error) {
	m := &infrav1.KKMachine{}
	key := client.ObjectKey{Name: name, Namespace: namespace}
	if err := c.Get(ctx, key, m); err != nil {
		return nil, err
	}
	return m, nil
}
