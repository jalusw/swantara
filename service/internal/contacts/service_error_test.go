package contacts

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestContactService_Update_PropagatesLookupError(t *testing.T) {
	ctx := context.Background()
	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	_, err := svc.Update(ctx, 1, &Contact{Name: "Acme"})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestContactService_EnableCustomer_PropagatesContactLookupError(t *testing.T) {
	ctx := context.Background()
	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	_, err := svc.EnableCustomer(ctx, 1, &CustomerProfile{})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestContactService_EnableCustomer_PropagatesRoleLookupError(t *testing.T) {
	ctx := context.Background()
	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return &Contact{Base: model.Base{ID: 1}}, nil
			},
		},
	}
	customers := CustomerProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*CustomerProfile, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, customers, SupplierProfileDAOMock{})

	_, err := svc.EnableCustomer(ctx, 1, &CustomerProfile{})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestContactService_DisableCustomer_PropagatesRoleLookupError(t *testing.T) {
	ctx := context.Background()
	customers := CustomerProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*CustomerProfile, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewContactService(ContactDAOMock{}, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, customers, SupplierProfileDAOMock{})

	if err := svc.DisableCustomer(ctx, 1); helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestContactService_DisableCustomer_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()
	customers := CustomerProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*CustomerProfile, error) {
			return &CustomerProfile{ContactID: 1, Active: true}, nil
		},
		CRUDMock: dao.CRUDMock[CustomerProfile]{
			UpdateFunc: func(_ context.Context, _ *CustomerProfile) (*CustomerProfile, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewContactService(ContactDAOMock{}, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, customers, SupplierProfileDAOMock{})

	if err := svc.DisableCustomer(ctx, 1); helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestContactService_EnableSupplier_PropagatesContactLookupError(t *testing.T) {
	ctx := context.Background()
	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	_, err := svc.EnableSupplier(ctx, 1, &SupplierProfile{})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestContactService_EnableSupplier_PropagatesRoleLookupError(t *testing.T) {
	ctx := context.Background()
	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return &Contact{Base: model.Base{ID: 1}}, nil
			},
		},
	}
	suppliers := SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*SupplierProfile, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, suppliers)

	_, err := svc.EnableSupplier(ctx, 1, &SupplierProfile{})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestContactService_DisableSupplier_PropagatesRoleLookupError(t *testing.T) {
	ctx := context.Background()
	suppliers := SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*SupplierProfile, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewContactService(ContactDAOMock{}, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, suppliers)

	if err := svc.DisableSupplier(ctx, 1); helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestContactService_DisableSupplier_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()
	suppliers := SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*SupplierProfile, error) {
			return &SupplierProfile{ContactID: 1, Active: true}, nil
		},
		CRUDMock: dao.CRUDMock[SupplierProfile]{
			UpdateFunc: func(_ context.Context, _ *SupplierProfile) (*SupplierProfile, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewContactService(ContactDAOMock{}, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, suppliers)

	if err := svc.DisableSupplier(ctx, 1); helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestContactService_SetDefaultAddress_PropagatesLookupError(t *testing.T) {
	ctx := context.Background()
	addresses := ContactAddressDAOMock{
		CRUDMock: dao.CRUDMock[ContactAddress]{
			FindFunc: func(_ context.Context, _ uint64) (*ContactAddress, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewContactService(ContactDAOMock{}, addresses, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	if err := svc.SetDefaultAddress(ctx, 1, 1); helper.AssertError(t, err, true, nil) {
		return
	}
}
