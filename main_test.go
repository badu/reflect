package reflector

import (
	"testing"
	"time"
)

type (

	NullString struct {
		String string
		Valid bool
	}

	Entity struct {
		Id        uint64
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt time.Time
	}

	Address struct {
		Entity
		Street NullString
	}

	User struct {
		Entity
		Username string
	}

	Customer struct {
		Entity
		Name      string
		Addresses []*Address
		Users     []*User
	}

	Price struct {
		Entity
		Type  int
		Value float64
	}

	PricesCollection []*Price

	Item struct {
		Entity
		Name   string
		Prices PricesCollection
	}

	Invoice struct {
		Entity
		Customer  *Customer
		Items     []*Item
		ExpiresAt time.Time
	}

	InvoiceInterface interface {
		Print()
	}
)

func (i Invoice) Print() {

}

func (i *Invoice) Send() {

}

func TestOne(t *testing.T) {
	r := &Reflector{}
	err := r.ComponentsScan(Invoice{})
	if err != nil {
		t.Fatalf("Error : %v", err)
	}

	t.Logf("%v", r.currentModel)
}
