/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import (
	"unicode/utf8"
	"unsafe"
)

// New returns a Value representing a pointer to a new zero value
// for the specified type. That is, the returned Value's *Type is PtrTo(Type).
func New(typ *RType) Value {
	if typ == nil {
		if willPrintDebug {
			panic("reflect: New(nil)")
		}
	}
	newPtr := unsafeNew(typ)
	return Value{Type: typ.PtrTo(), Ptr: newPtr, Flag: Flag(Ptr)}
}

// Constr returns a Value representing a new zero value for the specified type. No pointers.
func Constr(typ *RType) Value {
	return internalNew(typ)
}

// ReflectOn returns a new Value initialized to the concrete value stored in the interface i. ReflectOn(nil) returns the zero Value.
func ReflectOn(i interface{}) Value {
	if i == nil {
		if willPrintDebug {
			println("reflect.x.error : ReflectOn, provided param is nil")
		}
		return Value{}
	}
	// unpackEface converts the empty interface i to a Value.
	e := toIface(unsafe.Pointer(&i))
	// NOTE: don't read e.word until we know whether it is really a pointer or not.
	if e.Type == nil {
		return Value{}
	}
	f := Flag(e.Type.Kind())
	if e.Type.isDirectIface() {
		f |= pointerFlag // set the pointer flag
	}
	return Value{Type: e.Type, Ptr: e.word, Flag: f}
}

// ReflectOn returns a new Value initialized to the concrete value stored in the interface i, where interface i is actually a pointer to a concrete value.
func ReflectOnPtr(i interface{}) Value {
	if i == nil {
		panic("Bad usage : provided interface should something else than nil")
	}
	// unpackEface converts the empty interface i to a Value.
	e := toIface(unsafe.Pointer(&i))
	// NOTE: don't read e.word until we know whether it is really a pointer or not.

	if e.Type == nil {
		return Value{}
	}
	if e.Type.Kind() != Ptr {
		panic("Bad usage : provided interface should be a pointer to something")
	}
	if e.Type.isDirectIface() {
		panic("Bad usage : it shouldn't be a direct interface.")
	}

	ptrToV := e.word
	// The returned value's address is v's value.
	if ptrToV == nil {
		return Value{}
	}
	// if we got here, there is not a dereference, nor the pointer is nil - studying the type's pointer
	typ := e.Type.Deref()
	fl := Flag(e.Type.Kind())&exportFlag | pointerFlag | addressableFlag | Flag(typ.Kind())
	return Value{Type: typ, Ptr: ptrToV, Flag: fl}
}

// TypeOf used to be exactly this function.
// TypeOf returns the reflection Type that represents the dynamic type of i.
// If i is a nil interface value, TypeOf returns nil.
func TypeOf(i interface{}) *RType {
	result := (*toIface(unsafe.Pointer(&i))).Type
	if result == nil {
		return nil
	}
	return result
}

