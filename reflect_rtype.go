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

func (t *RType) isDirectIface() bool { return t.kind&kindDirectIface == 0 } // isDirectIface reports whether t is stored indirectly in an interface value.
func (t *RType) hasPointers() bool   { return t.kind&kindNoPointers == 0 }
func (t *RType) canHandleGC() bool   { return t.kind&kindGCProg == 0 }
func (t *RType) isAnon() bool        { return t.extraTypeFlag&hasNameFlag == 0 }
func (t *RType) hasExtraStar() bool  { return t.extraTypeFlag&hasExtraStarFlag != 0 }
func (t *RType) hasInfoFlag() bool   { return t.extraTypeFlag&hasExtraInfoFlag != 0 }
func (t *RType) hasName() bool       { return !t.isAnon() && len(t.nameOffsetStr().name()) > 0 }
func (t *RType) nameOffset(offset int32) name {
	return name{(*byte)(resolveNameOff(unsafe.Pointer(t), offset))}
}
func (t *RType) nameOffsetStr() name { return name{(*byte)(resolveNameOff(unsafe.Pointer(t), t.str))} }
func (t *RType) typeOffset(offset int32) *RType {
	return (*RType)(resolveTypeOff(unsafe.Pointer(t), offset))
}
func (t *RType) textOffset(offset int32) unsafe.Pointer {
	return resolveTextOff(unsafe.Pointer(t), offset)
}
func (t *RType) convToPtr() *ptrType         { return (*ptrType)(unsafe.Pointer(t)) }
func (t *RType) convToStruct() *structType   { return (*structType)(unsafe.Pointer(t)) }
func (t *RType) convToFn() *funcType         { return (*funcType)(unsafe.Pointer(t)) }
func (t *RType) convToIface() *ifaceType     { return (*ifaceType)(unsafe.Pointer(t)) }
func (t *RType) numIn() int                  { return int(t.convToFn().InLen) }
func (t *RType) numOut() int                 { return len(t.convToFn().outParams()) }
func (t *RType) ConvToMap() *mapType         { return (*mapType)(unsafe.Pointer(t)) }
func (t *RType) ConvToSlice() *sliceType     { return (*sliceType)(unsafe.Pointer(t)) }
func (t *RType) ConvToArray() *arrayType     { return (*arrayType)(unsafe.Pointer(t)) }
func (t *RType) ifaceMethods() []ifaceMethod { return t.convToIface().methods }
func (t *RType) Deref() *RType               { return (*ptrType)(unsafe.Pointer(t)).Type }
func (t *RType) NoOfIfaceMethods() int       { return len(t.ifaceMethods()) }
func (t *RType) Kind() Kind                  { return Kind(t.kind & kindMask) }
func (t *RType) Size() uintptr               { return t.size }
func (t *RType) FieldAlign() int             { return int(t.fieldAlign) }
func (t *RType) Align() int                  { return int(t.align) }
func (t *RType) String() string              { return string(t.nomen()) }
func (t *RType) IsExported() bool            { return t.kind&(1<<5|1<<6) == 0 }

/**
func (t *RType) GoString() string {
	result := ""

	if t.hasName() {
		result += "Name : " + string(t.nomen())
	} else {
		result += "Anonymous"
	}

	result += "\n\tKind : " + StringKind(t.Kind()) + " " + strconv.FormatUint(uint64(t.kind&kindMask), 10) + " " + strconv.FormatUint(uint64(t.kind), 10)
	result += "\n\tSize : " + strconv.FormatUint(uint64(t.size), 10)
	result += "\n\tAlign : " + strconv.FormatUint(uint64(t.align), 10)
	result += "\n\tFieldAlign : " + strconv.FormatUint(uint64(t.fieldAlign), 10)

	if t.Kind() == Func {
		result += "\n\tParams : " + strconv.FormatInt(int64(t.numIn()), 10) + " in , " + strconv.FormatInt(int64(t.numOut()), 10) + " out."
	}

	if t.hasExtraStar() {
		result += "*"
	}
	if t.isDirectIface() {
		result += "\n\tInterface"
	}
	if t.hasPointers() {
		result += "\n\tPointers"
	}
	if t.canHandleGC() {
		result += "\n\tGCHandler"
	}
	if t.hasInfoFlag() {
		result += "\n\tHasInfo"
	}
	if t.Kind() == Struct {
		if n := t.NoOfIfaceMethods(); n > 0 {
			result += "\n\t" + strconv.Itoa(n) + " methods vs " + strconv.Itoa(lenExportedMethods(t)) + " methods"
		}
	}

	return result
}
**/
func (t *RType) nomen() []byte {
	s := t.nameOffsetStr().name()
	if t.hasExtraStar() {
		return s[1:]
	}
	return s
}

