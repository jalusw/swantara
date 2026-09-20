package contacts

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
)

type ContactDAOMock struct {
	dao.CRUDMock[Contact]
	CreateWithDetailsFunc func(ctx context.Context, contact *Contact, addresses []*ContactAddress, banks []*ContactBankAccount, customer *CustomerProfile, supplier *SupplierProfile) (*Contact, error)
}

func (m ContactDAOMock) CreateWithDetails(ctx context.Context, contact *Contact, addresses []*ContactAddress, banks []*ContactBankAccount, customer *CustomerProfile, supplier *SupplierProfile) (*Contact, error) {
	if m.CreateWithDetailsFunc != nil {
		return m.CreateWithDetailsFunc(ctx, contact, addresses, banks, customer, supplier)
	}
	return contact, nil
}

type ContactAddressDAOMock struct {
	dao.CRUDMock[ContactAddress]
	ListByContactFunc func(ctx context.Context, contactID uint64) ([]*ContactAddress, error)
	SetDefaultFunc    func(ctx context.Context, contactID, addressID uint64) error
}

func (m ContactAddressDAOMock) ListByContact(ctx context.Context, contactID uint64) ([]*ContactAddress, error) {
	if m.ListByContactFunc != nil {
		return m.ListByContactFunc(ctx, contactID)
	}
	return []*ContactAddress{}, nil
}

func (m ContactAddressDAOMock) SetDefault(ctx context.Context, contactID, addressID uint64) error {
	if m.SetDefaultFunc != nil {
		return m.SetDefaultFunc(ctx, contactID, addressID)
	}
	return nil
}

type ContactBankAccountDAOMock struct {
	dao.CRUDMock[ContactBankAccount]
	ListByContactFunc func(ctx context.Context, contactID uint64) ([]*ContactBankAccount, error)
}

func (m ContactBankAccountDAOMock) ListByContact(ctx context.Context, contactID uint64) ([]*ContactBankAccount, error) {
	if m.ListByContactFunc != nil {
		return m.ListByContactFunc(ctx, contactID)
	}
	return []*ContactBankAccount{}, nil
}

type CustomerProfileDAOMock struct {
	dao.CRUDMock[CustomerProfile]
	FindByContactFunc func(ctx context.Context, contactID uint64) (*CustomerProfile, error)
}

func (m CustomerProfileDAOMock) FindByContact(ctx context.Context, contactID uint64) (*CustomerProfile, error) {
	if m.FindByContactFunc != nil {
		return m.FindByContactFunc(ctx, contactID)
	}
	return nil, nil
}

type SupplierProfileDAOMock struct {
	dao.CRUDMock[SupplierProfile]
	FindByContactFunc func(ctx context.Context, contactID uint64) (*SupplierProfile, error)
}

func (m SupplierProfileDAOMock) FindByContact(ctx context.Context, contactID uint64) (*SupplierProfile, error) {
	if m.FindByContactFunc != nil {
		return m.FindByContactFunc(ctx, contactID)
	}
	return nil, nil
}