// Convert returns the value v converted to type t.
// If the usual Go conversion rules do not allow conversion
// of the value v to type t, Convert panics.
func Convert(v Value, typ *RType) Value {
	destKind := typ.Kind()
	srcKind := v.Type.Kind()

	switch srcKind {
	case Int, Int8, Int16, Int32, Int64:
		switch destKind {
		case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, UintPtr:
			return makeInt(v.ro(), uint64(v.Int().Get()), typ) // convert operation: intXX -> [u]intXX
		case Float32, Float64:
			return makeFloat(v.ro(), float64(v.Int().Get()), typ) // convert operation: intXX -> floatXX
		case String:
			return makeString(v.ro(), string(v.Int().Get()), typ) // convert operation: intXX -> string
		}
	case Uint, Uint8, Uint16, Uint32, Uint64, UintPtr:
		switch destKind {
		case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, UintPtr:
			return makeInt(v.ro(), v.Uint().Get(), typ) // convert operation: uintXX -> [u]intXX
		case Float32, Float64:
			return makeFloat(v.ro(), float64(v.Uint().Get()), typ) // convert operation: uintXX -> floatXX
		case String:
			return makeString(v.ro(), string(v.Uint().Get()), typ) // convert operation: uintXX -> string
		}
	case Float32, Float64:
		switch destKind {
		case Int, Int8, Int16, Int32, Int64:
			return makeInt(v.ro(), uint64(int64(v.Float().Get())), typ) // convert operation: floatXX -> intXX
		case Uint, Uint8, Uint16, Uint32, Uint64, UintPtr:
			return makeInt(v.ro(), uint64(v.Float().Get()), typ) // convert operation: floatXX -> uintXX
		case Float32, Float64:
			return makeFloat(v.ro(), v.Float().Get(), typ) // convert operation: floatXX -> floatXX
		}
	case Complex64, Complex128:
		switch destKind {
		case Complex64, Complex128:
			return makeComplex(v.ro(), v.Complex().Get(), typ) // convert operation: complexXX -> complexXX
		}
	case String:
		sliceElem := (*sliceType)(unsafe.Pointer(typ)).ElemType
		if destKind == Slice && sliceElem.pkgPathLen() == 0 {
			switch sliceElem.Kind() {
			case Uint8:
				return makeBytes(v.ro(), []byte(*(*string)(v.Ptr)), typ) // convert operation: string -> []byte
			case Int32:
				return makeRunes(v.ro(), []rune(*(*string)(v.Ptr)), typ) // convert operation: string -> []rune
			}
		}
	case Slice:
		sliceElem := (*sliceType)(unsafe.Pointer(v.Type)).ElemType
		if destKind == String && sliceElem.pkgPathLen() == 0 {
			switch sliceElem.Kind() {
			case Uint8:
				return makeString(v.ro(), string(*(*[]byte)(v.Ptr)), typ) // convert operation: []byte -> string
			case Int32:
				return makeString(v.ro(), string(*(*[]rune)(v.Ptr)), typ) // // convert operation: []rune -> string
			}
		}
	}

	// dst and src have same underlying type.
	if v.Type.haveIdenticalUnderlyingType(typ, false) {
		return cvtDirect(v, typ)
	}
	derefType := (*ptrType)(unsafe.Pointer(v.Type)).Type
	destDerefType := (*ptrType)(unsafe.Pointer(typ)).Type
	// dst and src are unnamed pointer types with same underlying base type.
	if destKind == Ptr && !typ.hasName() &&
		srcKind == Ptr && !v.Type.hasName() &&
		derefType.haveIdenticalUnderlyingType(destDerefType, false) {
		return cvtDirect(v, typ)
	}

	if v.Type.implements(typ) {
		if srcKind == Interface {
			return cvtI2I(v, typ)
		}
		return cvtT2I(v, typ)
	}

	panic("reflect.Value.Convert: value of type ") // + TypeToString(v.Type) + " cannot be converted to type " + TypeToString(t))
}

// syntactic sugar
func ToMap(v Value) MapValue {
	if v.Type == nil {
		if willPrintDebug {
			println("Value->MapValue called on a nil type.")
		}
		return MapValue{}
	}
	if !v.IsValid() || !v.isExported() || v.Kind() != Map {
		if willPrintDebug {
			println("Value->MapValue kind `" + StringKind(v.Kind()) + "` not map, invalid or not exported.")
		}
		return MapValue{}
	}
	return MapValue{Value: v}
}

// syntactic sugar
func ToSlice(v Value) SliceValue {
	if v.Type == nil {
		if willPrintDebug {
			println("Value's Type is nil")
		}
		return SliceValue{}
	}
	k := v.Kind()
	if k != Array && k != Slice && k != String {
		if willPrintDebug {
			println("Cannot provide : kind not Array, Slice or String")
		}
		return SliceValue{}
	}
	return SliceValue{Value: v}
}

// syntactic sugar
func ToStruct(v Value) StructValue {
	if v.Type == nil {
		if willPrintDebug {
			panic("Value->StructValue called on a nil type.")
		}
	}

	if !v.IsValid() {
		if willPrintDebug {
			panic("Value->StructValue invalid")
		}
	}

	if v.hasMethodFlag() {
		if willPrintDebug {
			panic("Value->StructValue Has methods flag")
		}
	}

	// making sure is a struct : if it's an interface or something else and has methods, it's ok.
	if v.Kind() != Struct {
		if v.Type.Kind() == Interface {
			// Method on interface.
			intf := v.Type.convToIface()
			// the case of "interface{}"
			if len(intf.methods) == 0 {
				panic("Value->StructValue Interface has no methods.")
			}
		} else {
			methods := exportedMethods(v.Type)
			if len(methods) == 0 {
				panic("Value->StructValue has no methods.")
			}
		}
	}

	return StructValue{Value: v}
}