// Implements reports whether the type implements the interface type u.
// You always have to provide an Interface Kind of *Type
// Of course, providing nil, returns false
func (t *RType) Implements(u *RType) bool {
	if u == nil {
		return false
	}
	if u.Kind() != Interface {
		return false
	}
	return t.implements(u)
}

// AssignableTo reports whether a value of the type is assignable to type u.v
// Of course, providing nil, returns false
func (t *RType) AssignableTo(u *RType) bool {
	if u == nil {
		return false
	}
	return t.directlyAssignable(u) || t.implements(u)
}

// ConvertibleTo reports whether a value of the type is convertible to type u.
// Of course, providing nil, returns false
func (t *RType) ConvertibleTo(u *RType) bool {
	if u == nil {
		return false
	}
	return t.convertible(u)
}

// PtrTo returns the pointer type with element t.
// For example, if t represents type Foo, PtrTo(t) represents *Foo.
func (t *RType) PtrTo() *RType {
	if t.ptrToThis != 0 {
		return t.typeOffset(t.ptrToThis)
	}

	// Look in known types.
	typeName := byteSliceFromParams(star, t.nomen())
	for _, existingType := range typesByString(typeName) {
		// Attention : cannot use .Deref() here because below we need to return the pointer to the *Type
		pointerType := existingType.convToPtr()
		if pointerType.Type != t {
			continue
		}
		return &pointerType.RType
	}

	// Create a new ptrType starting with the description of an *ptr.
	proto := emptyPtrProto()
	proto.str = declareReflectName(newName(typeName))
	proto.ptrToThis = 0

	// For the type structures linked into the binary, the
	// compiler provides a good hash of the string.
	// Create a good hash for the new string by using
	// the FNV-1 hash's mixing function to combine the
	// old hash and the new "*".
	proto.hash = fnv1(t.hash, '*')
	proto.Type = t
	return &proto.RType
}

// directlyAssignable reports whether a value x of type V can be directly assigned (using memmove) to a value of type T.
// https://golang.org/doc/go_spec.html#Assignability
// Ignoring the interface rules (implemented elsewhere) and the ideal constant rules (no ideal constants at run time).
func (t *RType) directlyAssignable(dest *RType) bool {
	// x's type V is identical to T?
	if dest == t {
		return true
	}

	// Otherwise at least one of T and V must be unnamed
	// and they must have the same kind.
	if dest.hasName() && t.hasName() || dest.Kind() != t.Kind() {
		return false
	}
	// x's type T and V must  have identical underlying types.
	return t.haveIdenticalUnderlyingType(dest, true)
}

func (t *RType) haveIdenticalType(dest *RType, cmpTags bool) bool {
	if cmpTags {
		return dest == t
	}
	if dest.Name() != t.Name() || dest.Kind() != t.Kind() {
		return false
	}
	return t.haveIdenticalUnderlyingType(dest, false)
}

