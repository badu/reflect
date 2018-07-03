/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import "unsafe"

func (fn *funcType) inParams() []*RType {
	if fn.InLen == 0 {
		return nil
	}
	offset := unsafe.Sizeof(*fn)
	if fn.hasInfoFlag() {
		offset += additionalOffset
	}
	return (*[1 << 20]*RType)(funcOffset(fn, offset))[:fn.InLen]
}

func (fn *funcType) outParams() []*RType {
	outCount := fn.OutLen & (1<<15 - 1)
	if outCount == 0 {
		return nil
	}
	offset := unsafe.Sizeof(*fn)
	if fn.hasInfoFlag() {
		offset += additionalOffset
	}
	return (*[1 << 20]*RType)(funcOffset(fn, offset))[fn.InLen : fn.InLen+outCount]
}

func (fn *funcType) inParam(i int) *RType {
	offset := unsafe.Sizeof(*fn)
	if fn.hasInfoFlag() {
		offset += additionalOffset
	}
	if fn.InLen == 0 {
		return nil
	}
	return (*[1 << 20]*RType)(funcOffset(fn, offset))[:fn.InLen][i]
}

func (fn *funcType) outParam(i int) *RType {
	offset := unsafe.Sizeof(*fn)
	if fn.hasInfoFlag() {
		offset += additionalOffset
	}
	outCount := fn.OutLen & (1<<15 - 1) // fn.OutLen & (1<<15) != 0 meaning variadic
	if outCount == 0 {
		return nil
	}
	return (*[1 << 20]*RType)(funcOffset(fn, offset))[fn.InLen : fn.InLen+outCount][i]
}
