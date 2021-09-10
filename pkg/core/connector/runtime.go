package connector

import (
	"github.com/kubesphere/kubekey/pkg/core/common"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type BaseRuntime struct {
	ObjName   string
	connector Connector
	runner    *Runner
	workDir   string
	verbose   bool
	allHosts  []Host
	roleHosts map[string][]Host
}

func NewBaseRuntime(name string, connector Connector, verbose bool) BaseRuntime {
	return BaseRuntime{
		ObjName:   name,
		connector: connector,
		verbose:   verbose,
		allHosts:  make([]Host, 0, 0),
		roleHosts: make(map[string][]Host),
	}
}

func (b *BaseRuntime) GetObjName() string {
	return b.ObjName
}

func (b *BaseRuntime) SetObjName(name string) {
	b.ObjName = name
}

func (b *BaseRuntime) GetRunner() *Runner {
	return b.runner
}

func (b *BaseRuntime) SetRunner(r *Runner) {
	b.runner = r
}

func (b *BaseRuntime) GetConnector() Connector {
	return b.connector
}

func (b *BaseRuntime) SetConnector(c Connector) {
	b.connector = c
}

func (b *BaseRuntime) GenerateWorkDir() error {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "get current dir failed")
	}

	rootPath := filepath.Join(currentDir, common.KubeKey)
	if err := util.CreateDir(rootPath); err != nil {
		return errors.Wrap(err, "create work dir failed")
	}
	b.workDir = rootPath

	logDir := filepath.Join(rootPath, "logs")
	if err := util.CreateDir(logDir); err != nil {
		return errors.Wrap(err, "create logs dir failed")
	}

	for i := range b.allHosts {
		subPath := filepath.Join(rootPath, b.allHosts[i].GetName())
		if err := util.CreateDir(subPath); err != nil {
			return errors.Wrap(err, "create work dir failed")
		}
	}
	return nil
}

func (b *BaseRuntime) GetHostWorkDir() string {
	return filepath.Join(b.workDir, b.RemoteHost().GetName())
}

func (b *BaseRuntime) GetWorkDir() string {
	return b.workDir
}

func (b *BaseRuntime) GetAllHosts() []Host {
	return b.allHosts
}

func (b *BaseRuntime) SetAllHosts(hosts []Host) {
	b.allHosts = hosts
}

func (b *BaseRuntime) GetHostsByRole(role string) []Host {
	return b.roleHosts[role]
}

func (b *BaseRuntime) RemoteHost() Host {
	return b.GetRunner().Host
}

func (b *BaseRuntime) InitLogger() error {
	if b.GetWorkDir() == "" {
		if err := b.GenerateWorkDir(); err != nil {
			return err
		}
	}
	logDir := filepath.Join(b.GetWorkDir(), "logs")
	logger.Log = logger.NewLogger(logDir, b.verbose)
	return nil
}

func (b *BaseRuntime) Copy() Runtime {
	runtime := *b
	return &runtime
}

func (b *BaseRuntime) GenerateRoleMap() {
	for i := range b.allHosts {
		b.AppendRoleMap(b.allHosts[i])
	}
}

func (b *BaseRuntime) AppendHost(host Host) {
	b.allHosts = append(b.allHosts, host)
}

func (b *BaseRuntime) AppendRoleMap(host Host) {
	for _, r := range host.GetRoles() {
		if hosts, ok := b.roleHosts[r]; ok {
			hosts = append(hosts, host)
			b.roleHosts[r] = hosts
		} else {
			first := make([]Host, 0, 0)
			first = append(first, host)
			b.roleHosts[r] = first
		}
	}
}
