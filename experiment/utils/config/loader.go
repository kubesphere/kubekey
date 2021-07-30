package config

const (
	LOCAL    = "local"
	OPERATOR = "operator"
)

type Loader interface {
	Load(config *GlobalConfig) error
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

func (y YamlFileLoader) Load(config *GlobalConfig) error {
	return nil
}

type ConfigMapLoader struct {
}

func (c ConfigMapLoader) Load(config *GlobalConfig) error {
	return nil
}
