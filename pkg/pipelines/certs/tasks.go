package certs

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
)

type Certificate struct {
	Name          string
	Expires       string
	Residual      string
	AuthorityName string
	NodeName      string
}

type CaCertificate struct {
	AuthorityName string
	Expires       string
	Residual      string
	NodeName      string
}

var (
	certificateList = []string{
		"apiserver.crt",
		"apiserver-kubelet-client.crt",
		"front-proxy-client.crt",
	}
	caCertificateList = []string{
		"ca.crt",
		"front-proxy-ca.crt",
	}
	kubeConfigList = []string{
		"admin.conf",
		"controller-manager.conf",
		"scheduler.conf",
	}
	certificates   []*Certificate
	caCertificates []*CaCertificate
)

type ListClusterCerts struct {
	common.KubeAction
}

func (l *ListClusterCerts) Execute(runtime connector.Runtime) error {
	//for _, certFileName := range certificateList {
	//	certPath := filepath.Join(common.KubeCertDir, certFileName)
	//	certContext, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("cat %s", certPath), false)
	//	if err != nil {
	//		return errors.Wrap(err, "get cluster certs failed")
	//	}
	//	if cert, err := getCertInfo(certContext, certFileName, node.Name); err != nil {
	//		return err
	//	} else {
	//		certificates = append(certificates, cert)
	//	}
	//}
	return nil
}
