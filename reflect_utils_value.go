/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import (
	"runtime"
	"unsafe"
)

// arrayAt returns the i-th element of p,  an array whose elements are eltSize bytes wide.
// The array pointed at by p must have at least i+1 elements: it is invalid (but impossible to check here) to pass i >= len, because then the result will point outside the array. the benefit is to surface this assumption at the call site.)
func arrayAt(p unsafe.Pointer, i int, eltSize uintptr) unsafe.Pointer {
	return add(p, uintptr(i)*eltSize)
}
func align(x, n uintptr) uintptr                      { return (x + n - 1) &^ (n - 1) } // align returns the result of rounding x up to a multiple of n. n must be a power of two.
func convPtr(p unsafe.Pointer) unsafe.Pointer         { return *(*unsafe.Pointer)(p) }
func loadConvPtr(p unsafe.Pointer, x unsafe.Pointer)  { *(*unsafe.Pointer)(p) = x }
func convIface(p unsafe.Pointer) interface{}          { return *(*interface{})(p) }
func loadConvIface(p unsafe.Pointer, x interface{})   { *(*interface{})(p) = x }
func convIfaceMeth(p unsafe.Pointer) interface{}      { return *(*interface{ M() })(p) }
func convToSliceHeader(p unsafe.Pointer) *SliceHeader { return (*SliceHeader)(p) }
func internalNew(t *RType) Value {
	return Value{Type: t, Ptr: unsafeNew(t), Flag: Flag(t.Kind())&exportFlag | pointerFlag | addressableFlag | Flag(t.Kind())}
}

// makeInt returns a Value of type t equal to bits (possibly truncated),
// where t is a signed or unsigned int type.
func makeInt(f Flag, bits uint64, t *RType) Value {
	newPtr := unsafeNew(t)
	switch t.size {
	case 1:
		*(*uint8)(newPtr) = uint8(bits)
	case 2:
		*(*uint16)(newPtr) = uint16(bits)
	case 4:
		*(*uint32)(newPtr) = uint32(bits)
	case 8:
		*(*uint64)(newPtr) = bits
	}
	return Value{Type: t, Ptr: newPtr, Flag: f | pointerFlag | Flag(t.Kind())}
}

// makeFloat returns a Value of type t equal to v (possibly truncated to float32),
// where t is a float32 or float64 type.
func makeFloat(f Flag, v float64, t *RType) Value {
	newPtr := unsafeNew(t)
	switch t.size {
	case 4:
		*(*float32)(newPtr) = float32(v)
	case 8:
		*(*float64)(newPtr) = v
	}
	return Value{Type: t, Ptr: newPtr, Flag: f | pointerFlag | Flag(t.Kind())}
}

// makeComplex returns a Value of type t equal to v (possibly truncated to complex64),
// where t is a complex64 or complex128 type.
func makeComplex(f Flag, v complex128, t *RType) Value {
	newPtr := unsafeNew(t)
	switch t.size {
	case 8:
		*(*complex64)(newPtr) = complex64(v)
	case 16:
		*(*complex128)(newPtr) = v
	}
	return Value{Type: t, Ptr: newPtr, Flag: f | pointerFlag | Flag(t.Kind())}
}

func makeString(f Flag, s string, t *RType) Value {
	// was :
	// ptr := unsafeNew(t)
	// ret := value{Type: t.PtrTo(), ptr: ptr, flag: flag(Ptr)}.Deref()
	ret := internalNew(t)
	strVal := ret.String()
	if strVal.Debug == "" {
		strVal.Set(s)
	}
	ret.Flag = ret.Flag&^addressableFlag | f
	return ret
}

func makeBytes(f Flag, byt []byte, t *RType) Value {
	// was :
	// ptr := unsafeNew(t)
	// ret := value{Type: t.PtrTo(), ptr: ptr, flag: flag(Ptr)}.Deref()
	ret := internalNew(t)
	//ret.SetBytes(byt)
	*(*[]byte)(ret.Ptr) = byt
	ret.Flag = ret.Flag&^addressableFlag | f
	return ret
}

func makeRunes(f Flag, run []rune, t *RType) Value {
	// was :
	// ptr := unsafeNew(t)
	// ret := value{Type: t.PtrTo(), ptr: ptr, flag: flag(Ptr)}.Deref()
	ret := internalNew(t)
	//ret.SetRunes(run)
	*(*[]rune)(ret.Ptr) = run
	ret.Flag = ret.Flag&^addressableFlag | f
	return ret
}

