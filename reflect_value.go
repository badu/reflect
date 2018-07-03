/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import "unsafe"

// Kind returns v's Kind.
// If v is the zero Value (IsValid returns false), Kind returns Invalid.
func (v Value) Kind() Kind           { return Kind(v.Flag & kindMaskFlag) }
func (v Value) shiftMethodFlag() int { return int(v.Flag) >> methodShiftFlag }
func (v Value) IsValid() bool        { return v.Flag != 0 }
func (v Value) isPointer() bool      { return v.Flag&pointerFlag != 0 }
func (v Value) isExported() bool     { return v.Flag&exportFlag == 0 }
func (v Value) IsRO() bool           { return v.Flag&stickyROFlag != 0 }
func (v Value) hasMethodFlag() bool  { return v.Flag&methodFlag != 0 }
func (v Value) CanAddr() bool        { return v.Flag&addressableFlag != 0 }
func (v Value) CanSet() bool         { return v.Flag&(addressableFlag|exportFlag) == addressableFlag }

func (v Value) ro() Flag {
	if !v.isExported() {
		return stickyROFlag
	}
	return 0
}

// assignTo returns a value v that can be assigned directly to Type.
// It panics if v is not assignable to Type.
// For a conversion to an interface type, target is a suggested scratch space to use.
func (v Value) assignTo(dst *RType, target unsafe.Pointer) Value {
	if v.hasMethodFlag() {
		//v = v.makeMethodValue()
		panic("Value.assignTo : This is a method.")
	}

	switch {
	default:
		// Failed.
		panic("reflect.Value.assignTo: value of type " + TypeToString(v.Type) + " is not assignable to type " + TypeToString(dst))
	case v.Type.directlyAssignable(dst):
		// Overwrite type so that they match. Same memory layout, so no harm done.
		fl := v.Flag&(addressableFlag|pointerFlag) | v.ro()
		fl |= Flag(dst.Kind())
		return Value{Type: dst, Ptr: v.Ptr, Flag: fl}

	case v.Type.implements(dst):
		if target == nil {
			target = unsafeNew(dst)
		}
		if v.Kind() == Interface && v.IsNil() {
			// A nil ReadWriter passed to nil Reader is OK, but using ifaceE2I below will panic.
			// Avoid the panic by returning a nil dst (e.g., Reader) explicitly.
			return Value{Type: dst, Ptr: nil, Flag: Flag(Interface)}
		}
		assertE2I(v, dst, target)
		return Value{Type: dst, Ptr: target, Flag: pointerFlag | Flag(Interface)}
	}
}

func (v Value) valueInterface() interface{} {
	if v.hasMethodFlag() {
		// Value must be func kind ... not very usefull, since I've removed support (it has no method flag)
		return v.makeMethodValue().packEface()
	}

	if v.Kind() == Interface {
		// Special case: return the element inside the interface.
		// Empty interface has one layout, all interfaces with methods have a second layout.
		if v.Type.NoOfIfaceMethods() == 0 {
			// the case of "interface{}"
			return convIface(v.Ptr)
		}
		return convIfaceMeth(v.Ptr)
	}

	return v.packEface()
}

// packEface converts v to the empty interface.
func (v Value) packEface() interface{} {
	var i interface{}
	e := toIface(unsafe.Pointer(&i))
	// First, fill in the data portion of the interface.
	switch {
	case v.Type.isDirectIface():
		if !v.isPointer() {
			panic("reflect.x.error : packEface this is not a pointer")
		}
		// Value is indirect, and so is the interface we're making.
		valPtr := v.Ptr
		if v.CanAddr() {
			c := unsafeNew(v.Type)
			typedmemmove(v.Type, c, valPtr)
			valPtr = c
		}
		e.word = valPtr
	case v.isPointer():
		// Value is indirect, but interface is direct. We need to load the data at v.ptr into the interface data word.
		e.word = convPtr(v.Ptr)
	default:
		// Value is direct, and so is the interface.
		e.word = v.Ptr
	}
	// Now, fill in the type portion. We're very careful here not to have any operation between the e.word and e.Type assignments that would let the garbage collector observe the partially-built interface value.
	e.Type = v.Type
	return i
}

