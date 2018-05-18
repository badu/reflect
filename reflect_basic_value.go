/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import (
	"math"
)

func (v UintValue) CanSet() bool {
	return v.flag != 0 && v.flag&(addressableFlag|exportFlag) == addressableFlag
}

func (v UintValue) Kind() Kind { return Kind(v.flag & kindMaskFlag) }

// Uint returns v's underlying value, as a uint64.
func (v UintValue) Get() uint64 {
	if v.ptr == nil {
		if willPrintDebug {
			println("reflect.x.error : invalid `uint` (nil pointer)")
		}
		return 0
	}
	switch v.Kind() {
	case Uint8:
		return uint64(*(*uint8)(v.ptr))
	case Uint16:
		return uint64(*(*uint16)(v.ptr))
	case Uint32:
		return uint64(*(*uint32)(v.ptr))
	case Uint64:
		return *(*uint64)(v.ptr)
	case UintPtr:
		return uint64(*(*uintptr)(v.ptr))
	default:
		return uint64(*(*uint)(v.ptr))
	}
}

// SetUint sets v's underlying value to x.
// Fails silently if value is not settable
func (v UintValue) Set(value uint64) bool {
	if v.CanSet() {
		switch v.Kind() {
		case Uint8:
			*(*uint8)(v.ptr) = uint8(value)
		case Uint16:
			*(*uint16)(v.ptr) = uint16(value)
		case Uint32:
			*(*uint32)(v.ptr) = uint32(value)
		case Uint64:
			*(*uint64)(v.ptr) = uint64(value)
		case UintPtr:
			*(*uintptr)(v.ptr) = uintptr(value)
		default:
			*(*uint)(v.ptr) = uint(value)
		}
		return true
	}
	if willPrintDebug {
		println("reflect.x.error : trying to set not settable `uint`")
	}
	return false
}

func (v UintValue) Overflows(x uint64) bool {
	return x != (x<<(64-v.size))>>(64-v.size)
}

func (v Value) Uint() UintValue {
	k := v.Kind()
	if k < Uint || k > UintPtr {
		if willPrintDebug {
			panic("reflect.x.error : error attempting to convert `" + StringKind(k) + "` to `uint`")
		}
		// return empty and non settable value
		return UintValue{}
	}
	return UintValue{BasicValue{ptr: v.Ptr, flag: v.Flag, size: v.Type.size * 8}}
}

// generic int returns the largest available for conversions (to float, string)
func (v IntValue) CanSet() bool {
	return v.flag != 0 && v.flag&(addressableFlag|exportFlag) == addressableFlag
}

func (v IntValue) Kind() Kind { return Kind(v.flag & kindMaskFlag) }

func (v IntValue) Get() int64 {
	if v.ptr == nil {
		if willPrintDebug {
			println("reflect.x.error : invalid `int` (nil pointer)")
		}
		return 0
	}
	switch v.Kind() {
	case Int8:
		return int64(*(*int8)(v.ptr))
	case Int16:
		return int64(*(*int16)(v.ptr))
	case Int32:
		return int64(*(*int32)(v.ptr))
	case Int64:
		return *(*int64)(v.ptr)
	default:
		return int64(*(*int)(v.ptr))
	}
}

func (v IntValue) Set(x int64) bool {
	if v.CanSet() {
		switch v.Kind() {
		case Int8:
			*(*int8)(v.ptr) = int8(x)
		case Int16:
			*(*int16)(v.ptr) = int16(x)
		case Int32:
			*(*int32)(v.ptr) = int32(x)
		case Int64:
			*(*int64)(v.ptr) = x
		default:
			*(*int)(v.ptr) = int(x)
		}
		return true
	}
	if willPrintDebug {
		println("reflect.x.error : trying to set not settable `int`")
	}
	return false
}

func (v IntValue) Overflows(x int64) bool {
	return x != (x<<(64-v.size))>>(64-v.size)
}

