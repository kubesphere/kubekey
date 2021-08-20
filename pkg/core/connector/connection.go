package connector

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"io"
	"os"
)

type Connection interface {
	Exec(cmd string) (stdout string, stderr string, code int, err error)
	PExec(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (code int, err error)
	Fetch(local, remote string) error
	Scp(local, remote string) error
	RemoteFileExist(remote string) bool
	RemoteDirExist(remote string) (bool, error)
	MkDirAll(path string) error
	Chmod(path string, mode os.FileMode) error
	Close()
}

type Connector interface {
	Connect(host kubekeyapiv1alpha1.HostCfg) (Connection, error)
}