// MapOf returns the map type with the given key and element types.
// For example, if k represents int and e represents string,
// MapOf(k, e) represents map[int]string.
//
// If the key type is not a valid map key type (that is, if it does
// not implement Go's == operator), MapOf panics.
func MapOf(keyType, elemType *RType) *RType {
	if !ismapkey(keyType) {
		if willPrintDebug {
			panic("reflect.MapOf: invalid key type " + TypeToString(keyType))
		}
	}

	// Look in known types.
	typeName := byteSliceFromParams(mapStr, sqOpenPar, TypeToString(keyType), sqClosPar, TypeToString(elemType))
	for _, existingType := range typesByString(typeName) {
		mapType := existingType.ConvToMap()
		if mapType.KeyType == keyType && mapType.ElemType == elemType {
			return existingType
		}
	}

	// Make a map type.
	var imap interface{} = (map[unsafe.Pointer]unsafe.Pointer)(nil)
	proto := **(**mapType)(unsafe.Pointer(&imap))
	proto.str = declareReflectName(newName(typeName))
	proto.extraTypeFlag = 0
	proto.hash = fnv1(elemType.hash, 'm', byte(keyType.hash>>24), byte(keyType.hash>>16), byte(keyType.hash>>8), byte(keyType.hash))
	proto.KeyType = keyType
	proto.ElemType = elemType
	proto.bucket = bucketOf(keyType, elemType)
	if keyType.size > maxKeySize {
		proto.keySize = uint8(PtrSize)
		proto.indirectKey = 1
	} else {
		proto.keySize = uint8(keyType.size)
		proto.indirectKey = 0
	}

	if elemType.size > maxValSize {
		proto.valueSize = uint8(PtrSize)
		proto.indirectValue = 1
	} else {
		proto.valueSize = uint8(elemType.size)
		proto.indirectValue = 0
	}

	proto.bucketSize = uint16(proto.bucket.size)
	proto.reflexiveKey = isReflexive(keyType)
	proto.needsKeyUpdate = needKeyUpdate(keyType)
	proto.ptrToThis = 0

	return &proto.RType
}

// SliceOf returns the slice type with element type t.
// For example, if t represents int, SliceOf(t) represents []int.
func SliceOf(typ *RType) *RType {
	// Look in known types.
	typeName := byteSliceFromParams(sqOpenPar, sqClosPar, TypeToString(typ))
	for _, existingType := range typesByString(typeName) {
		sliceType := existingType.ConvToSlice()
		if sliceType.ElemType == typ {
			return existingType
		}
	}

	// Make a slice type.
	proto := emptySliceProto()
	proto.extraTypeFlag = 0
	proto.str = declareReflectName(newName(typeName))
	proto.hash = fnv1(typ.hash, '[')
	proto.ElemType = typ
	proto.ptrToThis = 0

	return &proto.RType
}

