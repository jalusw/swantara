package returns

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type OriginOrderLookupMock struct {
	CustomerOrderFunc func(ctx context.Context, orderID uint64) (uint64, error)
	VendorOrderFunc   func(ctx context.Context, orderID uint64) (uint64, error)
}

func (m OriginOrderLookupMock) CustomerOrder(ctx context.Context, orderID uint64) (uint64, error) {
	if m.CustomerOrderFunc != nil {
		return m.CustomerOrderFunc(ctx, orderID)
	}
	return 0, nil
}

func (m OriginOrderLookupMock) VendorOrder(ctx context.Context, orderID uint64) (uint64, error) {
	if m.VendorOrderFunc != nil {
		return m.VendorOrderFunc(ctx, orderID)
	}
	return 0, nil
}

type StockLocationLookupMock struct {
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.StockLocation], error)
}

func (m StockLocationLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{}}, nil
}

type StockLayerLookupMock struct {
	ListByMovementFunc func(ctx context.Context, moveID uint64) ([]*inventory.CostLayer, error)
}

func (m StockLayerLookupMock) ListByMovement(ctx context.Context, moveID uint64) ([]*inventory.CostLayer, error) {
	if m.ListByMovementFunc != nil {
		return m.ListByMovementFunc(ctx, moveID)
	}
	return []*inventory.CostLayer{}, nil
}

type ReturnValuerMock struct {
	RestockFunc          func(ctx context.Context, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error)
	ReturnToSupplierFunc func(ctx context.Context, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error)
}

func (m ReturnValuerMock) Restock(ctx context.Context, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error) {
	if m.RestockFunc != nil {
		return m.RestockFunc(ctx, moveID, unitCost, journalID, date)
	}
	return &inventory.CostLayer{MovementID: &moveID}, nil
}

func (m ReturnValuerMock) ReturnToSupplier(ctx context.Context, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error) {
	if m.ReturnToSupplierFunc != nil {
		return m.ReturnToSupplierFunc(ctx, moveID, journalID, date)
	}
	return &inventory.CostLayer{MovementID: &moveID}, nil
}

type CreditNoteEngineMock struct {
	CreateCreditNoteFunc       func(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error)
	CreateVendorCreditNoteFunc func(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error)
}

func (m CreditNoteEngineMock) CreateCreditNote(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
	if m.CreateCreditNoteFunc != nil {
		return m.CreateCreditNoteFunc(ctx, request)
	}
	return &accounting.Invoice{Base: model.Base{ID: 1}}, nil
}

func (m CreditNoteEngineMock) CreateVendorCreditNote(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
	if m.CreateVendorCreditNoteFunc != nil {
		return m.CreateVendorCreditNoteFunc(ctx, request)
	}
	return &accounting.Invoice{Base: model.Base{ID: 1}}, nil
}

type SaleOrderReturnRecorderMock struct {
	RecordReturnFunc func(ctx context.Context, orderID uint64, itemID uint64, qty float64) error
}

func (m SaleOrderReturnRecorderMock) RecordReturn(ctx context.Context, orderID uint64, itemID uint64, qty float64) error {
	if m.RecordReturnFunc != nil {
		return m.RecordReturnFunc(ctx, orderID, itemID, qty)
	}
	return nil
}

type PurchaseOrderReturnRecorderMock struct {
	RecordReturnFunc func(ctx context.Context, orderID uint64, itemID uint64, qty float64) error
}

func (m PurchaseOrderReturnRecorderMock) RecordReturn(ctx context.Context, orderID uint64, itemID uint64, qty float64) error {
	if m.RecordReturnFunc != nil {
		return m.RecordReturnFunc(ctx, orderID, itemID, qty)
	}
	return nil
}

func NewTestRMAService(
	rmas RMADAOMock,
	lines RMALineDAOMock,
	orders OriginOrderLookupMock,
	movements inventory.StockMovementDAOMock,
	locations StockLocationLookupMock,
	layers StockLayerLookupMock,
	valuer ReturnValuerMock,
	invoices accounting.InvoiceDAOMock,
	credits CreditNoteEngineMock,
	sequences sequence.DAOMock,
) RMAService {
	return NewRMAService(
		rmas,
		lines,
		orders,
		movements,
		locations,
		layers,
		valuer,
		invoices,
		credits,
		SaleOrderReturnRecorderMock{},
		PurchaseOrderReturnRecorderMock{},
		sequence.NewSequenceService(sequences),
		inventory.TransactionerMock{},
	)
}
