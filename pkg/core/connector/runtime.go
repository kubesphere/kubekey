package connector

type BaseRuntime struct {
	ObjName         string
	connector       Connector
	runner          *Runner
	DownloadCommand func(path, url string) string
	WorkDir         string
	ClusterHosts    []string
	AllHosts        []Host
	RoleHosts       map[string][]Host
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
	return b.WorkDir
}

func (b *BaseRuntime) SetWorkDir(str string) {
	b.WorkDir = str
}

func (b *BaseRuntime) GetAllHosts() []Host {
	return b.AllHosts
}

func (b *BaseRuntime) SetAllHosts(hosts []Host) {
	b.AllHosts = hosts
}

func (b *BaseRuntime) GetHostsByRole(role string) []Host {
	return b.RoleHosts[role]
}

func (b *BaseRuntime) RemoteHost() Host {
	return b.GetRunner().Host
}

func (b *BaseRuntime) Copy() Runtime {
	runtime := *b
	return &runtime
}

func (b *BaseRuntime) GenerateRoleMap() {
	for i := range b.AllHosts {
		b.AppendRoleMap(b.AllHosts[i])
	}
}

func (b *BaseRuntime) AppendHost(host Host) {
	b.AllHosts = append(b.AllHosts, host)
}

func (b *BaseRuntime) AppendRoleMap(host Host) {
	for _, r := range host.GetRoles() {
		if hosts, ok := b.RoleHosts[r]; ok {
			hosts = append(hosts, host)
			b.RoleHosts[r] = hosts
		} else {
			first := make([]Host, 0, 0)
			first = append(first, host)
			b.RoleHosts[r] = first
		}
	}
}
