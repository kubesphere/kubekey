package connector

import (
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
	Connect(host Host) (Connection, error)
}

type Runtime interface {
	GetRunner() *Runner
	SetRunner(r *Runner)
	GetConnector() Connector
	SetConnector(c Connector)
	GetWorkDir() string
	SetWorkDir(str string)
	GetAllHosts() []Host
	SetAllHosts([]Host)
	GetHostsByRole(role string) []Host
	RemoteHost() Host
	Copy() Runtime
}

type Host interface {
	GetName() string
	SetName(name string)
	GetAddress() string
	SetAddress(str string)
	GetPort() int
	SetPort(port int)
	GetUser() string
	SetUser(u string)
	GetPassword() string
	SetPassword(password string)
	GetPrivateKey() string
	SetPrivateKey(privateKey string)
	GetPrivateKeyPath() string
	SetPrivateKeyPath(path string)
	GetArch() string
	SetArch(arch string)
	GetRoles() []string
	SetRoles(roles []string)
	IsRole(role string) bool
	Copy() Host
}
