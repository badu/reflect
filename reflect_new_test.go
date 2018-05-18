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
	"time"
	"unsafe"
)

type (
	Function struct {
		Caller caller
	}

	caller interface {
		Call(args ...interface{}) error
	}

	callback func(args ...interface{}) error

	NullString struct {
		String string `json:"openedOn" sql:"type:TEXT NULL"`
		Valid  bool   `json:"-" sql:"-"`
	}

	Entity struct {
		Id        uint64    `sql:"AUTO_INCREMENT" orm:"primary_key" json:"id"`
		CreatedAt time.Time `json:"createdAt" sql:"DEFAULT:current_timestamp;NOT NULL"`
		UpdatedAt time.Time `json:"updatedAt" sql:"type:TIMESTAMP NULL"`
		DeletedAt time.Time `json:"deletedAt" sql:"type:TIMESTAMP NULL"`
		Persist   Function
		Delete    Function
		anonField bool
	}

	Address struct {
		Entity
		Street NullString `json:"street"`
	}

	User struct {
		Entity
		FirstName NullString `json:"firstName"`
		LastName  NullString `json:"lastName"`
		Username  string     `json:"username" sql:"VARCHAR(55);NOT NULL" valid:"required~Name is required,length(3|50)~Username is too short"`
	}

	Customer struct {
		Entity
		Name      string     `json:"name" valid:"required~Name is required,length(3|50)~Name is too short"`
		Addresses []*Address `json:"addresses"`
		Users     []*User    `json:"users"`
	}

	Price struct {
		Entity
		Type  int     `valid:"required~Type is required" json:"type"`
		Value float64 `valid:"required~Type is required" json:"value"`
	}

	PricesCollection []*Price

	Item struct {
		Entity
		Name   string           `json:"name" valid:"required~Name is required,length(3|50)~Name is too short"`
		Prices PricesCollection `json:"prices"`
	}

	Invoice struct {
		Entity
		Customer         *Customer `json:"customer"`
		Items            []*Item   `json:"items"`
		DueDate          time.Time `json:"dueDate" sql:"type:TIMESTAMP NULL"`
		PreviousInvoices []*Invoice
		priv             bool
		Name             string
	}

	InvoiceInterface interface {
		Print()
	}

	Principal struct {
		Entity   //embedded Persist and Delete
		Id       int
		Username string
	}
)

// a method for white list lookup
func (i Invoice) Print() {}
func (e *Entity) Print() {}

func (e Entity) SendMan() {
	println("Different Send...")
}

func (e *Entity) Inherited(newName string) string {
	println("Entity Send called : " + newName)
	return newName
}

func (i *Invoice) SetCustomer(customer *Customer) error {
	i.Customer = customer
	return nil
}

// a method for white list lookup
func (i *Invoice) Send(newName string) string {
	i.Name = newName
	println("Invoice Send called : " + newName)
	return i.Name
}

// Stringer implementation
func (i Invoice) String() string { return "Invoice Stringer" }

func (i Item) Print() {}

func (i Item) String() string { return "Item Stringer" }

func (p Price) Print() {}

func (p Price) String() string { return "Price Stringer" }

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
