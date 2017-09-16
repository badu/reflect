package reflector

import (
	"fmt"
	"testing"
	"time"
)

const (
	PackageQuoteItem Uint32 = 1
	ProductQuoteItem Uint32 = 2

	DiscountPercent Uint32 = 1
	DiscountFixed   Uint32 = 2
)

type (

	// ---
	Uint32 uint32

	NullUint64 struct {
		Valid  bool
		Uint64 uint64
	}
	NullTime struct {
		Valid bool
		Time  time.Time
	}
	Timestamp struct {
		time.Time
	}

	QuoteOrderCommon struct {
		Discount           *QuoteItemPrice
		Recurring          *QuoteItemPrice
		Selling            *QuoteItemPrice
		Hardware           *QuoteItemPrice
		Shipping           *QuoteItemPrice
		Total              *QuoteItemPrice
		CalculatedTaxes    *QuoteItemPrice
		CalculatedDiscount *QuoteItemPrice
		CalculatedTotal    *QuoteItemPrice
		RemoveTaxes        bool
		HasShipping        bool
	}

	QuoteItemPrice struct {
		Id                   uint64
		Type                 Uint32
		Value                float64
		Price                float64
		FromUnit             float64
		ToUnit               float64
		InventoryTaxDetailId NullUint64
	}

	QuoteItemPricesCollection []*QuoteItemPrice

	QuoteItem struct {
		Id                  uint64
		Type                Uint32
		ProductId           uint64
		Quantity            float64
		Prices              QuoteItemPricesCollection
		SetupType           Uint32
		OriginalType        Uint32
		OriginalName        string
		OriginalDescription NullString
		OriginalWeight      float64
		OriginalWeightUnit  string
		ShelfId             string
		ParentId            NullUint64
		Children            QuoteItemsCollection
		InStock             bool
		Discount            *QuoteItemPrice
		Selling             *QuoteItemPrice
		Recurring           *QuoteItemPrice
		Taxes               QuoteItemPricesCollection
		CalculatedDiscount  *QuoteItemPrice
		CalculatedTotal     *QuoteItemPrice
		CalculatedTaxes     *QuoteItemPrice
		Photo               string
		Thumb               string
	}

	QuoteItemsCollection []*QuoteItem

	Quote struct {
		QuoteOrderCommon
		Id            uint64
		Status        Uint32
		DisplayName   string
		Optional      NullString
		OneTime       time.Time
		CreatedAt     Timestamp
		UpdatedAt     NullTime
		ExpiresOn     NullTime
		OpenedOn      NullTime
		LastOpen      NullTime
		RemovedProp   NullString
		Version       uint32
		HeadVersionId uint64
		IsCurrent     bool
		Items         QuoteItemsCollection
		Prices        QuoteItemPricesCollection
	}
)

func (i QuoteItem) String() string {
	return fmt.Sprintf("[%d - %q] x %f pieces", i.Id, i.OriginalName, i.Quantity)
}

var (
	props = []string{"CreatedAt", "UpdatedAt", "Optional", "OneTime", "DisplayName", "ExpiresOn", "OpenedOn", "LastOpen", "RemoveTaxes", "TermsOptions", "Status", "ViewToken", "ShippingId", "RemovedProp", "Items"}
)

func Now() Timestamp {
	return Timestamp{time.Now()}
}

func TestComparator(t *testing.T) {
	oldQuote := &Quote{
		Id:          16,
		DisplayName: "Updated Quote #2",
		QuoteOrderCommon: QuoteOrderCommon{
			RemoveTaxes: false,
		},
		RemovedProp: NullString{
			String: "Removed",
			Valid:  true,
		},
		UpdatedAt: NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		Status: 4,
		Items: QuoteItemsCollection{
			{
				Id:           71,
				ProductId:    6,
				Quantity:     1,
				Type:         PackageQuoteItem,
				OriginalName: "Package 6 (Removed)",
			},
			{
				Id:           111,
				ProductId:    21,
				Quantity:     1,
				Type:         ProductQuoteItem,
				OriginalName: "Product 21",
			},
		},
		Prices: QuoteItemPricesCollection{
			{
				Id:    451,
				Type:  DiscountFixed,
				Value: 915,
			},
		},
	}
	newQuote := &Quote{
		Id:          16,
		DisplayName: "Updated Quote #5",
		QuoteOrderCommon: QuoteOrderCommon{
			RemoveTaxes: true,
		},
		Status:    5,
		CreatedAt: Now(),
		OneTime:   time.Now(),
		OpenedOn: NullTime{
			Valid: true,
			Time:  time.Now(),
		},

		Optional: NullString{
			Valid:  true,
			String: "Ceva",
		},
		Items: QuoteItemsCollection{
			{
				Id:           0,
				ProductId:    7,
				Quantity:     1,
				Type:         PackageQuoteItem,
				OriginalName: "Package 7 (Added)",
			},
			{
				Id:           111,
				ProductId:    21,
				Quantity:     1,
				Type:         ProductQuoteItem,
				OriginalName: "Product 21",
			},
		},
	}
	results, err := Compare(oldQuote, newQuote, props)
	if err != nil {
		t.Fatalf("Error : %v", err)
	}

	for _, difference := range results {
		if difference.IsSlice {
			if difference.OldValue != nil {
				item1, ok1 := difference.OldValue.(QuoteItem)
				if !ok1 {
					t.Fatalf("Error : cannot convert 1.")
				}
				fmt.Printf("%q removed %v\n", difference.FieldName, item1)
			}
			if difference.NewValue != nil {
				item2, ok2 := difference.NewValue.(QuoteItem)
				if !ok2 {
					t.Fatalf("Error : cannot convert 2 : %v", difference.NewValue)
				}
				fmt.Printf("%q added %v\n", difference.FieldName, item2)
			}
		}
	}

	var updatesMap map[string]interface{}
	if results != nil && len(results) > 0 {
		updatesMap, _ = MakeMapFromDifferences(results)
		for key, value := range updatesMap {
			t.Logf("Update %q %v", key, value)
		}
	}
}
