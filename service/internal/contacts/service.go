package contacts

import (
	"context"
	"strings"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type ContactService struct {
	contacts  ContactDAO
	addresses ContactAddressDAO
	customers CustomerProfileDAO
	suppliers SupplierProfileDAO
	banks     ContactBankAccountDAO
}

func NewContactService(
	contacts ContactDAO,
	addresses ContactAddressDAO,
	banks ContactBankAccountDAO,
	customers CustomerProfileDAO,
	suppliers SupplierProfileDAO,
) ContactService {
	return ContactService{
		contacts:  contacts,
		addresses: addresses,
		banks:     banks,
		customers: customers,
		suppliers: suppliers,
	}
}

func (s ContactService) List(ctx context.Context, q *query.Query) (*query.Page[Contact], error) {
	return s.contacts.List(ctx, q)
}

func (s ContactService) Find(ctx context.Context, id uint64) (*Contact, error) {
	return s.contacts.Find(ctx, id)
}

func (s ContactService) Delete(ctx context.Context, id uint64) error {
	return s.contacts.Delete(ctx, id)
}

func (s ContactService) ListAddressesByContact(ctx context.Context, contactID uint64) ([]*ContactAddress, error) {
	return s.addresses.ListByContact(ctx, contactID)
}

func (s ContactService) FindAddress(ctx context.Context, id uint64) (*ContactAddress, error) {
	return s.addresses.Find(ctx, id)
}

func (s ContactService) CreateAddress(ctx context.Context, address *ContactAddress) (*ContactAddress, error) {
	return s.addresses.Create(ctx, address)
}

func (s ContactService) UpdateAddress(ctx context.Context, address *ContactAddress) (*ContactAddress, error) {
	return s.addresses.Update(ctx, address)
}

func (s ContactService) DeleteAddress(ctx context.Context, id uint64) error {
	return s.addresses.Delete(ctx, id)
}

func (s ContactService) ListBankAccountsByContact(ctx context.Context, contactID uint64) ([]*ContactBankAccount, error) {
	return s.banks.ListByContact(ctx, contactID)
}

func (s ContactService) FindBankAccount(ctx context.Context, id uint64) (*ContactBankAccount, error) {
	return s.banks.Find(ctx, id)
}

func (s ContactService) CreateBankAccount(ctx context.Context, account *ContactBankAccount) (*ContactBankAccount, error) {
	return s.banks.Create(ctx, account)
}

func (s ContactService) UpdateBankAccount(ctx context.Context, account *ContactBankAccount) (*ContactBankAccount, error) {
	return s.banks.Update(ctx, account)
}

func (s ContactService) DeleteBankAccount(ctx context.Context, id uint64) error {
	return s.banks.Delete(ctx, id)
}

func (s ContactService) Create(
	ctx context.Context,
	contact *Contact,
	addresses []*ContactAddress,
	banks []*ContactBankAccount,
	customer *CustomerProfile,
	supplier *SupplierProfile,
) (*Contact, error) {
	if err := s.validateName(contact.Name); err != nil {
		return nil, err
	}
	if err := s.validateAddresses(addresses); err != nil {
		return nil, err
	}

	return s.contacts.CreateWithDetails(ctx, contact, addresses, banks, customer, supplier)
}

func (s ContactService) Update(ctx context.Context, contactID uint64, contact *Contact) (*Contact, error) {
	existing, err := s.contacts.Find(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrContactNotFound
	}
	if err := s.validateName(contact.Name); err != nil {
		return nil, err
	}

	existing.Name = contact.Name
	existing.DisplayName = contact.DisplayName
	existing.IsOrganization = contact.IsOrganization
	existing.ParentID = contact.ParentID
	existing.Email = contact.Email
	existing.Phone = contact.Phone
	existing.Mobile = contact.Mobile
	existing.Website = contact.Website
	existing.TaxID = contact.TaxID
	existing.Industry = contact.Industry
	existing.CurrencyCode = contact.CurrencyCode
	existing.Lang = contact.Lang
	existing.Active = contact.Active

	return s.contacts.Update(ctx, existing)
}

func (s ContactService) EnableCustomer(ctx context.Context, contactID uint64, extension *CustomerProfile) (*CustomerProfile, error) {
	existing, err := s.contacts.Find(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrContactNotFound
	}

	record, err := s.customers.FindByContact(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		extension.ContactID = contactID
		if !extension.Active {
			extension.Active = true
		}
		return s.customers.Create(ctx, extension)
	}

	record.CustomerPaymentTermID = extension.CustomerPaymentTermID
	record.CreditLimit = extension.CreditLimit
	record.ReceivableAccountID = extension.ReceivableAccountID
	record.Active = true
	return s.customers.Update(ctx, record)
}

func (s ContactService) DisableCustomer(ctx context.Context, contactID uint64) error {
	record, err := s.customers.FindByContact(ctx, contactID)
	if err != nil {
		return err
	}
	if record == nil {
		return nil
	}

	record.Active = false
	if _, err := s.customers.Update(ctx, record); err != nil {
		return err
	}
	return nil
}

func (s ContactService) EnableSupplier(ctx context.Context, contactID uint64, extension *SupplierProfile) (*SupplierProfile, error) {
	existing, err := s.contacts.Find(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrContactNotFound
	}

	record, err := s.suppliers.FindByContact(ctx, contactID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		extension.ContactID = contactID
		if !extension.Active {
			extension.Active = true
		}
		return s.suppliers.Create(ctx, extension)
	}

	record.SupplierPaymentTermID = extension.SupplierPaymentTermID
	record.PayableAccountID = extension.PayableAccountID
	record.Active = true
	return s.suppliers.Update(ctx, record)
}

func (s ContactService) DisableSupplier(ctx context.Context, contactID uint64) error {
	record, err := s.suppliers.FindByContact(ctx, contactID)
	if err != nil {
		return err
	}
	if record == nil {
		return nil
	}

	record.Active = false
	if _, err := s.suppliers.Update(ctx, record); err != nil {
		return err
	}
	return nil
}

func (s ContactService) SetDefaultAddress(ctx context.Context, contactID, addressID uint64) error {
	address, err := s.addresses.Find(ctx, addressID)
	if err != nil {
		return err
	}
	if address == nil {
		return ErrAddressNotFound
	}
	if address.ContactID != contactID {
		return ErrAddressNotForContact
	}

	return s.addresses.SetDefault(ctx, contactID, addressID)
}

func (s ContactService) validateName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrNameRequired
	}
	return nil
}

func (s ContactService) validateAddresses(addresses []*ContactAddress) error {
	seen := map[string]bool{}
	for _, address := range addresses {
		if address.Type != nil && !validAddressType(*address.Type) {
			return ErrInvalidAddressType
		}
		if address.IsDefault {
			key := ""
			if address.Type != nil {
				key = *address.Type
			}
			if seen[key] {
				return ErrMultipleDefaultAddresses
			}
			seen[key] = true
		}
	}
	return nil
}

func validAddressType(value string) bool {
	switch value {
	case AddressTypeBilling, AddressTypeShipping, AddressTypeOther:
		return true
	default:
		return false
	}
}
