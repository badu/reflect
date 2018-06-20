/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect_test

import (
	"fmt"
	. "github.com/badu/reflect"
	"testing"
	"unsafe"
)

// Call the received function.
func (f Function) Call(args ...interface{}) error {
	if !f.IsValid() {
		return fmt.Errorf("invalid function")
	}
	return f.Caller.Call(args...)
}

// IsValid returns true if f represents a Function.
// It returns false if f is the zero Value.
func (f Function) IsValid() bool {
	return f.Caller != nil
}

func (f callback) Call(args ...interface{}) error {
	return f(args...)
}

// Callback is the wrapper for function when sending.
func Callback(f func(args ...interface{}) error) Function {
	return Function{
		Caller: callback(f), //type cast to callback
	}
}

func TestInterface(t *testing.T) {
	iface := Tsmallv(1)
	k := ReflectOn(&iface)
	ToStruct(k).Methods(func(name []byte, index int, flag uintptr, in, out []*RType) {
		t.Logf("Ptr %d.%q %d params in %d params out", index, string(name), len(in), len(out))
	})
}

// Test scan invoice
func TestInvoiceFields(t *testing.T) {
	value := ReflectOn(Invoice{})
	t.Logf("%q", StringKind(value.Kind()))
	if value.Type.Kind() == Struct {
		value.Type.Fields(func(Type *RType, name []byte, tag []byte, pack []byte, embedded, exported bool, offset uintptr, index int) {
			t.Logf("%d. %q : %s ", index, string(name), StringKind(Type.Kind()))
			if len(tag) > 0 {
				t.Logf("Tag : %q", string(tag))
			}
			if len(pack) > 0 {
				t.Logf("Package : %q", string(pack))
			}
			t.Logf("Embedded : %t", embedded)
			t.Logf("Offset : %v", offset)
			t.Logf("Exported : %t", exported)
		})
	}
	t.Log("\n\n")
	ToStruct(value).Fields(func(Type *RType, name []byte, tag []byte, pack []byte, embedded, exported bool, offset uintptr, index int, P unsafe.Pointer) {
		t.Logf("%d. %q : %s @ %p", index, string(name), StringKind(Type.Kind()), P)
		if len(tag) > 0 {
			t.Logf("Tag : %q", string(tag))
		}
		if len(pack) > 0 {
			t.Logf("Package : %q", string(pack))
		}
		t.Logf("Embedded : %t", embedded)
		t.Logf("Offset : %v", offset)
		t.Logf("Exported : %t", exported)
	})
}

func TestInvoice(t *testing.T) {
	invPtr := &Invoice{}
	// method 1 : reflect on a pointer to something (reflect.New does that), deref, addr, struct, get method by name
	v := ReflectOnPtr(invPtr)
	m := ToStruct(v.Addr()).MethodByName("Send")
	rez, ok := m.Call([]Value{ReflectOn("First time")})
	if ok {
		t.Logf("REsULT : %#v", rez[0].String().Get())
	}
	f := ToStruct(v).FieldByName("Name")
	t.Logf("Invoice name : %q", f.String().Get())

	t.Logf("\n\n")

	inv := &Invoice{}
	// method 2 : reflect on a pointer to something (reflect.New does that), then iterate methods
	k := ReflectOn(inv)
	ToStruct(k).Methods(func(name []byte, index int, flag uintptr, in, out []*RType) {
		t.Logf("Ptr %d.%q %d params in %d params out", index, string(name), len(in), len(out))
		if len(in) > 0 {
			t.Logf("In params : %#v", in)
		}
		if len(out) > 0 {
			t.Logf("Out params : %#v", out)
		}
		if string(name) == "Send" {
			t.Log("Placing call.")
			//TODO : you can store Type and Flag in a separate struct, so you don't need to parse methods again. You just need the Ptr value of a Value.
			//Also, in and out params can be stored separately
			val := Value{Type: k.Type, Ptr: k.Ptr, Flag: flag}
			res, ok := val.Call([]Value{ReflectOn("Second timer.")})
			if !ok {
				t.Error("Error calling.")
			}
			t.Logf("Result : %q -> %#v", res[0].String().Get(), res)
		}
		if string(name) == "SetCustomer" {
			c := &Customer{Name: "Somebody SRL"}
			val := Value{Type: k.Type, Ptr: k.Ptr, Flag: flag}
			res, ok := val.Call([]Value{ReflectOn(c)})
			if !ok {
				t.Error("Error calling.")
			}
			t.Logf("Result : %#v", res)
		}
	})
	ToStruct(k.Deref()).Fields(func(Type *RType, name []byte, tag []byte, pack []byte, embedded, exported bool, offset uintptr, index int, P unsafe.Pointer) {
		if string(name) == "Name" {
			t.Logf("Iterator Invoice name : %q", *(*string)(P))
		}
	})

	t.Logf("Finally, %q and %q", inv.Name, invPtr.Name)
	// Trick
	m2 := ToStruct(k.Deref().Addr()).MethodByName("Send")
	rez2, ok := m2.Call([]Value{ReflectOn("Last time")})
	if ok {
		t.Logf("REsULT : %#v", rez2[0].String().Get())
	}
	t.Logf("Finally, %q and %q (%q)", inv.Name, invPtr.Name, inv.Customer.Name)
}