func (v Value) Int() IntValue {
	k := v.Kind()
	if k < Int || k > Int64 {
		if willPrintDebug {
			panic("reflect.x.error : error attempting to convert `" + StringKind(k) + "` to `int`")
		}
		// return empty and non settable value
		return IntValue{}
	}
	return IntValue{BasicValue{ptr: v.Ptr, flag: v.Flag, size: v.Type.size * 8}}
}

func (v FloatValue) CanSet() bool {
	return v.flag != 0 && v.flag&(addressableFlag|exportFlag) == addressableFlag
}
func (v FloatValue) Kind() Kind { return Kind(v.flag & kindMaskFlag) }

func (v FloatValue) Get() float64 {
	if v.ptr == nil {
		if willPrintDebug {
			println("reflect.x.error : invalid `float` (nil pointer)")
		}
		return 0
	}
	switch v.Kind() {
	case Float32:
		return float64(*(*float32)(v.ptr))
	default:
		return *(*float64)(v.ptr)
	}
}

func (v FloatValue) Set(x float64) bool {
	if v.CanSet() {
		switch v.Kind() {
		case Float32:
			*(*float32)(v.ptr) = float32(x)
		default:
			*(*float64)(v.ptr) = x
		}
		return true
	}
	if willPrintDebug {
		println("reflect.x.error : trying to set not settable `float`")
	}
	return false
}

func (v FloatValue) Overflows(x float64) bool {
	if Kind(v.flag&kindMaskFlag) == Float32 {
		if x < 0 {
			x = -x
		}
		return math.MaxFloat32 < x && x <= math.MaxFloat64
	}
	return false
}

func (v Value) Float() FloatValue {
	k := v.Kind()
	if k != Float64 && k != Float32 {
		if willPrintDebug {
			panic("reflect.x.error : error attempting to convert `" + StringKind(k) + "` to `float`")
		}
		// return empty and non settable value
		return FloatValue{}
	}
	return FloatValue{BasicValue{ptr: v.Ptr, flag: v.Flag, size: v.Type.size * 8}}
}

func (v ComplexValue) CanSet() bool {
	return v.flag != 0 && v.flag&(addressableFlag|exportFlag) == addressableFlag
}

func (v ComplexValue) Kind() Kind { return Kind(v.flag & kindMaskFlag) }

func (v ComplexValue) Get() complex128 {
	if v.ptr == nil {
		if willPrintDebug {
			println("reflect.x.error : invalid `complex` (nil pointer)")
		}
		return 0
	}
	switch v.Kind() {
	case Complex64:
		return complex128(*(*complex64)(v.ptr))
	default:
		return *(*complex128)(v.ptr)
	}
}

func (v ComplexValue) Set(x complex128) bool {
	if v.CanSet() {
		switch v.Kind() {
		case Complex64:
			*(*complex64)(v.ptr) = complex64(x)
		default:
			*(*complex128)(v.ptr) = x
		}
		return true
	}
	if willPrintDebug {
		println("reflect.x.error : trying to set not settable `complex`")
	}
	return false
}

func (v ComplexValue) Overflows(x complex128) bool {
	if v.Kind() == Complex64 {
		r, i := real(x), imag(x)
		if r < 0 {
			r = -r
		}
		if i < 0 {
			i = -i
		}
		return (math.MaxFloat32 < r && r <= math.MaxFloat64) || (math.MaxFloat32 < i && i <= math.MaxFloat64)
	}
	return false
}

func (v Value) Complex() ComplexValue {
	k := v.Kind()
	if k != Complex128 && k != Complex64 {
		if willPrintDebug {
			panic("reflect.x.error : error attempting to convert `" + StringKind(k) + "` to `complex`")
		}
		// return empty and non settable value
		return ComplexValue{}
	}
	return ComplexValue{BasicValue{ptr: v.Ptr, flag: v.Flag, size: v.Type.size * 8}}
}

func (v StringValue) CanSet() bool {
	return v.flag != 0 && v.flag&(addressableFlag|exportFlag) == addressableFlag
}

