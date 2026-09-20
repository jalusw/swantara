package returns

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	"gorm.io/gorm"
)

func TestOriginOrderLookup_CustomerOrder_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	findErr := errors.New("orders down")
	orders := saleOrderLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) { return nil, findErr },
	}
	lookup := NewOriginOrderLookup(orders, purchaseOrderLookupMock{})

	_, err := lookup.CustomerOrder(ctx, 1)
	helper.AssertError(t, err, true, findErr)
}

func TestOriginOrderLookup_VendorOrder_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	findErr := errors.New("orders down")
	orders := purchaseOrderLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) { return nil, findErr },
	}
	lookup := NewOriginOrderLookup(saleOrderLookupMock{}, orders)

	_, err := lookup.VendorOrder(ctx, 1)
	helper.AssertError(t, err, true, findErr)
}

func TestRMAService_Create_VendorReturnUsesVendorOrder(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		RMADAOMock{},
		RMALineDAOMock{},
		OriginOrderLookupMock{VendorOrderFunc: func(_ context.Context, _ uint64) (uint64, error) { return 8, nil }},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Number: "RMA/00002"}, nil
		}},
	)

	rma, err := svc.Create(ctx, CreateRMARequest{
		OrganizationID:  1,
		Type:            TypeVendorReturn,
		ContactID:       8,
		OriginOrderType: OriginPurchaseOrder,
		OriginOrderID:   2,
		Lines:           []RMALineRequest{{ItemID: 200, Qty: 1, Disposition: DispositionRestock}},
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if rma.State != StateDraft || rma.Name == nil || *rma.Name != "RMA/00002" {
		t.Errorf("rma = state %s name %v, want draft with RMA/00002", rma.State, rma.Name)
	}
}

func TestRMAService_Create_PropagatesOrderLookupError(t *testing.T) {
	ctx := context.Background()
	lookupErr := errors.New("orders down")
	svc := NewTestRMAService(
		RMADAOMock{},
		RMALineDAOMock{},
		OriginOrderLookupMock{CustomerOrderFunc: func(_ context.Context, _ uint64) (uint64, error) { return 0, lookupErr }},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Create(ctx, CreateRMARequest{
		Type:            TypeCustomerReturn,
		ContactID:       5,
		OriginOrderType: OriginSaleOrder,
		Lines:           []RMALineRequest{{ItemID: 100, Qty: 1, Disposition: DispositionRestock}},
	})
	helper.AssertError(t, err, true, lookupErr)
}

func TestRMAService_Create_PropagatesSequenceError(t *testing.T) {
	ctx := context.Background()
	seqErr := errors.New("sequence down")
	svc := NewTestRMAService(
		RMADAOMock{},
		RMALineDAOMock{},
		OriginOrderLookupMock{CustomerOrderFunc: func(_ context.Context, _ uint64) (uint64, error) { return 5, nil }},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return nil, seqErr
		}},
	)

	_, err := svc.Create(ctx, CreateRMARequest{
		Type:            TypeCustomerReturn,
		ContactID:       5,
		OriginOrderType: OriginSaleOrder,
		Lines:           []RMALineRequest{{ItemID: 100, Qty: 1, Disposition: DispositionRestock}},
	})
	helper.AssertError(t, err, true, seqErr)
}

func TestRMAService_Create_PropagatesPersistenceError(t *testing.T) {
	ctx := context.Background()
	persistErr := errors.New("insert failed")
	svc := NewTestRMAService(
		RMADAOMock{CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, _ *RMA, _ []*RMALine) (*RMA, error) {
			return nil, persistErr
		}},
		RMALineDAOMock{},
		OriginOrderLookupMock{CustomerOrderFunc: func(_ context.Context, _ uint64) (uint64, error) { return 5, nil }},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Number: "RMA/00001"}, nil
		}},
	)

	_, err := svc.Create(ctx, CreateRMARequest{
		Type:            TypeCustomerReturn,
		ContactID:       5,
		OriginOrderType: OriginSaleOrder,
		Lines:           []RMALineRequest{{ItemID: 100, Qty: 1, Disposition: DispositionRestock}},
	})
	helper.AssertError(t, err, true, persistErr)
}

