/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import "unsafe"

// During deepValueEqual, must keep track of checks that are
// in progress. The comparison algorithm assumes that all
// checks in progress are true when it reencounters them.
// Visited comparisons are stored in a map indexed by visit.
type visit struct {
	a1  unsafe.Pointer
	a2  unsafe.Pointer
	typ *RType
}

// Tests for deep equality using reflected types. The map argument tracks
// comparisons that have already been seen, which allows short circuiting on
// recursive types.
func deepValueEqual(v1, v2 Value, visited map[visit]bool, depth int) bool {
	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}
	if v1.Type != v2.Type {
		return false
	}

	// We want to avoid putting more in the visited map than we need to.
	// For any possible reference cycle that might be encountered,
	// hard(t) needs to return true for at least one of the types in the cycle.
	hard := func(k Kind) bool {
		switch k {
		case Map, Slice, Ptr, Interface:
			return true
		}
		return false
	}

	if v1.CanAddr() && v2.CanAddr() && hard(v1.Kind()) {
		addr1 := unsafe.Pointer(unsafeAddr(v1))
		addr2 := unsafe.Pointer(unsafeAddr(v2))
		if uintptr(addr1) > uintptr(addr2) {
			// Canonicalize order to reduce number of entries in visited.
			// Assumes non-moving garbage collector.
			addr1, addr2 = addr2, addr1
		}

		// Short circuit if references are already seen.
		typ := v1.Type
		v := visit{addr1, addr2, typ}
		if visited[v] {
			return true
		}

		// Remember for later.
		visited[v] = true
	}

	switch v1.Kind() {
	case Array:
		tv1 := v1.Type.ConvToArray()
		typ1 := tv1.ElemType
		fl1 := v1.Flag&(pointerFlag|addressableFlag) | v1.ro() | Flag(typ1.Kind())
		tv2 := v2.Type.ConvToArray()
		typ2 := tv2.ElemType
		fl2 := v2.Flag&(pointerFlag|addressableFlag) | v2.ro() | Flag(typ2.Kind())
		for i := 0; i < int(tv1.Len); i++ {
			offset1 := uintptr(i) * typ1.size
			offset2 := uintptr(i) * typ2.size
			val1 := add(v1.Ptr, offset1)
			val2 := add(v2.Ptr, offset2)
			if !deepValueEqual(Value{Type: typ1, Ptr: val1, Flag: fl1}, Value{Type: typ2, Ptr: val2, Flag: fl2}, visited, depth+1) {
				return false
			}
		}
		return true
	case Slice:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		s1 := (*sliceHeader)(v1.Ptr)
		s2 := (*sliceHeader)(v2.Ptr)
		if s1.Len != s2.Len {
			return false
		}
		if convToSliceHeader(v1.Ptr).Data == convToSliceHeader(v2.Ptr).Data {
			return true
		}
		typ1 := v1.Type.ConvToSlice().ElemType
		fl1 := addressableFlag | pointerFlag | v1.ro() | Flag(typ1.Kind())
		typ2 := v2.Type.ConvToSlice().ElemType
		fl2 := addressableFlag | pointerFlag | v2.ro() | Flag(typ2.Kind())

		for i := 0; i < s1.Len; i++ {
			val1 := arrayAt(s1.Data, i, typ1.size)
			val2 := arrayAt(s2.Data, i, typ2.size)
			if !deepValueEqual(Value{Type: typ1, Ptr: val1, Flag: fl1}, Value{Type: typ2, Ptr: val2, Flag: fl2}, visited, depth+1) {
				return false
			}
		}
		return true
	case Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		return deepValueEqual(v1.Iface(), v2.Iface(), visited, depth+1)
	case Ptr:
		if v1.pointer() == v2.pointer() {
			return true
		}
		return deepValueEqual(v1.Deref(), v2.Deref(), visited, depth+1)
	case Struct:
		sv1 := StructValue{Value: v1}
		sv2 := StructValue{Value: v2}
		numFields := len(sv1.Type.convToStruct().fields)
		for i, n := 0, numFields; i < n; i++ {
			if !deepValueEqual(sv1.Field(i), sv2.Field(i), visited, depth+1) {
				return false
			}
		}
		return true
	case Map:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if maplen(v1.pointer()) != maplen(v2.pointer()) {
			return false
		}
		if v1.pointer() == v2.pointer() {
			return true
		}
		v1map := MapValue{Value: v1}
		v2map := MapValue{Value: v2}
		for _, k := range v1map.MapKeys() {
			val1 := v1map.MapIndex(k)
			val2 := v2map.MapIndex(k)
			if !val1.IsValid() || !val2.IsValid() || !deepValueEqual(v1map.MapIndex(k), v2map.MapIndex(k), visited, depth+1) {
				return false
			}
		}
		return true
	case Chan:
		return false
	case Func:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		// Can't do better than this:
		return false
	default:
		// Normal equality suffices
		return v1.valueInterface() == v2.valueInterface()
	}
}

