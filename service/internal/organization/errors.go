package organization

import "errors"

var (
	ErrOrganizationNotFound = errors.New("organization not found")
	ErrNameRequired         = errors.New("organization name is required")
	ErrBaseCurrencyNotFound = errors.New("base currency does not exist")
	ErrParentNotFound       = errors.New("parent organization not found")
	ErrParentCycle          = errors.New("organization parent would create a cycle")
	ErrCountryCodeInvalid   = errors.New("unsupported country code")
	ErrModuleUnknown        = errors.New("unknown module")
)
