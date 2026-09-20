package contacts

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestNewContactDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewContactDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewContactAddressDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewContactAddressDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewContactBankAccountDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewContactBankAccountDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewCustomerProfileDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewCustomerProfileDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewSupplierProfileDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewSupplierProfileDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestContactDAO_CreateWithDetails(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock)
		input   func() (*Contact, []*ContactAddress, []*ContactBankAccount, *CustomerProfile, *SupplierProfile)
		check   func(t *testing.T, created *Contact, addresses []*ContactAddress, banks []*ContactBankAccount, customer *CustomerProfile, supplier *SupplierProfile)
		wantErr bool
	}{
		{
			name: "creates contact and roles",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contacts"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contact_addresses"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contact_bank_accounts"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contact_customers"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contact_suppliers"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
				mock.ExpectCommit()
				return db, mock
			},
			input: func() (*Contact, []*ContactAddress, []*ContactBankAccount, *CustomerProfile, *SupplierProfile) {
				return &Contact{Name: "Acme Corp"},
					[]*ContactAddress{{Type: helper.Ptr(AddressTypeBilling)}},
					[]*ContactBankAccount{{BankName: helper.Ptr("Bank X")}},
					&CustomerProfile{Active: true},
					&SupplierProfile{Active: true}
			},
			check: func(t *testing.T, created *Contact, addresses []*ContactAddress, banks []*ContactBankAccount, customer *CustomerProfile, supplier *SupplierProfile) {
				if created.ID != 1 {
					t.Errorf("contact id = %d, want 1", created.ID)
				}
				if addresses[0].ContactID != 1 {
					t.Errorf("address contact_id = %d, want 1", addresses[0].ContactID)
				}
				if banks[0].ContactID != 1 {
					t.Errorf("bank contact_id = %d, want 1", banks[0].ContactID)
				}
				if customer.ContactID != 1 {
					t.Errorf("customer contact_id = %d, want 1", customer.ContactID)
				}
				if supplier.ContactID != 1 {
					t.Errorf("supplier contact_id = %d, want 1", supplier.ContactID)
				}
			},
			wantErr: false,
		},
		{
			name: "without details",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contacts"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
				return db, mock
			},
			input: func() (*Contact, []*ContactAddress, []*ContactBankAccount, *CustomerProfile, *SupplierProfile) {
				return &Contact{Name: "Solo"}, nil, nil, nil, nil
			},
			check: func(t *testing.T, created *Contact, addresses []*ContactAddress, banks []*ContactBankAccount, customer *CustomerProfile, supplier *SupplierProfile) {
				if created.ID != 1 {
					t.Errorf("contact id = %d, want 1", created.ID)
				}
			},
			wantErr: false,
		},
		{
			name: "rolls back on contact error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contacts"`)).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
				return db, mock
			},
			input: func() (*Contact, []*ContactAddress, []*ContactBankAccount, *CustomerProfile, *SupplierProfile) {
				return &Contact{Name: "Acme"}, nil, nil, nil, nil
			},
			check:   nil,
			wantErr: true,
		},
		{
			name: "rolls back on address error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contacts"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contact_addresses"`)).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
				return db, mock
			},
			input: func() (*Contact, []*ContactAddress, []*ContactBankAccount, *CustomerProfile, *SupplierProfile) {
				return &Contact{Name: "Acme"}, []*ContactAddress{{}}, nil, nil, nil
			},
			check:   nil,
			wantErr: true,
		},
		{
			name: "rolls back on bank error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contacts"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contact_bank_accounts"`)).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
				return db, mock
			},
			input: func() (*Contact, []*ContactAddress, []*ContactBankAccount, *CustomerProfile, *SupplierProfile) {
				return &Contact{Name: "Acme"}, nil, []*ContactBankAccount{{}}, nil, nil
			},
			check:   nil,
			wantErr: true,
		},
		{
			name: "rolls back on customer error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contacts"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contact_customers"`)).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
				return db, mock
			},
			input: func() (*Contact, []*ContactAddress, []*ContactBankAccount, *CustomerProfile, *SupplierProfile) {
				return &Contact{Name: "Acme"}, nil, nil, &CustomerProfile{}, nil
			},
			check:   nil,
			wantErr: true,
		},
		{
			name: "rolls back on supplier error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contacts"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contact_customers"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "contact_suppliers"`)).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
				return db, mock
			},
			input: func() (*Contact, []*ContactAddress, []*ContactBankAccount, *CustomerProfile, *SupplierProfile) {
				return &Contact{Name: "Acme"}, nil, nil, &CustomerProfile{}, &SupplierProfile{}
			},
			check:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.setup(t)
			ctx := context.Background()
			contacts := NewContactDAO(db)

			contact, addresses, banks, customer, supplier := tt.input()
			created, err := contacts.CreateWithDetails(ctx, contact, addresses, banks, customer, supplier)

			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if tt.check != nil {
				tt.check(t, created, addresses, banks, customer, supplier)
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestContactAddressDAO_ListByContact(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock)
		wantLen int
		wantErr bool
	}{
		{
			name: "returns addresses",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "contact_addresses" WHERE contact_id = $1`)).
					WithArgs(uint64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "contact_addresses" WHERE contact_id = $1`)).
					WithArgs(uint64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "contact_id", "type"}).
						AddRow(1, 1, AddressTypeBilling).
						AddRow(2, 1, AddressTypeShipping))
				return db, mock
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "propagates error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "contact_addresses"`)).
					WithArgs(uint64(1)).
					WillReturnError(errors.New("db down"))
				return db, mock
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.setup(t)
			ctx := context.Background()
			addresses := NewContactAddressDAO(db)

			items, err := addresses.ListByContact(ctx, 1)

			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if !tt.wantErr && len(items) != tt.wantLen {
				t.Errorf("items = %+v, want %d items", items, tt.wantLen)
			}
			if !tt.wantErr && tt.name == "returns addresses" {
				if items[0].Type == nil || *items[0].Type != AddressTypeBilling {
					t.Errorf("first item type = %+v, want %s", items[0].Type, AddressTypeBilling)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestContactAddressDAO_SetDefault(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "sets single default",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "contact_addresses" SET`)).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "contact_addresses" SET`)).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
				return db, mock
			},
			wantErr: false,
		},
		{
			name: "rolls back on clear error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "contact_addresses" SET`)).
					WillReturnError(errors.New("update failed"))
				mock.ExpectRollback()
				return db, mock
			},
			wantErr: true,
		},
		{
			name: "rolls back on set error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "contact_addresses" SET`)).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "contact_addresses" SET`)).
					WillReturnError(errors.New("update failed"))
				mock.ExpectRollback()
				return db, mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.setup(t)
			ctx := context.Background()
			addresses := NewContactAddressDAO(db)

			err := addresses.SetDefault(ctx, 1, 2)

			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestContactBankAccountDAO_ListByContact(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock)
		wantLen int
		wantErr bool
	}{
		{
			name: "returns accounts",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "contact_bank_accounts" WHERE contact_id = $1`)).
					WithArgs(uint64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "contact_bank_accounts" WHERE contact_id = $1`)).
					WithArgs(uint64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "contact_id", "bank_name"}).
						AddRow(1, 1, "Bank X"))
				return db, mock
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "propagates error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "contact_bank_accounts"`)).
					WithArgs(uint64(1)).
					WillReturnError(errors.New("db down"))
				return db, mock
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.setup(t)
			ctx := context.Background()
			banks := NewContactBankAccountDAO(db)

			items, err := banks.ListByContact(ctx, 1)

			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if !tt.wantErr && len(items) != tt.wantLen {
				t.Errorf("items = %+v, want %d items", items, tt.wantLen)
			}
			if !tt.wantErr && tt.name == "returns accounts" {
				if items[0].BankName == nil || *items[0].BankName != "Bank X" {
					t.Errorf("bank_name = %+v, want Bank X", items[0].BankName)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestCustomerProfileDAO_FindByContact(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock)
		contact uint64
		wantNil bool
		wantErr bool
	}{
		{
			name: "finds role",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "contact_customers" WHERE contact_id = $1 LIMIT $2`)).
					WithArgs(uint64(1), 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "contact_id", "active"}).AddRow(1, 1, true))
				return db, mock
			},
			contact: 1,
			wantNil: false,
			wantErr: false,
		},
		{
			name: "missing role returns nil",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "contact_customers" WHERE contact_id = $1 LIMIT $2`)).
					WithArgs(uint64(9), 1).
					WillReturnError(gorm.ErrRecordNotFound)
				return db, mock
			},
			contact: 9,
			wantNil: true,
			wantErr: false,
		},
		{
			name: "propagates error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "contact_customers" WHERE contact_id = $1 LIMIT $2`)).
					WithArgs(uint64(1), 1).
					WillReturnError(errors.New("db down"))
				return db, mock
			},
			contact: 1,
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.setup(t)
			ctx := context.Background()
			customers := NewCustomerProfileDAO(db)

			record, err := customers.FindByContact(ctx, tt.contact)

			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if tt.wantNil && record != nil {
				t.Errorf("record = %+v, want nil", record)
			}
			if !tt.wantNil && !tt.wantErr && record.ContactID != tt.contact {
				t.Errorf("record.ContactID = %d, want %d", record.ContactID, tt.contact)
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestSupplierProfileDAO_FindByContact(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock)
		contact uint64
		wantNil bool
		wantErr bool
	}{
		{
			name: "finds role",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "contact_suppliers" WHERE contact_id = $1 LIMIT $2`)).
					WithArgs(uint64(1), 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "contact_id", "active"}).AddRow(1, 1, true))
				return db, mock
			},
			contact: 1,
			wantNil: false,
			wantErr: false,
		},
		{
			name: "missing role returns nil",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "contact_suppliers" WHERE contact_id = $1 LIMIT $2`)).
					WithArgs(uint64(9), 1).
					WillReturnError(gorm.ErrRecordNotFound)
				return db, mock
			},
			contact: 9,
			wantNil: true,
			wantErr: false,
		},
		{
			name: "propagates error",
			setup: func(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
				db, mock := query.NewMockDB(t)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "contact_suppliers" WHERE contact_id = $1 LIMIT $2`)).
					WithArgs(uint64(1), 1).
					WillReturnError(errors.New("db down"))
				return db, mock
			},
			contact: 1,
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.setup(t)
			ctx := context.Background()
			suppliers := NewSupplierProfileDAO(db)

			record, err := suppliers.FindByContact(ctx, tt.contact)

			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if tt.wantNil && record != nil {
				t.Errorf("record = %+v, want nil", record)
			}
			if !tt.wantNil && !tt.wantErr && record.ContactID != tt.contact {
				t.Errorf("record.ContactID = %d, want %d", record.ContactID, tt.contact)
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}