func TestRMAService_Confirm_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	findErr := errors.New("db down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) { return nil, findErr }),
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Confirm(ctx, 1)
	helper.AssertError(t, err, true, findErr)
}

func TestRMAService_Confirm_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()
	updateErr := errors.New("db down")
	svc := NewTestRMAService(
		RMADAOMock{CRUDMock: dao.CRUDMock[RMA]{
			FindFunc: func(_ context.Context, _ uint64) (*RMA, error) {
				return &RMA{Base: model.Base{ID: 1}, State: StateDraft}, nil
			},
			UpdateFunc: func(_ context.Context, _ *RMA) (*RMA, error) { return nil, updateErr },
		}},
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Confirm(ctx, 1)
	helper.AssertError(t, err, true, updateErr)
}

func TestRMAService_Cancel_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	findErr := errors.New("db down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) { return nil, findErr }),
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Cancel(ctx, 1)
	helper.AssertError(t, err, true, findErr)
}

func TestRMAService_Cancel_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()
	updateErr := errors.New("db down")
	svc := NewTestRMAService(
		RMADAOMock{CRUDMock: dao.CRUDMock[RMA]{
			FindFunc: func(_ context.Context, _ uint64) (*RMA, error) {
				return &RMA{Base: model.Base{ID: 1}, State: StateDraft}, nil
			},
			UpdateFunc: func(_ context.Context, _ *RMA) (*RMA, error) { return nil, updateErr },
		}},
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Cancel(ctx, 1)
	helper.AssertError(t, err, true, updateErr)
}

func TestRMAService_Done_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	findErr := errors.New("db down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) { return nil, findErr }),
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Done(ctx, 1)
	helper.AssertError(t, err, true, findErr)
}

func TestRMAService_Done_PropagatesUpdateError(t *testing.T) {
	ctx := context.Background()
	updateErr := errors.New("db down")
	svc := NewTestRMAService(
		RMADAOMock{CRUDMock: dao.CRUDMock[RMA]{
			FindFunc: func(_ context.Context, _ uint64) (*RMA, error) {
				return &RMA{Base: model.Base{ID: 1}, State: StateRefunded}, nil
			},
			UpdateFunc: func(_ context.Context, _ *RMA) (*RMA, error) { return nil, updateErr },
		}},
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Done(ctx, 1)
	helper.AssertError(t, err, true, updateErr)
}

func TestRMAService_Receive_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	findErr := errors.New("db down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) { return nil, findErr }),
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, time.Now())
	helper.AssertError(t, err, true, findErr)
}

func TestRMAService_Receive_ReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) { return nil, nil }),
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, time.Now())
	helper.AssertError(t, err, true, ErrRMANotFound)
}

func TestRMAService_Receive_RejectsWrongState(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, State: StateDraft}, nil
		}),
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, time.Now())
	helper.AssertError(t, err, true, ErrRMAState)
}

func TestRMAService_Receive_PropagatesLineListError(t *testing.T) {
	ctx := context.Background()
	listErr := errors.New("lines down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) { return nil, listErr }},
		OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, time.Now())
	helper.AssertError(t, err, true, listErr)
}

func TestRMAService_Receive_PropagatesMoveCreateError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	moveErr := errors.New("movements down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, SrcLocationID: 10, DstLocationID: 20, State: inventory.MovementStateDone}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, _ *inventory.StockMovement) (*inventory.StockMovement, error) {
				return nil, moveErr
			}},
		},
		StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, moveErr)
}

func TestRMAService_Receive_PropagatesReturnToSupplierError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	valueErr := errors.New("valuation down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Type: TypeVendorReturn, OriginOrderID: helper.Ptr(uint64(7)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 200, Qty: 5, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 22}, ItemID: 200, SrcLocationID: 30, DstLocationID: 40, State: inventory.MovementStateDone}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 51
				return movement, nil
			}},
		},
		StockLocationLookupMock{}, StockLayerLookupMock{},
		ReturnValuerMock{ReturnToSupplierFunc: func(_ context.Context, _ uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
			return nil, valueErr
		}},
		accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 6, date)
	helper.AssertError(t, err, true, valueErr)
}

