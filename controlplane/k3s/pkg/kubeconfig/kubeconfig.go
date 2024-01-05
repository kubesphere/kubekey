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

// Package kubeconfig implements the kubeconfig generation logic.
package kubeconfig

import (
	"context"
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/certs"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v3/util/secret"
)

var (
	// ErrDependentCertificateNotFound is returned when the dependent certificate is not found.
	ErrDependentCertificateNotFound = errors.New("could not find secret ca")
)

// New creates a new Kubeconfig using the cluster name and specified endpoint.
func New(clusterName, endpoint string, clientCACert *x509.Certificate, clientCAKey crypto.Signer, serverCACert *x509.Certificate) (*api.Config, error) {
	cfg := &certs.Config{
		CommonName:   "kubernetes-admin",
		Organization: []string{"system:masters"},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientKey, err := certs.NewPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create private key")
	}

	clientCert, err := cfg.NewSignedCert(clientKey, clientCACert, clientCAKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign certificate")
	}

	userName := fmt.Sprintf("%s-admin", clusterName)
	contextName := fmt.Sprintf("%s@%s", userName, clusterName)

	return &api.Config{
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                   endpoint,
				CertificateAuthorityData: certs.EncodeCertPEM(serverCACert),
			},
		},
		Contexts: map[string]*api.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: userName,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			userName: {
				ClientKeyData:         certs.EncodePrivateKeyPEM(clientKey),
				ClientCertificateData: certs.EncodeCertPEM(clientCert),
			},
		},
		CurrentContext: contextName,
	}, nil
}

// CreateSecret creates the Kubeconfig secret for the given cluster.
func CreateSecret(ctx context.Context, c client.Client, cluster *clusterv1.Cluster) error {
	name := util.ObjectKey(cluster)
	return CreateSecretWithOwner(ctx, c, name, cluster.Spec.ControlPlaneEndpoint.String(), metav1.OwnerReference{
		APIVersion: clusterv1.GroupVersion.String(),
		Kind:       "Cluster",
		Name:       cluster.Name,
		UID:        cluster.UID,
	})
}

// CreateSecretWithOwner creates the Kubeconfig secret for the given cluster name, namespace, endpoint, and owner reference.
func CreateSecretWithOwner(ctx context.Context, c client.Client, clusterName client.ObjectKey, endpoint string, owner metav1.OwnerReference) error {
	server := fmt.Sprintf("https://%s", endpoint)
	out, err := generateKubeconfig(ctx, c, clusterName, server)
	if err != nil {
		return err
	}

	return c.Create(ctx, GenerateSecretWithOwner(clusterName, out, owner))
}

// GenerateSecret returns a Kubernetes secret for the given Cluster and kubeconfig data.
func GenerateSecret(cluster *clusterv1.Cluster, data []byte) *corev1.Secret {
	name := util.ObjectKey(cluster)
	return GenerateSecretWithOwner(name, data, metav1.OwnerReference{
		APIVersion: clusterv1.GroupVersion.String(),
		Kind:       "Cluster",
		Name:       cluster.Name,
		UID:        cluster.UID,
	})
}

// GenerateSecretWithOwner returns a Kubernetes secret for the given Cluster name, namespace, kubeconfig data, and ownerReference.
func GenerateSecretWithOwner(clusterName client.ObjectKey, data []byte, owner metav1.OwnerReference) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name(clusterName.Name, secret.Kubeconfig),
			Namespace: clusterName.Namespace,
			Labels: map[string]string{
				clusterv1.ClusterNameLabel: clusterName.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				owner,
			},
		},
		Data: map[string][]byte{
			secret.KubeconfigDataName: data,
		},
	}
}

func generateKubeconfig(ctx context.Context, c client.Client, clusterName client.ObjectKey, endpoint string) ([]byte, error) {
	clusterCA, err := secret.GetFromNamespacedName(ctx, c, clusterName, secret.ClusterCA)
	if err != nil {
		if apierrors.IsNotFound(errors.Cause(err)) {
			return nil, ErrDependentCertificateNotFound
		}
		return nil, err
	}

	clientClusterCA, err := secret.GetFromNamespacedName(ctx, c, clusterName, secret.ClientClusterCA)
	if err != nil {
		if apierrors.IsNotFound(errors.Cause(err)) {
			return nil, ErrDependentCertificateNotFound
		}
		return nil, err
	}

	clientCACert, err := certs.DecodeCertPEM(clientClusterCA.Data[secret.TLSCrtDataName])
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode CA Cert")
	} else if clientCACert == nil {
		return nil, errors.New("certificate not found in config")
	}

	clientCAKey, err := certs.DecodePrivateKeyPEM(clientClusterCA.Data[secret.TLSKeyDataName])
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode private key")
	} else if clientCAKey == nil {
		return nil, errors.New("CA private key not found")
	}

	serverCACert, err := certs.DecodeCertPEM(clusterCA.Data[secret.TLSCrtDataName])
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode CA Cert")
	} else if serverCACert == nil {
		return nil, errors.New("certificate not found in config")
	}

	cfg, err := New(clusterName.Name, endpoint, clientCACert, clientCAKey, serverCACert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate a kubeconfig")
	}

	out, err := clientcmd.Write(*cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize config to yaml")
	}
	return out, nil
}
