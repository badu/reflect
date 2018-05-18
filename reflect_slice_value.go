/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

// Index returns v's i'th element.
func (v SliceValue) Index(i int) Value {
	switch v.Kind() {
	case Array:
		tt := v.Type.ConvToArray()
		if uint(i) >= uint(tt.Len) {
			if willPrintDebug {
				println("reflect.SliceValue.Index: array index out of range")
			}
			return Value{}
		}
		typ := tt.ElemType
		offset := uintptr(i) * typ.size
		// Either pointerFlag is set and v.ptr points at array, or pointerFlag is not set and v.ptr is the actual array data.
		// In the former case, we want v.ptr + offset.
		// In the latter case, we must be doing Index(0), so offset = 0, so v.ptr + offset is still the correct address.
		val := add(v.Ptr, offset)
		fl := v.Flag&(pointerFlag|addressableFlag) | v.ro() | Flag(typ.Kind()) // bits same as overall array
		return Value{Type: typ, Ptr: val, Flag: fl}

	case Slice:
		// Element flag same as Deref of ptr.
		// Addressable, indirect, possibly read-only.
		s := (*sliceHeader)(v.Ptr)
		if uint(i) >= uint(s.Len) {
			if willPrintDebug {
				println("reflect.SliceValue.Index: slice index out of range")
			}
			return Value{}
		}
		typ := v.Type.ConvToSlice().ElemType
		val := arrayAt(s.Data, i, typ.size)
		fl := addressableFlag | pointerFlag | v.ro() | Flag(typ.Kind())
		return Value{Type: typ, Ptr: val, Flag: fl}

	case String:
		s := (*stringHeader)(v.Ptr)
		if uint(i) >= uint(s.Len) {
			if willPrintDebug {
				println("reflect.SliceValue.Index: string index out of range")
			}
			return Value{}
		}
		p := arrayAt(s.Data, i, 1)
		fl := v.ro() | Flag(Uint8) | pointerFlag
		return Value{Type: uint8Type, Ptr: p, Flag: fl}

	default:
		// kind checks are performed in public ToSlice(), so this should NEVER happen
		if willPrintDebug {
			println("reflect.SliceValue.Slice : unknown kind `" + StringKind(v.Kind()) + "`. How did you got here?")
		}
		return Value{}
	}
}

// Len returns v's length.
func (v SliceValue) Len() int {
	switch v.Kind() {
	case Array:
		return int(v.Type.ConvToArray().Len)
	case Slice:
		return (*sliceHeader)(v.Ptr).Len // Slice is bigger than a word; assume pointerFlag.
	case String:
		return (*stringHeader)(v.Ptr).Len // String is bigger than a word; assume pointerFlag.
	default:
		if willPrintDebug {
			println("reflect.SliceValue.Len : unknown kind `" + StringKind(v.Kind()) + "`. How did you got here?")
		}
		return 0 // The length of "unknown"
	}

}

// Bytes returns v's underlying value.
func (v SliceValue) Bytes() []byte {
	if v.Kind() != Slice {
		if willPrintDebug {
			println("reflect.SliceValue.Bytes: kind not slice (`" + StringKind(v.Kind()) + "`)")
		}
		return nil
	}
	if v.Type.ConvToSlice().ElemType.Kind() != Uint8 {
		if willPrintDebug {
			println("reflect.SliceValue.Bytes of non-byte slice")
		}
		return nil
	}
	// Slice is always bigger than a word; assume pointerFlag.
	return *(*[]byte)(v.Ptr)
}

// SetBytes sets v's underlying value.
func (v SliceValue) SetBytes(x []byte) {
	if !v.IsValid() || !v.CanSet() || v.Kind() != Slice {
		if willPrintDebug {
			println("reflect.SliceValue.SetBytes: kind not slice (`" + StringKind(v.Kind()) + "`) or invalid or not settable")
		}
		return
	}
	if v.Type.ConvToSlice().ElemType.Kind() != Uint8 {
		if willPrintDebug {
			println("reflect.Value.SetBytes of non-byte slice")
		}
		return
	}
	*(*[]byte)(v.Ptr) = x
}

