package subscription

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

type SubscriptionDAOMock struct {
	dao.CRUDMock[Subscription]
	ListDueFunc  func(ctx context.Context, asOf time.Time) ([]*Subscription, error)
	UpdateTxFunc func(ctx context.Context, tx *gorm.DB, subscription *Subscription) (*Subscription, error)
}

func (m SubscriptionDAOMock) ListDue(ctx context.Context, asOf time.Time) ([]*Subscription, error) {
	if m.ListDueFunc != nil {
		return m.ListDueFunc(ctx, asOf)
	}
	return []*Subscription{}, nil
}

func (m SubscriptionDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, subscription *Subscription) (*Subscription, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, subscription)
	}
	return subscription, nil
}

type SubscriptionLineDAOMock struct {
	dao.CRUDMock[SubscriptionLine]
	ListBySubscriptionFunc func(ctx context.Context, subscriptionID uint64) ([]*SubscriptionLine, error)
}

func (m SubscriptionLineDAOMock) ListBySubscription(ctx context.Context, subscriptionID uint64) ([]*SubscriptionLine, error) {
	if m.ListBySubscriptionFunc != nil {
		return m.ListBySubscriptionFunc(ctx, subscriptionID)
	}
	return []*SubscriptionLine{}, nil
}

type InvoiceEngineMock struct {
	CreateFunc func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

func (m InvoiceEngineMock) Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, request)
	}
	return &accounting.Invoice{Base: model.Base{ID: 1}, ContactID: request.ContactID, AmountTotal: amount.Zero()}, nil
}

func (m InvoiceEngineMock) CreateTx(ctx context.Context, tx *gorm.DB, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	return m.Create(ctx, request)
}

type PosterMock struct {
	PostTxFunc func(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
}

func (m PosterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	return m.PostTx(ctx, nil, request)
}

func (m PosterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return &accounting.JournalEntry{Base: model.Base{ID: 5}}, nil
}

func (m PosterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{Base: model.Base{ID: 6}}, nil
}

func (m PosterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{Base: model.Base{ID: 6}}, nil
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

type IncomeAccountResolverMock struct {
	ResolveIncomeAccountFunc       func(ctx context.Context, variantID uint64) (uint64, error)
	ResolveVariantOrganizationFunc func(ctx context.Context, variantID uint64) (*uint64, error)
}

func (m IncomeAccountResolverMock) ResolveIncomeAccount(ctx context.Context, variantID uint64) (uint64, error) {
	if m.ResolveIncomeAccountFunc != nil {
		return m.ResolveIncomeAccountFunc(ctx, variantID)
	}
	return 200, nil
}

func (m IncomeAccountResolverMock) ResolveVariantOrganization(ctx context.Context, variantID uint64) (*uint64, error) {
	if m.ResolveVariantOrganizationFunc != nil {
		return m.ResolveVariantOrganizationFunc(ctx, variantID)
	}
	return nil, nil
}

type SubscriptionConfigSourceMock struct {
	JournalIDFunc                func(ctx context.Context, organizationID uint64) (uint64, error)
	DeferredRevenueAccountIDFunc func(ctx context.Context, organizationID uint64) (uint64, error)
}

func (m SubscriptionConfigSourceMock) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.JournalIDFunc != nil {
		return m.JournalIDFunc(ctx, organizationID)
	}
	return 10, nil
}

func (m SubscriptionConfigSourceMock) DeferredRevenueAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.DeferredRevenueAccountIDFunc != nil {
		return m.DeferredRevenueAccountIDFunc(ctx, organizationID)
	}
	return 300, nil
}