// Addr returns a pointer value representing the address of v.
// Addr is typically used to obtain a pointer to a struct field or slice element in order to call a method that requires a pointer receiver.
// If value is a named type and is addressable, start with its address, so that if the type has pointer methods, we find them.
func (v Value) Addr() Value {
	if !v.CanAddr() {
		if willPrintDebug {
			panic("reflect.Value.Addr: called on a NON addressable value")
		}
	}
	return Value{Type: v.Type.PtrTo(), Ptr: v.Ptr, Flag: v.ro() | Flag(Ptr)}
}

// Iface returns the value that the interface v contains.
func (v Value) Iface() Value {
	switch v.Kind() {
	case Interface:
		var eface interface{}
		if v.Type.NoOfIfaceMethods() == 0 {
			// the case of "interface{}"
			eface = convIface(v.Ptr)
		} else {
			eface = convIfaceMeth(v.Ptr)
		}
		// unpackEface converts the empty interface 'eface' to a Value.
		e := toIface(unsafe.Pointer(&eface))
		// NOTE: don't read e.word until we know whether it is really a pointer or not.
		t := e.Type
		if t == nil {
			if willPrintDebug {
				println("Failed to unpack interface.")
			}
			// it's invalid
			return Value{}
		}
		f := Flag(t.Kind())
		if t.isDirectIface() {
			f |= pointerFlag
		}
		x := Value{Type: t, Ptr: e.word, Flag: f}
		if x.IsValid() {
			x.Flag |= v.ro()
		}
		return x
	default:
		if willPrintDebug {
			panic("reflect.Value.Iface: NOT an interface (kind:`" + StringKind(v.Kind()) + "` flag:`" + v.flagsToString() + "`)")
		}
		return v
	}
}

// Deref returns the value that the pointer v points to.
// If v is a nil pointer, Deref returns a zero Value.
// If v is not a pointer, Deref returns v.

// PtrTo returns the pointer type with element t.
// For example, if t represents type Foo, PtrTo(t) represents *Foo.
func (v Value) Deref() Value {
	switch v.Kind() {
	case Ptr:
		ptrToV := v.Ptr
		if v.isPointer() {
			ptrToV = convPtr(ptrToV)
		}
		// The returned value's address is v's value.
		if ptrToV == nil {
			if willPrintDebug {
				panic("reflect.Value.Deref: RETURNING EMPTY VALUE")
			}
			return Value{}
		}
		// if we got here, there is not a dereference, nor the pointer is nil - studying the type's pointer
		typ := v.Type.Deref()
		fl := v.Flag&exportFlag | pointerFlag | addressableFlag | Flag(typ.Kind())
		return Value{Type: typ, Ptr: ptrToV, Flag: fl}
	default:
		if willPrintDebug {
			panic("reflect.Value.Deref: NOT a pointer (kind:`" + StringKind(v.Kind()) + "` flag:`" + v.flagsToString() + "`)")
		}
		return v
	}
}

// IsNil reports whether its argument v is nil. The argument must be
// a chan, func, interface, map, pointer, or slice value; if it is
// not, IsNil panics. Note that IsNil is not always equivalent to a
// regular comparison with nil in Go. For example, if v was created
// by calling ReflectOn with an uninitialized interface variable i,
// i==nil will be true but v.IsNil will panic as v will be the zero
// Value.
func (v Value) IsNil() bool {
	switch v.Kind() {
	case Chan, Func, Map, Ptr:
		if v.hasMethodFlag() {
			return false
		}
		ptrToV := v.Ptr
		if v.isPointer() {
			ptrToV = convPtr(ptrToV)
		}
		return ptrToV == nil
	case Interface, Slice:
		// Both interface and slice are nil if first word is 0.
		// Both are always bigger than a word; assume pointerFlag.
		return convPtr(v.Ptr) == nil
	default:
		if willPrintDebug {
			panic("reflect.Value.IsNil: unknown type")
		}
		return true
	}
}