// runes returns v's underlying value.
func (v SliceValue) Runes() []rune {
	if v.Type.ConvToSlice().ElemType.Kind() != Int32 {
		if willPrintDebug {
			println("reflect.SliceValue.Runes of non-rune slice")
		}
		return nil
	}
	// Slice is always bigger than a word; assume pointerFlag.
	return *(*[]rune)(v.Ptr)
}

// SetRunes sets v's underlying value.
func (v SliceValue) SetRunes(x []rune) {
	if !v.IsValid() || !v.CanSet() {
		return
	}
	if v.Type.ConvToSlice().ElemType.Kind() != Int32 {
		if willPrintDebug {
			println("reflect.SliceValue.SetRunes of non-rune slice")
		}
		return
	}
	*(*[]rune)(v.Ptr) = x
}

// Cap returns v's capacity.
func (v SliceValue) Cap() int {
	switch v.Kind() {
	case Array:
		return int(v.Type.ConvToArray().Len)
	case Slice:
		return (*sliceHeader)(v.Ptr).Cap // Slice is always bigger than a word; assume pointerFlag.
	case String:
		return (*stringHeader)(v.Ptr).Len
	default:
		// kind checks are performed in public ToSlice(), so this should NEVER happen
		if willPrintDebug {
			println("reflect.Iterable.Slice : unknown kind `" + StringKind(v.Kind()) + "`. How did you got here?")
		}
		return 0 // The Cap of "unknown"
	}
}

// SetLen sets v's length to n.
func (v SliceValue) SetLen(n int) {
	if v.Kind() != Slice || !v.IsValid() || !v.CanSet() {
		if willPrintDebug {
			println("reflect.SliceValue.SetLen: kind not slice (`" + StringKind(v.Kind()) + "`) or invalid or not settable")
		}
		return
	}
	header := (*sliceHeader)(v.Ptr)
	if uint(n) > uint(header.Cap) {
		if willPrintDebug {
			println("reflect.SliceValue.SetLen: slice length out of range")
		}
		return
	}
	header.Len = n
}

// SetCap sets v's capacity to n.
func (v SliceValue) SetCap(n int) {
	if v.Kind() != Slice || !v.IsValid() || !v.CanSet() {
		if willPrintDebug {
			println("reflect.SliceValue.SetCap: kind not slice (`" + StringKind(v.Kind()) + "`) or invalid or not settable")
		}
		return
	}
	header := (*sliceHeader)(v.Ptr)
	if n < header.Len || n > header.Cap {
		if willPrintDebug {
			println("reflect.SliceValue.SetCap: slice capacity out of range in SetCap")
		}
		return
	}
	header.Cap = n
}

// Slice returns v[i:j].
func (v SliceValue) Slice(i, j int) SliceValue {
	var (
		cap   int
		slice *sliceType
		base  ptr
	)
	switch v.Kind() {
	case Array:
		if !v.CanAddr() {
			if willPrintDebug {
				println("reflect.Value.Slice: slice of unaddressable array")
			}
			return v
		}
		array := v.Type.ConvToArray()
		cap = int(array.Len)
		slice = array.SliceType.ConvToSlice()
		base = v.Ptr
	case Slice:
		slice = v.Type.ConvToSlice()
		header := (*sliceHeader)(v.Ptr)
		base = header.Data
		cap = header.Cap
	case String:
		header := (*stringHeader)(v.Ptr)
		if i < 0 || j < i || j > header.Len {
			if willPrintDebug {
				println("reflect.Value.Slice: string slice index out of bounds")
			}
			return v
		}
		var finalHeader stringHeader
		if i < header.Len {
			finalHeader = stringHeader{arrayAt(header.Data, i, 1), j - i}
		}
		return SliceValue{Value: Value{Type: v.Type, Ptr: ptr(&finalHeader), Flag: v.Flag}}
	default:
		// kind checks are performed in public ToSlice(), so this should NEVER happen
		if willPrintDebug {
			println("reflect.Iterable.Slice : unknown kind `" + StringKind(v.Kind()) + "`. How did you got here?")
		}
		return v
	}

	if i < 0 || j < i || j > cap {
		if willPrintDebug {
			println("reflect.Value.Slice: index out of bounds")
		}
		return v
	}

	// Declare slice so that gc can see the base pointer in it.
	var slicePtrs []ptr

	// Reinterpret as *sliceHeader to edit.
	header := (*sliceHeader)(ptr(&slicePtrs))
	header.Len = j - i
	header.Cap = cap - i
	if cap-i > 0 {
		header.Data = arrayAt(base, i, slice.ElemType.size)
	} else {
		// do not advance pointer, to avoid pointing beyond end of slice
		header.Data = base
	}
	// make a flag to mark a pointer to a slice
	fl := v.ro() | pointerFlag | Flag(Slice)
	return SliceValue{Value: Value{Type: &slice.RType, Ptr: ptr(&slicePtrs), Flag: fl}}
}

