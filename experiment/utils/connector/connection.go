package connector

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"io"
)

type Connection interface {
	Exec(cmd string) (stdout string, stderr string, code int, err error)
	PExec(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (code int, err error)
	Scp(src, dst string) error
	Close()
}

type Connector interface {
	Connect(host kubekeyapiv1alpha1.HostCfg) (Connection, error)
}