// Test calling ourselves from a struct
func TestCaller(t *testing.T) {

	// let's say principal is built inside another package and returned to where it's needed
	principal := &Principal{
		Id:       1,
		Username: "foo",
	}

	// that package will create callbacks to itself
	principal.Persist = Callback(func(args ...interface{}) error {
		t.Logf("Persist Called Inside Another Package")
		return fmt.Errorf("Test error")
	})
	principal.Delete = Callback(func(args ...interface{}) error {
		t.Logf("Delete Called Inside Another Package with Arguments : %v", args)
		return nil
	})

	//now, the very struct calls methods which otherwise were undefined
	err := principal.Persist.Call()
	if err != nil {
		t.Logf("Error Persisting : %v", err)
	}

	err = principal.Delete.Call("Passed param 1", 1, Invoice{})
	if err != nil {
		t.Errorf("Error Deleting : %v", err)
	}
}

func TestExplainRelation(t *testing.T) {
	type MyByte byte
	type MyRune rune

	var myrs []MyRune
	var mybs []MyByte
	var str = "abc"

	typMyRuneSlice := TypeOf(myrs)
	typMyBytesSlice := TypeOf(mybs)
	typString := TypeOf(str)

	t.Logf("Rune convertible to string : %t", typMyRuneSlice.ConvertibleTo(typString))    // false
	t.Logf("String convertible to rune : %t", typString.ConvertibleTo(typMyRuneSlice))    // false
	t.Logf("[]byte convertible to string : %t", typMyBytesSlice.ConvertibleTo(typString)) // false
	t.Logf("String convertible to []byte : %t", typString.ConvertibleTo(typMyBytesSlice)) // false
	/**
	From https://github.com/go101/go101/wiki/Some-Details-About-Byte-Slices-And-Rune-Slices-In-Go
	Quoted:

	So I looks both the implementations of gc and gccgo violate the restrictions of Go type system, with gccgo violates more than gc.
	Though, personally, I don't think the violations are harmful.
	I hope gc can violate more as gccgo, so that the violations can be viewed as unintended semantics sugars.
	In fact, the reflect package also violates some restrictions of Go type system.
	Go type system forbids us converting a []MyByte value to []byte, but with the help of the method Bytes() of reflect.Value, such conversions are possible.
	*/
	h := []byte("Hello")
	hc := append(h, " C"...)
	_ = append(h, " Go"...)
	t.Logf("Cap = %d, Len = %d, Result = %s\n", cap(h), len(h), hc)
	/**
		What! What happens here? The reason is gccgo and gc use different memory allocation policies when allocating the memory block for hosting the elements of the result slice of []byte("Hello"). Both implementations don't violate Go specification and other official Go documentations. In other words, the capacity of the result slice of []byte("Hello") is compiler dependent.
	Fix below :
	*/

	h2 := []byte("Hello")
	h2 = h2[:len(h2):len(h2)] // need this line
	hc2 := append(h2, " C"...)
	_ = append(h2, " Go"...)
	t.Logf("Cap = %d, Len = %d, Result = %s\n", cap(h2), len(h2), hc2)
}

func TestFastIToA(t *testing.T) {
	t.Logf("%s", I2A(301293, -1))
}