// ArrayOf returns the array type with the given count and element type.
// For example, if t represents int, ArrayOf(t,5) represents [5]int.
//
// If the resulting type would be larger than the available address space, ArrayOf panics.
func ArrayOf(elem *RType, count int) *RType {
	// Look in known types.
	typeName := byteSliceFromParams(sqOpenPar, I2A(count, -1), sqClosPar, TypeToString(elem))
	for _, existingType := range typesByString(typeName) {
		arrayType := existingType.ConvToArray()
		if arrayType.ElemType == elem {
			return existingType
		}
	}

	// Make an array type.
	proto := emptyArrayProto()
	proto.extraTypeFlag = 0
	proto.str = declareReflectName(newName(typeName))
	proto.hash = fnv1(elem.hash, sqOpenPar)
	for n := uint32(count); n > 0; n >>= 8 {
		proto.hash = fnv1(proto.hash, byte(n))
	}
	proto.hash = fnv1(proto.hash, sqClosPar)
	proto.ElemType = elem
	proto.ptrToThis = 0
	if elem.size > 0 {
		max := ^uintptr(0) / elem.size
		if uintptr(count) > max {
			if willPrintDebug {
				panic("reflect.ArrayOf: array size would exceed virtual address space")
			}
			return nil
		}
	}
	proto.size = elem.size * uintptr(count)
	if count > 0 && elem.ptrData != 0 {
		proto.ptrData = elem.size*uintptr(count-1) + elem.ptrData
	}
	proto.align = elem.align
	proto.fieldAlign = elem.fieldAlign
	proto.Len = uintptr(count)
	proto.SliceType = SliceOf(elem)

	proto.kind &^= kindNoPointers
	switch {
	case !elem.hasPointers() || proto.size == 0:
		// No pointers.
		proto.kind |= kindNoPointers
		proto.gcData = nil
		proto.ptrData = 0

	case count == 1:
		// In memory, 1-element array looks just like the element.
		proto.kind |= elem.kind & kindGCProg
		proto.gcData = elem.gcData
		proto.ptrData = elem.ptrData

	case elem.canHandleGC() && proto.size <= maxPtrMaskBytes*8*PtrSize:
		// Element is small with pointer mask; array is still small.
		// Create direct pointer mask by turning each 1 bit in elem
		// into count 1 bits in larger mask.
		mask := make([]byte, (proto.ptrData/PtrSize+7)/8)
		elemMask := (*[1 << 30]byte)(unsafe.Pointer(elem.gcData))[:]
		elemWords := elem.size / PtrSize
		for j := uintptr(0); j < elem.ptrData/PtrSize; j++ {
			if (elemMask[j/8]>>(j%8))&1 != 0 {
				for i := uintptr(0); i < proto.Len; i++ {
					k := i*elemWords + j
					mask[k/8] |= 1 << (k % 8)
				}
			}
		}
		proto.gcData = &mask[0]

	default:
		// Create program that emits one element
		// and then repeats to make the array.
		prog := []byte{0, 0, 0, 0} // will be length of prog
		elemGC := (*[1 << 30]byte)(unsafe.Pointer(elem.gcData))[:]
		elemPtrs := elem.ptrData / PtrSize
		if elem.canHandleGC() {
			// Element is small with pointer mask; use as literal bits.
			mask := elemGC
			// Emit 120-bit chunks of full bytes (max is 127 but we avoid using partial bytes).
			var n uintptr
			for n = elemPtrs; n > 120; n -= 120 {
				prog = append(prog, 120)
				prog = append(prog, mask[:15]...)
				mask = mask[15:]
			}
			prog = append(prog, byte(n))
			prog = append(prog, mask[:(n+7)/8]...)
		} else {
			// Element has GC program; emit one element.
			elemProg := elemGC[4 : 4+*(*uint32)(unsafe.Pointer(&elemGC[0]))-1]
			prog = append(prog, elemProg...)
		}
		// Pad from ptrdata to size.
		elemWords := elem.size / PtrSize
		if elemPtrs < elemWords {
			// Emit literal 0 bit, then repeat as needed.
			prog = append(prog, 0x01, 0x00)
			if elemPtrs+1 < elemWords {
				prog = append(prog, 0x81)
				prog = appendVarint(prog, elemWords-elemPtrs-1)
			}
		}
		// Repeat count-1 times.
		if elemWords < 0x80 {
			prog = append(prog, byte(elemWords|0x80))
		} else {
			prog = append(prog, 0x80)
			prog = appendVarint(prog, elemWords)
		}
		prog = appendVarint(prog, uintptr(count)-1)
		prog = append(prog, 0)
		*(*uint32)(unsafe.Pointer(&prog[0])) = uint32(len(prog) - 4)
		proto.kind |= kindGCProg
		proto.gcData = &prog[0]
		proto.ptrData = proto.size // overestimate but ok; must match program
	}

	esize := elem.size
	ealg := elem.alg

	proto.alg = new(algo)
	if ealg.equal != nil {
		eequal := ealg.equal
		proto.alg.equal = func(p, q unsafe.Pointer) bool {
			for i := 0; i < count; i++ {
				pi := arrayAt(p, i, esize)
				qi := arrayAt(q, i, esize)
				if !eequal(pi, qi) {
					return false
				}

			}
			return true
		}
	}
	if ealg.hash != nil {
		ehash := ealg.hash
		proto.alg.hash = func(ptr unsafe.Pointer, seed uintptr) uintptr {
			o := seed
			for i := 0; i < count; i++ {
				o = ehash(arrayAt(ptr, i, esize), o)
			}
			return o
		}
	}

	switch {
	case count == 1 && !elem.isDirectIface():
		// array of 1 direct iface type can be direct
		proto.kind |= kindDirectIface
	default:
		proto.kind &^= kindDirectIface
	}

	return &proto.RType
}

