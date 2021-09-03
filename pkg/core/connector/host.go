package connector

type BaseHost struct {
	Name            string            `yaml:"name,omitempty" json:"name,omitempty"`
	Address         string            `yaml:"address,omitempty" json:"address,omitempty"`
	InternalAddress string            `yaml:"internalAddress,omitempty" json:"internalAddress,omitempty"`
	Port            int               `yaml:"port,omitempty" json:"port,omitempty"`
	User            string            `yaml:"user,omitempty" json:"user,omitempty"`
	Password        string            `yaml:"password,omitempty" json:"password,omitempty"`
	PrivateKey      string            `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`
	PrivateKeyPath  string            `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`
	Arch            string            `yaml:"arch,omitempty" json:"arch,omitempty"`
	Roles           []string          `json:"-"`
	RoleTable       map[string]bool   `json:"-"`
	Labels          map[string]string `json:"-"`
}

func (b *BaseHost) GetName() string {
	return b.Name
}

func (b *BaseHost) SetName(name string) {
	b.Name = name
}

func (b *BaseHost) GetAddress() string {
	return b.Address
}

func (b *BaseHost) SetAddress(str string) {
	b.Address = str
}

func (b *BaseHost) GetInternalAddress() string {
	return b.InternalAddress
}

func (b *BaseHost) SetInternalAddress(str string) {
	b.InternalAddress = str
}

func (b *BaseHost) GetPort() int {
	return b.Port
}

func (b *BaseHost) SetPort(port int) {
	b.Port = port
}

func (b *BaseHost) GetUser() string {
	return b.User
}

func (b *BaseHost) SetUser(u string) {
	b.User = u
}

func (b *BaseHost) GetPassword() string {
	return b.Password
}

func (b *BaseHost) SetPassword(password string) {
	b.Password = password
}

func (b *BaseHost) GetPrivateKey() string {
	return b.PrivateKey
}

func (b *BaseHost) SetPrivateKey(privateKey string) {
	b.PrivateKey = privateKey
}

func (b *BaseHost) GetPrivateKeyPath() string {
	return b.PrivateKeyPath
}

func (b *BaseHost) SetPrivateKeyPath(path string) {
	b.PrivateKeyPath = path
}

func (b *BaseHost) GetArch() string {
	return b.Arch
}

func (b *BaseHost) SetArch(arch string) {
	b.Arch = arch
}

func (b *BaseHost) GetRoles() []string {
	return b.Roles
}

func (b *BaseHost) SetRoles(roles []string) {
	b.Roles = roles
}

func (b *BaseHost) SetRole(role string) {
	b.RoleTable[role] = true
	b.Roles = append(b.Roles, role)
}

func (b *BaseHost) IsRole(role string) bool {
	if res, ok := b.RoleTable[role]; ok {
		return res
	} else {
		return false
	}
}

func (b *BaseHost) SetLabel(k, v string) {
	b.Labels[k] = v
}

func (b *BaseHost) GetLabel(k string) (string, bool) {
	v, ok := b.Labels[k]
	return v, ok
}

func (b *BaseHost) Copy() Host {
	host := *b
	return &host
}
