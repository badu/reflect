/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import (
	"bytes"
	"unsafe"
)

func toIface(t unsafe.Pointer) *ifaceRtype { return (*ifaceRtype)(t) }
func funcOffset(t *funcType, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(t)) + x)
}
func structFieldOffset(f *structField) uintptr       { return f.offsetEmbed >> 1 }
func isEmbedded(f *structField) bool                 { return f.offsetEmbed&1 != 0 }
func declareReflectName(n name) int32                { return addReflectOff(unsafe.Pointer(n.bytes)) } // It returns a new nameOff that can be used to refer to the pointer.
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer { return unsafe.Pointer(uintptr(p) + x) }         // add returns p+x.

func byteSliceFromParams(params ...interface{}) []byte {
	result := make([]byte, 0)
	for _, param := range params {
		switch v := param.(type) {
		case string:
			result = append(result, []byte(v)...)
		case byte:
			result = append(result, v)
		case []byte:
			result = append(result, v...)
		default:
			panic("reflect.x.error : bad usage of the name builder.")
		}
	}
	return result
}

// fnv1 incorporates the list of bytes into the hash x using the FNV-1 hash function.
func fnv1(x uint32, list ...byte) uint32 {
	for _, b := range list {
		x = x*16777619 ^ uint32(b)
	}
	return x
}

// typesByString returns the subslice of typelinks() whose elements have
// the given string representation.
// It may be empty (no known types with that string) or may have
// multiple elements (multiple types with that string).
func typesByString(target []byte) []*RType {
	sections, offset := typeLinks()
	var results []*RType
	var currentType *RType
	var search []byte
	for offsI, offs := range offset {
		section := sections[offsI]

		// We are looking for the first index i where the string becomes >= target.
		// This is a copy of sort.Search, with f(h) replaced by (*Type[h].String() >= target).
		i, j := 0, len(offs)
		for i < j {
			h := i + (j-i)/2 // avoid overflow when computing h
			// i ≤ h < j
			currentType = (*RType)(unsafe.Pointer(uintptr(section) + uintptr(offs[h])))
			search = currentType.nameOffsetStr().name()
			if currentType.hasExtraStar() {
				search = search[1:]
			}
			// Compare(a, b []byte) int
			// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
			// if search >= target {
			if bytes.Compare(search, target) >= 0 {
				j = h // preserves f(j) == true
			} else {
				i = h + 1 // preserves f(i-1) == false
			}
		}
		// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.

		// Having found the first, linear scan forward to find the last.
		// We could do a second binary search, but the caller is going
		// to do a linear scan anyway.
		for k := i; k < len(offs); k++ {
			currentType = (*RType)(unsafe.Pointer(uintptr(section) + uintptr(offs[k])))
			search = currentType.nameOffsetStr().name()
			if currentType.hasExtraStar() {
				search = search[1:]
			}
			// if search != target {
			if !bytes.Equal(search, target) {
				break
			}
			results = append(results, currentType)
		}
	}
	return results
}

func appendVarint(x []byte, v uintptr) []byte {
	for ; v >= 0x80; v >>= 7 {
		x = append(x, byte(v|0x80))
	}
	x = append(x, byte(v))
	return x
}

func emptyFuncProto() funcType {
	var ifunc interface{} = (func())(nil)
	prototype := *(**funcType)(unsafe.Pointer(&ifunc))
	return *prototype
}

func emptySliceProto() sliceType {
	var islice interface{} = ([]unsafe.Pointer)(nil)
	prototype := *(**sliceType)(unsafe.Pointer(&islice))
	return *prototype
}

func emptyArrayProto() arrayType {
	var iarray interface{} = [1]unsafe.Pointer{}
	prototype := *(**arrayType)(unsafe.Pointer(&iarray))
	return *prototype
}

func emptyPtrProto() ptrType {
	var iptr interface{} = (*unsafe.Pointer)(nil)
	prototype := *(**ptrType)(unsafe.Pointer(&iptr))
	return *prototype
}

// append a bit to the bitmap.
func appendBitVector(vec *bitVector, bit uint8) {
	if vec.num%8 == 0 {
		vec.data = append(vec.data, 0)
	}
	vec.data[vec.num/8] |= bit << (vec.num % 8)
	vec.num++
}

// funcStr builds a string representation of a funcType.
func funcStr(ft *funcType) []byte {
	result := make([]byte, 0, 64)
	result = append(result, "func("...)
	for i, t := range ft.inParams() {
		if i > 0 {
			result = append(result, ", "...)
		}
		result = append(result, t.nomen()...)
	}
	result = append(result, ')')
	out := ft.outParams()

	switch len(out) {
	case 0: // do nothing
	case 1:
		result = append(result, ' ')
	default:
		result = append(result, " ("...)
	}
	for i, t := range out {
		if i > 0 {
			result = append(result, ", "...)
		}
		result = append(result, t.nomen()...)
	}
	if len(out) > 1 {
		result = append(result, ')')
	}
	return result
}

