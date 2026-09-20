package pos

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type TransactionerMock struct {
	RunFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error
}

func (m TransactionerMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, fn)
	}
	return fn(nil)
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

type PaymentAccountLookupMock struct {
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.POSPaymentAccount], error)
}

func (m PaymentAccountLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.POSPaymentAccount], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.POSPaymentAccount]{Items: []*reference.POSPaymentAccount{}}, nil
}

type ShipEngineMock struct {
	ShipFunc    func(ctx context.Context, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error)
	ShipTxFunc  func(ctx context.Context, tx *gorm.DB, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error)
	RestockFunc func(ctx context.Context, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error)
}

func (m ShipEngineMock) Ship(ctx context.Context, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error) {
	if m.ShipFunc != nil {
		return m.ShipFunc(ctx, moveID, journalID, date)
	}
	return &inventory.CostLayer{MovementID: &moveID}, nil
}

func (m ShipEngineMock) ShipTx(ctx context.Context, tx *gorm.DB, moveID uint64, journalID uint64, date time.Time) (*inventory.CostLayer, error) {
	if m.ShipTxFunc != nil {
		return m.ShipTxFunc(ctx, tx, moveID, journalID, date)
	}
	return m.Ship(ctx, moveID, journalID, date)
}

func (m ShipEngineMock) Restock(ctx context.Context, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error) {
	if m.RestockFunc != nil {
		return m.RestockFunc(ctx, moveID, unitCost, journalID, date)
	}
	return &inventory.CostLayer{MovementID: &moveID}, nil
}

func (m ShipEngineMock) RestockTx(ctx context.Context, tx *gorm.DB, moveID uint64, unitCost amount.Amount, journalID uint64, date time.Time) (*inventory.CostLayer, error) {
	return m.Restock(ctx, moveID, unitCost, journalID, date)
}

type InvoiceEngineMock struct {
	CreateFunc           func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
	CreateCreditNoteFunc func(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error)
}

func (m InvoiceEngineMock) Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, request)
	}
	return &accounting.Invoice{Base: model.Base{ID: 1}, ContactID: request.ContactID}, nil
}

func (m InvoiceEngineMock) CreateTx(ctx context.Context, tx *gorm.DB, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	return m.Create(ctx, request)
}

func (m InvoiceEngineMock) CreateCreditNote(ctx context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
	if m.CreateCreditNoteFunc != nil {
		return m.CreateCreditNoteFunc(ctx, request)
	}
	return &accounting.Invoice{Base: model.Base{ID: 2}, ContactID: request.OriginalInvoiceID}, nil
}

func (m InvoiceEngineMock) CreateCreditNoteTx(ctx context.Context, tx *gorm.DB, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
	return m.CreateCreditNote(ctx, request)
}

type POSSessionDAOMock struct {
	dao.CRUDMock[POSSession]
	ListInOrganizationFunc func(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[POSSession], error)
}

func (m POSSessionDAOMock) ListInOrganization(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[POSSession], error) {
	if m.ListInOrganizationFunc != nil {
		return m.ListInOrganizationFunc(ctx, q, organizationID)
	}
	return m.List(ctx, q)
}

type POSOrderDAOMock struct {
	dao.CRUDMock[POSOrder]
	CreateWithLinesAndPaymentsFunc   func(ctx context.Context, order *POSOrder, lines []*POSOrderLine, payments []*POSPayment) (*POSOrder, error)
	CreateWithLinesAndPaymentsTxFunc func(ctx context.Context, tx *gorm.DB, order *POSOrder, lines []*POSOrderLine, payments []*POSPayment) (*POSOrder, error)
	ListInOrganizationFunc           func(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[POSOrder], error)
}

func (m POSOrderDAOMock) CreateWithLinesAndPayments(ctx context.Context, order *POSOrder, lines []*POSOrderLine, payments []*POSPayment) (*POSOrder, error) {
	if m.CreateWithLinesAndPaymentsFunc != nil {
		return m.CreateWithLinesAndPaymentsFunc(ctx, order, lines, payments)
	}
	return order, nil
}

func (m POSOrderDAOMock) CreateWithLinesAndPaymentsTx(ctx context.Context, tx *gorm.DB, order *POSOrder, lines []*POSOrderLine, payments []*POSPayment) (*POSOrder, error) {
	if m.CreateWithLinesAndPaymentsTxFunc != nil {
		return m.CreateWithLinesAndPaymentsTxFunc(ctx, tx, order, lines, payments)
	}
	return m.CreateWithLinesAndPayments(ctx, order, lines, payments)
}

func (m POSOrderDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, order *POSOrder) (*POSOrder, error) {
	return m.Update(ctx, order)
}

func (m POSOrderDAOMock) ListInOrganization(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[POSOrder], error) {
	if m.ListInOrganizationFunc != nil {
		return m.ListInOrganizationFunc(ctx, q, organizationID)
	}
	return m.List(ctx, q)
}

type POSOrderLineDAOMock struct {
	dao.CRUDMock[POSOrderLine]
	ListByOrderFunc func(ctx context.Context, orderID uint64) ([]*POSOrderLine, error)
}

func (m POSOrderLineDAOMock) ListByOrder(ctx context.Context, orderID uint64) ([]*POSOrderLine, error) {
	if m.ListByOrderFunc != nil {
		return m.ListByOrderFunc(ctx, orderID)
	}
	return []*POSOrderLine{}, nil
}

type POSPaymentDAOMock struct {
	dao.CRUDMock[POSPayment]
	ListByOrderFunc  func(ctx context.Context, orderID uint64) ([]*POSPayment, error)
	SumBySessionFunc func(ctx context.Context, sessionID uint64) ([]PaymentMethodTotal, error)
}

func (m POSPaymentDAOMock) ListByOrder(ctx context.Context, orderID uint64) ([]*POSPayment, error) {
	if m.ListByOrderFunc != nil {
		return m.ListByOrderFunc(ctx, orderID)
	}
	return []*POSPayment{}, nil
}

func (m POSPaymentDAOMock) SumBySession(ctx context.Context, sessionID uint64) ([]PaymentMethodTotal, error) {
	if m.SumBySessionFunc != nil {
		return m.SumBySessionFunc(ctx, sessionID)
	}
	return []PaymentMethodTotal{}, nil
}
