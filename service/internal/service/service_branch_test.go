package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestServiceService_CreateOrder_PropagatesOrderCreateError(t *testing.T) {
	ctx := context.Background()

	orders := ServiceOrderDAOMock{
		CreateFunc: func(_ context.Context, _ *ServiceOrder) (*ServiceOrder, error) {
			return nil, errors.New("insert failed")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})

	_, err := svc.CreateOrder(ctx, CreateOrderRequest{Name: "X", Type: OrderTypeRepair, Lines: []LineRequest{{Type: LineTypePart, Qty: 1}}})
	if err == nil {
		t.Error("expected order create error to propagate")
	}
}

func TestServiceService_Complete_SkipsPostingForZeroCostPart(t *testing.T) {
	ctx := context.Background()

	var posted int
	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: OrderStateInProgress}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, order *ServiceOrder) (*ServiceOrder, error) {
			return order, nil
		},
	}
	lines := ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return []*ServiceOrderLine{{Type: LineTypePart, Qty: 0, UnitCost: 25}}, nil
		},
	}
	poster := ServicePosterMock{
		PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted++
			return &accounting.JournalEntry{}, nil
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, lines, poster, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})

	order, err := svc.Complete(ctx, CompleteRequest{OrderID: 1, JournalID: 1, Date: time.Now(), COGSAccountID: 2, StockValuationAccountID: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.State != OrderStateDone {
		t.Errorf("order = %+v, want done", order)
	}
	if posted != 0 {
		t.Errorf("postings = %d, want 0 for zero-cost part line", posted)
	}
}

func TestServiceService_CreateOrder_PropagatesLineAndTxErrors(t *testing.T) {
	ctx := context.Background()

	orders := ServiceOrderDAOMock{
		CreateFunc: func(_ context.Context, order *ServiceOrder) (*ServiceOrder, error) {
			order.ID = 1
			return order, nil
		},
	}
	lines := ServiceOrderLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ServiceOrderLine) (*ServiceOrderLine, error) {
			return nil, errors.New("insert failed")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, lines, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})

	_, err := svc.CreateOrder(ctx, CreateOrderRequest{Name: "X", Type: OrderTypeRepair, Lines: []LineRequest{{Type: LineTypePart, Qty: 1}}})
	if err == nil {
		t.Error("expected line creation error to propagate")
	}
}

func TestServiceService_AddLine_BranchValidation(t *testing.T) {
	ctx := context.Background()

	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateNew}, nil
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})

	_, err := svc.AddLine(ctx, AddLineRequest{OrderID: 1, Line: LineRequest{Type: "bogus", Qty: 1}})
	if !errors.Is(err, ErrOrderLineInvalidType) {
		t.Errorf("bad line type err = %v, want ErrOrderLineInvalidType", err)
	}
	_, err = svc.AddLine(ctx, AddLineRequest{OrderID: 1, Line: LineRequest{Type: LineTypePart, Qty: 0}})
	if !errors.Is(err, ErrOrderLineInvalidQty) {
		t.Errorf("zero qty err = %v, want ErrOrderLineInvalidQty", err)
	}
	_, err = svc.AddLine(ctx, AddLineRequest{OrderID: 1, Line: LineRequest{Type: LineTypePart, Qty: 1, UnitCost: -1}})
	if !errors.Is(err, ErrOrderLineInvalidPrice) {
		t.Errorf("negative price err = %v, want ErrOrderLineInvalidPrice", err)
	}
}

func TestServiceService_AddLine_PropagatesFindAndTxErrors(t *testing.T) {
	ctx := context.Background()

	findErr := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, findErr, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.AddLine(ctx, AddLineRequest{OrderID: 1, Line: LineRequest{Type: LineTypePart, Qty: 1}}); err == nil {
		t.Error("expected order find error to propagate")
	}

	lines := ServiceOrderLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ServiceOrderLine) (*ServiceOrderLine, error) {
			return nil, errors.New("insert failed")
		},
	}
	svc = NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateNew}, nil
		},
	}, lines, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.AddLine(ctx, AddLineRequest{OrderID: 1, Line: LineRequest{Type: LineTypePart, Qty: 1}}); err == nil {
		t.Error("expected line creation error to propagate")
	}
}

func TestServiceService_OrderLookup_PropagatesFindAndNotFound(t *testing.T) {
	ctx := context.Background()

	findErr := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, findErr, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Start(ctx, 1); err == nil {
		t.Error("expected order find error to propagate")
	}

	missing := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return nil, nil
		},
	}
	svc = NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, missing, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Start(ctx, 1); !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("err = %v, want ErrOrderNotFound", err)
	}
}

func TestServiceService_CancelOrder_PropagatesFindError(t *testing.T) {
	ctx := context.Background()

	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.CancelOrder(ctx, 1); err == nil {
		t.Error("expected order find error to propagate")
	}
}

func TestServiceService_Transition_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()

	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateNew}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *ServiceOrder) (*ServiceOrder, error) {
			return nil, errors.New("update failed")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.CancelOrder(ctx, 1); err == nil {
		t.Error("expected update error to propagate")
	}
}

func TestServiceService_Schedule_PropagatesFindAndTransitionErrors(t *testing.T) {
	ctx := context.Background()

	findErr := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, findErr, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Schedule(ctx, ScheduleRequest{OrderID: 1}); err == nil {
		t.Error("expected order find error to propagate")
	}

	done := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateDone}, nil
		},
	}
	svc = NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, done, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Schedule(ctx, ScheduleRequest{OrderID: 1}); !errors.Is(err, ErrOrderInvalidState) {
		t.Errorf("err = %v, want ErrOrderInvalidState", err)
	}
}

