/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import (
	"strconv"
	"unsafe"
)

// valueToString returns a textual representation of the reflection value val.
// For debugging only.
func ValueToString(v Value) string {
	var str string

	if !v.IsValid() {
		return "<zero Value>"
	}
	switch v.Kind() {
	case Int, Int8, Int16, Int32, Int64:
		return strconv.FormatInt(v.Int().Get(), 10)
	case Uint, Uint8, Uint16, Uint32, Uint64, UintPtr:
		return strconv.FormatUint(v.Uint().Get(), 10)
	case Float32, Float64:
		return strconv.FormatFloat(v.Float().Get(), 'g', -1, 64)
	case Complex64, Complex128:
		c := v.Complex().Get()
		return strconv.FormatFloat(real(c), 'g', -1, 64) + "+" + strconv.FormatFloat(imag(c), 'g', -1, 64) + "i"
	case String:
		strVal := v.String()
		if strVal.Debug != "" {
			return strVal.Debug
		}
		return strVal.Get()
	case Bool:
		if v.Bool().Get() {
			return "true"
		} else {
			return "false"
		}
	case Ptr:
		str = TypeToString(v.Type) + "("
		if v.IsNil() {
			str += "0"
		} else {
			str += "&" + ValueToString(v.Deref())
		}
		str += ")"
		return str
	case Array, Slice:
		str += TypeToString(v.Type)
		str += "{"
		slice := ToSlice(v)
		for i := 0; i < slice.Len(); i++ {
			if i > 0 {
				str += ", "
			}
			str += ValueToString(slice.Index(i))
		}
		str += "}"
		return str
	case Map:
		str = TypeToString(v.Type)
		str += "{"
		str += "<can't iterate on maps>"
		str += "}"
		return str
	case Struct:
		vx := StructValue{Value: v}
		numFields := len(v.Type.convToStruct().fields)
		str += TypeToString(v.Type)
		str += "{"
		for i, n := 0, numFields; i < n; i++ {
			if i > 0 {
				str += ", "
			}
			str += ValueToString(vx.Field(i))
		}
		str += "}"
		return str
	case Interface:
		return TypeToString(v.Type) + "(" + ValueToString(v.Iface()) + ")"
	case Func:
		return TypeToString(v.Type) + "(" + strconv.FormatUint(uint64(v.Pointer()), 10) + ")"
	case Chan:
		str = TypeToString(v.Type)
		return str
	default:
		return "ValueToString ERROR : can't print type " + TypeToString(v.Type)
	}
}

// NewAt returns a Value representing a pointer to a value of the
// specified type, using p as that pointer.
func NewAt(typ *RType, p unsafe.Pointer) Value {
	return Value{Type: typ.PtrTo(), Ptr: p, Flag: Flag(Ptr)}
}

// MakeRO returns a copy of v with the read-only flag set.
func MakeRO(val Value) Value {
	v := val
	v.Flag |= stickyROFlag
	return v
}

func FuncLayout(tt, trcvr *RType) (frametype *RType, argSize, retOffset uintptr, stack []byte, gc []byte, ptrs bool) {
	t := tt
	var rcvr *RType
	if trcvr != nil {
		rcvr = trcvr
	}
	var ft *RType
	var s *bitVector
	if rcvr != nil {
		ft, argSize, retOffset, s = funcLayout(t, rcvr)
	} else {
		ft, argSize, retOffset, s = funcLayout(t, nil)
	}
	frametype = ft
	for i := uint32(0); i < s.num; i++ {
		stack = append(stack, s.data[i/8]>>(i%8)&1)
	}
	if !ft.canHandleGC() {
		panic("can't handle gc programs")
	}
	gcdata := (*[1000]byte)(unsafe.Pointer(ft.gcData))
	for i := uintptr(0); i < ft.ptrData/PtrSize; i++ {
		gc = append(gc, gcdata[i/8]>>(i%8)&1)
	}
	ptrs = ft.hasPointers()
	return
}

func TypeLinks() []string {
	var r []string
	sections, offset := typeLinks()
	for i, offs := range offset {
		rodata := sections[i]
		for _, off := range offs {
			typ := (*RType)(resolveTypeOff(unsafe.Pointer(rodata), off))
			r = append(r, TypeToString(typ))
		}
	}
	return r
}

var GCBits = gcbits

//go:linkname gcbits reflect.gcbits
func gcbits(interface{}) []byte // provided by runtime

func MapBucketOf(x, y *RType) *RType {
	return bucketOf(x, y)
}

func CachedBucketOf(t *RType) *RType {
	m := t
	if m.Kind() != Map {
		//panic("not map")
		return nil
	}
	return m.ConvToMap().bucket
}

type EmbedWithUnexpMeth struct{}

func (EmbedWithUnexpMeth) f() {}

type pinUnexpMeth interface {
	f()
}

var pinUnexpMethI = pinUnexpMeth(EmbedWithUnexpMeth{})

func FirstMethodNameBytes(ti *RType) *byte {
	_ = pinUnexpMethI
	t := ti
	m, ok := methods(t)
	if !ok {
		panic("there are no methods.")
	}
	mname := t.nameOffset(m[0].nameOffset)
	if *mname.data(0)&(1<<2) == 0 {
		panic("method name does not have pkgPath *string")
	}
	return mname.bytes
}

type OtherPkgFields struct {
	OtherExported   int
	otherUnexported int
}

func IsExported(t *RType) bool {
	n := t.nameOffsetStr()
	return n.isExported()
}

func ResolveReflectName(s string) {
	declareReflectName(newName([]byte(s)))
}

type Buffer struct {
	buf []byte
}
