/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect

import (
	"bytes"
)

func (v StructValue) Methods(inspect MethodInspectFn) {
	// we're sure that it is a struct : check is performed in ToStruct()
	if v.hasMethodFlag() {
		if willPrintDebug {
			panic("reflect.StructValue.Methods: Has methods flag")
		}
	}

	if v.Kind() != Ptr {
		if willPrintDebug {
			panic("Doesn't work like that (we don't want a pointer)")
		}
	}

	methods := exportedMethods(v.Type)

	for i := range methods {
		p := methods[i]
		mType := v.Type.typeOffset(p.typeOffset)
		fnType := mType.convToFn()
		input := fnType.inParams()
		output := fnType.outParams()

		fl := v.Flag & (stickyROFlag | pointerFlag) // Clear embedROFlag
		fl |= Flag(Func)
		fl |= Flag(i)<<methodShiftFlag | methodFlag

		inspect(v.Type.nameOffset(p.nameOffset).name(), i, fl, input, output)
	}
}

// Method returns a function value corresponding to v's i'th method.
// The arguments to a Call on the returned function should not include
// a receiver; the returned function will always use v as the receiver.
// Method panics if i is out of range or if v is a nil interface value.
func (v StructValue) Method(index int) Value {
	// we're sure that it is a struct : check is performed in ToStruct()
	if v.hasMethodFlag() {
		if willPrintDebug {
			panic("reflect.StructValue.Method: Has methods flag")
		}
	}

	if v.Type.Kind() == Interface {
		if v.IsNil() {
			if willPrintDebug {
				panic("reflect.StructValue.Method: Interface Method on nil interface value")
			}
		}
		if uint(index) >= uint(v.Type.NoOfIfaceMethods()) {
			if willPrintDebug {
				panic("reflect.StructValue.Method: Interface Method index out of range")
			}
		}

		it := v.Type.convToIface()
		if it == nil {
			if willPrintDebug {
				println("reflect.StructValue.Method: NIL Interface")
			}
			return Value{}
		}

		var p *ifaceMethod
		for idx := range it.methods {
			p = &it.methods[idx]
			if idx == index {
				methodName := it.nameOffset(p.nameOffset)
				if !methodName.isExported() {
					panic("reflect.StructValue.Method: unexported method")
				}
				if p.typeOffset == 0 {
					panic("reflect.x.error : method type is zero. Apply fix.")
				}
				fl := v.Flag & (stickyROFlag | pointerFlag) // Clear embedROFlag
				fl |= Flag(Func)
				fl |= Flag(index)<<methodShiftFlag | methodFlag
				return Value{Type: v.Type, Ptr: v.Ptr, Flag: fl}
			}
		}
		panic("reflect.StructValue.Method: " + I2A(index, -1) + " method not found")
	} else {
		if uint(index) >= uint(lenExportedMethods(v.Type)) {
			if willPrintDebug {
				panic("reflect.StructValue.Method: Method index out of range")
			}
		}
	}

	fl := v.Flag & (stickyROFlag | pointerFlag) // Clear embedROFlag
	fl |= Flag(Func)
	fl |= Flag(index)<<methodShiftFlag | methodFlag

	return Value{Type: v.Type, Ptr: v.Ptr, Flag: fl}
}

// MethodByName returns a function value corresponding to the method
// of v with the given name.
// The arguments to a Call on the returned function should not include
// a receiver; the returned function will always use v as the receiver.
// It returns the zero Value if no method was found.
func (v StructValue) MethodByName(name string) Value {
	// we're sure that it is a struct : check is performed in ToStruct()
	if v.hasMethodFlag() {
		if willPrintDebug {
			panic("reflect.StructValue.MethodByName: Has methods flag")
		}
	}

	if v.Type.Kind() == Interface {
		if v.IsNil() {
			if willPrintDebug {
				panic("reflect.StructValue.MethodByName: Method on nil interface value")
			}
		}

		it := v.Type.convToIface()
		if it == nil {
			if willPrintDebug {
				println("reflect.StructValue.MethodByName: NIL Interface")
			}
			return Value{}
		}

		var p *ifaceMethod
		for i := range it.methods {
			p = &it.methods[i]
			if string(it.nameOffset(p.nameOffset).name()) == name {
				methodName := it.nameOffset(p.nameOffset)
				if !methodName.isExported() {
					panic("reflect.StructValue.MethodByName: unexported method")
				}
				if p.typeOffset == 0 {
					panic("reflect.x.error : method type is zero. Apply fix.")
				}
				fl := v.Flag & (stickyROFlag | pointerFlag) // Clear embedROFlag
				fl |= Flag(Func)
				fl |= Flag(i)<<methodShiftFlag | methodFlag
				return Value{Type: v.Type, Ptr: v.Ptr, Flag: fl}
			}
		}
		panic("reflect.StructValue.MethodByName: " + name + "method not found")
	}

	methods := exportedMethods(v.Type)
	byteName := []byte(name)
	for i := range methods {
		p := methods[i]
		if p.typeOffset == 0 {
			panic("reflect.x.error : method type is zero. Apply fix.")
		}
		fl := v.Flag & (stickyROFlag | pointerFlag) // Clear embedROFlag
		fl |= Flag(Func)
		fl |= Flag(i)<<methodShiftFlag | methodFlag
		if bytes.Equal(byteName, v.Type.nameOffset(p.nameOffset).name()) {
			return Value{Type: v.Type, Ptr: v.Ptr, Flag: fl}
		}
	}
	return Value{}
}

