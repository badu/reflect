package reflector

import (
	"fmt"
	"testing"
	"time"
)

type (
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
func (i Invoice) Print() {

}

// a method for white list lookup
func (i *Invoice) Send() {

}

// Stringer implementation
func (i Invoice) String() string {
	return "Invoice Stringer"
}

// Test scan invoice
func TestInvoice(t *testing.T) {
	r := &Reflector{}
	r.MethodsLookup = []string{"Print", "Send"}
	err := r.ComponentsScan(Invoice{})
	if err != nil {
		t.Fatalf("Error : %v", err)
	}

	t.Logf("%v", r.currentModel)
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