func exportedMethods(t *RType) []method {
	all, ok := methods(t)
	if !ok {
		return nil
	}
	allExported := true
	for _, method := range all {
		methodName := t.nameOffset(method.nameOffset)
		if !methodName.isExported() {
			allExported = false
			break
		}
	}
	var methods []method
	if allExported {
		methods = all
	} else {
		methods = make([]method, 0, len(all))
		for _, m := range all {
			methodName := t.nameOffset(m.nameOffset)
			if methodName.isExported() {
				if m.typeOffset == 0 {
					panic("reflect.x.error : method type is zero. Apply fix.")
				}
				methods = append(methods, m)
			}
		}
		methods = methods[:len(methods):len(methods)]
	}

	return methods
}

func lenExportedMethods(t *RType) int {
	all, ok := methods(t)
	if !ok {
		return 0
	}
	count := 0
	for _, method := range all {
		methodName := t.nameOffset(method.nameOffset)
		if methodName.isExported() {
			if method.typeOffset == 0 {
				panic("reflect.x.error : method type is zero. Apply fix.")
			}
			count++
		}
	}
	return count
}

func methods(t *RType) ([]method, bool) {
	if !t.hasInfoFlag() {
		return nil, false
	}

	var ut *uncommonType
	switch t.Kind() {
	case Struct:
		ut = &(*uncommonStruct)(unsafe.Pointer(t)).u
	case Ptr:
		ut = &(*uncommonPtr)(unsafe.Pointer(t)).u
	case Func:
		ut = &(*uncommonFunc)(unsafe.Pointer(t)).u
	case Slice:
		ut = &(*uncommonSlice)(unsafe.Pointer(t)).u
	case Array:
		ut = &(*uncommonArray)(unsafe.Pointer(t)).u
	case Interface:
		ut = &(*uncommonInterface)(unsafe.Pointer(t)).u
	default:
		ut = &(*uncommonConcrete)(unsafe.Pointer(t)).u
	}

	if ut.mCount == 0 {
		return nil, false
	}

	return (*[1 << 16]method)(unsafe.Pointer(uintptr(unsafe.Pointer(ut)) + uintptr(ut.mOffset)))[:ut.mCount:ut.mCount], true
}

// funcLayout computes a struct type representing the layout of the function arguments and return values for the function type t.
// If rcvr != nil, rcvr specifies the type of the receiver.
// The returned type exists only for GC, so we only fill out GC relevant info.
// Currently, that's just size and the GC program. We also fill in the name for possible debugging use.
func funcLayout(t *RType, rcvr *RType) (*RType, uintptr, uintptr, *bitVector) {
	if t.Kind() != Func {
		panic("reflect.x.error : funcLayout of non-func type")
	}
	if rcvr != nil && rcvr.Kind() == Interface {
		panic("reflect.x.error : funcLayout with interface receiver.")
	}

	funcType := t.convToFn()

	// compute gc program & stack bitmap for arguments
	ptrBitVector := new(bitVector)
	var offset uintptr
	if rcvr != nil {
		// Reflect uses the "interface" calling convention for
		// methods, where receivers take one word of argument
		// space no matter how big they actually are.
		if rcvr.isDirectIface() || rcvr.hasPointers() {
			appendBitVector(ptrBitVector, 1)
		}
		offset += PtrSize
	}

	for _, in := range funcType.inParams() {
		offset += -offset & uintptr(in.align-1)
		if in.hasPointers() {
			in.addTypeBits(ptrBitVector, offset)
		}
		offset += in.size
	}
	argN := ptrBitVector.num
	argSize := offset
	if IsAMD64p32 {
		offset += -offset & (8 - 1)
	}
	offset += -offset & (PtrSize - 1)
	retOffset := offset

	for _, out := range funcType.outParams() {
		offset += -offset & uintptr(out.align-1)
		if out.hasPointers() {
			out.addTypeBits(ptrBitVector, offset)
		}
		offset += out.size
	}
	offset += -offset & (PtrSize - 1)

	// build dummy Type holding gc program
	result := &RType{
		align:   PtrSize,
		size:    offset,
		ptrData: uintptr(ptrBitVector.num) * PtrSize,
	}

	if IsAMD64p32 {
		result.align = 8
	}

	if ptrBitVector.num > 0 {
		result.gcData = &ptrBitVector.data[0]
	} else {
		result.kind |= kindNoPointers
	}

	ptrBitVector.num = argN

	var fnSign []byte

	if rcvr != nil {
		fnSign = byteSliceFromParams(methStr, openPar, rcvr.nomen(), closePar, openPar, t.nomen(), closePar)
	} else {
		fnSign = byteSliceFromParams(fnStr, openPar, t.nomen(), closePar)
	}

	result.str = declareReflectName(newName(fnSign))

	return result, argSize, retOffset, ptrBitVector
}

// from strconv - contains reports whether the string contains the byte c.
func contains(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}

// from strconv
func unhex(b byte) (v rune, ok bool) {
	c := rune(b)
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}
	return
}

// methodName returns the name of the calling method,
// assumed to be two stack frames above.
/**
func methodName() string {
	pc, _, line, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown method"
	}
	return f.Name() + " line : " + strconv.Itoa(line)
}
**/

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding - Found in "log" package
func I2A(i int, wid int) string {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	return string(b[bp:])
}