// Set assigns x to the value v.
// As in Go, x's value must be assignable to v's type.
func (v Value) Set(toX Value) bool {
	x := toX

	if !v.IsValid() || !v.CanSet() {
		if willPrintDebug {
			panic("reflect.Value.Set : value is not settable.")
		}
		return false
	}
	// do not let unexported x leak
	if !x.IsValid() || !x.isExported() {
		if willPrintDebug {
			panic("reflect.Value.Set : parameter is not exported.")
		}
		return false
	}
	var target unsafe.Pointer
	if v.Kind() == Interface {
		target = v.Ptr
	}
	x = x.assignTo(v.Type, target)
	if x.isPointer() {
		typedmemmove(v.Type, v.Ptr, x.Ptr)
	} else {
		loadConvPtr(v.Ptr, x.Ptr)
	}
	return true
}

// pointer returns the underlying pointer represented by v.
// v.Kind() must be ptr, Map, Chan, Func, or UnsafePointer
func (v Value) pointer() unsafe.Pointer {
	if v.Type.size != PtrSize || !v.Type.hasPointers() {
		if willPrintDebug {
			panic("reflect.Value.pointer: called pointer on a NON pointer Value")
		}
		return nil
	}
	if v.isPointer() {
		return convPtr(v.Ptr)
	}
	return v.Ptr
}

// Pointer returns v's value as a uintptr.
// It returns uintptr instead of ptr so that
// code using reflect cannot obtain ptrs
// without importing the unsafe package explicitly.
//
// If v's Kind is Func, the returned pointer is an underlying
// code pointer, but not necessarily enough to identify a
// single function uniquely. The only guarantee is that the
// result is zero if and only if v is a nil func Value.
//
// If v's Kind is Slice, the returned pointer is to the first
// element of the slice. If the slice is nil the returned value
// is 0.  If the slice is empty but non-nil the return value is non-zero.
func (v Value) Pointer() uintptr {
	switch v.Kind() {
	case Map, Ptr, UnsafePointer:
		return uintptr(v.pointer())
	case Func:
		if v.hasMethodFlag() {
			// As the doc comment says, the returned pointer is an
			// underlying code pointer but not necessarily enough to
			// identify a single function uniquely. All method expressions
			// created via reflect have the same underlying code pointer,
			// so their Pointers are equal. The function used here must
			// match the one used in makeMethodValue.
			f := methodValueCall
			return **(**uintptr)(unsafe.Pointer(&f))
		}
		p := v.pointer()
		// Non-nil func value points at data block.
		// First word of data block is actual code.
		if p != nil {
			p = convPtr(p)
		}
		return uintptr(p)
	case Slice:
		return convToSliceHeader(v.Ptr).Data
	default:
		panic("reflect.PointerValue: unknown kind")
	}
}

// CanInterface reports whether Interface can be used without panicking.
// Of course, if Value is not valid, it cannot interface.
func (v Value) CanInterface() bool {
	if !v.IsValid() {
		if willPrintDebug {
			panic("reflect.Value.CanInterface: called on a value without a flag")
		}
	}
	return v.isExported()
}

// Interface returns v's current value as an interface{}.
// It is equivalent to:
// 	var i interface{} = (v's underlying value)
// Given a reflect.Value we can recover an interface value using the Interface method; in effect the method packs the type and value information back into an interface representation and returns the result
// y := v.Interface().(float64) // y will have type float64.
// fmt.Println(y)
// TODO : shouldn't this be called "Concrete" ?
func (v Value) Interface() interface{} {
	if !v.IsValid() {
		if willPrintDebug {
			panic("reflect.Value.Interface: called on a value without a flag")
		}
		return nil
	}
	if !v.isExported() {
		if willPrintDebug {
			panic("Value is not exported. How do you interface?")
		}
	}
	return v.valueInterface()
}

