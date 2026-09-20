package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestServiceService_CreateOrder_Success(t *testing.T) {
	ctx := context.Background()

	var createdLines int
	orders := ServiceOrderDAOMock{
		CreateFunc: func(_ context.Context, order *ServiceOrder) (*ServiceOrder, error) {
			order.ID = 1
			return order, nil
		},
	}
	lines := ServiceOrderLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, line *ServiceOrderLine) (*ServiceOrderLine, error) {
			createdLines++
			return line, nil
		},
	}
	svc := NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, lines, posterMock{}, invoiceBuilderMock{}, txMock{})

	order, err := svc.CreateOrder(ctx, CreateOrderRequest{
		OrganizationID: 10,
		Name:           "Repair",
		Type:           OrderTypeRepair,
		Lines: []LineRequest{
			{Type: LineTypePart, Qty: 1, UnitCost: 10, UnitPrice: 20},
			{Type: LineTypeLabor, Qty: 2, UnitPrice: 50},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID != 1 || order.State != OrderStateNew || order.OrganizationID == nil || *order.OrganizationID != 10 {
		t.Errorf("order = %+v", order)
	}
	if createdLines != 2 {
		t.Errorf("createdLines = %d, want 2", createdLines)
	}
}

func TestServiceService_CreateOrder_Validation(t *testing.T) {
	svc := NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})

	if _, err := svc.CreateOrder(context.Background(), CreateOrderRequest{Name: "", Type: OrderTypeRepair}); !errors.Is(err, ErrOrderNameRequired) {
		t.Errorf("empty name err = %v, want ErrOrderNameRequired", err)
	}
	if _, err := svc.CreateOrder(context.Background(), CreateOrderRequest{Name: "X", Type: "bogus"}); !errors.Is(err, ErrOrderInvalidType) {
		t.Errorf("bad type err = %v, want ErrOrderInvalidType", err)
	}
	if _, err := svc.CreateOrder(context.Background(), CreateOrderRequest{Name: "X", Type: OrderTypeRepair}); !errors.Is(err, ErrOrderEmptyLines) {
		t.Errorf("no lines err = %v, want ErrOrderEmptyLines", err)
	}
	badLine := CreateOrderRequest{Name: "X", Type: OrderTypeRepair, Lines: []LineRequest{{Type: "bogus", Qty: 1}}}
	if _, err := svc.CreateOrder(context.Background(), badLine); !errors.Is(err, ErrOrderLineInvalidType) {
		t.Errorf("bad line type err = %v, want ErrOrderLineInvalidType", err)
	}
	badQty := CreateOrderRequest{Name: "X", Type: OrderTypeRepair, Lines: []LineRequest{{Type: LineTypePart, Qty: 0}}}
	if _, err := svc.CreateOrder(context.Background(), badQty); !errors.Is(err, ErrOrderLineInvalidQty) {
		t.Errorf("zero qty err = %v, want ErrOrderLineInvalidQty", err)
	}
	badPrice := CreateOrderRequest{Name: "X", Type: OrderTypeRepair, Lines: []LineRequest{{Type: LineTypePart, Qty: 1, UnitCost: -1}}}
	if _, err := svc.CreateOrder(context.Background(), badPrice); !errors.Is(err, ErrOrderLineInvalidPrice) {
		t.Errorf("negative cost err = %v, want ErrOrderLineInvalidPrice", err)
	}
}

func TestServiceService_AddLine(t *testing.T) {
	ctx := context.Background()

	var created *ServiceOrderLine
	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateScheduled}, nil
		},
	}
	lines := ServiceOrderLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, line *ServiceOrderLine) (*ServiceOrderLine, error) {
			created = line
			return line, nil
		},
	}
	svc := NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, lines, posterMock{}, invoiceBuilderMock{}, txMock{})

	line, err := svc.AddLine(ctx, AddLineRequest{OrderID: 1, Line: LineRequest{Type: LineTypeLabor, Qty: 1, UnitPrice: 100}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line.ServiceOrderID != 1 || line.Type != LineTypeLabor {
		t.Errorf("line = %+v", line)
	}
	if created == nil {
		t.Error("line not persisted")
	}

	done := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateDone}, nil
		},
	}
	svc = NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, done, lines, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.AddLine(ctx, AddLineRequest{OrderID: 1, Line: LineRequest{Type: LineTypePart, Qty: 1}}); !errors.Is(err, ErrOrderInvalidState) {
		t.Errorf("done order err = %v, want ErrOrderInvalidState", err)
	}
}