// UnsafeAddr returns a pointer to v's data.
// It is for advanced clients that also import the "unsafe" package.
// It panics if v is not addressable.
func unsafeAddr(v Value) uintptr {
	if v.Type == nil {
		if willPrintDebug {
			panic("reflect.x.error : unaddressable nil type")
		}
	}
	if !v.CanAddr() {
		if willPrintDebug {
			panic("reflect.x.error : unaddressable value")
		}
	}
	return uintptr(v.Ptr)
}

// DeepEqual reports whether x and y are ``deeply equal,'' defined as follows.
// Two values of identical type are deeply equal if one of the following cases applies.
// Values of distinct types are never deeply equal.
//
// Array values are deeply equal when their corresponding elements are deeply equal.
//
// Struct values are deeply equal if their corresponding fields,
// both exported and unexported, are deeply equal.
//
// Func values are deeply equal if both are nil; otherwise they are not deeply equal.
//
// Interface values are deeply equal if they hold deeply equal concrete values.
//
// Map values are deeply equal when all of the following are true:
// they are both nil or both non-nil, they have the same length,
// and either they are the same map object or their corresponding keys
// (matched using Go equality) map to deeply equal values.
//
// Pointer values are deeply equal if they are equal using Go's == operator
// or if they point to deeply equal values.
//
// Slice values are deeply equal when all of the following are true:
// they are both nil or both non-nil, they have the same length,
// and either they point to the same initial entry of the same underlying array
// (that is, &x[0] == &y[0]) or their corresponding elements (up to length) are deeply equal.
// Note that a non-nil empty slice and a nil slice (for example, []byte{} and []byte(nil))
// are not deeply equal.
//
// Other values - numbers, bools, strings, and channels - are deeply equal
// if they are equal using Go's == operator.
//
// In general DeepEqual is a recursive relaxation of Go's == operator.
// However, this idea is impossible to implement without some inconsistency.
// Specifically, it is possible for a value to be unequal to itself,
// either because it is of func type (uncomparable in general)
// or because it is a floating-point NaN value (not equal to itself in floating-point comparison),
// or because it is an array, struct, or interface containing
// such a value.
// On the other hand, pointer values are always equal to themselves,
// even if they point at or contain such problematic values,
// because they compare equal using Go's == operator, and that
// is a sufficient condition to be deeply equal, regardless of content.
// DeepEqual has been defined so that the same short-cut applies
// to slices and maps: if x and y are the same slice or the same map,
// they are deeply equal regardless of content.
//
// As DeepEqual traverses the data values it may find a cycle. The
// second and subsequent times that DeepEqual compares two pointer
// values that have been compared before, it treats the values as
// equal rather than examining the values to which they point.
// This ensures that DeepEqual terminates.
func DeepEqual(x, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}
	v1 := ReflectOn(x)
	v2 := ReflectOn(y)
	if v1.Type != v2.Type {
		return false
	}
	return deepValueEqual(v1, v2, make(map[visit]bool), 0)
}
