/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import (
	"unsafe"
)

// NewAt returns a Value representing a pointer to a value of the
// specified type, using p as that pointer.
func NewAt(typ *RType, p ptr) Value {
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
