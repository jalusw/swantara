package reference

import "errors"

var (
	ErrCurrencyNotFound         = errors.New("currency not found")
	ErrRateValidFromMissing     = errors.New("fx rate valid_from is required")
	ErrUnitNotFound             = errors.New("unit not found")
	ErrUnitGroupNotFound        = errors.New("unit category not found")
	ErrUnitNameTaken            = errors.New("unit name already exists in category")
	ErrInvalidFactor            = errors.New("unit factor must be greater than zero")
	ErrUnitGroupMismatch        = errors.New("units belong to different categories")
	ErrTermNotFound             = errors.New("payment term not found")
	ErrInvalidTermLines         = errors.New("payment term lines must sum to 100 percent and include at most one balance line")
	ErrMultipleBalanceLines     = errors.New("payment term may only contain one balance line")
	ErrTermOrganizationRequired = errors.New("payment term organization is required")
	ErrTermOrganizationMismatch = errors.New("payment term organization mismatch")
	ErrDuplicateTermName        = errors.New("payment term name already exists for organization")
	ErrPaymentTermNotInOrg      = errors.New("payment term does not belong to organization")
	ErrDuplicateCode            = errors.New("dimension account code already exists for organization")
	ErrParentNotFound           = errors.New("parent dimension account not found")

	ErrAccountNotFound    = errors.New("account not found")
	ErrInvalidAccountType = errors.New("account type is not valid")
	ErrDuplicateAccount   = errors.New("account code already exists for organization")
	ErrInvalidParent      = errors.New("parent account does not exist")

	ErrJournalNotFound       = errors.New("journal not found")
	ErrInvalidJournalType    = errors.New("journal type is not valid")
	ErrJournalDefaultAccount = errors.New("journal default account does not exist")

	ErrTaxNotFound       = errors.New("tax not found")
	ErrInvalidTaxType    = errors.New("tax type is not valid")
	ErrInvalidTaxScope   = errors.New("tax scope is not valid")
	ErrTaxAmountMissing  = errors.New("percent and fixed taxes require an amount")
	ErrInvalidTaxAccount = errors.New("tax account does not exist")

	ErrTaxYearNotFound = errors.New("tax year not found")
	ErrInvalidTaxYear  = errors.New("tax year date range is invalid")

	ErrCarrierName            = errors.New("carrier name is required")
	ErrCarrierNotFound        = errors.New("carrier not found")
	ErrCarrierDeliveryProduct = errors.New("carrier delivery item does not exist")
)
