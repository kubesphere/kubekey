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

package kkinstance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
)

const semaphoreInformationKey = "lock-information"

// Mutex uses a ConfigMap to synchronize KKInstance.
type Mutex struct {
	client client.Client
}

// NewMutex returns a lock that can be held by a KKInstance.
func NewMutex(client client.Client) *Mutex {
	return &Mutex{
		client: client,
	}
}

// Lock allows a control plane node to be the first and only node to run kubeadm init.
func (m *Mutex) Lock(ctx context.Context, cluster *clusterv1.Cluster, kkInstance *infrav1.KKInstance) bool {
	sema := newSemaphore()
	cmName := configMapName(cluster.Name)
	log := ctrl.LoggerFrom(ctx, "ConfigMap", klog.KRef(cluster.Namespace, cmName))
	err := m.client.Get(ctx, client.ObjectKey{
		Namespace: cluster.Namespace,
		Name:      cmName,
	}, sema.ConfigMap)
	switch {
	case apierrors.IsNotFound(err):
		break
	case err != nil:
		log.Error(err, "Failed to acquire init lock")
		return false
	default: // Successfully found an existing config map.
		info, err := sema.information()
		if err != nil {
			log.Error(err, "Failed to get information about the existing init lock")
			return false
		}
		// The kkinstance requesting the lock is the kkinstance that created the lock, therefore the lock is acquired.
		if info.KKInstanceName == kkInstance.Name {
			return true
		}

		// If the kkinstance that created the lock can not be found unlock the mutex.
		if err := m.client.Get(ctx, client.ObjectKey{
			Namespace: cluster.Namespace,
			Name:      info.KKInstanceName,
		}, &clusterv1.Machine{}); err != nil {
			log.Error(err, "Failed to get kkinstance holding lock")
			if apierrors.IsNotFound(err) {
				m.Unlock(ctx, cluster)
			}
		}
		log.Info(fmt.Sprintf("Waiting for KKInstance %s to reconcile", info.KKInstanceName))
		return false
	}

	// Adds owner reference, namespace and name
	sema.setMetadata(cluster)
	// Adds the additional information
	if err := sema.setInformation(&information{KKInstanceName: kkInstance.Name}); err != nil {
		log.Error(err, "Failed to acquire init lock while setting semaphore information")
		return false
	}

	log.Info("Attempting to acquire the lock")
	err = m.client.Create(ctx, sema.ConfigMap)
	switch {
	case apierrors.IsAlreadyExists(err):
		log.Info("Cannot acquire the lock. The lock has been acquired by someone else")
		return false
	case err != nil:
		log.Error(err, "Error acquiring the init lock")
		return false
	default:
		return true
	}
}

// Unlock releases the lock.
func (m *Mutex) Unlock(ctx context.Context, cluster *clusterv1.Cluster) bool {
	sema := newSemaphore()
	cmName := configMapName(cluster.Name)
	log := ctrl.LoggerFrom(ctx, "ConfigMap", klog.KRef(cluster.Namespace, cmName))
	err := m.client.Get(ctx, client.ObjectKey{
		Namespace: cluster.Namespace,
		Name:      cmName,
	}, sema.ConfigMap)
	switch {
	case apierrors.IsNotFound(err):
		log.Info("Control plane lock not found, it may have been released already")
		return true
	case err != nil:
		log.Error(err, "Error unlocking the control plane lock")
		return false
	default:
		// Delete the config map semaphore if there is no error fetching it
		if err := m.client.Delete(ctx, sema.ConfigMap); err != nil {
			if apierrors.IsNotFound(err) {
				return true
			}
			log.Error(err, "Error deleting the config map underlying the control plane lock")
			return false
		}
		return true
	}
}

type information struct {
	KKInstanceName string `json:"kkInstanceName"`
}

type semaphore struct {
	*corev1.ConfigMap
}

func newSemaphore() *semaphore {
	return &semaphore{&corev1.ConfigMap{}}
}

func configMapName(clusterName string) string {
	return fmt.Sprintf("%s-lock", clusterName)
}

func (s semaphore) information() (*information, error) {
	li := &information{}
	if err := json.Unmarshal([]byte(s.Data[semaphoreInformationKey]), li); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal semaphore information")
	}
	return li, nil
}

func (s semaphore) setInformation(information *information) error {
	b, err := json.Marshal(information)
	if err != nil {
		return errors.Wrap(err, "failed to marshal semaphore information")
	}
	s.Data = map[string]string{}
	s.Data[semaphoreInformationKey] = string(b)
	return nil
}

func (s *semaphore) setMetadata(cluster *clusterv1.Cluster) {
	s.ObjectMeta = metav1.ObjectMeta{
		Namespace: cluster.Namespace,
		Name:      configMapName(cluster.Name),
		Labels: map[string]string{
			clusterv1.ClusterLabelName: cluster.Name,
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion: cluster.APIVersion,
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			},
		},
	}
}
