package reflector

import (
	"testing"
	"time"
)

type (
	Entity struct {
		Id        uint64
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt time.Time
	}

	Address struct {
		Entity
		Street string
	}

	Customer struct {
		Entity
		Name      string
		Addresses []*Address
	}

	Price struct {
		Entity
		Type  int
		Value float64
	}

	Item struct {
		Entity
		Name   string
		Prices []*Price
	}

	Invoice struct {
		Entity
		Customer *Customer
		Items    []*Item
	}
)

func TestOne(t *testing.T) {
	r := &Reflector{}
	r.ComponentsScan(Invoice{})
}
