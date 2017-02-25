package reflector

import "fmt"

// Call the received function.
func (f Function) Call(args ...interface{}) error {
	if !f.IsValid() {
		return fmt.Errorf("invalid function")
	}
	return f.Caller.Call(args...)
}

// IsValid returns true if f represents a Function.
// It returns false if f is the zero Value.
func (f Function) IsValid() bool {
	return f.Caller != nil
}
