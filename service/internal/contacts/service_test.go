package contacts

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func ptr[T any](v T) *T {
	return &v
}

func TestContactService_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		contact   *Contact
		addresses []*ContactAddress
		wantErrV  error
	}{
		{
			name:      "creates contact with details",
			contact:   &Contact{Name: "Acme Corp"},
			addresses: []*ContactAddress{{Type: ptr(AddressTypeBilling), IsDefault: true}},
		},
		{
			name:     "rejects empty name",
			contact:  &Contact{Name: ""},
			wantErrV: ErrNameRequired,
		},
		{
			name:      "rejects unknown address type",
			contact:   &Contact{Name: "Acme Corp"},
			addresses: []*ContactAddress{{Type: ptr("warehouse")}},
			wantErrV:  ErrInvalidAddressType,
		},
		{
			name:      "rejects multiple default addresses",
			contact:   &Contact{Name: "Acme Corp"},
			addresses: []*ContactAddress{{Type: ptr(AddressTypeBilling), IsDefault: true}, {Type: ptr(AddressTypeBilling), IsDefault: true}},
			wantErrV:  ErrMultipleDefaultAddresses,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var persisted *Contact
			contacts := ContactDAOMock{
				CreateWithDetailsFunc: func(_ context.Context, contact *Contact, addresses []*ContactAddress, _ []*ContactBankAccount, _ *CustomerProfile, _ *SupplierProfile) (*Contact, error) {
					contact.ID = 1
					persisted = contact
					for _, address := range addresses {
						address.ContactID = 1
					}
					return contact, nil
				},
			}
			svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

			created, err := svc.Create(ctx, tt.contact, tt.addresses, nil, nil, nil)
			if helper.AssertError(t, err, tt.wantErrV != nil, tt.wantErrV) {
				return
			}
			if created == nil || persisted == nil {
				t.Fatal("expected created contact, got nil")
			}
		})
	}
}

func TestContactService_Update_NotFound(t *testing.T) {
	ctx := context.Background()

	contacts := ContactDAOMock{
		CRUDMock: dao.CRUDMock[Contact]{
			FindFunc: func(_ context.Context, _ uint64) (*Contact, error) {
				return nil, nil
			},
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	_, err := svc.Update(ctx, 1, &Contact{Name: "Acme"})
	if helper.AssertError(t, err, true, ErrContactNotFound) {
		return
	}
}

func TestContactService_EnableCustomer_GetOrCreate(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		existing   *CustomerProfile
		wantCreate bool
	}{
		{
			name:       "creates customer role for new contact role",
			existing:   nil,
			wantCreate: true,
		},
		{
			name:       "updates existing customer role",
			existing:   &CustomerProfile{ContactID: 1, Active: false},
			wantCreate: false,
		},
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
			customers := CustomerProfileDAOMock{
				FindByContactFunc: func(_ context.Context, _ uint64) (*CustomerProfile, error) {
					return tt.existing, nil
				},
				CRUDMock: dao.CRUDMock[CustomerProfile]{
					CreateFunc: func(_ context.Context, record *CustomerProfile) (*CustomerProfile, error) {
						if !tt.wantCreate {
							t.Error("expected update path, but Create was called")
						}
						return record, nil
					},
					UpdateFunc: func(_ context.Context, record *CustomerProfile) (*CustomerProfile, error) {
						if tt.wantCreate {
							t.Error("expected create path, but Update was called")
						}
						if !record.Active {
							t.Error("expected customer role to be activated")
						}
						return record, nil
					},
				},
			}
			svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, customers, SupplierProfileDAOMock{})

			_, err := svc.EnableCustomer(ctx, 1, &CustomerProfile{})
			if helper.AssertError(t, err, false, nil) {
				return
			}
		})
	}
}

func TestContactService_DisableSupplier(t *testing.T) {
	ctx := context.Background()

	suppliers := SupplierProfileDAOMock{
		FindByContactFunc: func(_ context.Context, contactID uint64) (*SupplierProfile, error) {
			if contactID != 1 {
				return nil, nil
			}
			return &SupplierProfile{ContactID: 1, Active: true}, nil
		},
		CRUDMock: dao.CRUDMock[SupplierProfile]{
			UpdateFunc: func(_ context.Context, record *SupplierProfile) (*SupplierProfile, error) {
				if record.Active {
					t.Error("expected supplier role to be deactivated")
				}
				return record, nil
			},
		},
	}
	svc := NewContactService(ContactDAOMock{}, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, suppliers)

	if err := svc.DisableSupplier(ctx, 1); helper.AssertError(t, err, false, nil) {
		return
	}
}

func TestContactService_Create_WithCustomerAndSupplierRoles(t *testing.T) {
	ctx := context.Background()

	var persistedCustomer *CustomerProfile
	var persistedSupplier *SupplierProfile
	contacts := ContactDAOMock{
		CreateWithDetailsFunc: func(_ context.Context, contact *Contact, _ []*ContactAddress, _ []*ContactBankAccount, customer *CustomerProfile, supplier *SupplierProfile) (*Contact, error) {
			contact.ID = 1
			persistedCustomer = customer
			persistedSupplier = supplier
			return contact, nil
		},
	}
	svc := NewContactService(contacts, ContactAddressDAOMock{}, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	created, err := svc.Create(ctx, &Contact{Name: "Acme Corp"}, nil, nil, &CustomerProfile{Active: true}, &SupplierProfile{Active: true})
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}
	if persistedCustomer == nil || !persistedCustomer.Active {
		t.Error("expected active customer role to be persisted")
	}
	if persistedSupplier == nil || !persistedSupplier.Active {
		t.Error("expected active supplier role to be persisted")
	}
}

func TestContactService_SetDefaultAddress_Success(t *testing.T) {
	ctx := context.Background()

	addresses := ContactAddressDAOMock{
		CRUDMock: dao.CRUDMock[ContactAddress]{
			FindFunc: func(_ context.Context, addressID uint64) (*ContactAddress, error) {
				if addressID != 1 {
					return nil, nil
				}
				return &ContactAddress{Base: model.Base{ID: 1}, ContactID: 1}, nil
			},
		},
		SetDefaultFunc: func(_ context.Context, contactID, addressID uint64) error {
			if contactID != 1 || addressID != 1 {
				t.Errorf("SetDefault(%d, %d), want (1, 1)", contactID, addressID)
			}
			return nil
		},
	}
	svc := NewContactService(ContactDAOMock{}, addresses, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	if err := svc.SetDefaultAddress(ctx, 1, 1); !helper.AssertError(t, err, false, nil) {
		return
	}
}

func TestContactService_SetDefaultAddress_ValidatesOwnership(t *testing.T) {
	ctx := context.Background()

	addresses := ContactAddressDAOMock{
		CRUDMock: dao.CRUDMock[ContactAddress]{
			FindFunc: func(_ context.Context, addressID uint64) (*ContactAddress, error) {
				if addressID != 1 {
					return nil, nil
				}
				return &ContactAddress{ContactID: 2}, nil
			},
		},
	}
	svc := NewContactService(ContactDAOMock{}, addresses, ContactBankAccountDAOMock{}, CustomerProfileDAOMock{}, SupplierProfileDAOMock{})

	if err := svc.SetDefaultAddress(ctx, 1, 1); !helper.AssertError(t, err, true, ErrAddressNotForContact) {
		return
	}
}