// grow grows the slice s so that it can hold extra more values, allocating
// more capacity if needed. It also returns the old and new slice lengths.
func grow(s SliceValue, extra int) (SliceValue, int, int) {
	i0 := s.Len()
	i1 := i0 + extra
	if i1 < i0 {
		panic("reflect.x.error : grow: slice overflow")
	}
	m := s.Cap()
	if i1 <= m {
		return s.Slice(0, i1), i0, i1
	}
	if m == 0 {
		m = extra
	} else {
		for m < i1 {
			if i0 < 1024 {
				m += m
			} else {
				m += m / 4
			}
		}
	}
	t := MakeSlice(s.Type, i1, m)
	Copy(t, s)
	return t, i0, i1
}

func bucketOf(ktyp, etyp *RType) *RType {
	// See comment on hmap.overflow in ../runtime/hashmap.go.
	var kind uint8
	if !ktyp.hasPointers() && !etyp.hasPointers() &&
		ktyp.size <= maxKeySize && etyp.size <= maxValSize {
		kind = kindNoPointers
	}

	if ktyp.size > maxKeySize {
		ktyp = ktyp.PtrTo()
	}
	if etyp.size > maxValSize {
		etyp = etyp.PtrTo()
	}

	// Prepare GC data if any.
	// A bucket is at most bucketSize*(1+maxKeySize+maxValSize)+2*ptrSize bytes,
	// or 2072 bytes, or 259 pointer-size words, or 33 bytes of pointer bitmap.
	// Note that since the key and value are known to be <= 128 bytes,
	// they're guaranteed to have bitmaps instead of GC programs.
	var gcdata *byte
	var ptrdata uintptr
	var overflowPad uintptr

	// On NaCl, pad if needed to make overflow end at the proper struct alignment.
	// On other systems, align > ptrSize is not possible.
	if IsAMD64p32 && (ktyp.align > PtrSize || etyp.align > PtrSize) {
		overflowPad = PtrSize
	}
	size := bucketSize*(1+ktyp.size+etyp.size) + overflowPad + PtrSize
	if size&uintptr(ktyp.align-1) != 0 || size&uintptr(etyp.align-1) != 0 {
		panic("reflect.x.error : bad size computation in MapOf")
	}

	if kind != kindNoPointers {
		nptr := (bucketSize*(1+ktyp.size+etyp.size) + PtrSize) / PtrSize
		mask := make([]byte, (nptr+7)/8)
		base := bucketSize / PtrSize

		if ktyp.hasPointers() {
			if !ktyp.canHandleGC() {
				panic("reflect.x.error : unexpected GC program in MapOf")
			}
			kmask := (*[16]byte)(unsafe.Pointer(ktyp.gcData))
			for i := uintptr(0); i < ktyp.ptrData/PtrSize; i++ {
				if (kmask[i/8]>>(i%8))&1 != 0 {
					for j := uintptr(0); j < bucketSize; j++ {
						word := base + j*ktyp.size/PtrSize + i
						mask[word/8] |= 1 << (word % 8)
					}
				}
			}
		}
		base += bucketSize * ktyp.size / PtrSize

		if etyp.hasPointers() {
			if !etyp.canHandleGC() {
				panic("reflect.x.error : unexpected GC program in MapOf")
			}
			emask := (*[16]byte)(unsafe.Pointer(etyp.gcData))
			for i := uintptr(0); i < etyp.ptrData/PtrSize; i++ {
				if (emask[i/8]>>(i%8))&1 != 0 {
					for j := uintptr(0); j < bucketSize; j++ {
						word := base + j*etyp.size/PtrSize + i
						mask[word/8] |= 1 << (word % 8)
					}
				}
			}
		}
		base += bucketSize * etyp.size / PtrSize
		base += overflowPad / PtrSize

		word := base
		mask[word/8] |= 1 << (word % 8)
		gcdata = &mask[0]
		ptrdata = (word + 1) * PtrSize

		// overflow word must be last
		if ptrdata != size {
			panic("reflect.x.error : bad layout computation in MapOf")
		}
	}

	b := &RType{
		align:   PtrSize,
		size:    size,
		kind:    kind,
		ptrData: ptrdata,
		gcData:  gcdata,
	}
	if overflowPad > 0 {
		b.align = 8
	}

	b.str = declareReflectName(newName(byteSliceFromParams(bucketStr, openPar, ktyp.nomen(), comma, etyp.nomen(), closePar)))
	return b
}