func (t *RType) haveIdenticalUnderlyingType(dest *RType, cmpTags bool) bool {
	if dest == t {
		return true
	}

	kind := dest.Kind()
	if kind != t.Kind() {
		return false
	}

	// Non-composite types of equal kind have same underlying type (the predefined instance of the type).
	if Bool <= kind && kind <= Complex128 || kind == String || kind == UnsafePointer {
		return true
	}

	// Composite types.
	switch kind {
	case Array:
		destArray := dest.ConvToArray()
		return destArray.Len == t.ConvToArray().Len && t.haveIdenticalType(destArray.ElemType, cmpTags)
	case Func:
		destFn := dest.convToFn()
		srcFn := t.convToFn()
		if destFn.OutLen != srcFn.OutLen || destFn.InLen != srcFn.InLen {
			return false
		}
		for i := 0; i < destFn.numIn(); i++ {
			if !srcFn.inParam(i).haveIdenticalType(destFn.inParam(i), cmpTags) {
				return false
			}
		}
		for i := 0; i < destFn.numOut(); i++ {
			if !srcFn.outParam(i).haveIdenticalType(destFn.outParam(i), cmpTags) {
				return false
			}
		}
		return true
	case Interface:
		destMethods := dest.ifaceMethods()
		srcMethods := t.ifaceMethods()
		// the case of "interface{}"
		if len(destMethods) == 0 && len(srcMethods) == 0 {
			return true
		}
		// Might have the same methods but still need a run time conversion.
		return false
	case Map:
		return t.ConvToMap().KeyType.haveIdenticalType(dest.ConvToMap().KeyType, cmpTags) && t.ConvToMap().ElemType.haveIdenticalType(dest.ConvToMap().ElemType, cmpTags)
	case Slice:
		return t.ConvToSlice().ElemType.haveIdenticalType(dest.ConvToSlice().ElemType, cmpTags)
	case Ptr:
		return t.Deref().haveIdenticalType(dest.Deref(), cmpTags)
	case Struct:
		destStruct := dest.convToStruct()
		srcStruct := t.convToStruct()
		if len(destStruct.fields) != len(srcStruct.fields) {
			return false
		}
		if !bytes.Equal(destStruct.pkgPath.name(), srcStruct.pkgPath.name()) {
			return false
		}
		for i := range destStruct.fields {
			destField := &destStruct.fields[i]
			srcField := &srcStruct.fields[i]
			if !bytes.Equal(destField.name.name(), srcField.name.name()) {
				return false
			}
			if !srcField.Type.haveIdenticalType(destField.Type, cmpTags) {
				return false
			}
			if cmpTags && !bytes.Equal(destField.name.tag(), srcField.name.tag()) {
				return false
			}
			if destField.offsetEmbed != srcField.offsetEmbed {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (t *RType) pkg() (int32, bool) {
	if !t.hasInfoFlag() {
		return 0, false
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
	return ut.pkgPath, true
}

// implements reports whether the type V implements the interface type T.
func (t *RType) implements(dest *RType) bool {
	if dest.Kind() != Interface {
		return false
	}
	destIntf := dest.convToIface()
	// the case of "interface{}"
	if len(destIntf.methods) == 0 {
		return true
	}

	// The same algorithm applies in both cases, but the method tables for an interface type and a concrete type are different, so the code is duplicated.
	// In both cases the algorithm is a linear scan over the two lists - T's methods and V's methods - simultaneously.
	// Since method tables are stored in a unique sorted order (alphabetical, with no duplicate method names), the scan through V's methods must hit a match for each of T's methods along the way, or else V does not implement T.
	// This lets us run the scan in overall linear time instead of the quadratic time  a naive search would require.
	// See also ../runtime/iface.go.
	if t.Kind() == Interface {
		srcIntf := t.convToIface()
		i := 0
		for j := 0; j < len(srcIntf.methods); j++ {
			destMethod := &destIntf.methods[i]
			destMethodName := destIntf.nameOffset(destMethod.nameOffset)
			srcMethod := &srcIntf.methods[j]
			srcMethodName := t.nameOffset(srcMethod.nameOffset)
			if bytes.Equal(srcMethodName.name(), destMethodName.name()) &&
				t.typeOffset(srcMethod.typeOffset) == destIntf.typeOffset(destMethod.typeOffset) {
				if !destMethodName.isExported() {
					destPkgPath := destMethodName.pkgPath()
					if len(destPkgPath) == 0 {
						destPkgPath = destIntf.pkgPath.name()
					}
					srcPkgPath := srcMethodName.pkgPath()
					if len(srcPkgPath) == 0 {
						srcPkgPath = srcIntf.pkgPath.name()
					}
					if !bytes.Equal(destPkgPath, srcPkgPath) {
						continue
					}
				}
				if i++; i >= len(destIntf.methods) {
					return true
				}
			}
		}
		return false
	}

	vmethods, ok := methods(t)
	if !ok {
		return false
	}
	origPkgPath := make([]byte, 0)
	pkg, ok := t.pkg()
	if ok {
		origPkgPath = t.nameOffset(pkg).name()
	}
	i := 0
	for j := 0; j < len(vmethods); j++ {
		destMethod := &destIntf.methods[i]
		destMethodName := destIntf.nameOffset(destMethod.nameOffset)
		srcMethod := vmethods[j]
		srcMethodName := t.nameOffset(srcMethod.nameOffset)
		if bytes.Equal(srcMethodName.name(), destMethodName.name()) &&
			t.typeOffset(srcMethod.typeOffset) == destIntf.typeOffset(destMethod.typeOffset) {
			if !destMethodName.isExported() {
				destPkgPath := destMethodName.pkgPath()
				if len(destPkgPath) == 0 {
					destPkgPath = destIntf.pkgPath.name()
				}
				srcPkgPath := srcMethodName.pkgPath()
				if len(srcPkgPath) == 0 {
					srcPkgPath = origPkgPath
				}
				if !bytes.Equal(destPkgPath, srcPkgPath) {
					continue
				}
			}
			if i++; i >= len(destIntf.methods) {
				return true
			}
		}
	}
	return false
}

// convertOp returns the function to convert a value of type src to a value of type dst. If the conversion is illegal, convertOp returns nil.
func (t *RType) convertible(dst *RType) bool {
	destKind := dst.Kind()
	srcKind := t.Kind()

	switch srcKind {
	case Int, Int8, Int16, Int32, Int64:
		switch destKind {
		case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, UintPtr:
			return true
		case Float32, Float64:
			return true
		case String:
			return true
		}
	case Uint, Uint8, Uint16, Uint32, Uint64, UintPtr:
		switch destKind {
		case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, UintPtr:
			return true
		case Float32, Float64:
			return true
		case String:
			return true
		}
	case Float32, Float64:
		switch destKind {
		case Int, Int8, Int16, Int32, Int64:
			return true
		case Uint, Uint8, Uint16, Uint32, Uint64, UintPtr:
			return true
		case Float32, Float64:
			return true
		}
	case Complex64, Complex128:
		switch destKind {
		case Complex64, Complex128:
			return true
		}
	case String:
		sliceElem := dst.ConvToSlice().ElemType
		if destKind == Slice && sliceElem.pkgPathLen() == 0 {
			switch sliceElem.Kind() {
			case Uint8:
				return true
			case Int32:
				return true
			}
		}
	case Slice:
		sliceElem := t.ConvToSlice().ElemType
		if destKind == String && sliceElem.pkgPathLen() == 0 {
			switch sliceElem.Kind() {
			case Uint8:
				return true
			case Int32:
				return true
			}
		}
	}

	// dst and src have same underlying type.
	if t.haveIdenticalUnderlyingType(dst, false) {
		return true
	}

	// dst and src are unnamed pointer types with same underlying base type.
	if destKind == Ptr && !dst.hasName() &&
		srcKind == Ptr && !t.hasName() &&
		t.Deref().haveIdenticalUnderlyingType(dst.Deref(), false) {
		return true
	}

	return t.implements(dst)
}

func (t *RType) Comparable() bool {
	return t.alg != nil && t.alg.equal != nil
}

// Methods applicable only to some types, depending on Kind.
// The methods allowed for each kind are:
//
// 	Int*, Uint*, Float*, Complex*: Bits
// 	Array: Deref, Len
// 	Chan: ChanDir, Deref
// 	Func: In, NumIn, Out, NumOut, IsVariadic.
// 	Map: Key, Deref
// 	ptr: Deref
// 	Slice: Deref
// 	Struct: Field, FieldByIndex, FieldByName, FieldByNameFunc, NumField

// Bits returns the size of the type in bits.
func (t *RType) Bits() int {
	if t == nil {
		return 0
	}
	k := t.Kind()
	if k < Int || k > Complex128 {
		if willPrintDebug {
			println("reflect.Bits : of non-arithmetic Type.")
		}
		return 0
	}
	return int(t.size) * 8
}

func (t *RType) addTypeBits(vec *bitVector, offset uintptr) {
	switch t.Kind() {
	case Chan, Func, Map, Ptr, Slice, String, UnsafePointer:
		// 1 pointer at start of representation
		for vec.num < uint32(offset/uintptr(PtrSize)) {
			appendBitVector(vec, 0)
		}
		appendBitVector(vec, 1)
	case Interface:
		// 2 pointers
		for vec.num < uint32(offset/uintptr(PtrSize)) {
			appendBitVector(vec, 0)
		}
		appendBitVector(vec, 1)
		appendBitVector(vec, 1)
	case Array:
		// repeat inner type
		tArray := t.ConvToArray()
		for i := 0; i < int(tArray.Len); i++ {
			if tArray.ElemType.hasPointers() {
				tArray.ElemType.addTypeBits(vec, offset+uintptr(i)*tArray.ElemType.size)
			}
		}
	case Struct:
		// apply fields
		structType := t.convToStruct()
		for i := range structType.fields {
			field := &structType.fields[i]
			if field.Type.hasPointers() {
				field.Type.addTypeBits(vec, offset+structFieldOffset(field))
			}
		}
	}
}

// used only in tests
func (t *RType) PkgPath() string {
	if t.isAnon() {
		return ""
	}
	pk, ok := t.pkg()
	if !ok {
		return ""
	}
	return string(t.nameOffset(pk).name())
}

func (t *RType) pkgPathLen() int {
	if t.isAnon() {
		return 0
	}
	pk, ok := t.pkg()
	if !ok {
		return 0
	}
	return t.nameOffset(pk).nameLen()
}

func (t *RType) Name() string {
	if t.isAnon() {
		return ""
	}
	s := t.nameOffsetStr().name()
	i := len(s) - 1
	for i >= 0 {
		if s[i] == '.' {
			break
		}
		i--
	}
	// if we have extra star, and it's the full name, then set it to avoid first char (which is the star)
	if t.hasExtraStar() && i == -1 {
		i = 0
	}
	return string(s[i+1:])
}

func (t *RType) Fields(inspect InspectTypeFn) {
	if t.Kind() != Struct {
		if willPrintDebug {
			println("reflect.Field: Requested field of non-struct type")
		}
		return
	}
	structType := (*structType)(unsafe.Pointer(t))
	for i := range structType.fields {
		field := &structType.fields[i]
		fn := field.name
		var PkgPath []byte

		if !fn.isExported() {
			PkgPath = structType.pkgPath.name()
		}

		var Tag []byte
		if tag := fn.tag(); len(tag) > 0 {
			Tag = tag
		}

		inspect(field.Type, fn.name(), Tag, PkgPath, field.offsetEmbed&1 != 0, fn.isExported(), field.offsetEmbed>>1, i)
	}
}

func (t *RType) StructFields() []structField {
	return t.convToStruct().fields
}
