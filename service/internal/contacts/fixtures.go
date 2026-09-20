package contacts

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func ContactFixture(opts ...func(*Contact) *Contact) *Contact {
	now := time.Now()
	p := &Contact{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 100))),
		Name:           gofakeit.Company(),
		DisplayName:    helper.Ptr(gofakeit.Company()),
		IsOrganization: gofakeit.Bool(),
		ParentID:       helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Email:          helper.Ptr(gofakeit.Email()),
		Phone:          helper.Ptr(gofakeit.Phone()),
		Mobile:         helper.Ptr(gofakeit.Phone()),
		Website:        helper.Ptr(gofakeit.URL()),
		TaxID:          helper.Ptr(gofakeit.AppName()),
		Industry:       helper.Ptr(gofakeit.AppName()),
		CurrencyCode:   helper.Ptr("USD"),
		Lang:           "en",
		Active:         gofakeit.Bool(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func ContactAddressFixture(opts ...func(*ContactAddress) *ContactAddress) *ContactAddress {
	now := time.Now()
	pa := &ContactAddress{
		Base:        model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ContactID:   uint64(gofakeit.Number(1, 10000)),
		Type:        helper.Ptr(gofakeit.Word()),
		Line1:       helper.Ptr(gofakeit.Address().Address),
		Line2:       helper.Ptr(gofakeit.Address().Address),
		City:        helper.Ptr(gofakeit.Address().City),
		State:       helper.Ptr(gofakeit.Address().State),
		PostalCode:  helper.Ptr(gofakeit.Address().Zip),
		CountryCode: helper.Ptr(gofakeit.Address().Country),
		IsDefault:   gofakeit.Bool(),
	}
	for _, opt := range opts {
		opt(pa)
	}
	return pa
}

func ContactBankAccountFixture(opts ...func(*ContactBankAccount) *ContactBankAccount) *ContactBankAccount {
	now := time.Now()
	ba := &ContactBankAccount{
		Base:          model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ContactID:     uint64(gofakeit.Number(1, 10000)),
		AccountHolder: helper.Ptr(gofakeit.Name()),
		BankName:      helper.Ptr(gofakeit.AppName()),
		IBAN:          helper.Ptr(gofakeit.AppName()),
		SwiftBIC:      helper.Ptr(gofakeit.AppName()),
		AccountNumber: helper.Ptr(gofakeit.AppName()),
		RoutingNumber: helper.Ptr(gofakeit.AppName()),
		CurrencyCode:  helper.Ptr("USD"),
	}
	for _, opt := range opts {
		opt(ba)
	}
	return ba
}

func CustomerProfileFixture(opts ...func(*CustomerProfile) *CustomerProfile) *CustomerProfile {
	now := time.Now()
	pc := &CustomerProfile{
		Base:                  model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ContactID:             uint64(gofakeit.Number(1, 10000)),
		CustomerPaymentTermID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		CreditLimit:           helper.Ptr(gofakeit.Float64Range(1000, 100000)),
		ReceivableAccountID:   helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Active:                gofakeit.Bool(),
	}
	for _, opt := range opts {
		opt(pc)
	}
	return pc
}

func SupplierProfileFixture(opts ...func(*SupplierProfile) *SupplierProfile) *SupplierProfile {
	now := time.Now()
	ps := &SupplierProfile{
		Base:                  model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ContactID:             uint64(gofakeit.Number(1, 10000)),
		SupplierPaymentTermID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		PayableAccountID:      helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Active:                gofakeit.Bool(),
	}
	for _, opt := range opts {
		opt(ps)
	}
	return ps
}
