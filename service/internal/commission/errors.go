package commission

import "errors"

var (
	ErrPlanNotFound       = errors.New("commission plan not found")
	ErrPlanInactive       = errors.New("commission plan is inactive")
	ErrPlanInvalidBasis   = errors.New("commission plan basis is invalid")
	ErrRuleNotFound       = errors.New("commission rule not found")
	ErrRuleInvalid        = errors.New("commission rule must set a rate percentage or fixed amount")
	ErrRuleRange          = errors.New("commission rule range is invalid")
	ErrAssignmentNotFound = errors.New("commission assignment not found")
	ErrAssignmentOverlap  = errors.New("salesperson already assigned to this plan for the period")
	ErrEntryNotFound      = errors.New("commission entry not found")
	ErrEntryNotConfirmed  = errors.New("commission entry must be confirmed to settle")
	ErrEntryNotDraft      = errors.New("commission entry must be draft to confirm")
	ErrEntryNotOpen       = errors.New("commission entry is not open to cancel")
	ErrNoRuleMatches      = errors.New("no commission rule matches the amount")
	ErrAccountsRequired   = errors.New("commission accrual requires expense and payable accounts")
	ErrSourceMissing      = errors.New("commission source must identify an invoice or payment")
	ErrInvoiceNotFound    = errors.New("commission source invoice not found")
	ErrEntryDuplicate     = errors.New("commission entry already exists for the source")
)
