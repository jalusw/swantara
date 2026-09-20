package contacts

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestContactService_Update_Success(t *testing.T) {
	ctx := context.Background()

	var updated *Contact
	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return &Contact{Base: model.Base{ID: 1}, Name: "Old"}, nil
			},
			UpdateFunc: func(_ context.Context, record *Contact) (*Contact, error) {
				updated = record
				return record, nil
			},
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	incoming := &Contact{Name: "New Name", DisplayName: helper.Ptr("New Display"), Email: helper.Ptr("new@example.com"), Active: false}
	_, err := svc.Update(ctx, 1, incoming)
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if updated == nil || updated.Name != "New Name" || updated.Email == nil || *updated.Email != "new@example.com" || updated.Active {
		t.Errorf("updated = %+v, want copied fields", updated)
	}
}

func TestContactService_Update_InvalidName(t *testing.T) {
	ctx := context.Background()
	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return &Contact{Base: model.Base{ID: 1}}, nil
			},
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})
	_, err := svc.Update(ctx, 1, &Contact{Name: "  "})
	if !errors.Is(err, ErrNameRequired) {
		t.Fatalf("err = %v, want ErrNameRequired", err)
	}
}

func TestContactService_EnableSupplier_GetOrCreate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		existing   *SupplierProfile
		wantCreate bool
	}{
		{name: "creates supplier role for new contact role", existing: nil, wantCreate: true},
		{name: "updates existing supplier role", existing: &SupplierProfile{ContactID: 1, Active: false}, wantCreate: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contacts := ContactDAOMock{
				CRUDMock: dao.CRUDMock[Contact]{
					FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
						return &Contact{Base: model.Base{ID: 1}}, nil
					},
				},
			}
			suppliers := SupplierProfileDAOMock{
				FindByContactFunc: func(_ context.Context, _ uint64) (*SupplierProfile, error) {
					return tt.existing, nil
				},
				CRUDMock: dao.CRUDMock[SupplierProfile]{
					CreateFunc: func(_ context.Context, record *SupplierProfile) (*SupplierProfile, error) {
						if !tt.wantCreate {
							t.Error("expected update path, but Create was called")
						}
						if !record.Active {
							t.Error("expected supplier role to be activated")
						}
						return record, nil
					},
					UpdateFunc: func(_ context.Context, record *SupplierProfile) (*SupplierProfile, error) {
						if tt.wantCreate {
							t.Error("expected create path, but Update was called")
						}
						if !record.Active {
							t.Error("expected supplier role to be activated")
						}
						return record, nil
					},
				},
			}
			svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, suppliers)

			_, err := svc.EnableSupplier(ctx, 1, &SupplierProfile{})
			if helper.AssertError(t, err, false, nil) {
				return
			}
		})
	}
}

func TestContactService_EnableSupplier_ContactNotFound(t *testing.T) {
	ctx := context.Background()
	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return nil, nil
			},
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})
	_, err := svc.EnableSupplier(ctx, 1, &SupplierProfile{})
	if !errors.Is(err, ErrContactNotFound) {
		t.Fatalf("err = %v, want ErrContactNotFound", err)
	}
}

func TestContactService_EnableCustomer_ContactNotFound(t *testing.T) {
	ctx := context.Background()
	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return nil, nil
			},
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})
	_, err := svc.EnableCustomer(ctx, 1, &CustomerProfile{})
	if !errors.Is(err, ErrContactNotFound) {
		t.Fatalf("err = %v, want ErrContactNotFound", err)
	}
}

func TestContactService_DisableCustomer(t *testing.T) {
	ctx := context.Background()

	disabled := false
	customers := CustomerProfileDAOMock{
		FindByContactFunc: func(_ context.Context, contactID uint64) (*CustomerProfile, error) {
			if contactID != 1 {
				return nil, nil
			}
			return &CustomerProfile{ContactID: 1, Active: true}, nil
		},
		CRUDMock: dao.CRUDMock[CustomerProfile]{
			UpdateFunc: func(_ context.Context, record *CustomerProfile) (*CustomerProfile, error) {
				if record.Active {
					t.Error("expected customer role to be deactivated")
				}
				disabled = true
				return record, nil
			},
		},
	}
	svc := NewContactService(ContactDAOMock{}, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, customers, SupplierProfileDAOMock{})

	if err := svc.DisableCustomer(ctx, 1); helper.AssertError(t, err, false, nil) {
		return
	}
	if !disabled {
		t.Error("expected customer role update")
	}

	if err := svc.DisableCustomer(ctx, 99); err != nil {
		t.Errorf("missing role should be a no-op, got %v", err)
	}
}

func TestContactService_DisableSupplier_MissingRoleIsNoOp(t *testing.T) {
	ctx := context.Background()
	suppliers := SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, _ uint64) (*SupplierProfile, error) {
			return nil, nil
		},
	}
	svc := NewContactService(ContactDAOMock{}, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, suppliers)
	if err := svc.DisableSupplier(ctx, 99); err != nil {
		t.Errorf("missing role should be a no-op, got %v", err)
	}
}

func TestContactService_SetDefaultAddress_NotFound(t *testing.T) {
	ctx := context.Background()
	addresses := ContactAddressDAOMock{
		CRUDMock: dao.CRUDMock[ContactAddress]{
			FindFunc: func(_ context.Context, _ uint64) (*ContactAddress, error) {
				return nil, nil
			},
		},
	}
	svc := NewContactService(ContactDAOMock{}, addresses, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})
	if err := svc.SetDefaultAddress(ctx, 1, 1); !errors.Is(err, ErrAddressNotFound) {
		t.Fatalf("err = %v, want ErrAddressNotFound", err)
	}
}

func TestContactService_ValidateAddresses(t *testing.T) {
	svc := NewContactService(ContactDAOMock{}, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	if err := svc.validateName("  Acme  "); err != nil {
		t.Errorf("valid name err = %v", err)
	}
	if err := svc.validateAddresses([]*ContactAddress{{Type: ptr(AddressTypeShipping), IsDefault: true}, {Type: ptr(AddressTypeBilling)}}); err != nil {
		t.Errorf("valid addresses err = %v", err)
	}
	if err := svc.validateAddresses([]*ContactAddress{{IsDefault: true}, {Type: nil, IsDefault: true}}); !errors.Is(err, ErrMultipleDefaultAddresses) {
		t.Errorf("duplicate default nil-type err = %v, want ErrMultipleDefaultAddresses", err)
	}
}