// isReflexive reports whether the == operation on the type is reflexive.
// That is, x == x for all values x of type t.
func isReflexive(t *RType) bool {
	switch t.Kind() {
	case Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, UintPtr, Chan, Ptr, String, UnsafePointer:
		return true
	case Float32, Float64, Complex64, Complex128, Interface:
		return false
	case Array:
		return isReflexive(t.ConvToArray().ElemType)
	case Struct:
		for _, field := range t.convToStruct().fields {
			if !isReflexive(field.Type) {
				return false
			}
		}
		return true
	default:
		// Func, Map, Slice, Invalid
		panic("reflect.x.error : isReflexive called on non-key type (Func, Map, Slice, Invalid)")
	}
}

// needKeyUpdate reports whether map overwrites require the key to be copied.
func needKeyUpdate(t *RType) bool {
	switch t.Kind() {
	case Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, UintPtr, Chan, Ptr, UnsafePointer:
		return false
	case Float32, Float64, Complex64, Complex128, Interface, String:
		// Float keys can be updated from +0 to -0.
		// String keys can be updated to use a smaller backing store.
		// Interfaces might have floats of strings in them.
		return true
	case Array:
		return needKeyUpdate(t.ConvToArray().ElemType)
	case Struct:
		for _, field := range t.convToStruct().fields {
			if needKeyUpdate(field.Type) {
				return true
			}
		}
		return false
	default:
		// Func, Map, Slice, Invalid
		panic("reflect.x.error : needKeyUpdate called on non-key type (Func, Map, Slice, Invalid)")
	}
}

func assertE2I(v Value, dst *RType, target unsafe.Pointer) {
	// TODO : @badu - Type links to methods
	//to be read "if NumMethod(dst) == 0{"
	if (dst.Kind() == Interface && dst.NoOfIfaceMethods() == 0) || (dst.Kind() != Interface && lenExportedMethods(dst) == 0) {
		// the case of "interface{}"
		loadConvIface(target, v.valueInterface())
	} else {
		ifaceE2I(dst, v.valueInterface(), target)
	}
}

// convert operation: direct copy
func cvtDirect(v Value, typ *RType) Value {
	f := v.Flag
	valPtr := v.Ptr
	if v.CanAddr() {
		// indirect, mutable word - make a copy
		c := unsafeNew(typ)
		typedmemmove(typ, c, valPtr)
		valPtr = c
		f &^= addressableFlag
	}
	return Value{Type: typ, Ptr: valPtr, Flag: v.ro() | f}
}

// convert operation: concrete -> interface
func cvtT2I(v Value, typ *RType) Value {
	target := unsafeNew(typ)
	assertE2I(v, typ, target)
	return Value{Type: typ, Ptr: target, Flag: v.ro() | pointerFlag | Flag(Interface)}
}

// convert operation: interface -> interface
func cvtI2I(v Value, typ *RType) Value {
	if v.IsNil() {
		ret := Zero(typ)
		ret.Flag |= v.ro()
		return ret
	}
	switch v.Kind() {
	case Ptr:
		return cvtT2I(v.Deref(), typ)
	case Interface:
		return cvtT2I(v.Iface(), typ)
	default:
		return cvtT2I(v, typ)
	}
}

// callMethod is the call implementation used by a function returned by makeMethodValue (used by v.Method(i).Interface()).
// It is a streamlined version of the usual reflect call: the caller has already laid out the argument frame for us, so we don't have to deal with individual Values for each argument.
// It is in this file so that it can be next to the two similar functions above.
// The remainder of the makeMethodValue implementation is in makefunc.go.
//
// NOTE: This function must be marked as a "wrapper" in the generated code, so that the linker can make it work correctly for panic and recover.
// The gc compilers know to do that for the name "reflect.callMethod".
func callMethod(ctx *methodValue, framePtr unsafe.Pointer) {
	rcvrType, t, fn, _ := ctx.rcvrVal.methodReceiver(ctx.method)
	frameType, argSize, retOffset, _ := funcLayout(t, rcvrType)
	// Make a new frame that is one word bigger so we can store the receiver.
	args := unsafeNew(frameType)

	// Copy in receiver and rest of args.
	// Avoid constructing out-of-bounds pointers if there are no args.
	ctx.rcvrVal.storeRcvr(args)
	if argSize-PtrSize > 0 {
		typedmemmovepartial(frameType, add(args, PtrSize), framePtr, PtrSize, argSize-PtrSize)
	}

	// Call.
	call(frameType, fn, args, uint32(frameType.size), uint32(retOffset))

	// Copy return values. On amd64p32, the beginning of return values  is 64-bit aligned, so the caller's frame layout (which doesn't have a receiver) is different from the layout of the fn call, which has  a receiver.
	// Ignore any changes to args and just copy return values.
	// Avoid constructing out-of-bounds pointers if there are no return values.
	if frameType.size-retOffset > 0 {
		callerRetOffset := retOffset - PtrSize
		if IsAMD64p32 {
			callerRetOffset = align(argSize-PtrSize, 8)
		}
		typedmemmovepartial(frameType,
			add(framePtr, callerRetOffset),
			add(args, retOffset),
			retOffset,
			frameType.size-retOffset)
	}

	// This is untyped because the frame is really a stack, even though it's a heap object.
	memclrNoHeapPointers(args, frameType.size)

	// Without the KeepAlive call, the finalizer could run at the start of syscall.Read, closing the file descriptor before syscall.Read makes the actual system call.
	runtime.KeepAlive(ctx)
}