// Get returns the value associated with key in the tag string.
// If there is no such key in the tag, Get returns the empty string.
// If the tag does not have the conventional format, the value
// returned by Get is unspecified. To determine whether a tag is
// explicitly set to the empty string, use Lookup.
func GetTagNamed(tag string, key string) string {
	v, _ := TagLookup(tag, key)
	return v
}

// Lookup returns the value associated with key in the tag string.
// If the key is present in the tag the value (which may be empty)
// is returned. Otherwise the returned value will be the empty string.
// The ok return value reports whether the value was explicitly set in
// the tag string. If the tag does not have the conventional format,
// the value returned by Lookup is unspecified.
func TagLookup(tag string, key string) (string, bool) {
	// When modifying this code, also update the validateStructTag code
	// in cmd/vet/structtag.go.

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := tag[:i]
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := tag[:i+1]
		tag = tag[i+1:]

		if key == name {
			value, err := Unquote(qvalue)
			if err != nil {
				break
			}
			return value, true
		}
	}
	return "", false
}

// Zero returns a Value representing the zero value for the specified type.
// The result is different from the zero value of the Value struct,
// which represents no value at all.
// For example, Zero(TypeOf(42)) returns a Value with Kind Int and value 0.
// The returned value is neither addressable nor settable.
func Zero(typ *RType) Value {
	if typ == nil {
		if willPrintDebug {
			panic("reflect: Zero(nil)")
		}
	}
	if typ.isDirectIface() {
		return Value{Type: typ, Ptr: unsafeNew(typ), Flag: Flag(typ.Kind()) | pointerFlag}
	}
	return Value{Type: typ, Ptr: nil, Flag: Flag(typ.Kind())}
}

// MakeMap creates a new map with the specified type.
func MakeMap(typ *RType) MapValue {
	return MakeMapWithSize(typ, 0)
}

// MakeMapWithSize creates a new map with the specified type
// and initial space for approximately n elements.
func MakeMapWithSize(typ *RType, n int) MapValue {
	if typ.Kind() != Map {
		if willPrintDebug {
			panic("reflect.MakeMapWithSize of non-map type")
		}
	}
	m := makemap(typ, n)
	return MapValue{Value: Value{Type: typ, Ptr: m, Flag: Flag(Map)}}
}

// Copy copies the contents of src into dst until either
// dst has been filled or src has been exhausted.
// It returns the number of elements copied.
// Dst and src each must have kind Slice or Array, and
// dst and src must have the same element type.
//
// As a special case, src can have kind String if the element type of dst is kind Uint8.
func Copy(dest, src SliceValue) (int, bool) {
	dKind := dest.Kind()
	if dKind != Array && dKind != Slice {
		if willPrintDebug {
			panic("reflect.Copy: destination not array or slice")
		}
	}
	if dKind == Array {
		if !dest.IsValid() || !dest.CanSet() {
			if willPrintDebug {
				panic("reflect.Copy: destination must be assignable")
			}
		}
	}
	if !dest.IsValid() || !dest.isExported() {
		if willPrintDebug {
			panic("reflect.Copy: destination must be exported")
		}
	}
	if !src.IsValid() || !src.isExported() {
		if willPrintDebug {
			panic("reflect.Copy: source must be exported")
		}
	}

	destKind := dest.Type.Kind()

	sKind := src.Kind()
	var stringCopy bool
	if sKind != Array && sKind != Slice {
		hasUTF8 := false
		if destKind == Array {
			hasUTF8 = dest.Type.ConvToArray().ElemType.Kind() == Uint8
		} else if destKind == Slice {
			hasUTF8 = dest.Type.ConvToSlice().ElemType.Kind() == Uint8
		}
		stringCopy = sKind == String && hasUTF8
		if !stringCopy {
			if willPrintDebug {
				panic("reflect.CopySlice: source not array, slice or string")
			}
		}
	}

	var de *RType
	if destKind == Array {
		de = dest.Type.ConvToArray().ElemType
	} else if destKind == Slice {
		de = dest.Type.ConvToSlice().ElemType
	}

	if !stringCopy {
		var se *RType
		if src.Type.Kind() == Array {
			se = src.Type.ConvToArray().ElemType
		} else if src.Type.Kind() == Slice {
			se = src.Type.ConvToSlice().ElemType
		}
		if de != se {
			if willPrintDebug {
				panic("Unmatched types " + TypeToString(de) + " != " + TypeToString(se))
			}
		}
	}

	var ds, ss sliceHeader
	if dKind == Array {
		ds.Data = dest.Ptr
		ds.Len = dest.Len()
		ds.Cap = ds.Len
	} else {
		ds = *(*sliceHeader)(dest.Ptr)
	}
	if sKind == Array {
		ss.Data = src.Ptr
		ss.Len = src.Len()
		ss.Cap = ss.Len
	} else if sKind == Slice {
		ss = *(*sliceHeader)(src.Ptr)
	} else {
		sh := *(*stringHeader)(src.Ptr)
		ss.Data = sh.Data
		ss.Len = sh.Len
		ss.Cap = sh.Len
	}

	return typedslicecopy(de, ds, ss), true
}

