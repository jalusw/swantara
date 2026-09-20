package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CurrencyRateResolver interface {
	Rate(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error)
}

type ClosingJournalResolver interface {
	ClosingJournalID(ctx context.Context, organizationID uint64) (uint64, error)
}

type FxAccountResolver interface {
	FxGainAccountID(ctx context.Context, organizationID uint64) (uint64, error)
	FxLossAccountID(ctx context.Context, organizationID uint64) (uint64, error)
}

type OrganizationLookup interface {
	Find(ctx context.Context, id uint64) (*reference.Organization, error)
}

type PaymentTermSplit struct {
	Amount   amount.Amount
	DueDate  time.Time
	Sequence int
}

type PaymentTermSplitter interface {
	Splits(ctx context.Context, termID uint64, total amount.Amount, date time.Time) ([]PaymentTermSplit, error)
}

type TaxRuleMapper interface {
	Find(ctx context.Context, id uint64) (*TaxRule, error)
	Resolve(ctx context.Context, positionID uint64, srcTaxID *uint64, srcAccountID uint64) (*ResolveResult, error)
}

type ContactCreditLimiter interface {
	CreditLimit(ctx context.Context, contactID uint64) (*float64, error)
}

type StockMovementPostRequest struct {
	OrganizationID uint64
	JournalID      uint64
	Date           time.Time
	Ref            string
	OriginType     string
	OriginID       uint64
	Description    string
	Lines          []PostingLine
}
