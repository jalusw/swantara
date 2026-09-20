package contacts

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestContactDAOMockDefaults(t *testing.T) {
	ctx := context.Background()

	contact := &Contact{Name: "Acme"}
	created, err := (ContactDAOMock{}).CreateWithDetails(ctx, contact, nil, nil, nil, nil)
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if created != contact {
		t.Error("CreateWithDetails default should pass through the contact")
	}

	addresses, err := (ContactAddressDAOMock{}).ListByContact(ctx, 1)
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if len(addresses) != 0 {
		t.Errorf("ListByContact default = %+v, want empty", addresses)
	}

	if err := (ContactAddressDAOMock{}).SetDefault(ctx, 1, 1); err != nil {
		t.Errorf("SetDefault default error = %v, want nil", err)
	}

	banks, err := (ContactBankAccountDAOMock{}).ListByContact(ctx, 1)
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if len(banks) != 0 {
		t.Errorf("ListByContact default = %+v, want empty", banks)
	}

	customer, err := (CustomerProfileDAOMock{}).FindByContact(ctx, 1)
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if customer != nil {
		t.Errorf("FindByContact default = %+v, want nil", customer)
	}

	supplier, err := (SupplierProfileDAOMock{}).FindByContact(ctx, 1)
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if supplier != nil {
		t.Errorf("FindByContact default = %+v, want nil", supplier)
	}
}

func TestContactDAOMock_ListByContact_UsesConfiguredFunction(t *testing.T) {
	ctx := context.Background()

	addresses := ContactAddressDAOMock{
		ListByContactFunc: func(_ context.Context, _ uint64) ([]*ContactAddress, error) {
			return []*ContactAddress{{Base: model.Base{ID: 1}}}, nil
		},
	}
	items, err := addresses.ListByContact(ctx, 1)
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Errorf("items = %+v, want one address", items)
	}

	banks := ContactBankAccountDAOMock{
		ListByContactFunc: func(_ context.Context, _ uint64) ([]*ContactBankAccount, error) {
			return []*ContactBankAccount{{Base: model.Base{ID: 2}}}, nil
		},
	}
	accounts, err := banks.ListByContact(ctx, 1)
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if len(accounts) != 1 || accounts[0].ID != 2 {
		t.Errorf("accounts = %+v, want one bank account", accounts)
	}
}
