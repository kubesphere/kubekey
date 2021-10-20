package kubernetes

type UpgradeStep int

const (
	ToV121 UpgradeStep = iota + 1
	ToV122
)

var UpgradeStepList = []UpgradeStep{
	ToV121,
	ToV122,
}

func (u UpgradeStep) String() string {
	switch u {
	case ToV121:
		return "to v1.21"
	case ToV122:
		return "to v1.22"
	default:
		return "invalid option"
	}
}
