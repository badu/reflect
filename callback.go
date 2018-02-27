package reflector

func (f callback) Call(args ...interface{}) error {
	return f(args...)
}
