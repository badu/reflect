/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package old

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
