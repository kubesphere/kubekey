package modules

type CustomModule struct {
	BaseModule
}

func (t *CustomModule) Is() string {
	return CustomModuleType
}