func (v Value) flagsToString() string {
	result := ""
	if v.isPointer() {
		result += "[pointer]"
	}
	if v.isExported() {
		result += "[exported]"
	}
	if v.IsRO() {
		result += "[isRo]"
	}
	if v.CanAddr() {
		result += "[addressable]"
	}
	if v.hasMethodFlag() {
		result += "[method]"
	}
	return result
}

/**
func (v Value) GoString() string {
	k := v.Kind()
	result := "Value Kind : `" + StringKind(k) + "`" + v.Type.GoString()
	switch k {
	case Struct:
		result += "\nValueFlags : " + v.flagsToString()

	case Ptr:
		result += "\nValueFlags : " + v.flagsToString()
		deref := v.Type.Deref()
		result += "\nDerefTypeName : `" + deref.Name()
		result += "\nDerefTypeKind : `" + StringKind(deref.Kind()) + "`"
	case Func:
		result += "\nValueFlags : " + v.flagsToString()
		result += "\nTypeKind : " + StringKind(v.Type.Kind())
		//result += "\nIn :" + strconv.Itoa(v.Type.numIn()) + " Out : " + strconv.Itoa(v.Type.numOut())
	}
	return result
}
**/
// makeMethodValue converts v from the rcvr+method index representation
// of a method value to an actual method func value, which is
// basically the receiver value with a special bit set, into a true
// func value - a value holding an actual func. The output is
// semantically equivalent to the input as far as the user of package
// reflect can tell, but the true func representation can be handled
// by code like Convert and Interface and Assign.

// makeMethodValue converts v from the rcvr+method index representation
// of a method value to an actual method func value, which is
// basically the receiver value with a special bit set, into a true
// func value - a value holding an actual func. The output is
// semantically equivalent to the input as far as the user of package
// reflect can tell, but the true func representation can be handled
// by code like Convert and Interface and Assign.
func (v Value) makeMethodValue() Value {
	if !v.hasMethodFlag() {
		panic("reflect.x.error :invalid use of makeMethodValue")
	}

	// Ignoring the methodFlag bit, v describes the receiver, not the method type.
	fl := v.Flag & (exportFlag | addressableFlag | pointerFlag)
	fl |= Flag(v.Type.Kind())
	rcvr := Value{Type: v.Type, Ptr: v.Ptr, Flag: fl}

	// v.Type returns the actual type of the method value.
	funcType := v.MethodType()

	// Indirect Go func value (dummy) to obtain
	// actual code address. (A Go func value is a pointer
	// to a C function pointer. https://golang.org/s/go11func.)
	dummy := methodValueCall
	code := **(**uintptr)(unsafe.Pointer(&dummy))

	// methodValue contains a stack map for use by the runtime
	_, _, _, stack := funcLayout(funcType, nil)

	fv := &methodValue{
		fnUintPtr: code,
		stack:     stack,
		method:    v.shiftMethodFlag(),
		rcvrVal:   rcvr,
	}

	// Cause panic if method is not appropriate.
	// The panic would still happen during the call if we omit this,
	// but we want Interface() and other operations to fail early.
	_, _, _, ok := fv.rcvrVal.methodReceiver(fv.method)
	if !ok && willPrintDebug {
		panic("INTERNAL ERROR ON makeMethodValue : methodReceiver call failed.")
	}
	return Value{Type: funcType, Ptr: unsafe.Pointer(fv), Flag: v.Flag&exportFlag | Flag(Func)}
}

