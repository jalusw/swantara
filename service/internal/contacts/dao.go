package contacts

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type ContactDAO interface {
	dao.CRUD[Contact]
	CreateWithDetails(
		ctx context.Context,
		contact *Contact,
		addresses []*ContactAddress,
		banks []*ContactBankAccount,
		customer *CustomerProfile,
		supplier *SupplierProfile,
	) (*Contact, error)
	CreateWithDetailsTx(
		ctx context.Context,
		tx *gorm.DB,
		contact *Contact,
		addresses []*ContactAddress,
		banks []*ContactBankAccount,
		customer *CustomerProfile,
		supplier *SupplierProfile,
	) (*Contact, error)
}

type contactDAO struct {
	dao.Base[Contact]
	db *gorm.DB
}

func NewContactDAO(db *gorm.DB) ContactDAO {
	return contactDAO{Base: dao.NewBase[Contact](db), db: db}
}

func (d contactDAO) CreateWithDetails(
	ctx context.Context,
	contact *Contact,
	addresses []*ContactAddress,
	banks []*ContactBankAccount,
	customer *CustomerProfile,
	supplier *SupplierProfile,
) (*Contact, error) {
	var created *Contact
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		created, err = d.CreateWithDetailsTx(ctx, tx, contact, addresses, banks, customer, supplier)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (d contactDAO) CreateWithDetailsTx(
	ctx context.Context,
	tx *gorm.DB,
	contact *Contact,
	addresses []*ContactAddress,
	banks []*ContactBankAccount,
	customer *CustomerProfile,
	supplier *SupplierProfile,
) (*Contact, error) {
	if err := tx.WithContext(ctx).Create(contact).Error; err != nil {
		return nil, err
	}
	for _, address := range addresses {
		address.ContactID = contact.ID
		if err := tx.WithContext(ctx).Create(address).Error; err != nil {
			return nil, err
		}
	}
	for _, bank := range banks {
		bank.ContactID = contact.ID
		if err := tx.WithContext(ctx).Create(bank).Error; err != nil {
			return nil, err
		}
	}
	if customer != nil {
		customer.ContactID = contact.ID
		if err := tx.WithContext(ctx).Create(customer).Error; err != nil {
			return nil, err
		}
	}
	if supplier != nil {
		supplier.ContactID = contact.ID
		if err := tx.WithContext(ctx).Create(supplier).Error; err != nil {
			return nil, err
		}
	}
	return contact, nil
}

type ContactAddressDAO interface {
	dao.CRUD[ContactAddress]
	ListByContact(ctx context.Context, contactID uint64) ([]*ContactAddress, error)
	SetDefault(ctx context.Context, contactID, addressID uint64) error
}

type contactAddressDAO struct {
	dao.Base[ContactAddress]
	db *gorm.DB
}

func NewContactAddressDAO(db *gorm.DB) ContactAddressDAO {
	return contactAddressDAO{Base: dao.NewBase[ContactAddress](db), db: db}
}

func (d contactAddressDAO) ListByContact(ctx context.Context, contactID uint64) ([]*ContactAddress, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "contact_id", Operator: query.Equal, Value: contactID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d contactAddressDAO) SetDefault(ctx context.Context, contactID, addressID uint64) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&ContactAddress{}).
			Where("contact_id = ?", contactID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&ContactAddress{}).
			Where("id = ?", addressID).
			Update("is_default", true).Error
	})
}

type ContactBankAccountDAO interface {
	dao.CRUD[ContactBankAccount]
	ListByContact(ctx context.Context, contactID uint64) ([]*ContactBankAccount, error)
}

type contactBankAccountDAO struct {
	dao.Base[ContactBankAccount]
}

func NewContactBankAccountDAO(db *gorm.DB) ContactBankAccountDAO {
	return contactBankAccountDAO{Base: dao.NewBase[ContactBankAccount](db)}
}

func (d contactBankAccountDAO) ListByContact(ctx context.Context, contactID uint64) ([]*ContactBankAccount, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "contact_id", Operator: query.Equal, Value: contactID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type CustomerProfileDAO interface {
	dao.CRUD[CustomerProfile]
	FindByContact(ctx context.Context, contactID uint64) (*CustomerProfile, error)
}

type contactCustomerDAO struct {
	dao.Base[CustomerProfile]
}

func NewCustomerProfileDAO(db *gorm.DB) CustomerProfileDAO {
	return contactCustomerDAO{Base: dao.NewBase[CustomerProfile](db)}
}

func (d contactCustomerDAO) FindByContact(ctx context.Context, contactID uint64) (*CustomerProfile, error) {
	return d.Search(ctx, "contact_id", contactID)
}

type SupplierProfileDAO interface {
	dao.CRUD[SupplierProfile]
	FindByContact(ctx context.Context, contactID uint64) (*SupplierProfile, error)
}

type contactSupplierDAO struct {
	dao.Base[SupplierProfile]
}

func NewSupplierProfileDAO(db *gorm.DB) SupplierProfileDAO {
	return contactSupplierDAO{Base: dao.NewBase[SupplierProfile](db)}
}

func (d contactSupplierDAO) FindByContact(ctx context.Context, contactID uint64) (*SupplierProfile, error) {
	return d.Search(ctx, "contact_id", contactID)
}
