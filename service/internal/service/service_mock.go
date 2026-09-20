package service

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"gorm.io/gorm"
)

type ServicePosterMock struct {
	PostFunc      func(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
	PostTxFunc    func(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
	ReverseFunc   func(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
	ReverseTxFunc func(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
}

func (m ServicePosterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostFunc != nil {
		return m.PostFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m ServicePosterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return m.Post(ctx, request)
}

func (m ServicePosterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseFunc != nil {
		return m.ReverseFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m ServicePosterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseTxFunc != nil {
		return m.ReverseTxFunc(ctx, tx, request)
	}
	return m.Reverse(ctx, request)
}

type ServiceInvoiceBuilderMock struct {
	CreateFunc func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

func (m ServiceInvoiceBuilderMock) Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, request)
	}
	return &accounting.Invoice{}, nil
}

type ServiceTransactionerMock struct {
	RunFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error
}

func (m ServiceTransactionerMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, fn)
	}
	return fn(nil)
}

func NewTestServiceService(
	equipments EquipmentDAO,
	contracts ServiceContractDAO,
	orders ServiceOrderDAO,
	lines ServiceOrderLineDAO,
	poster accounting.Poster,
	invoices InvoiceBuilder,
	tx db.Transactioner,
) ServiceService {
	return NewServiceService(equipments, contracts, orders, lines, poster, invoices, tx)
}

func NewTestMaintenanceService(
	plans MaintenancePlanDAO,
	equipments EquipmentDAO,
	orders ServiceOrderDAO,
	tx db.Transactioner,
) MaintenanceService {
	return NewMaintenanceService(plans, equipments, orders, tx)
}
