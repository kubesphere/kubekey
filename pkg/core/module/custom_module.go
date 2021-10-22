package module

type CustomModule struct {
	BaseModule
}

func (c *CustomModule) Init() {
	if c.Name == "" {
		c.Name = DefaultCustomModuleName
	}
}

func (c *CustomModule) Is() string {
	return CustomModuleType
}