// Sets v's underlying value to x.
func (v StringValue) Set(x string) bool {
	if v.CanSet() {
		*(*string)(v.ptr) = x
		return true
	}
	if willPrintDebug {
		println("reflect.x.error : trying to set not settable `string`")
	}
	return false
}

func (v StringValue) Get() string {
	if v.ptr == nil {
		if willPrintDebug {
			println("reflect.x.error : invalid `string` (nil pointer)")
		}
		return ""
	}
	return *(*string)(v.ptr)
}

// String returns the string v's underlying value, as a string.
// String is a special case because of Go's String method convention.
// Unlike the other getters, it does not panic if v's Kind is not String.
// Instead, it returns a string of the form "<T value>" where T is v's type.
// The fmt package treats Values specially. It does not call their String
// method implicitly but instead prints the concrete values they hold.
func (v Value) String() StringValue {
	switch v.Kind() {
	case Invalid:
		return StringValue{Debug: "<invalid Value>"}
	case String:
		return StringValue{BasicValue: BasicValue{ptr: v.Ptr, flag: v.Flag, size: v.Type.size * 8}}
	default:
		if willPrintDebug {
			println("reflect.x.warning : error attempting to convert `" + StringKind(v.Kind()) + "` to `string`")
		}
		// If you call String on a reflect.Value of other type, it's better to
		// print something than to panic. Useful in debugging.
		if v.hasMethodFlag() {
			return StringValue{Debug: "<" + TypeToString(v.MethodType()) + " Value>"}
		} else {
			return StringValue{Debug: "<" + TypeToString(v.Type) + " Value>"}
		}

	}
}

// Bool returns v's underlying value.
func (v BoolValue) CanSet() bool {
	return v.flag != 0 && v.flag&(addressableFlag|exportFlag) == addressableFlag
}

func (v BoolValue) Get() bool {
	if v.ptr == nil {
		if willPrintDebug {
			println("reflect.x.error : invalid `bool` (nil pointer)")
		}
		return false
	}
	return *(*bool)(v.ptr)
}

// SetBool sets v's underlying value.
func (v BoolValue) Set(x bool) bool {
	if v.CanSet() {
		*(*bool)(v.ptr) = x
		return true
	}
	if willPrintDebug {
		println("reflect.x.error : trying to set not settable `bool`")
	}
	return false
}

func (v Value) Bool() BoolValue {
	if v.Kind() == Bool {
		return BoolValue{BasicValue{ptr: v.Ptr, flag: v.Flag, size: v.Type.size * 8}}
	}
	return BoolValue{}
}

func (v Value) UnsafePointer() PointerValue {
	switch v.Kind() {
	case UnsafePointer:
		return PointerValue{BasicValue{ptr: v.Ptr, flag: v.Flag, size: v.Type.size * 8}}
	case Ptr:
		return PointerValue{BasicValue{ptr: v.pointer(), flag: v.Flag, size: v.Type.size * 8}}
	default:
		if willPrintDebug {
			panic("reflect.x.error : error attempting to convert `" + StringKind(v.Kind()) + "` to `unsafe.Pointer` or `ptr`")
		}
		return PointerValue{}
	}
}

func (v PointerValue) Get() ptr {
	if v.ptr == nil {
		if willPrintDebug {
			println("reflect.x.error : invalid `unsafe.Pointer` (nil pointer)")
		}
		return nil
	}
	if v.flag&pointerFlag != 0 {
		return convPtr(v.ptr)
	}
	return v.ptr
}

func (v PointerValue) CanSet() bool {
	return v.flag != 0 && v.flag&(addressableFlag|exportFlag) == addressableFlag
}

// SetPointer sets the ptr value v to x.
func (v PointerValue) Set(x ptr) bool {
	if v.CanSet() {
		loadConvPtr(v.ptr, x)
		return true
	}
	if willPrintDebug {
		println("reflect.x.error : trying to set not settable `unsafe.Pointer`")
	}
	return false

}