func TestServiceService_Schedule_And_Start(t *testing.T) {
	ctx := context.Background()

	var updatedTx *ServiceOrder
	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateNew}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, order *ServiceOrder) (*ServiceOrder, error) {
			updatedTx = order
			return order, nil
		},
		UpdateFunc: func(_ context.Context, order *ServiceOrder) (*ServiceOrder, error) {
			return order, nil
		},
	}
	svc := NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})

	date := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	scheduled, err := svc.Schedule(ctx, ScheduleRequest{OrderID: 1, ScheduledDate: date, TechnicianID: helper.Ptr(uint64(5))})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheduled.State != OrderStateScheduled || scheduled.ScheduledDate == nil || *scheduled.TechnicianID != 5 {
		t.Errorf("scheduled = %+v", scheduled)
	}
	if updatedTx == nil || updatedTx.State != OrderStateScheduled {
		t.Error("schedule transition not persisted via UpdateTx")
	}

	orders.FindFunc = func(_ context.Context, _ uint64) (*ServiceOrder, error) {
		return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateScheduled}, nil
	}
	svc = NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	started, err := svc.Start(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if started.State != OrderStateInProgress {
		t.Errorf("started = %+v, want in_progress", started)
	}
}

func TestServiceService_Transition_Failure(t *testing.T) {
	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateDone}, nil
		},
	}
	svc := NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.Start(context.Background(), 1); !errors.Is(err, ErrOrderInvalidState) {
		t.Errorf("err = %v, want ErrOrderInvalidState", err)
	}
}

func TestServiceService_Complete_Validation(t *testing.T) {
	svc := NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})

	if _, err := svc.Complete(context.Background(), CompleteRequest{}); !errors.Is(err, ErrOrderRequiresJournal) {
		t.Errorf("no journal err = %v, want ErrOrderRequiresJournal", err)
	}
	if _, err := svc.Complete(context.Background(), CompleteRequest{JournalID: 1, COGSAccountID: 0}); !errors.Is(err, ErrOrderRequiresAccounts) {
		t.Errorf("no accounts err = %v, want ErrOrderRequiresAccounts", err)
	}

	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateNew}, nil
		},
	}
	svc = NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.Complete(context.Background(), CompleteRequest{OrderID: 1, JournalID: 1, COGSAccountID: 2, StockValuationAccountID: 3}); !errors.Is(err, ErrOrderInvalidState) {
		t.Errorf("new order err = %v, want ErrOrderInvalidState", err)
	}
}

func TestServiceService_Bill_Validations(t *testing.T) {
	ctx := context.Background()

	svc := NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.Bill(ctx, BillRequest{}); !errors.Is(err, ErrOrderRequiresJournal) {
		t.Errorf("no journal err = %v, want ErrOrderRequiresJournal", err)
	}
	if _, err := svc.Bill(ctx, BillRequest{JournalID: 1}); !errors.Is(err, ErrOrderRequiresAccounts) {
		t.Errorf("no revenue account err = %v, want ErrOrderRequiresAccounts", err)
	}

	notDone := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateScheduled}, nil
		},
	}
	svc = NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, notDone, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.Bill(ctx, BillRequest{OrderID: 1, JournalID: 1, RevenueAccountID: 2}); !errors.Is(err, ErrOrderNotDone) {
		t.Errorf("not done err = %v, want ErrOrderNotDone", err)
	}

	noContact := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateDone}, nil
		},
	}
	svc = NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, noContact, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.Bill(ctx, BillRequest{OrderID: 1, JournalID: 1, RevenueAccountID: 2}); !errors.Is(err, ErrOrderRequiresContact) {
		t.Errorf("no contact err = %v, want ErrOrderRequiresContact", err)
	}

	noBillable := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateDone, OrganizationID: helper.Ptr(uint64(10)), ContactID: helper.Ptr(uint64(5))}, nil
		},
	}
	lines := ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*ServiceOrderLine, error) {
			return []*ServiceOrderLine{{Billable: false}}, nil
		},
	}
	svc = NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, noBillable, lines, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.Bill(ctx, BillRequest{OrderID: 1, JournalID: 1, RevenueAccountID: 2}); !errors.Is(err, ErrOrderNoLines) {
		t.Errorf("no billable lines err = %v, want ErrOrderNoLines", err)
	}
}