func NewSlice(ofType *RType) SliceValue {
	if ofType == nil {
		if willPrintDebug {
			panic("reflect: New(nil)")
		}
	}
	newPtr := unsafeNew(ofType)
	return SliceValue{Value: Value{Type: ofType.PtrTo(), Ptr: newPtr, Flag: Flag(Ptr)}}
}

// MakeSlice creates a new zero-initialized slice value for the specified slice type, length, and capacity.
func MakeSlice(ofType *RType, len, cap int) SliceValue {
	if ofType.Kind() != Slice {
		if willPrintDebug {
			panic("reflect.MakeSlice of non-slice type")
		}
	}
	if len < 0 {
		if willPrintDebug {
			panic("reflect.MakeSlice: negative len")
		}
	}
	if cap < 0 {
		if willPrintDebug {
			panic("reflect.MakeSlice: negative cap")
		}
	}
	if len > cap {
		if willPrintDebug {
			panic("reflect.MakeSlice: len > cap")
		}
	}
	s := sliceHeader{unsafeNewArray(ofType.ConvToSlice().ElemType, cap), len, cap}
	return SliceValue{Value: Value{Type: ofType, Ptr: unsafe.Pointer(&s), Flag: pointerFlag | Flag(Slice)}}
}

// Swapper returns a function that swaps the elements in the provided
// slice.
//
// Swapper panics if the provided interface is not a slice.
// Used in `sort` package, slice.go
func Swapper(slice interface{}) func(i, j int) {
	possibleSlice := ReflectOn(slice)
	v := ToSlice(possibleSlice)
	if !v.IsValid() {
		panic("reflect.Swapper: parameter is not a slice, but " + StringKind(possibleSlice.Kind()))
	}
	// Fast path for slices of size 0 and 1. Nothing to swap.
	switch v.Len() {
	case 0:
		return func(i, j int) { panic("reflect.Swapper: slice index out of range") }
	case 1:
		return func(i, j int) {
			if i != 0 || j != 0 {
				panic("reflect.Swapper: slice index out of range")
			}
		}
	}

	typ := v.Type.ConvToSlice().ElemType
	size := typ.size
	hasPtr := typ.hasPointers()

	// Some common & small cases, without using memmove:
	if hasPtr {
		if size == PtrSize {
			ps := *(*[]unsafe.Pointer)(v.Ptr)
			return func(i, j int) { ps[i], ps[j] = ps[j], ps[i] }
		}
		if typ.Kind() == String {
			ss := *(*[]string)(v.Ptr)
			return func(i, j int) { ss[i], ss[j] = ss[j], ss[i] }
		}
	} else {
		switch size {
		case 8:
			is := *(*[]int64)(v.Ptr)
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		case 4:
			is := *(*[]int32)(v.Ptr)
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		case 2:
			is := *(*[]int16)(v.Ptr)
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		case 1:
			is := *(*[]int8)(v.Ptr)
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		}
	}

	s := (*sliceHeader)(v.Ptr)
	tmp := unsafeNew(typ) // swap scratch space

	return func(i, j int) {
		if uint(i) >= uint(s.Len) || uint(j) >= uint(s.Len) {
			panic("reflect.Swapper: slice index out of range")
		}
		val1 := arrayAt(s.Data, i, size)
		val2 := arrayAt(s.Data, j, size)
		typedmemmove(typ, tmp, val1)
		typedmemmove(typ, val1, val2)
		typedmemmove(typ, val2, tmp)
	}
}

