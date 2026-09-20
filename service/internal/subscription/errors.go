package subscription

import "errors"

var (
	ErrSubscriptionNotFound                = errors.New("subscription not found")
	ErrSubscriptionState                   = errors.New("subscription state is invalid for this operation")
	ErrSubscriptionNoLines                 = errors.New("subscription must have at least one line")
	ErrSubscriptionNoPlan                  = errors.New("subscription plan is required")
	ErrSubscriptionPlanNotFound            = errors.New("subscription plan not found")
	ErrSubscriptionPlanOrganization        = errors.New("subscription plan does not belong to the organization")
	ErrSubscriptionLineProduct             = errors.New("subscription line item is required")
	ErrSubscriptionLineProductOrganization = errors.New("subscription line item does not belong to the organization")
	ErrSubscriptionLineQty                 = errors.New("subscription line quantity must be greater than zero")
	ErrSubscriptionLinePrice               = errors.New("subscription line unit price must not be negative")
	ErrSubscriptionLineDiscount            = errors.New("subscription line discount must be between 0 and 100")
	ErrSubscriptionNotDue                  = errors.New("subscription is not due for invoicing")
	ErrSubscriptionAlreadyInvoiced         = errors.New("subscription already invoiced for this period")
	ErrSubscriptionContact                 = errors.New("subscription contact is required")
	ErrSubscriptionContactOrganization     = errors.New("subscription contact does not belong to the organization")
	ErrSubscriptionCurrency                = errors.New("subscription currency is required")
	ErrSubscriptionPriceBook               = errors.New("subscription price_book is required")
	ErrSubscriptionPriceBookOrganization   = errors.New("subscription price_book does not belong to the organization")
	ErrSubscriptionConfig                  = errors.New("subscription is not configured for invoicing")
)
