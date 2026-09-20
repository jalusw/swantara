package contacts

import "errors"

const (
	AddressTypeBilling  = "billing"
	AddressTypeShipping = "shipping"
	AddressTypeOther    = "other"
)

var (
	ErrContactNotFound          = errors.New("contact not found")
	ErrNameRequired             = errors.New("contact name is required")
	ErrInvalidAddressType       = errors.New("address type must be one of billing, shipping, other")
	ErrMultipleDefaultAddresses = errors.New("only one default address is allowed per contact")
	ErrAddressNotFound          = errors.New("address not found")
	ErrBankAccountNotFound      = errors.New("bank account not found")
	ErrAddressNotForContact     = errors.New("address does not belong to contact")
	ErrBankAccountNotForContact = errors.New("bank account does not belong to contact")
)