func TypeToString(t *RType) string {
	return string(t.nomen())
}

func StringKind(k Kind) string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind" + I2A(int(k), -1)
}

// UnquoteChar decodes the first character or byte in the escaped string
// or character literal represented by the string s.
// It returns four values:
//
//	1) value, the decoded Unicode code point or byte value;
//	2) multibyte, a boolean indicating whether the decoded character requires a multibyte UTF-8 representation;
//	3) tail, the remainder of the string after the character; and
//	4) an error that will be nil if the character is syntactically valid.
//
// The second argument, quote, specifies the type of literal being parsed
// and therefore which escaped quote character is permitted.
// If set to a single quote, it permits the sequence \' and disallows unescaped '.
// If set to a double quote, it permits \" and disallows unescaped ".
// If set to zero, it does not permit either escape and allows both quote characters to appear unescaped.
func UnquoteChar(s string, quote byte) (value rune, multibyte bool, tail string, err error) {
	// easy cases
	switch c := s[0]; {
	case c == quote && (quote == '\'' || quote == '"'):
		err = ErrSyntax
		return
	case c >= utf8.RuneSelf:
		r, size := utf8.DecodeRuneInString(s)
		return r, true, s[size:], nil
	case c != '\\':
		return rune(s[0]), false, s[1:], nil
	}

	// hard case: c is backslash
	if len(s) <= 1 {
		err = ErrSyntax
		return
	}
	c := s[1]
	s = s[2:]

	switch c {
	case 'a':
		value = '\a'
	case 'b':
		value = '\b'
	case 'f':
		value = '\f'
	case 'n':
		value = '\n'
	case 'r':
		value = '\r'
	case 't':
		value = '\t'
	case 'v':
		value = '\v'
	case 'x', 'u', 'U':
		n := 0
		switch c {
		case 'x':
			n = 2
		case 'u':
			n = 4
		case 'U':
			n = 8
		}
		var v rune
		if len(s) < n {
			err = ErrSyntax
			return
		}
		for j := 0; j < n; j++ {
			x, ok := unhex(s[j])
			if !ok {
				err = ErrSyntax
				return
			}
			v = v<<4 | x
		}
		s = s[n:]
		if c == 'x' {
			// single-byte string, possibly not UTF-8
			value = v
			break
		}
		if v > utf8.MaxRune {
			err = ErrSyntax
			return
		}
		value = v
		multibyte = true
	case '0', '1', '2', '3', '4', '5', '6', '7':
		v := rune(c) - '0'
		if len(s) < 2 {
			err = ErrSyntax
			return
		}
		for j := 0; j < 2; j++ { // one digit already; two more
			x := rune(s[j]) - '0'
			if x < 0 || x > 7 {
				err = ErrSyntax
				return
			}
			v = (v << 3) | x
		}
		s = s[2:]
		if v > 255 {
			err = ErrSyntax
			return
		}
		value = v
	case '\\':
		value = '\\'
	case '\'', '"':
		if c != quote {
			err = ErrSyntax
			return
		}
		value = rune(c)
	default:
		err = ErrSyntax
		return
	}
	tail = s
	return
}

// Unquote interprets s as a single-quoted, double-quoted,
// or backquoted Go string literal, returning the string value
// that s quotes.  (If s is single-quoted, it would be a Go
// character literal; Unquote returns the corresponding
// one-character string.)
func Unquote(s string) (string, error) {
	n := len(s)
	if n < 2 {
		return "", ErrSyntax
	}
	quote := s[0]
	if quote != s[n-1] {
		return "", ErrSyntax
	}
	s = s[1 : n-1]

	if quote == '`' {
		if contains(s, '`') {
			return "", ErrSyntax
		}
		if contains(s, '\r') {
			// -1 because we know there is at least one \r to remove.
			buf := make([]byte, 0, len(s)-1)
			for i := 0; i < len(s); i++ {
				if s[i] != '\r' {
					buf = append(buf, s[i])
				}
			}
			return string(buf), nil
		}
		return s, nil
	}
	if quote != '"' && quote != '\'' {
		return "", ErrSyntax
	}
	if contains(s, '\n') {
		return "", ErrSyntax
	}

	// Is it trivial? Avoid allocation.
	if !contains(s, '\\') && !contains(s, quote) {
		switch quote {
		case '"':
			return s, nil
		case '\'':
			r, size := utf8.DecodeRuneInString(s)
			if size == len(s) && (r != utf8.RuneError || size != 1) {
				return s, nil
			}
		}
	}

	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	for len(s) > 0 {
		c, multibyte, ss, err := UnquoteChar(s, quote)
		if err != nil {
			return "", err
		}
		s = ss
		if c < utf8.RuneSelf || !multibyte {
			buf = append(buf, byte(c))
		} else {
			n := utf8.EncodeRune(runeTmp[:], c)
			buf = append(buf, runeTmp[:n]...)
		}
		if quote == '\'' && len(s) != 0 {
			// single-quoted must be single character
			return "", ErrSyntax
		}
	}
	return string(buf), nil
}

