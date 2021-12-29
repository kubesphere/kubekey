package etcd

import (
	"crypto/x509"
	"fmt"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/utils/certs"
	"k8s.io/client-go/util/cert"
	certutil "k8s.io/client-go/util/cert"
	netutils "k8s.io/utils/net"
	"net"
	"path/filepath"
	"strings"
)

// KubekeyCertEtcdCA is the definition of the root CA used by the hosted etcd server.
func KubekeyCertEtcdCA() *certs.KubekeyCert {
	return &certs.KubekeyCert{
		Name:     "etcd-ca",
		LongName: "self-signed CA to provision identities for etcd",
		BaseName: "ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				CommonName: "etcd-ca",
			},
		},
	}
}

// KubekeyCertEtcdAdmin is the definition of the cert for etcd admin.
func KubekeyCertEtcdAdmin(hostname string, altNames *certutil.AltNames) *certs.KubekeyCert {
	l := strings.Split(hostname, ".")
	return &certs.KubekeyCert{
		Name:     "etcd-admin",
		LongName: "certificate for etcd admin",
		BaseName: fmt.Sprintf("admin-%s", hostname),
		CAName:   "etcd-ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
				AltNames:   *altNames,
				CommonName: fmt.Sprintf("etcd-admin-%s", l[0]),
			},
		},
	}
}

// KubekeyCertEtcdMember is the definition of the cert for etcd member.
func KubekeyCertEtcdMember(hostname string, altNames *certutil.AltNames) *certs.KubekeyCert {
	l := strings.Split(hostname, ".")
	return &certs.KubekeyCert{
		Name:     "etcd-member",
		LongName: "certificate for etcd member",
		BaseName: fmt.Sprintf("member-%s", hostname),
		CAName:   "etcd-ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
				AltNames:   *altNames,
				CommonName: fmt.Sprintf("etcd-member-%s", l[0]),
			},
		},
	}
}

// KubekeyCertEtcdMember is the definition of the cert for etcd client.
func KubekeyCertEtcdClient(hostname string, altNames *certutil.AltNames) *certs.KubekeyCert {
	l := strings.Split(hostname, ".")
	return &certs.KubekeyCert{
		Name:     "etcd-client",
		LongName: "certificate for etcd client",
		BaseName: fmt.Sprintf("node-%s", hostname),
		CAName:   "etcd-ca",
		Config: certs.CertConfig{
			Config: certutil.Config{
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
				AltNames:   *altNames,
				CommonName: fmt.Sprintf("etcd-node-%s", l[0]),
			},
		},
	}
}

type FetchCerts struct {
	common.KubeAction
}

func (f *FetchCerts) Execute(runtime connector.Runtime) error {
	src := "/etc/ssl/etcd/ssl"
	dst := fmt.Sprintf("%s/pki/etcd", runtime.GetWorkDir())
	if v, ok := f.PipelineCache.Get(common.ETCDCluster); ok {
		c := v.(*EtcdCluster)

		if c.clusterExist {
			if certs, err := runtime.GetRunner().SudoCmd("ls /etc/ssl/etcd/ssl/ | grep .pem", false); err != nil {
				return err
			} else {
				certsList := strings.Split(certs, "\r\n")
				if len(certsList) > 0 {
					for _, cert := range certsList {
						if err := runtime.GetRunner().Fetch(filepath.Join(dst, cert), filepath.Join(src, cert)); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

type GenerateCerts struct {
	common.KubeAction
}

func (g *GenerateCerts) Execute(runtime connector.Runtime) error {
	var pkiPath string
	if g.KubeConf.Arg.CertificatesDir == "" {
		pkiPath = fmt.Sprintf("%s/pki/etcd", runtime.GetWorkDir())
	}

	var altName cert.AltNames

	dnsList := []string{"localhost", "etcd.kube-system.svc.cluster.local", "etcd.kube-system.svc", "etcd.kube-system", "etcd"}
	ipList := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}

	if g.KubeConf.Cluster.ControlPlaneEndpoint.Domain == "" {
		dnsList = append(dnsList, kubekeyapiv1alpha2.DefaultLBDomain)
	} else {
		dnsList = append(dnsList, g.KubeConf.Cluster.ControlPlaneEndpoint.Domain)
	}

	for _, host := range g.KubeConf.Cluster.Hosts {
		dnsList = append(dnsList, host.Name)
		internalAddress := netutils.ParseIPSloppy(host.InternalAddress)
		if internalAddress != nil {
			ipList = append(ipList, internalAddress)
		}
	}

	altName.DNSNames = dnsList
	altName.IPs = ipList

	files := []string{"ca.pem", "ca-key.pem"}

	// CA
	certsList := []*certs.KubekeyCert{KubekeyCertEtcdCA()}

	// Certs
	for _, host := range runtime.GetAllHosts() {
		if host.IsRole(common.ETCD) {
			certsList = append(certsList, KubekeyCertEtcdAdmin(host.GetName(), &altName))
			files = append(files, []string{fmt.Sprintf("admin-%s.pem", host.GetName()), fmt.Sprintf("admin-%s-key.pem", host.GetName())}...)
			certsList = append(certsList, KubekeyCertEtcdMember(host.GetName(), &altName))
			files = append(files, []string{fmt.Sprintf("member-%s.pem", host.GetName()), fmt.Sprintf("member-%s-key.pem", host.GetName())}...)
		}
		if host.IsRole(common.Master) {
			certsList = append(certsList, KubekeyCertEtcdClient(host.GetName(), &altName))
			files = append(files, []string{fmt.Sprintf("node-%s.pem", host.GetName()), fmt.Sprintf("node-%s-key.pem", host.GetName())}...)
		}
	}

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