func TestRMAService_Receive_PropagatesApplyError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	applyErr := errors.New("movements down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 2, Disposition: DispositionScrap}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, SrcLocationID: 10, DstLocationID: 20, State: inventory.MovementStateDone}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 50
				return movement, nil
			}},
			FindForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ uint64) (*inventory.StockMovement, error) {
				return &inventory.StockMovement{Base: model.Base{ID: 50}, State: inventory.MovementStateConfirmed}, nil
			},
			ApplyTxFunc: func(_ context.Context, _ *gorm.DB, _ *inventory.StockMovement) (*inventory.StockMovement, error) {
				return nil, applyErr
			},
		},
		StockLocationLookupMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 99}}}}, nil
		}},
		StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, applyErr)
}

func TestRMAService_Receive_PropagatesRestockError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	valueErr := errors.New("valuation down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, SrcLocationID: 10, DstLocationID: 20, State: inventory.MovementStateDone}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 50
				return movement, nil
			}},
		},
		StockLocationLookupMock{},
		StockLayerLookupMock{ListByMovementFunc: func(_ context.Context, moveID uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: model.Base{ID: 1}, MovementID: &moveID, UnitCost: helper.Ptr(10.0)}}, nil
		}},
		ReturnValuerMock{RestockFunc: func(_ context.Context, _ uint64, _ amount.Amount, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
			return nil, valueErr
		}},
		accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, valueErr)
}

func TestRMAService_Receive_PropagatesRecordReturnError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	recordErr := errors.New("recorder down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, SrcLocationID: 10, DstLocationID: 20, State: inventory.MovementStateDone}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 50
				return movement, nil
			}},
		},
		StockLocationLookupMock{},
		StockLayerLookupMock{ListByMovementFunc: func(_ context.Context, moveID uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: model.Base{ID: 1}, MovementID: &moveID, UnitCost: helper.Ptr(10.0)}}, nil
		}},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)
	svc.saleReturns = SaleOrderReturnRecorderMock{RecordReturnFunc: func(_ context.Context, _ uint64, _ uint64, _ float64) error { return recordErr }}

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, recordErr)
}

func TestRMAService_Receive_PropagatesLineUpdateError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	updateErr := errors.New("db down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Type: TypeVendorReturn, OriginOrderID: helper.Ptr(uint64(7)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{
			CRUDMock: dao.CRUDMock[RMALine]{UpdateFunc: func(_ context.Context, _ *RMALine) (*RMALine, error) { return nil, updateErr }},
			ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
				return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 200, Qty: 5, Disposition: DispositionRestock}}, nil
			},
		},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 22}, ItemID: 200, SrcLocationID: 30, DstLocationID: 40, State: inventory.MovementStateDone}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 51
				return movement, nil
			}},
		},
		StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 6, date)
	helper.AssertError(t, err, true, updateErr)
}

func TestRMAService_Receive_PropagatesRMAUpdateError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	updateErr := errors.New("db down")
	svc := NewTestRMAService(
		RMADAOMock{CRUDMock: dao.CRUDMock[RMA]{
			FindFunc: func(_ context.Context, _ uint64) (*RMA, error) {
				return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
			},
			UpdateFunc: func(_ context.Context, _ *RMA) (*RMA, error) { return nil, updateErr },
		}},
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, SrcLocationID: 10, DstLocationID: 20, State: inventory.MovementStateDone}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 50
				return movement, nil
			}},
		},
		StockLocationLookupMock{},
		StockLayerLookupMock{ListByMovementFunc: func(_ context.Context, moveID uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: model.Base{ID: 1}, MovementID: &moveID, UnitCost: helper.Ptr(10.0)}}, nil
		}},
		ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, updateErr)
}

func TestRMAService_Receive_PropagatesScrapLocationError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	locErr := errors.New("locations down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 2, Disposition: DispositionScrap}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, SrcLocationID: 10, DstLocationID: 20, State: inventory.MovementStateDone}}, nil
			},
		},
		StockLocationLookupMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return nil, locErr
		}},
		StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, locErr)
}

