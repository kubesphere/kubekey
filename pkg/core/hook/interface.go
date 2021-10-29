package hook

type Interface interface {
	Try() error
	Catch(err error) error
	Finally()
}
