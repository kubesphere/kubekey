package prepare

type FastPrepare struct {
	BasePrepare
	Inject func() (bool, error)
}

func (b *FastPrepare) PreCheck() (bool, error) {
	return b.Inject()
}