// NumMethod returns the number of exported methods in the value's method set.
func (v StructValue) NumMethod() int {
	// we're sure that it is a struct : check is performed in ToStruct()
	if v.Type.Kind() == Interface {
		return v.Type.NoOfIfaceMethods()
	}

	return lenExportedMethods(v.Type)
}

func (v StructValue) Fields(inspect InspectValueFn) {
	// we're sure that it is a struct : check is performed in ToStruct()
	structType := v.Type.convToStruct()
	for i := range structType.fields {
		field := &structType.fields[i]
		var PkgPath []byte
		if !field.name.isExported() {
			PkgPath = structType.pkgPath.name()
		}
		var Tag []byte
		if tag := field.name.tag(); len(tag) > 0 {
			Tag = tag
		}
		inspect(field.Type, field.name.name(), Tag, PkgPath, isEmbedded(field), field.name.isExported(), structFieldOffset(field), i, add(v.Ptr, structFieldOffset(field)))
	}
}

// Field returns the i'th field of the struct v.
func (v StructValue) Field(i int) Value {
	// we're sure that it is a struct : check is performed in ToStruct()
	structType := v.Type.convToStruct()
	if uint(i) >= uint(len(structType.fields)) {
		if willPrintDebug {
			panic("reflect.Value.Field: Field index out of range")
		}
	}

	field := &structType.fields[i]
	typ := field.Type

	// Inherit permission bits from v, but clear embedROFlag.
	fl := v.Flag&(stickyROFlag|pointerFlag|addressableFlag) | Flag(typ.Kind())
	// Using an unexported field forces exportFlag.
	if !field.name.isExported() {
		if isEmbedded(field) {
			fl |= embedROFlag
		} else {
			fl |= stickyROFlag
		}
	}
	// Either pointerFlag is set and v.ptr points at struct, or pointerFlag is not set and v.ptr is the actual struct data.
	// In the former case, we want v.ptr + offset.
	// In the latter case, we must have field.offset = 0, so v.ptr + field.offset is still the correct address.
	fieldPtr := add(v.Ptr, structFieldOffset(field))
	return Value{Type: typ, Ptr: fieldPtr, Flag: fl}
}

// FieldByIndex returns the nested field corresponding to index.
func (v StructValue) FieldByIndex(index []int) Value {
	// we're sure that it is a struct : check is performed in ToStruct()
	if len(index) == 1 {
		return v.Field(index[0])
	}
	for i, x := range index {
		if i > 0 {
			if v.Kind() == Ptr {
				if v.IsNil() {
					if willPrintDebug {
						panic("reflect.Value.FieldByIndex: indirection through nil pointer to embedded struct")
					}
				}
				deref := v.Type.Deref()
				if deref.Kind() == Struct {
					v.Value = v.Deref()
				}
			}
		}
		v.Value = v.Field(x)
	}
	return v.Value
}

// FieldByName returns the struct field with the given name.
// It returns the zero Value if no field was found.
func (v StructValue) FieldByName(name string) Value {
	// we're sure that it is a struct : check is performed in ToStruct()
	byteName := []byte(name)
	structType := v.Type.convToStruct()
	for i := range structType.fields {
		field := &structType.fields[i]
		if bytes.Equal(field.name.name(), byteName) {
			return v.Field(i)
		}
	}
	return Value{}
}

// NumField returns the number of fields in the struct v.
func (v StructValue) NumField() int {
	// TODO : disable / deprecate, returns all fields (exported and unexported)
	return len(v.Type.convToStruct().fields)
}
