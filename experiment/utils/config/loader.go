package config

const (
	LOCAL    = "local"
	OPERATOR = "operator"
)

type Loader interface {
	Load(manager *Manager) error
}

func NewLoader(flag string) Loader {
	switch flag {
	case LOCAL:
		return &YamlFileLoader{}
	case OPERATOR:
		return &ConfigMapLoader{}
	default:
		return &YamlFileLoader{}
	}
}

type YamlFileLoader struct {
}

func (y YamlFileLoader) Load(manager *Manager) error {
	return nil
}

type ConfigMapLoader struct {
}

func (c ConfigMapLoader) Load(manager *Manager) error {
	return nil
}
