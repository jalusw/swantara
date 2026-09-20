package products

import "errors"

var (
	ErrItemNotFound             = errors.New("item not found")
	ErrVariantNotFound          = errors.New("item variant not found")
	ErrPriceBookNotFound        = errors.New("price_book not found")
	ErrSkuTaken                 = errors.New("sku already exists")
	ErrInvalidAppliesTo         = errors.New("invalid price_book rule applies_to scope")
	ErrInvalidComputeType       = errors.New("invalid price_book rule compute type")
	ErrInvalidRuleScope         = errors.New("price_book rule scope requires a item or category")
	ErrInvalidRuleValue         = errors.New("price_book rule discount must be between 0 and 100 and fixed price cannot be negative")
	ErrPriceUnavailable         = errors.New("no price rule resolves for the given variant")
	ErrItemCategory             = errors.New("item category does not exist")
	ErrItemCategoryOrganization = errors.New("item category does not belong to the organization")
	ErrVariantTemplate          = errors.New("item variant template does not exist")
	ErrEmptyAttributeMatrix     = errors.New("variant generation requires at least one attribute")
	ErrSupplierNotFound         = errors.New("supplier supplier contact not found")
	ErrSupplierNotSupplier      = errors.New("supplier contact is not an active supplier")
	ErrInvalidValidity          = errors.New("supplier item validity window is invalid")
	ErrInvalidMinQty            = errors.New("supplier item minimum quantity cannot be negative")
	ErrNoValidOffer             = errors.New("no valid supplier offer matches the variant, quantity and date")
	ErrNoIncomeAccount          = errors.New("item category has no income account configured")
	ErrNoExpenseAccount         = errors.New("item category has no expense account configured")
)
