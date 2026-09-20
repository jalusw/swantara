package reference

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func CurrencyFixture(opts ...func(*Currency) *Currency) *Currency {
	c := &Currency{
		Base:          model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: gofakeit.Date(), UpdatedAt: gofakeit.Date()},
		Code:          "USD",
		Name:          "US Dollar",
		Symbol:        helper.Ptr("$"),
		DecimalPlaces: 2,
		Rounding:      0.01,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func OrganizationFixture(opts ...func(*Organization) *Organization) *Organization {
	hour := gofakeit.Number(0, 23)
	o := &Organization{
		Base:              model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: gofakeit.Date(), UpdatedAt: gofakeit.Date()},
		Name:              gofakeit.Company(),
		LegalName:         helper.Ptr(gofakeit.Company()),
		ParentID:          helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		BaseCurrency:      "USD",
		CountryCode:       helper.Ptr("US"),
		TaxID:             helper.Ptr(gofakeit.AppName()),
		Logo:              helper.Ptr(gofakeit.URL()),
		Timezone:          "UTC",
		TaxYearStartMonth: 1,
		AutoCheckoutHour:  &hour,
		RoundingMinutes:   0,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func AccountFixture(opts ...func(*Account) *Account) *Account {
	a := &Account{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: gofakeit.Date(), UpdatedAt: gofakeit.Date()},
		OrganizationID: uint64(gofakeit.Number(1, 10000)),
		Code:           gofakeit.AppName(),
		Name:           gofakeit.AppName(),
		Type:           "asset",
		Reconcilable:   false,
		CurrencyCode:   helper.Ptr("USD"),
		ParentID:       helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Active:         true,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func JournalFixture(opts ...func(*Journal) *Journal) *Journal {
	j := &Journal{
		Base:             model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: gofakeit.Date(), UpdatedAt: gofakeit.Date()},
		OrganizationID:   uint64(gofakeit.Number(1, 10000)),
		Name:             gofakeit.AppName(),
		Code:             helper.Ptr(gofakeit.AppName()),
		Type:             "general",
		DefaultAccountID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		CurrencyCode:     helper.Ptr("USD"),
		BankAccountID:    helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		SequenceID:       helper.Ptr(uint64(gofakeit.Number(1, 10000))),
	}
	for _, opt := range opts {
		opt(j)
	}
	return j
}

func TaxFixture(opts ...func(*Tax) *Tax) *Tax {
	t := &Tax{
		Base:               model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: gofakeit.Date(), UpdatedAt: gofakeit.Date()},
		OrganizationID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:               gofakeit.AppName(),
		Amount:             helper.Ptr(gofakeit.Float64Range(1, 100)),
		Type:               TaxTypePercent,
		Scope:              TaxScopeSale,
		PriceInclude:       false,
		TaxAccountID:       helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		RefundTaxAccountID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Active:             true,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func WarehouseFixture(opts ...func(*Warehouse) *Warehouse) *Warehouse {
	w := &Warehouse{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: gofakeit.Date(), UpdatedAt: gofakeit.Date()},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:           gofakeit.AppName(),
		Code:           helper.Ptr(gofakeit.AppName()),
		Line1:          helper.Ptr(gofakeit.Address().Street),
		Line2:          helper.Ptr(""),
		City:           helper.Ptr(gofakeit.Address().City),
		State:          helper.Ptr(gofakeit.Address().State),
		PostalCode:     helper.Ptr(gofakeit.Address().Zip),
		CountryCode:    helper.Ptr("US"),
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func StockLocationFixture(opts ...func(*StockLocation) *StockLocation) *StockLocation {
	sl := &StockLocation{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: gofakeit.Date(), UpdatedAt: gofakeit.Date()},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		WarehouseID:    helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:           gofakeit.AppName(),
		Code:           helper.Ptr(gofakeit.AppName()),
		ParentID:       helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Usage:          "internal",
		Barcode:        helper.Ptr(gofakeit.AppName()),
	}
	for _, opt := range opts {
		opt(sl)
	}
	return sl
}

func UnitFixture(opts ...func(*Unit) *Unit) *Unit {
	u := &Unit{
		Base:       model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: gofakeit.Date(), UpdatedAt: gofakeit.Date()},
		CategoryID: uint64(gofakeit.Number(1, 10000)),
		Name:       gofakeit.AppName(),
		Factor:     1.0,
		UnitType:   "reference",
		Rounding:   0.01,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

func PaymentTermFixture(opts ...func(*PaymentTerm) *PaymentTerm) *PaymentTerm {
	pt := &PaymentTerm{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: gofakeit.Date(), UpdatedAt: gofakeit.Date()},
		OrganizationID: uint64(gofakeit.Number(1, 10000)),
		Name:           gofakeit.AppName(),
		Note:           helper.Ptr(gofakeit.Sentence(3)),
		Code:           helper.Ptr(gofakeit.AppName()),
		IsActive:       true,
		TemplateKey:    helper.Ptr(gofakeit.AppName()),
	}
	for _, opt := range opts {
		opt(pt)
	}
	return pt
}