// BytesToString effectively converts bytes to string
// nolint: gas
func BytesToString(src []byte) string {
	return *(*string)(unsafe.Pointer(&src))
}

// StringToBytes effectively converts string to bytes
// nolint: gas
func StringToBytes(src string) []byte {
	strstruct := StringStructOf(&src)
	return *(*[]byte)(unsafe.Pointer(&sliceHeader{
		Data: strstruct.Data,
		Len:  strstruct.Len,
		Cap:  strstruct.Len,
	}))
}

func StringStructOf(src *string) *stringHeader {
	return (*stringHeader)(unsafe.Pointer(src))
}

// Atoi parses an int from a string s.
// The bool result reports whether s is a number representable by a value of type int.
func Atoi(src string) (int, bool) {
	if len(src) == 0 {
		return 0, false
	}

	negative := false
	if src[0] == '-' {
		negative = true
		src = src[1:]
	}

	unsignedResult := uint(0)
	for i := 0; i < len(src); i++ {
		c := src[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		if unsignedResult > maxUint/10 {
			// overflow
			return 0, false
		}
		unsignedResult *= 10
		un1 := unsignedResult + uint(c) - '0'
		if un1 < unsignedResult {
			// overflow
			return 0, false
		}
		unsignedResult = un1
	}

	if !negative && unsignedResult > uint(maxInt) {
		return 0, false
	}
	if negative && unsignedResult > uint(maxInt)+1 {
		return 0, false
	}

	result := int(unsignedResult)
	if negative {
		result = -result
	}

	return result, true
}

// Atoi32 is like Atoi but for integers
// that fit into an int32.
func Atoi32(src string) (int32, bool) {
	if n, ok := Atoi(src); n == int(int32(n)) {
		return int32(n), ok
	}
	return 0, false
}

// MakeFunc returns a new function of the given Type
// that wraps the function fn. When called, that new function
// does the following:
//
//	- converts its arguments to a slice of Values.
//	- runs results := fn(args).
//	- returns the results as a slice of Values, one per formal result.
//
// The implementation fn can assume that the argument Value slice
// has the number and type of arguments given by typ.
// If typ describes a variadic function, the final Value is itself
// a slice representing the variadic arguments, as in the
// body of a variadic function. The result Value slice returned by fn
// must have the number and type of results given by typ.
//
// The Value.Call method allows the caller to invoke a typed function
// in terms of Values; in contrast, MakeFunc allows the caller to implement
// a typed function in terms of Values.
//
// The Examples section of the documentation includes an illustration
// of how to use MakeFunc to build a swap function for different types.
//
func MakeFunc(typ *RType, fn func(args []Value) (results []Value)) Value {
	if typ.Kind() != Func {
		panic("reflect: call of MakeFunc with non-Func type")
	}

	ftyp := (*funcType)(unsafe.Pointer(typ))

	// Indirect Go func value (dummy) to obtain
	// actual code address. (A Go func value is a pointer
	// to a C function pointer. https://golang.org/s/go11func.)
	dummy := stubFunction
	code := **(**uintptr)(unsafe.Pointer(&dummy))

	// makeFuncImpl contains a stack map for use by the runtime
	_, _, _, stack := funcLayout(typ, nil)

	impl := &makeFuncImpl{code: code, stack: stack, typ: ftyp, fn: fn}

	return Value{Type: typ, Ptr: unsafe.Pointer(impl), Flag: Flag(Func)}
}