// methodReceiver returns information about the receiver
// described by v. The Value v may or may not have the
// methodFlag bit set, so the kind cached in v.flag should
// not be used.
// The return value rcvrtype gives the method's actual receiver type.
// The return value t gives the method type signature (without the receiver).
// The return value fn is a pointer to the method code.
func (v Value) methodReceiver(i int) (*RType, *RType, unsafe.Pointer, bool) {
	if v.Type.Kind() == Interface {
		iface := v.Type.convToIface()
		if uint(i) >= uint(len(iface.methods)) {
			if willPrintDebug {
				panic("reflect.Value,methodReceiver: x error - invalid method index")
			}
			return nil, nil, nil, false
		}
		method := &iface.methods[i]
		if !iface.nameOffset(method.nameOffset).isExported() {
			if willPrintDebug {
				panic("reflect.Value,methodReceiver : x error -unexported method")
			}
			return nil, nil, nil, false
		}

		concrete := (*concreteRtype)(v.Ptr)
		if concrete.iTab == nil {
			if willPrintDebug {
				panic("reflect.Value,methodReceiver: x error - method on nil interface value")
			}
			return nil, nil, nil, false
		}
		return concrete.iTab.Type, iface.typeOffset(method.typeOffset), unsafe.Pointer(&concrete.iTab.fun[i]), true
	}

	methods := exportedMethods(v.Type)
	if uint(i) >= uint(len(methods)) {
		if willPrintDebug {
			panic("reflect.Value,methodReceiver: x error - invalid method index")
		}
		return nil, nil, nil, false
	}
	method := methods[i]
	if !v.Type.nameOffset(method.nameOffset).isExported() {
		if willPrintDebug {
			panic("reflect.Value,methodReceiver: x error - unexported method")
		}
		return nil, nil, nil, false
	}
	ifaceFn := v.Type.textOffset(method.ifaceCall)
	return v.Type, v.Type.typeOffset(method.typeOffset), unsafe.Pointer(&ifaceFn), true
}

// Call calls the function v with the input arguments in.
// For example, if len(in) == 3, v.Call(in) represents the Go call v(in[0], in[1], in[2]).
// It returns the output results as Values.
// As in Go, each input argument must be assignable to the type of the function's corresponding input parameter.
// If v is a variadic function, Call creates the variadic slice parameter itself, copying in the corresponding values.
func (v Value) Call(valArgs []Value) ([]Value, bool) {
	// Get function pointer, type.
	t := v.Type
	var (
		fnPtr    unsafe.Pointer
		rcvr     Value
		rcvrType *RType
		ok       bool
	)
	if v.hasMethodFlag() {
		rcvr = v
		rcvrType, t, fnPtr, ok = v.methodReceiver(v.shiftMethodFlag())
		if !ok {
			return nil, false
		}
	} else if v.isPointer() {
		fnPtr = convPtr(v.Ptr)
	} else {
		fnPtr = v.Ptr
	}

	if fnPtr == nil {
		if willPrintDebug {
			panic("reflect.StructValue.Call: nil function")
		}
		return nil, false
	}

	n := t.numIn()
	if len(valArgs) < n {
		if willPrintDebug {
			panic("reflect.StructValue.Call: too few input arguments")
		}
		return nil, false
	}
	if len(valArgs) > n {
		if willPrintDebug {
			panic("reflect.StructValue.Call: too many input arguments (is variadic? will not work)")
		}
		return nil, false
	}

	for _, x := range valArgs {
		if x.Kind() == Invalid {
			if willPrintDebug {
				panic("reflect: using zero Value argument")
			}
			return nil, false
		}
	}
	srcFn := t.convToFn()
	for i := 0; i < n; i++ {
		if xt, targ := valArgs[i].Type, srcFn.inParam(i); !xt.AssignableTo(targ) {
			if willPrintDebug {
				panic("reflect.StructValue.Call: using " + TypeToString(xt) + " as type " + TypeToString(targ))
			}
			return nil, false
		}
	}

	nin := len(valArgs)
	if nin != t.numIn() {
		if willPrintDebug {
			panic("reflect.StructValue.Call: wrong argument count")
		}
		return nil, false
	}
	numResults := t.numOut()

	// Compute frame type.
	frameType, _, returnOffset, _ := funcLayout(t, rcvrType)

	// Allocate a chunk of memory for frame.
	args := unsafeNew(frameType)

	offset := uintptr(0)

	// Copy inputs into args.
	if rcvrType != nil {
		rcvr.storeRcvr(args)
		offset = PtrSize
	}

	for i, pin := range valArgs {
		v := pin
		if !v.IsValid() || !v.isExported() {
			if willPrintDebug {
				panic("reflect.Value.Call: parameter must be exported")
			}
			return nil, false
		}

		targ := srcFn.inParam(i)
		alignUPtr := uintptr(targ.align)
		offset = (offset + alignUPtr - 1) &^ (alignUPtr - 1)
		n := targ.size
		if n == 0 {
			// Not safe to compute args+offset pointing at 0 bytes,
			// because that might point beyond the end of the frame,
			// but we still need to call assignTo to check if it's assignable.
			v.assignTo(targ, nil)
			continue
		}
		addr := add(args, offset)
		v = v.assignTo(targ, addr)
		if v.isPointer() {
			typedmemmove(targ, addr, v.Ptr)
		} else {
			loadConvPtr(addr, v.Ptr)
		}
		offset += n
	}

	// Call.
	call(frameType, fnPtr, args, uint32(frameType.size), uint32(returnOffset))

	if numResults == 0 {
		// This is untyped because the frame is really a  stack, even though it's a heap object.
		memclrNoHeapPointers(args, frameType.size)
		return nil, true
	}
	// Zero the now unused input area of args,
	// because the Values returned by this function contain pointers to the args object,
	// and will thus keep the args object alive indefinitely.
	memclrNoHeapPointers(args, returnOffset)
	// Wrap Values around return values in args.
	results := make([]Value, numResults)
	offset = returnOffset
	for i := 0; i < numResults; i++ {
		tv := srcFn.outParam(i)
		a := uintptr(tv.align)
		offset = (offset + a - 1) &^ (a - 1)
		if tv.size != 0 {
			results[i] = Value{Type: tv, Ptr: add(args, offset), Flag: pointerFlag | Flag(tv.Kind())}
		} else {
			// For zero-sized return value, args+offset may point to the next object.
			// In this case, return the zero value instead.
			results[i] = Zero(tv)
		}
		offset += tv.size
	}

	return results, true
}