func TestServiceService_CancelOrder(t *testing.T) {
	ctx := context.Background()

	var updatedTx *ServiceOrder
	orders := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateNew}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, order *ServiceOrder) (*ServiceOrder, error) {
			updatedTx = order
			return order, nil
		},
	}
	svc := NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, orders, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})

	cancelled, err := svc.CancelOrder(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cancelled.State != OrderStateCancelled || updatedTx == nil {
		t.Errorf("cancelled = %+v, updatedTx = %v", cancelled, updatedTx)
	}

	done := ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceOrder, error) {
			return &ServiceOrder{Base: model.Base{ID: 1}, State: OrderStateInvoiced}, nil
		},
	}
	svc = NewServiceService(EquipmentDAOMock{}, ServiceContractDAOMock{}, done, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.CancelOrder(ctx, 1); !errors.Is(err, ErrOrderInvalidState) {
		t.Errorf("invoiced err = %v, want ErrOrderInvalidState", err)
	}
}

func TestServiceService_Contract_Extra(t *testing.T) {
	ctx := context.Background()

	var created *ServiceContract
	contracts := ServiceContractDAOMock{
		CreateFunc: func(_ context.Context, contract *ServiceContract) (*ServiceContract, error) {
			created = contract
			return contract, nil
		},
	}
	svc := NewServiceService(EquipmentDAOMock{}, contracts, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})

	if _, err := svc.CreateContract(ctx, CreateContractRequest{Name: ""}); !errors.Is(err, ErrContractNameRequired) {
		t.Errorf("empty name err = %v, want ErrContractNameRequired", err)
	}

	start := time.Now()
	draft, err := svc.CreateContract(ctx, CreateContractRequest{Name: "C1", DateStart: &start})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.State != ContractStateDraft {
		t.Errorf("draft = %+v, want draft state when dates incomplete", draft)
	}
	if created == nil || created.State != ContractStateDraft {
		t.Error("contract not persisted")
	}
}

func TestServiceService_CancelContract_InvalidState(t *testing.T) {
	contracts := ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceContract, error) {
			return &ServiceContract{Base: model.Base{ID: 1}, State: ContractStateDraft}, nil
		},
	}
	svc := NewServiceService(EquipmentDAOMock{}, contracts, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.CancelContract(context.Background(), 1); !errors.Is(err, ErrContractInvalidState) {
		t.Errorf("err = %v, want ErrContractInvalidState", err)
	}
}

func TestServiceService_ActivateContract(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	contracts := ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*ServiceContract, error) {
			return &ServiceContract{Base: model.Base{ID: 1}, State: ContractStateDraft, DateStart: &now, DateEnd: &now}, nil
		},
		UpdateFunc: func(_ context.Context, contract *ServiceContract) (*ServiceContract, error) {
			return contract, nil
		},
	}
	svc := NewServiceService(EquipmentDAOMock{}, contracts, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	activated, err := svc.ActivateContract(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activated.State != ContractStateActive {
		t.Errorf("activated = %+v, want active", activated)
	}

	contracts.FindFunc = func(_ context.Context, _ uint64) (*ServiceContract, error) {
		return &ServiceContract{Base: model.Base{ID: 1}, State: ContractStateActive}, nil
	}
	svc = NewServiceService(EquipmentDAOMock{}, contracts, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.ActivateContract(ctx, 1); !errors.Is(err, ErrContractInvalidState) {
		t.Errorf("active err = %v, want ErrContractInvalidState", err)
	}

	contracts.FindFunc = func(_ context.Context, _ uint64) (*ServiceContract, error) {
		return &ServiceContract{Base: model.Base{ID: 1}, State: ContractStateDraft}, nil
	}
	svc = NewServiceService(EquipmentDAOMock{}, contracts, ServiceOrderDAOMock{}, ServiceOrderLineDAOMock{}, posterMock{}, invoiceBuilderMock{}, txMock{})
	if _, err := svc.ActivateContract(ctx, 1); !errors.Is(err, ErrContractInvalidState) {
		t.Errorf("missing dates err = %v, want ErrContractInvalidState", err)
	}
}

func TestValidateLineType(t *testing.T) {
	if err := validateLineType(LineTypePart); err != nil {
		t.Errorf("part err = %v", err)
	}
	if err := validateLineType("bogus"); !errors.Is(err, ErrOrderLineInvalidType) {
		t.Errorf("bogus err = %v, want ErrOrderLineInvalidType", err)
	}
}
