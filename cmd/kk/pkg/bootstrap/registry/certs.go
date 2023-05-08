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

package registry

import (
	"crypto/x509"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/cert"
	certutil "k8s.io/client-go/util/cert"
	netutils "k8s.io/utils/net"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/utils/certs"
)

const (
	RegistryCertificateBaseName = "dockerhub.kubekey.local"
	LocalCertsDir               = "localCertsDir"
	CertsFileList               = "certsFileList"
)

// KubekeyCertEtcdCA is the definition of the root CA used by the hosted etcd server.
func KubekeyCertRegistryCA() *certs.KubekeyCert {
	return &certs.KubekeyCert{
		Name:     "registry-ca",
		LongName: "self-signed CA to provision identities for registry",
		BaseName: "ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				CommonName: "registry-ca",
			},
		},
	}
}

// KubekeyCertEtcdAdmin is the definition of the cert for etcd admin.
func KubekeyCertRegistryServer(baseName string, altNames *certutil.AltNames) *certs.KubekeyCert {
	return &certs.KubekeyCert{
		Name:     "registry-server",
		LongName: "certificate for registry server",
		BaseName: baseName,
		CAName:   "registry-ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
				AltNames:   *altNames,
				CommonName: baseName,
			},
		},
	}
}

type FetchCerts struct {
	common.KubeAction
}

func (f *FetchCerts) Execute(runtime connector.Runtime) error {
	src := "/etc/ssl/registry/ssl"
	dst := fmt.Sprintf("%s/pki/registry", runtime.GetWorkDir())

	certs, err := runtime.GetRunner().SudoCmd("ls /etc/ssl/registry/ssl/ | grep .pem", false)
	if err != nil {
		return nil
	}

	certsList := strings.Split(certs, "\r\n")
	if len(certsList) > 0 {
		for _, cert := range certsList {
			if err := runtime.GetRunner().Fetch(filepath.Join(dst, cert), filepath.Join(src, cert)); err != nil {
				return errors.Wrap(err, fmt.Sprintf("Fetch %s failed", filepath.Join(src, cert)))
			}
		}
	}

	return nil
}

type GenerateCerts struct {
	common.KubeAction
}

func (g *GenerateCerts) Execute(runtime connector.Runtime) error {

	pkiPath := fmt.Sprintf("%s/pki/registry", runtime.GetWorkDir())

	var altName cert.AltNames

	dnsList := []string{"localhost", g.KubeConf.Cluster.Registry.PrivateRegistry, runtime.GetHostsByRole(common.Registry)[0].GetName()}
	ipList := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback, netutils.ParseIPSloppy(runtime.GetHostsByRole(common.Registry)[0].GetInternalAddress())}

	altName.DNSNames = dnsList
	altName.IPs = ipList

	files := []string{"ca.pem", "ca-key.pem", fmt.Sprintf("%s.pem", g.KubeConf.Cluster.Registry.PrivateRegistry), fmt.Sprintf("%s-key.pem", g.KubeConf.Cluster.Registry.PrivateRegistry)}

	// CA
	certsList := []*certs.KubekeyCert{KubekeyCertRegistryCA()}

	// Certs
	certsList = append(certsList, KubekeyCertRegistryServer(g.KubeConf.Cluster.Registry.PrivateRegistry, &altName))

	var lastCACert *certs.KubekeyCert
	for _, c := range certsList {
		if c.CAName == "" {
			err := certs.GenerateCA(c, pkiPath, g.KubeConf)
			if err != nil {
				return err
			}
			lastCACert = c
		} else {
			err := certs.GenerateCerts(c, lastCACert, pkiPath, g.KubeConf)
			if err != nil {
				return err
			}
		}
	}

	g.ModuleCache.Set(LocalCertsDir, pkiPath)
	g.ModuleCache.Set(CertsFileList, files)

	return nil
}