// v is a method receiver. Store at p the word which is used to
// encode that receiver at the start of the argument list.
// Reflect uses the "interface" calling convention for
// methods, which always uses one word to record the receiver.
func (v Value) storeRcvr(p unsafe.Pointer) {
	if v.Type.Kind() == Interface {
		// the interface data word becomes the receiver word
		loadConvPtr(p, (*concreteRtype)(v.Ptr).word)
	} else if v.isPointer() && !v.Type.isDirectIface() {
		loadConvPtr(p, convPtr(v.Ptr))
	} else {
		loadConvPtr(p, v.Ptr)
	}
}

func (v Value) MethodType() *RType {
	// Method value.
	// v.Type describes the receiver, not the method type.
	shift := v.shiftMethodFlag()
	if v.Type.Kind() == Interface {
		// Method on interface.
		intf := v.Type.convToIface()
		if uint(shift) >= uint(len(intf.methods)) {
			if willPrintDebug {
				panic("reflect.Value.MethodType: x error: invalid interface method index (interface)")
			}
			return nil
		}
		method := &intf.methods[shift]
		return v.Type.typeOffset(method.typeOffset)
	}

	// Method on concrete type.
	methods := exportedMethods(v.Type)
	if uint(shift) >= uint(len(methods)) {
		if willPrintDebug {
			panic("reflect.Value.MethodType: x error: invalid concrete method index")
		}
		return nil
	}
	method := methods[shift]
	return v.Type.typeOffset(method.typeOffset)
}

func (v Value) In(i int) *RType {
	if v.hasMethodFlag() {
		return v.MethodType().convToFn().inParam(i)
	}
	return v.Type.convToFn().inParam(i)
}

func (v Value) NumIn() int {
	if v.hasMethodFlag() {
		return v.MethodType().numIn()
	}
	return v.Type.numIn()
}

func (v Value) Out(i int) *RType {
	if v.hasMethodFlag() {
		return v.MethodType().convToFn().outParam(i)
	}
	return v.Type.convToFn().outParam(i)
}

func (v Value) NumOut() int {
	if v.hasMethodFlag() {
		return v.MethodType().numOut()
	}
	return v.Type.numOut()
}