// callReflect is the call implementation used by a function
// returned by MakeFunc. In many ways it is the opposite of the
// method Value.call above. The method above converts a call using Values
// into a call of a function with a concrete argument frame, while
// callReflect converts a call of a function with a concrete argument
// frame into a call using Values.
// It is in this file so that it can be next to the call method above.
// The remainder of the MakeFunc implementation is in makefunc.go.
//
// NOTE: This function must be marked as a "wrapper" in the generated code,
// so that the linker can make it work correctly for panic and recover.
// The gc compilers know to do that for the name "reflect.callReflect".
func callReflect(ctxt *makeFuncImpl, frame unsafe.Pointer) {
	ftyp := ctxt.typ
	f := ctxt.fn

	// Copy argument frame into Values.
	ptr := frame
	off := uintptr(0)
	in := make([]Value, 0, int(ftyp.InLen))
	for _, typ := range ftyp.inParams() {
		off += -off & uintptr(typ.align-1)
		v := Value{Type: typ, Ptr: nil, Flag: Flag(typ.Kind())}
		if typ.isDirectIface() {
			// value cannot be inlined in interface data.
			// Must make a copy, because f might keep a reference to it,
			// and we cannot let f keep a reference to the stack frame
			// after this function returns, not even a read-only reference.
			v.Ptr = unsafeNew(typ)
			if typ.size > 0 {
				typedmemmove(typ, v.Ptr, add(ptr, off))
			}
			v.Flag |= pointerFlag
		} else {
			v.Ptr = *(*unsafe.Pointer)(add(ptr, off))
		}
		in = append(in, v)
		off += typ.size
	}

	// Call underlying function.
	out := f(in)
	numOut := ftyp.OutLen
	if len(out) != int(numOut) {
		panic("reflect: wrong return count from function created by MakeFunc. Are you using variadic functions?")

	}

	// Copy results back into argument frame.
	if numOut > 0 {
		off += -off & (PtrSize - 1)
		if IsAMD64p32 {
			off = align(off, 8)
		}
		for i, typ := range ftyp.outParams() {
			v := out[i]
			if v.Type != typ {
				// TODO : on MakeFunc it panics here if the signature of the returned function is wrong
				panic("reflect: function created by MakeFunc using " + funcName(f) + " returned wrong type: have " + out[i].Type.String() + " for " + typ.String())
			}
			if v.Flag&exportFlag != 0 {
				panic("reflect: function created by MakeFunc using " + funcName(f) + " returned value obtained from unexported field")
			}
			off += -off & uintptr(typ.align-1)
			if typ.size == 0 {
				continue
			}
			addr := add(ptr, off)
			if v.Flag&pointerFlag != 0 {
				typedmemmove(typ, addr, v.Ptr)
			} else {
				*(*unsafe.Pointer)(addr) = v.Ptr
			}
			off += typ.size
		}
	}

	// runtime.getArgInfo expects to be able to find ctxt on the
	// stack when it finds our caller, makeFuncStub. Make sure it
	// doesn't get garbage collected.
	runtime.KeepAlive(ctxt)
}

// funcName returns the name of f, for use in error messages.
func funcName(f func([]Value) []Value) string {
	pc := *(*uintptr)(unsafe.Pointer(&f))
	rf := runtime.FuncForPC(pc)
	if rf != nil {
		return rf.Name()
	}
	return "closure"
}