func TestRMAService_Receive_PropagatesCostLookupError(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	layerErr := errors.New("layers down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, SrcLocationID: 10, DstLocationID: 20, State: inventory.MovementStateDone}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 50
				return movement, nil
			}},
		},
		StockLocationLookupMock{},
		StockLayerLookupMock{ListByMovementFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return nil, layerErr
		}},
		ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, layerErr)
}

func TestRMAService_Refund_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	findErr := errors.New("db down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) { return nil, findErr }),
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Refund(ctx, 1, 5, time.Now(), "REF")
	helper.AssertError(t, err, true, findErr)
}

func TestRMAService_Refund_ReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) { return nil, nil }),
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Refund(ctx, 1, 5, time.Now(), "REF")
	helper.AssertError(t, err, true, ErrRMANotFound)
}

func TestRMAService_Refund_RejectsWrongState(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, State: StateConfirmed}, nil
		}),
		RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{},
	)

	_, err := svc.Refund(ctx, 1, 5, time.Now(), "REF")
	helper.AssertError(t, err, true, ErrRMAState)
}

func TestRMAService_Refund_PropagatesCreditNoteError(t *testing.T) {
	ctx := context.Background()
	creditErr := errors.New("credits down")
	organizationID := uint64(1)
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: &organizationID, Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateReceived}, nil
		}),
		RMALineDAOMock{},
		OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{},
		accounting.InvoiceDAOMock{FindBySaleOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 30}, OrganizationID: &organizationID}, nil
		}},
		CreditNoteEngineMock{CreateCreditNoteFunc: func(_ context.Context, _ accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
			return nil, creditErr
		}},
		sequence.DAOMock{},
	)

	_, err := svc.Refund(ctx, 1, 5, time.Now(), "REF")
	helper.AssertError(t, err, true, creditErr)
}

func TestRMAService_Refund_PropagatesLineListError(t *testing.T) {
	ctx := context.Background()
	listErr := errors.New("lines down")
	organizationID := uint64(1)
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: &organizationID, Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateReceived}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) { return nil, listErr }},
		OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{},
		accounting.InvoiceDAOMock{FindBySaleOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 30}, OrganizationID: &organizationID}, nil
		}},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Refund(ctx, 1, 5, time.Now(), "REF")
	helper.AssertError(t, err, true, listErr)
}

func TestRMAService_Refund_PropagatesLineUpdateError(t *testing.T) {
	ctx := context.Background()
	updateErr := errors.New("db down")
	organizationID := uint64(1)
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: &organizationID, Type: TypeCustomerReturn, OriginOrderID: helper.Ptr(uint64(9)), State: StateReceived}, nil
		}),
		RMALineDAOMock{
			CRUDMock: dao.CRUDMock[RMALine]{UpdateFunc: func(_ context.Context, _ *RMALine) (*RMALine, error) { return nil, updateErr }},
			ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
				return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 100, Qty: 2}}, nil
			},
		},
		OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{},
		accounting.InvoiceDAOMock{FindBySaleOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 30}, OrganizationID: &organizationID}, nil
		}},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Refund(ctx, 1, 5, time.Now(), "REF")
	helper.AssertError(t, err, true, updateErr)
}

func TestRMAService_Refund_PropagatesRMAUpdateError(t *testing.T) {
	ctx := context.Background()
	updateErr := errors.New("db down")
	organizationID := uint64(1)
	svc := NewTestRMAService(
		RMADAOMock{CRUDMock: dao.CRUDMock[RMA]{
			FindFunc: func(_ context.Context, _ uint64) (*RMA, error) {
				return &RMA{Base: model.Base{ID: 1}, OrganizationID: &organizationID, Type: TypeVendorReturn, OriginOrderID: helper.Ptr(uint64(7)), State: StateReceived}, nil
			},
			UpdateFunc: func(_ context.Context, _ *RMA) (*RMA, error) { return nil, updateErr },
		}},
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 200, Qty: 5}}, nil
		}},
		OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{},
		accounting.InvoiceDAOMock{FindByPurchaseOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 40}, OrganizationID: &organizationID}, nil
		}},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Refund(ctx, 1, 6, time.Now(), "REF")
	helper.AssertError(t, err, true, updateErr)
}