func TestServiceService_Start_PropagatesFindError(t *testing.T) {
	ctx := context.Background()

	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Start(ctx, 1); err == nil {
		t.Error("expected order find error to propagate")
	}
}

func TestServiceService_Complete_PropagatesFindAndLineErrors(t *testing.T) {
	ctx := context.Background()

	findErr := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, findErr, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Complete(ctx, CompleteRequest{OrderID: 1, JournalID: 1, COGSAccountID: 2, StockValuationAccountID: 3}); err == nil {
		t.Error("expected order find error to propagate")
	}

	linesErr := ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return nil, errors.New("lines failed")
		},
	}
	svc = NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateInProgress}, nil
		},
	}, linesErr, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Complete(ctx, CompleteRequest{OrderID: 1, JournalID: 1, COGSAccountID: 2, StockValuationAccountID: 3}); err == nil {
		t.Error("expected line list error to propagate")
	}
}

func TestServiceService_Complete_PersistsResolution(t *testing.T) {
	ctx := context.Background()

	var updated *ServiceOrder
	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: OrderStateInProgress}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, order *ServiceOrder) (*ServiceOrder, error) {
			updated = order
			return order, nil
		},
	}
	lines := ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return []*ServiceOrderLine{{Type: LineTypeLabor, Qty: 1, UnitCost: 0}}, nil
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, lines, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})

	order, err := svc.Complete(ctx, CompleteRequest{
		OrderID: 1, JournalID: 1, Date: time.Now(),
		COGSAccountID: 2, StockValuationAccountID: 3, Resolution: "replaced motor",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Resolution != "replaced motor" || updated == nil || updated.Resolution != "replaced motor" {
		t.Errorf("order = %+v, updated = %v, want resolution persisted", order, updated)
	}
}

func TestServiceService_Complete_PropagatesPosterAndUpdateErrors(t *testing.T) {
	ctx := context.Background()

	baseOrder := func() ServiceOrderDAOMock {
		return ServiceOrderDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
				return &ServiceOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: OrderStateInProgress}, nil
			},
		}
	}
	lines := ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return []*ServiceOrderLine{{Type: LineTypePart, Qty: 2, UnitCost: 25}}, nil
		},
	}

	posterErr := ServicePosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return nil, errors.New("post failed")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, baseOrder(), lines, posterErr, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Complete(ctx, CompleteRequest{OrderID: 1, JournalID: 1, Date: time.Now(), COGSAccountID: 2, StockValuationAccountID: 3}); err == nil {
		t.Error("expected posting error to propagate")
	}

	orders := baseOrder()
	orders.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, _ *ServiceOrder) (*ServiceOrder, error) {
		return nil, errors.New("update failed")
	}
	posterOk := ServicePosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{}, nil
		},
	}
	svc = NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, lines, posterOk, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Complete(ctx, CompleteRequest{OrderID: 1, JournalID: 1, Date: time.Now(), COGSAccountID: 2, StockValuationAccountID: 3}); err == nil {
		t.Error("expected update error to propagate")
	}
}

func TestServiceService_Bill_PropagatesFindLineAndInvoiceErrors(t *testing.T) {
	ctx := context.Background()

	findErr := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, findErr, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Bill(ctx, BillRequest{OrderID: 1, JournalID: 1, RevenueAccountID: 2}); err == nil {
		t.Error("expected order find error to propagate")
	}

	linesErr := ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return nil, errors.New("lines failed")
		},
	}
	svc = NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: OrderStateDone, ContactID: helper.Ptr(uint64(5))}, nil
		},
	}, linesErr, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.Bill(ctx, BillRequest{OrderID: 1, JournalID: 1, RevenueAccountID: 2}); err == nil {
		t.Error("expected line list error to propagate")
	}

	invoiceErr := ServiceInvoiceBuilderMock{
		CreateFunc: func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
			return nil, errors.New("invoice failed")
		},
	}
	svc = NewTestServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: OrderStateDone, ContactID: helper.Ptr(uint64(5))}, nil
		},
	}, ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return []*ServiceOrderLine{{Type: LineTypePart, Qty: 1, UnitPrice: 80, Billable: true}}, nil
		},
	}, ServicePosterMock{}, invoiceErr, ServiceTransactionerMock{})
	if _, err := svc.Bill(ctx, BillRequest{OrderID: 1, JournalID: 1, RevenueAccountID: 2}); err == nil {
		t.Error("expected invoice creation error to propagate")
	}
}

func TestServiceService_CancelContract_PropagatesFindAndNotFound(t *testing.T) {
	ctx := context.Background()

	findErr := ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceContract, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, findErr, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.CancelContract(ctx, 1); err == nil {
		t.Error("expected contract find error to propagate")
	}

	missing := ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceContract, error) {
			return nil, nil
		},
	}
	svc = NewTestServiceService(EquipmentDAOMock{}, missing, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.CancelContract(ctx, 1); !errors.Is(err, ErrContractNotFound) {
		t.Errorf("err = %v, want ErrContractNotFound", err)
	}
}

func TestServiceService_ActivateContract_PropagatesFindAndNotFound(t *testing.T) {
	ctx := context.Background()

	findErr := ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceContract, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewTestServiceService(EquipmentDAOMock{}, findErr, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.ActivateContract(ctx, 1); err == nil {
		t.Error("expected contract find error to propagate")
	}

	missing := ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceContract, error) {
			return nil, nil
		},
	}
	svc = NewTestServiceService(EquipmentDAOMock{}, missing, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, ServicePosterMock{}, ServiceInvoiceBuilderMock{}, ServiceTransactionerMock{})
	if _, err := svc.ActivateContract(ctx, 1); !errors.Is(err, ErrContractNotFound) {
		t.Errorf("err = %v, want ErrContractNotFound", err)
	}
}
