package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestServiceService_Complete_PostsCOGSForPartLines(t *testing.T) {
	var posted accounting.PostRequest
	svc := NewServiceService(
		EquipmentDAOMock{},
		ServiceContractDAOMock{},
		ServiceOrderDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
				return &ServiceOrder{Base: model.Base{ID: 8}, OrganizationID: helper.Ptr(uint64(10)), State: OrderStateInProgress, Name: "Fix printer"}, nil
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, order *ServiceOrder) (*ServiceOrder, error) { return order, nil },
		},
		ServiceOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return []*ServiceOrderLine{
				{Type: LineTypePart, Qty: 2, UnitCost: 25},
				{Type: LineTypeLabor, Qty: 1, UnitCost: 0},
			}, nil
		}},
		posterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = request
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		}},
		invoiceBuilderMock{},
		txMock{},
	)

	order, err := svc.Complete(context.Background(), CompleteRequest{
		OrderID: 8, JournalID: 3, Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		COGSAccountID: 200, StockValuationAccountID: 300,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.State != OrderStateDone {
		t.Errorf("order = %+v, want done", order)
	}
	if !posted.Lines[0].Debit.Equal(amount.FromFloat64(50)) || posted.Lines[0].AccountID != 200 {
		t.Errorf("cogs line = %+v, want Dr 50 on 200", posted.Lines[0])
	}
	if !posted.Lines[1].Credit.Equal(amount.FromFloat64(50)) || posted.Lines[1].AccountID != 300 {
		t.Errorf("inventory line = %+v, want Cr 50 on 300", posted.Lines[1])
	}
}

func TestServiceService_Bill_OnlyBillableLinesFlowToInvoice(t *testing.T) {
	var request accounting.CreateInvoiceRequest
	svc := NewServiceService(
		EquipmentDAOMock{},
		ServiceContractDAOMock{},
		ServiceOrderDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
				return &ServiceOrder{Base: model.Base{ID: 8}, OrganizationID: helper.Ptr(uint64(10)), State: OrderStateDone, ContactID: helper.Ptr(uint64(77)), Name: "Fix printer"}, nil
			},
			UpdateFunc: func(_ context.Context, order *ServiceOrder) (*ServiceOrder, error) { return order, nil },
		},
		ServiceOrderLineDAOMock{ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return []*ServiceOrderLine{
				{Type: LineTypePart, ItemID: helper.Ptr(uint64(11)), Qty: 1, UnitPrice: 80, Billable: true},
				{Type: LineTypeLabor, Description: "Repair labor", Qty: 2, UnitPrice: 50, Billable: true},
				{Type: LineTypePart, Qty: 1, UnitPrice: 10, Billable: false},
			}, nil
		}},
		posterMock{},
		invoiceBuilderMock{CreateFunc: func(_ context.Context, r accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
			request = r
			return &accounting.Invoice{Base: model.Base{ID: 500}}, nil
		}},
		txMock{},
	)

	order, err := svc.Bill(context.Background(), BillRequest{OrderID: 8, JournalID: 3, RevenueAccountID: 400})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.State != OrderStateInvoiced {
		t.Errorf("order = %+v, want invoiced", order)
	}
	if order.InvoiceID == nil || *order.InvoiceID != 500 {
		t.Errorf("invoice_id = %v, want 500", order.InvoiceID)
	}
	if len(request.Lines) != 2 {
		t.Fatalf("invoice lines = %d, want 2 (warranty/non-billable excluded)", len(request.Lines))
	}
	if request.Lines[0].Qty != 1 || request.Lines[0].UnitPrice != 80 {
		t.Errorf("line[0] = %+v, want qty 1 price 80", request.Lines[0])
	}
	if request.Lines[1].Qty != 2 || request.Lines[1].UnitPrice != 50 {
		t.Errorf("line[1] = %+v, want qty 2 price 50", request.Lines[1])
	}
}

func TestServiceService_Bill_RequiresDoneOrderWithBillableLines(t *testing.T) {
	svc := NewServiceService(
		EquipmentDAOMock{},
		ServiceContractDAOMock{},
		ServiceOrderDAOMock{FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 8}, State: OrderStateNew, ContactID: helper.Ptr(uint64(77))}, nil
		}},
		ServiceOrderLineDAOMock{},
		posterMock{},
		invoiceBuilderMock{},
		txMock{},
	)

	_, err := svc.Bill(context.Background(), BillRequest{OrderID: 8, JournalID: 3, RevenueAccountID: 400})
	if !errors.Is(err, ErrOrderNotDone) {
		t.Fatalf("err = %v, want ErrOrderNotDone", err)
	}
}

func TestServiceService_CreateOrder_RejectsInvalidTypeAndLines(t *testing.T) {
	svc := NewServiceService(
		EquipmentDAOMock{},
		ServiceContractDAOMock{},
		ServiceOrderDAOMock{},
		ServiceOrderLineDAOMock{},
		posterMock{},
		invoiceBuilderMock{},
		txMock{},
	)

	_, err := svc.CreateOrder(context.Background(), CreateOrderRequest{Name: "X", Type: "bogus", Lines: []LineRequest{{Type: LineTypeLabor, Qty: 1}}})
	if !errors.Is(err, ErrOrderInvalidType) {
		t.Fatalf("err = %v, want ErrOrderInvalidType", err)
	}

	_, err = svc.CreateOrder(context.Background(), CreateOrderRequest{Name: "X", Type: OrderTypeRepair, Lines: []LineRequest{{Type: LineTypePart, Qty: -1}}})
	if !errors.Is(err, ErrOrderLineInvalidQty) {
		t.Fatalf("err = %v, want ErrOrderLineInvalidQty", err)
	}
}

func TestServiceService_Contract_StateTransitions(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	svc := NewServiceService(
		EquipmentDAOMock{},
		ServiceContractDAOMock{
			CreateFunc: func(_ context.Context, contract *ServiceContract) (*ServiceContract, error) {
				contract.ID = 3
				return contract, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*ServiceContract, error) {
				return &ServiceContract{Base: model.Base{ID: 3}, State: ContractStateActive}, nil
			},
			UpdateFunc: func(_ context.Context, contract *ServiceContract) (*ServiceContract, error) { return contract, nil },
		},
		ServiceOrderDAOMock{},
		ServiceOrderLineDAOMock{},
		posterMock{},
		invoiceBuilderMock{},
		txMock{},
	)

	contract, err := svc.CreateContract(context.Background(), CreateContractRequest{
		OrganizationID: 10, Name: "Gold SLA", DateStart: &start, DateEnd: &end,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contract.State != ContractStateActive {
		t.Errorf("contract = %+v, want active when dated", contract)
	}

	cancelled, err := svc.CancelContract(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancelled.State != ContractStateCancelled {
		t.Errorf("contract = %+v, want cancelled", cancelled)
	}
}

type posterMock struct {
	PostFunc      func(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
	PostTxFunc    func(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
	ReverseFunc   func(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
	ReverseTxFunc func(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
}

func (m posterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostFunc != nil {
		return m.PostFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m posterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return m.Post(ctx, request)
}

func (m posterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseFunc != nil {
		return m.ReverseFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m posterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseTxFunc != nil {
		return m.ReverseTxFunc(ctx, tx, request)
	}
	return m.Reverse(ctx, request)
}

type invoiceBuilderMock struct {
	CreateFunc func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

func (m invoiceBuilderMock) Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, request)
	}
	return &accounting.Invoice{}, nil
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
