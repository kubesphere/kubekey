package hook

func Call(hook Interface) error {
	err := hook.Catch(hook.Try())
	hook.Finally()
	return err
}
