package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type PosterMock struct {
	PostFunc    func(ctx context.Context, request PostRequest) (*JournalEntry, error)
	PostTxFunc  func(ctx context.Context, tx *gorm.DB, request PostRequest) (*JournalEntry, error)
	ReverseFunc func(ctx context.Context, request ReverseRequest) (*JournalEntry, error)
}

func (m PosterMock) Post(ctx context.Context, request PostRequest) (*JournalEntry, error) {
	if m.PostFunc != nil {
		return m.PostFunc(ctx, request)
	}
	return &JournalEntry{}, nil
}

func (m PosterMock) PostTx(ctx context.Context, tx *gorm.DB, request PostRequest) (*JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return m.Post(ctx, request)
}

func (m PosterMock) Reverse(ctx context.Context, request ReverseRequest) (*JournalEntry, error) {
	if m.ReverseFunc != nil {
		return m.ReverseFunc(ctx, request)
	}
	return &JournalEntry{}, nil
}

func (m PosterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request ReverseRequest) (*JournalEntry, error) {
	return m.Reverse(ctx, request)
}

type AccountLookupMock struct {
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error)
}

func (m AccountLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
}

type TransactionerMock struct {
	RunFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error
}

func (m TransactionerMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, fn)
	}
	return fn(nil)
}

type SequenceDAOMock struct {
	ReserveFunc func(ctx context.Context, organizationID uint64, code string, now time.Time) (*sequence.Reservation, error)
}

func (m SequenceDAOMock) Reserve(ctx context.Context, organizationID uint64, code string, now time.Time) (*sequence.Reservation, error) {
	if m.ReserveFunc != nil {
		return m.ReserveFunc(ctx, organizationID, code, now)
	}
	return &sequence.Reservation{Value: 1, Number: "INV/00001"}, nil
}

func (m SequenceDAOMock) ReserveTx(ctx context.Context, tx *gorm.DB, organizationID uint64, code string, now time.Time) (*sequence.Reservation, error) {
	return m.Reserve(ctx, organizationID, code, now)
}

type TaxPeriodDAOMock struct {
	dao.CRUDMock[TaxPeriod]
	ListByOrganizationFunc func(ctx context.Context, organizationID uint64) ([]*TaxPeriod, error)
	FindByDateFunc         func(ctx context.Context, organizationID uint64, date time.Time) (*TaxPeriod, error)
}

func (m TaxPeriodDAOMock) ListByOrganization(ctx context.Context, organizationID uint64) ([]*TaxPeriod, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, organizationID)
	}
	return []*TaxPeriod{}, nil
}

func (m TaxPeriodDAOMock) FindByDate(ctx context.Context, organizationID uint64, date time.Time) (*TaxPeriod, error) {
	if m.FindByDateFunc != nil {
		return m.FindByDateFunc(ctx, organizationID, date)
	}
	return nil, nil
}

func (m TaxPeriodDAOMock) UpdateTx(ctx context.Context, _ *gorm.DB, period *TaxPeriod) (*TaxPeriod, error) {
	return m.Update(ctx, period)
}

type JournalResolverMock struct {
	JournalIDFunc func(ctx context.Context, organizationID uint64) (uint64, error)
}

func (m JournalResolverMock) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.JournalIDFunc != nil {
		return m.JournalIDFunc(ctx, organizationID)
	}
	return 10, nil
}

type PaymentTermSplitterMock struct {
	SplitsFunc func(ctx context.Context, termID uint64, total amount.Amount, date time.Time) ([]PaymentTermSplit, error)
}

func (m PaymentTermSplitterMock) Splits(ctx context.Context, termID uint64, total amount.Amount, date time.Time) ([]PaymentTermSplit, error) {
	if m.SplitsFunc != nil {
		return m.SplitsFunc(ctx, termID, total, date)
	}
	return []PaymentTermSplit{{Amount: total, DueDate: date}}, nil
}

type PaymentCreatorMock struct {
	CreateFunc         func(ctx context.Context, request CreatePaymentRequest) (*Payment, error)
	CreateOutboundFunc func(ctx context.Context, request CreatePaymentRequest) (*Payment, error)
}

func (m PaymentCreatorMock) Create(ctx context.Context, request CreatePaymentRequest) (*Payment, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, request)
	}
	return &Payment{}, nil
}

func (m PaymentCreatorMock) CreateOutbound(ctx context.Context, request CreatePaymentRequest) (*Payment, error) {
	if m.CreateOutboundFunc != nil {
		return m.CreateOutboundFunc(ctx, request)
	}
	return &Payment{}, nil
}

type ContactCreditLimiterMock struct {
	CreditLimitFunc func(ctx context.Context, contactID uint64) (*float64, error)
}

func (m ContactCreditLimiterMock) CreditLimit(ctx context.Context, contactID uint64) (*float64, error) {
	if m.CreditLimitFunc != nil {
		return m.CreditLimitFunc(ctx, contactID)
	}
	return nil, nil
}

type TaxRuleMapperMock struct {
	FindFunc    func(ctx context.Context, id uint64) (*TaxRule, error)
	ResolveFunc func(ctx context.Context, positionID uint64, srcTaxID *uint64, srcAccountID uint64) (*ResolveResult, error)
}

func (m TaxRuleMapperMock) Find(ctx context.Context, id uint64) (*TaxRule, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m TaxRuleMapperMock) Resolve(ctx context.Context, positionID uint64, srcTaxID *uint64, srcAccountID uint64) (*ResolveResult, error) {
	if m.ResolveFunc != nil {
		return m.ResolveFunc(ctx, positionID, srcTaxID, srcAccountID)
	}
	account := srcAccountID
	return &ResolveResult{Account: &account, TaxAccount: srcTaxID}, nil
}
