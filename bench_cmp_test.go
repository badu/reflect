/*
 * Copyright 2009-2018 The Go Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package reflect_test

import (
	"github.com/badu/reflect"
	systemReflect "reflect"
	"testing"
	"time"
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

/**

running:
1. go test -c bench_cmp_test.go
2. reflect.test.exe -test.bench=. -test.benchtime=10ms -test.memprofile=mem.out -test.memprofilerate=1
3. pprof -http=:6061 -sample_index=alloc_objects mem.out

*/
func BenchmarkReflect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := reflect.ReflectOn(Invoice{})
		value.Type.Fields(func(Type *reflect.RType, name []byte, tag []byte, pack []byte, embedded, exported bool, offset uintptr, index int) {

		})
	}
}

func BenchmarkOldReflect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value := systemReflect.ValueOf(Invoice{})
		for j := 0; j < value.Type().NumField(); j++ {
			value.Type().Field(j)
		}
	}
}
