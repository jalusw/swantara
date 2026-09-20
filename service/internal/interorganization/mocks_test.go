package interorganization_test

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func baseID(id uint64) model.Base { return model.Base{ID: id} }

type purchaseOrderCreatorMock struct {
	CreateFunc func(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error)
}

func (m purchaseOrderCreatorMock) Create(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order, lines)
	}
	return order, nil
}

type posterMock struct {
	PostTxFunc func(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
}

func (m posterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

func (m posterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m posterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

func (m posterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

type txMock struct {
	RunFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error
}

func (m txMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, fn)
	}
	return fn(nil)
}

type rateSourceMock struct {
	RateFunc func(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error)
}

func (m rateSourceMock) Rate(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error) {
	if m.RateFunc != nil {
		return m.RateFunc(ctx, currencyCode, organizationID, rateType, date)
	}
	return amount.Zero(), amount.ErrInvalidRate
}

type organizationLookupMock struct {
	FindFunc func(ctx context.Context, id uint64) (*reference.Organization, error)
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.Organization], error)
}

func (m organizationLookupMock) Find(ctx context.Context, id uint64) (*reference.Organization, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m organizationLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
}

type accountLookupMock struct {
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error)
}

func (m accountLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
}

type taxPeriodDAOMock struct {
	dao.CRUDMock[accounting.TaxPeriod]
	FindByDateFunc         func(ctx context.Context, organizationID uint64, date time.Time) (*accounting.TaxPeriod, error)
	ListByOrganizationFunc func(ctx context.Context, organizationID uint64) ([]*accounting.TaxPeriod, error)
	UpdateTxFunc           func(ctx context.Context, tx *gorm.DB, period *accounting.TaxPeriod) (*accounting.TaxPeriod, error)
}

func (m taxPeriodDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, period *accounting.TaxPeriod) (*accounting.TaxPeriod, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, period)
	}
	return period, nil
}

func (m taxPeriodDAOMock) FindByDate(ctx context.Context, organizationID uint64, date time.Time) (*accounting.TaxPeriod, error) {
	if m.FindByDateFunc != nil {
		return m.FindByDateFunc(ctx, organizationID, date)
	}
	return nil, nil
}

func (m taxPeriodDAOMock) ListByOrganization(ctx context.Context, organizationID uint64) ([]*accounting.TaxPeriod, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, organizationID)
	}
	return nil, nil
}