// Slice3 is the 3-index form of the slice operation: it returns v[i:j:k].
func (v SliceValue) Slice3(i, j, k int) SliceValue {
	var (
		cap   int
		slice *sliceType
		base  ptr
	)
	switch v.Kind() {
	case Array:
		if !v.CanAddr() {
			if willPrintDebug {
				println("reflect.SliceValue.Slice3: slice of unaddressable array")
			}
			return v
		}
		array := v.Type.ConvToArray()
		cap = int(array.Len)
		slice = array.SliceType.ConvToSlice()
		base = v.Ptr
	case Slice:
		slice = v.Type.ConvToSlice()
		s := (*sliceHeader)(v.Ptr)
		base = s.Data
		cap = s.Cap
	case String:
		//return v[i:j:k] of String
		header := (*stringHeader)(v.Ptr)
		if i < 0 || j < i || j > header.Len {
			if willPrintDebug {
				println("reflect.Value.Slice: string slice index out of bounds")
			}
			return v
		}
		var finalHeader stringHeader
		if i < header.Len {
			finalHeader = stringHeader{arrayAt(header.Data, i, 1), j - i}
		}
		return SliceValue{Value: Value{Type: v.Type, Ptr: ptr(&finalHeader), Flag: v.Flag}}
	default:
		// kind checks are performed in public ToSlice(), so this should NEVER happen
		if willPrintDebug {
			println("reflect.Iterable.Slice : unknown kind `" + StringKind(v.Kind()) + "`. How did you got here?")
		}
		return v
	}

	if i < 0 || j < i || k < j || k > cap {
		if willPrintDebug {
			println("reflect.SliceValue.Slice3: index out of bounds")
		}
		return v
	}

	// Declare slice so that the garbage collector can see the base pointer in it.
	var slicePtrs []ptr

	// Reinterpret as *sliceHeader to edit.
	header := (*sliceHeader)(ptr(&slicePtrs))
	header.Len = j - i
	header.Cap = k - i
	if k-i > 0 {
		header.Data = arrayAt(base, i, slice.ElemType.size)
	} else {
		// do not advance pointer, to avoid pointing beyond end of slice
		header.Data = base
	}
	// make a flag to mark a pointer to a slice
	fl := v.ro() | pointerFlag | Flag(Slice)
	return SliceValue{Value: Value{Type: &slice.RType, Ptr: ptr(&slicePtrs), Flag: fl}}
}

// AppendSlice appends a slice t to a slice s and returns the resulting slice.
// The slices s and t must have the same element type.
func (v SliceValue) AppendWithSlice(slice SliceValue) SliceValue {
	// TODO : check if it works for String
	// TODO : do nothing if Array
	if v.Type.ConvToSlice().ElemType != slice.Type.ConvToSlice().ElemType {
		if willPrintDebug {
			println("reflect.SliceValue.AppendWithSlice : unmatched types " + TypeToString(v.Type.ConvToSlice().ElemType) + " and " + TypeToString(slice.Type.ConvToSlice().ElemType))
		}
		return v
	}
	src, i0, i1 := grow(v, slice.Len())
	Copy(src.Slice(i0, i1), slice)
	return src
}

// Append appends the values x to a slice s and returns the resulting slice.
// As in Go, each x's value must be assignable to the slice's element type.
func (v SliceValue) Append(values ...Value) SliceValue {
	// TODO : check if it works for String
	// TODO : do nothing if Array
	src, i0, i1 := grow(v, len(values))
	for i, j := i0, 0; i < i1; i, j = i+1, j+1 {
		src.Index(i).Set(values[j])
	}
	return src
}
