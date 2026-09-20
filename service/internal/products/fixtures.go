package products

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func ItemFixture(opts ...func(*Item) *Item) *Item {
	now := time.Now()
	pt := &Item{
		Base:            model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID:  nil,
		Name:            gofakeit.AppName(),
		CategoryID:      nil,
		Type:            "item",
		UnitID:          nil,
		PurchaseUnitID:  nil,
		ListPrice:       gofakeit.Float64Range(1, 10000),
		StandardCost:    gofakeit.Float64Range(1, 10000),
		IsPurchasable:   true,
		IsSellable:      true,
		IsManufactured:  false,
		Tracking:        "none",
		Weight:          gofakeit.Float64Range(0, 100),
		Volume:          gofakeit.Float64Range(0, 100),
		HsCode:          nil,
		DescriptionSale: nil,
		DescriptionPur:  nil,
		Active:          true,
	}
	for _, opt := range opts {
		opt(pt)
	}
	return pt
}

func ItemVariantFixture(opts ...func(*ItemVariant) *ItemVariant) *ItemVariant {
	now := time.Now()
	pv := &ItemVariant{
		Base:          model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ItemID:        uint64(gofakeit.Number(1, 10000)),
		Sku:           nil,
		Barcode:       nil,
		ExtraCost:     0,
		Active:        true,
		AttributeJSON: nil,
	}
	for _, opt := range opts {
		opt(pv)
	}
	return pv
}

func PriceBookFixture(opts ...func(*PriceBook) *PriceBook) *PriceBook {
	now := time.Now()
	pl := &PriceBook{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		Name:           gofakeit.AppName(),
		CurrencyCode:   nil,
		OrganizationID: nil,
		Active:         true,
	}
	for _, opt := range opts {
		opt(pl)
	}
	return pl
}

func PriceRuleFixture(opts ...func(*PriceRule) *PriceRule) *PriceRule {
	now := time.Now()
	r := &PriceRule{
		Base:        model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		PriceBookID: uint64(gofakeit.Number(1, 10000)),
		AppliesTo:   "item",
		ItemID:      nil,
		CategoryID:  nil,
		MinQty:      0,
		ComputeType: "fixed",
		FixedPrice:  nil,
		DiscountPct: nil,
		DateStart:   nil,
		DateEnd:     nil,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}
