package connector

type BaseRuntime struct {
	ObjName   string
	connector Connector
	runner    *Runner
	workDir   string
	allHosts  []Host
	roleHosts map[string][]Host
}

func NewBaseRuntime(name string, connector Connector, workDir string) BaseRuntime {
	return BaseRuntime{
		ObjName:   name,
		connector: connector,
		workDir:   workDir,
		allHosts:  make([]Host, 0, 0),
		roleHosts: make(map[string][]Host),
	}
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

func (b *BaseRuntime) GetWorkDir() string {
	return b.workDir
}

func (b *BaseRuntime) SetWorkDir(str string) {
	b.workDir = str
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
