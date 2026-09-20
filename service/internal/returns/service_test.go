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

func TestOriginOrderLookup_CustomerOrder_ReturnsContactForConfirmedOrder(t *testing.T) {
	ctx := context.Background()
	orders := saleOrderLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: model.Base{ID: 1}, State: sales.OrderStateConfirmed, ContactID: 5}, nil
		},
	}
	lookup := NewOriginOrderLookup(orders, purchaseOrderLookupMock{})

	contact, err := lookup.CustomerOrder(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contact != 5 {
		t.Errorf("contact = %d, want 5", contact)
	}
}

func TestOriginOrderLookup_CustomerOrder_RejectsUnconfirmedOrder(t *testing.T) {
	ctx := context.Background()
	orders := saleOrderLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
			return &sales.SaleOrder{Base: model.Base{ID: 1}, State: sales.OrderStateDraft, ContactID: 5}, nil
		},
	}
	lookup := NewOriginOrderLookup(orders, purchaseOrderLookupMock{})

	_, err := lookup.CustomerOrder(ctx, 1)
	if helper.AssertError(t, err, true, ErrRMAOrder) {
		return
	}
}

func TestOriginOrderLookup_CustomerOrder_RejectsMissingAndErroringOrder(t *testing.T) {
	ctx := context.Background()
	orders := saleOrderLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) { return nil, nil },
	}
	lookup := NewOriginOrderLookup(orders, purchaseOrderLookupMock{})

	_, err := lookup.CustomerOrder(ctx, 1)
	if helper.AssertError(t, err, true, ErrRMAOrder) {
		return
	}

	sentinel := errors.New("lookup down")
	orders = saleOrderLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) { return nil, sentinel },
	}
	lookup = NewOriginOrderLookup(orders, purchaseOrderLookupMock{})

	_, err = lookup.CustomerOrder(ctx, 1)
	if helper.AssertError(t, err, true, sentinel) {
		return
	}
}

func TestOriginOrderLookup_VendorOrder_ReturnsVendorForConfirmedOrder(t *testing.T) {
	ctx := context.Background()
	orders := purchaseOrderLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: model.Base{ID: 1}, State: procurement.PurchaseOrderStateDone, SupplierID: 7}, nil
		},
	}
	lookup := NewOriginOrderLookup(saleOrderLookupMock{}, orders)

	supplier, err := lookup.VendorOrder(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if supplier != 7 {
		t.Errorf("supplier = %d, want 7", supplier)
	}
}

func TestOriginOrderLookup_VendorOrder_RejectsUnconfirmedOrder(t *testing.T) {
	ctx := context.Background()
	orders := purchaseOrderLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
			return &procurement.PurchaseOrder{Base: model.Base{ID: 1}, State: procurement.PurchaseOrderStateDraft, SupplierID: 7}, nil
		},
	}
	lookup := NewOriginOrderLookup(saleOrderLookupMock{}, orders)

	_, err := lookup.VendorOrder(ctx, 1)
	if helper.AssertError(t, err, true, ErrRMAOrder) {
		return
	}
}

func TestOriginOrderLookup_VendorOrder_RejectsMissingOrder(t *testing.T) {
	ctx := context.Background()
	orders := purchaseOrderLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) { return nil, nil },
	}
	lookup := NewOriginOrderLookup(saleOrderLookupMock{}, orders)

	_, err := lookup.VendorOrder(ctx, 1)
	if helper.AssertError(t, err, true, ErrRMAOrder) {
		return
	}
}

type saleOrderLookupMock struct {
	FindFunc func(ctx context.Context, id uint64) (*sales.SaleOrder, error)
}

func (m saleOrderLookupMock) Find(ctx context.Context, id uint64) (*sales.SaleOrder, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

type purchaseOrderLookupMock struct {
	FindFunc func(ctx context.Context, id uint64) (*procurement.PurchaseOrder, error)
}

func (m purchaseOrderLookupMock) Find(ctx context.Context, id uint64) (*procurement.PurchaseOrder, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func TestRMAService_Create_SetsDraftStateAndName(t *testing.T) {
	ctx := context.Background()
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
			return &sequence.Reservation{Number: "RMA/00001"}, nil
		}},
	)

	rma, err := svc.Create(ctx, CreateRMARequest{
		OrganizationID:  1,
		Type:            TypeCustomerReturn,
		ContactID:       5,
		OriginOrderType: OriginSaleOrder,
		OriginOrderID:   1,
		Reason:          "Defective",
		Lines:           []RMALineRequest{{ItemID: 100, Qty: 2, Disposition: DispositionRestock}},
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if rma.State != StateDraft || rma.Name == nil || *rma.Name != "RMA/00001" {
		t.Errorf("rma = state %s name %v, want draft with RMA/00001", rma.State, rma.Name)
	}
	if rma.OriginOrderType == nil || *rma.OriginOrderType != OriginSaleOrder {
		t.Errorf("origin order type = %v, want sale_order", rma.OriginOrderType)
	}
}

func TestRMAService_Create_RejectsInvalidType(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(RMADAOMock{}, RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	_, err := svc.Create(ctx, CreateRMARequest{
		OrganizationID:  1,
		Type:            "swap",
		ContactID:       5,
		OriginOrderType: OriginSaleOrder,
		OriginOrderID:   1,
		Lines:           []RMALineRequest{{ItemID: 100, Qty: 1, Disposition: DispositionRestock}},
	})
	helper.AssertError(t, err, true, ErrRMAType)
}

func TestRMAService_Create_RejectsInvalidDisposition(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(RMADAOMock{}, RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	_, err := svc.Create(ctx, CreateRMARequest{
		OrganizationID:  1,
		Type:            TypeCustomerReturn,
		ContactID:       5,
		OriginOrderType: OriginSaleOrder,
		OriginOrderID:   1,
		Lines:           []RMALineRequest{{ItemID: 100, Qty: 1, Disposition: "burn"}},
	})
	helper.AssertError(t, err, true, ErrRMADisposition)
}

func TestRMAService_Create_RejectsContactMismatch(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		RMADAOMock{},
		RMALineDAOMock{},
		OriginOrderLookupMock{CustomerOrderFunc: func(_ context.Context, _ uint64) (uint64, error) { return 6, nil }},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Create(ctx, CreateRMARequest{
		OrganizationID:  1,
		Type:            TypeCustomerReturn,
		ContactID:       5,
		OriginOrderType: OriginSaleOrder,
		OriginOrderID:   1,
		Lines:           []RMALineRequest{{ItemID: 100, Qty: 1, Disposition: DispositionRestock}},
	})
	helper.AssertError(t, err, true, ErrRMAContactMismatch)
}

func TestRMAService_Confirm_TransitionsToConfirmed(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, State: StateDraft}, nil
		}),
		RMALineDAOMock{},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	updated, err := svc.Confirm(ctx, 1)
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	if updated.State != StateConfirmed {
		t.Errorf("state = %s, want confirmed", updated.State)
	}
}

func TestRMAService_Confirm_RejectsWrongState(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, State: StateReceived}, nil
		}),
		RMALineDAOMock{},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Confirm(ctx, 1)
	helper.AssertError(t, err, true, ErrRMAState)
}

func TestRMAService_Receive_CustomerReturnRestock(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	organizationID := uint64(1)
	restocked := false
	restockCost := amount.Amount{}
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{
				Base:            model.Base{ID: 1},
				OrganizationID:  &organizationID,
				Type:            TypeCustomerReturn,
				OriginOrderType: helper.Ptr(OriginSaleOrder),
				OriginOrderID:   helper.Ptr(uint64(9)),
				State:           StateConfirmed,
			}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, originType string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{
					Base:          model.Base{ID: 21},
					ItemID:        100,
					Qty:           3,
					SrcLocationID: 10,
					DstLocationID: 20,
					UnitID:        helper.Ptr(uint64(2)),
					State:         inventory.MovementStateDone,
				}}, nil
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
		ReturnValuerMock{RestockFunc: func(_ context.Context, moveID uint64, unitCost amount.Amount, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
			restocked = true
			restockCost = unitCost
			return &inventory.CostLayer{MovementID: &moveID}, nil
		}},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	rma, err := svc.Receive(ctx, 1, 5, date)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if rma.State != StateReceived {
		t.Errorf("state = %s, want received", rma.State)
	}
	if !restocked {
		t.Error("restock valuer was not called")
	}
	if !restockCost.Equal(amount.FromFloat64(10)) {
		t.Errorf("restock unit cost = %v, want 10", restockCost)
	}
}

func TestRMAService_Receive_CustomerReturnScrap(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	organizationID := uint64(1)
	applied := false
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{
				Base:            model.Base{ID: 1},
				OrganizationID:  &organizationID,
				Type:            TypeCustomerReturn,
				OriginOrderType: helper.Ptr(OriginSaleOrder),
				OriginOrderID:   helper.Ptr(uint64(9)),
				State:           StateConfirmed,
			}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 100, Qty: 2, Disposition: DispositionScrap}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{
					Base:          model.Base{ID: 21},
					ItemID:        100,
					Qty:           3,
					SrcLocationID: 10,
					DstLocationID: 20,
					State:         inventory.MovementStateDone,
				}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 50
				return movement, nil
			}},
			FindForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, moveID uint64) (*inventory.StockMovement, error) {
				return &inventory.StockMovement{Base: model.Base{ID: moveID}, State: inventory.MovementStateConfirmed}, nil
			},
			ApplyTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				applied = true
				movement.State = inventory.MovementStateDone
				return movement, nil
			},
		},
		StockLocationLookupMock{ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
			if q.Filters[0].Field == "usage" && q.Filters[0].Value == "scrap" {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 99}}}}, nil
			}
			return &query.Page[reference.StockLocation]{}, nil
		}},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	rma, err := svc.Receive(ctx, 1, 5, date)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if rma.State != StateReceived || !applied {
		t.Errorf("state = %s applied = %v, want received and movement applied", rma.State, applied)
	}
}

func TestRMAService_Receive_VendorReturn(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	organizationID := uint64(1)
	returned := false
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{
				Base:            model.Base{ID: 1},
				OrganizationID:  &organizationID,
				Type:            TypeVendorReturn,
				OriginOrderType: helper.Ptr(OriginPurchaseOrder),
				OriginOrderID:   helper.Ptr(uint64(7)),
				State:           StateConfirmed,
			}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 200, Qty: 5, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, originType string, _ uint64) ([]*inventory.StockMovement, error) {
				if originType != OriginPurchaseOrder {
					t.Errorf("origin type = %s, want purchase_order", originType)
				}
				return []*inventory.StockMovement{{
					Base:          model.Base{ID: 22},
					ItemID:        200,
					Qty:           5,
					SrcLocationID: 30,
					DstLocationID: 40,
					State:         inventory.MovementStateDone,
				}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				if movement.SrcLocationID != 40 || movement.DstLocationID != 30 {
					t.Errorf("reverse movement = %d -> %d, want 40 -> 30", movement.SrcLocationID, movement.DstLocationID)
				}
				movement.ID = 51
				return movement, nil
			}},
		},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{ReturnToSupplierFunc: func(_ context.Context, moveID uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
			returned = true
			return &inventory.CostLayer{MovementID: &moveID}, nil
		}},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	rma, err := svc.Receive(ctx, 1, 6, date)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if rma.State != StateReceived || !returned {
		t.Errorf("state = %s returned = %v, want received and returned to supplier", rma.State, returned)
	}
}

func TestRMAService_Receive_MissingMove(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	organizationID := uint64(1)
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{
				Base:            model.Base{ID: 1},
				OrganizationID:  &organizationID,
				Type:            TypeCustomerReturn,
				OriginOrderType: helper.Ptr(OriginSaleOrder),
				OriginOrderID:   helper.Ptr(uint64(9)),
				State:           StateConfirmed,
			}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
			return nil, nil
		}},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, ErrRMAMoveNotFound)
}

func TestRMAService_Refund_CustomerReturnCreatesCreditNote(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)
	organizationID := uint64(1)
	credited := false
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{
				Base:            model.Base{ID: 1},
				OrganizationID:  &organizationID,
				Type:            TypeCustomerReturn,
				OriginOrderType: helper.Ptr(OriginSaleOrder),
				OriginOrderID:   helper.Ptr(uint64(9)),
				State:           StateReceived,
			}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{FindBySaleOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 30}, OrganizationID: &organizationID, Type: accounting.InvoiceTypeCustomerInvoice}, nil
		}},
		CreditNoteEngineMock{CreateCreditNoteFunc: func(_ context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
			credited = true
			if request.OriginalInvoiceID != 30 {
				t.Errorf("original invoice = %d, want 30", request.OriginalInvoiceID)
			}
			return &accounting.Invoice{Base: model.Base{ID: 31}, OrganizationID: &organizationID, Type: accounting.InvoiceTypeCustomerCredit}, nil
		}},
		sequence.DAOMock{},
	)

	rma, err := svc.Refund(ctx, 1, 5, date, "RMA-1")
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}
	if rma.State != StateRefunded || !credited {
		t.Errorf("state = %s credited = %v, want refunded and credit note created", rma.State, credited)
	}
}

func TestRMAService_Refund_VendorReturnUsesVendorCreditNote(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)
	organizationID := uint64(1)
	credited := false
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{
				Base:            model.Base{ID: 1},
				OrganizationID:  &organizationID,
				Type:            TypeVendorReturn,
				OriginOrderType: helper.Ptr(OriginPurchaseOrder),
				OriginOrderID:   helper.Ptr(uint64(7)),
				State:           StateReceived,
			}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 200, Qty: 5, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{FindByPurchaseOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 40}, OrganizationID: &organizationID, Type: accounting.InvoiceTypeSupplierBill}, nil
		}},
		CreditNoteEngineMock{CreateVendorCreditNoteFunc: func(_ context.Context, request accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
			credited = true
			if request.OriginalInvoiceID != 40 {
				t.Errorf("original invoice = %d, want 40", request.OriginalInvoiceID)
			}
			return &accounting.Invoice{Base: model.Base{ID: 41}, OrganizationID: &organizationID, Type: accounting.InvoiceTypeSupplierCredit}, nil
		}},
		sequence.DAOMock{},
	)

	rma, err := svc.Refund(ctx, 1, 6, date, "RMA-2")
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}
	if rma.State != StateRefunded || !credited {
		t.Errorf("state = %s credited = %v, want refunded and supplier credit note created", rma.State, credited)
	}
}

func TestRMAService_Refund_NoInvoice(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)
	organizationID := uint64(1)
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{
				Base:            model.Base{ID: 1},
				OrganizationID:  &organizationID,
				Type:            TypeCustomerReturn,
				OriginOrderType: helper.Ptr(OriginSaleOrder),
				OriginOrderID:   helper.Ptr(uint64(9)),
				State:           StateReceived,
			}, nil
		}),
		RMALineDAOMock{},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{FindBySaleOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return nil, nil
		}},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Refund(ctx, 1, 5, date, "")
	helper.AssertError(t, err, true, ErrRMAInvoice)
}

func TestRMAService_Done_AfterRefund(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, State: StateRefunded}, nil
		}),
		RMALineDAOMock{},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	updated, err := svc.Done(ctx, 1)
	if err != nil {
		t.Fatalf("done failed: %v", err)
	}
	if updated.State != StateDone {
		t.Errorf("state = %s, want done", updated.State)
	}
}

func TestRMAService_Cancel_FromConfirmed(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, State: StateConfirmed}, nil
		}),
		RMALineDAOMock{},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	updated, err := svc.Cancel(ctx, 1)
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	if updated.State != StateCancelled {
		t.Errorf("state = %s, want cancelled", updated.State)
	}
}

func TestRMAService_Cancel_RejectsReceived(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, State: StateReceived}, nil
		}),
		RMALineDAOMock{},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Cancel(ctx, 1)
	helper.AssertError(t, err, true, ErrRMAState)
}

func TestRMAService_Receive_CustomerReturnRecordsOriginSale(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	organizationID := uint64(1)
	recorded := false
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{
				Base:            model.Base{ID: 1},
				OrganizationID:  &organizationID,
				Type:            TypeCustomerReturn,
				OriginOrderType: helper.Ptr(OriginSaleOrder),
				OriginOrderID:   helper.Ptr(uint64(9)),
				State:           StateConfirmed,
			}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{
					Base:          model.Base{ID: 21},
					ItemID:        100,
					Qty:           3,
					SrcLocationID: 10,
					DstLocationID: 20,
					State:         inventory.MovementStateDone,
				}}, nil
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
		ReturnValuerMock{RestockFunc: func(_ context.Context, moveID uint64, _ amount.Amount, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
			return &inventory.CostLayer{MovementID: &moveID}, nil
		}},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)
	svc.saleReturns = SaleOrderReturnRecorderMock{RecordReturnFunc: func(_ context.Context, orderID uint64, itemID uint64, qty float64) error {
		if orderID != 9 || itemID != 100 || qty != 2 {
			t.Errorf("record return = (%d, %d, %v), want (9, 100, 2)", orderID, itemID, qty)
		}
		recorded = true
		return nil
	}}

	_, err := svc.Receive(ctx, 1, 5, date)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if !recorded {
		t.Error("expected origin sale order return to be recorded")
	}
}

func TestRMAService_Receive_VendorReturnRecordsOriginPurchase(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	organizationID := uint64(1)
	recorded := false
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{
				Base:            model.Base{ID: 1},
				OrganizationID:  &organizationID,
				Type:            TypeVendorReturn,
				OriginOrderType: helper.Ptr(OriginPurchaseOrder),
				OriginOrderID:   helper.Ptr(uint64(7)),
				State:           StateConfirmed,
			}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, RMAID: 1, ItemID: 200, Qty: 5, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return []*inventory.StockMovement{{
					Base:          model.Base{ID: 22},
					ItemID:        200,
					Qty:           5,
					SrcLocationID: 30,
					DstLocationID: 40,
					State:         inventory.MovementStateDone,
				}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 51
				return movement, nil
			}},
		},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{ReturnToSupplierFunc: func(_ context.Context, moveID uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
			return &inventory.CostLayer{MovementID: &moveID}, nil
		}},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)
	svc.purchaseReturn = PurchaseOrderReturnRecorderMock{RecordReturnFunc: func(_ context.Context, orderID uint64, itemID uint64, qty float64) error {
		if orderID != 7 || itemID != 200 || qty != 5 {
			t.Errorf("record return = (%d, %d, %v), want (7, 200, 5)", orderID, itemID, qty)
		}
		recorded = true
		return nil
	}}

	_, err := svc.Receive(ctx, 1, 5, date)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if !recorded {
		t.Error("expected origin purchase order return to be recorded")
	}
}

func TestRMAService_Create_RejectsWithoutLines(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(RMADAOMock{}, RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	_, err := svc.Create(ctx, CreateRMARequest{Type: TypeCustomerReturn})
	helper.AssertError(t, err, true, ErrRMALines)
}

func TestRMAService_Create_RejectsUnknownOriginOrderType(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(RMADAOMock{}, RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	_, err := svc.Create(ctx, CreateRMARequest{
		Type:            TypeCustomerReturn,
		OriginOrderType: "drop",
		Lines:           []RMALineRequest{{ItemID: 100, Qty: 1, Disposition: DispositionRestock}},
	})
	helper.AssertError(t, err, true, ErrRMAOrder)
}

func TestRMAService_ValidateLines_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(RMADAOMock{}, RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	tests := []struct {
		name    string
		lines   []RMALineRequest
		wantErr error
	}{
		{name: "no item", lines: []RMALineRequest{{Qty: 1, Disposition: DispositionRestock}}, wantErr: ErrRMALineProduct},
		{name: "non positive qty", lines: []RMALineRequest{{ItemID: 100, Qty: 0, Disposition: DispositionRestock}}, wantErr: ErrRMALineQty},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, CreateRMARequest{Type: TypeCustomerReturn, OriginOrderType: OriginSaleOrder, Lines: tt.lines})
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestRMAService_StateTransitions_NotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
		return nil, nil
	}), RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	if _, err := svc.Confirm(ctx, 1); !errors.Is(err, ErrRMANotFound) {
		t.Errorf("Confirm = %v, want ErrRMANotFound", err)
	}
	if _, err := svc.Cancel(ctx, 1); !errors.Is(err, ErrRMANotFound) {
		t.Errorf("Cancel = %v, want ErrRMANotFound", err)
	}
	if _, err := svc.Done(ctx, 1); !errors.Is(err, ErrRMANotFound) {
		t.Errorf("Done = %v, want ErrRMANotFound", err)
	}
}

func TestRMAService_Done_RejectsNonRefundedState(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
		return &RMA{Base: model.Base{ID: 1}, State: StateReceived}, nil
	}), RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	_, err := svc.Done(ctx, 1)
	helper.AssertError(t, err, true, ErrRMAState)
}

func TestRMAService_Done_MarksDone(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
		return &RMA{Base: model.Base{ID: 1}, State: StateRefunded}, nil
	}), RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	done, err := svc.Done(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if done.State != StateDone {
		t.Errorf("state = %s, want done", done.State)
	}
}

func TestRMAService_Receive_RejectsWithoutLines(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
		return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
	}), RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	_, err := svc.Receive(ctx, 1, 5, time.Now())
	helper.AssertError(t, err, true, ErrRMALines)
}

func TestRMAService_Receive_ScrapWithoutLocation(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
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
				return []*inventory.StockMovement{{
					Base:          model.Base{ID: 21},
					ItemID:        100,
					SrcLocationID: 10,
					DstLocationID: 20,
					State:         inventory.MovementStateDone,
				}}, nil
			},
		},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, ErrRMAScrapLocation)
}

func TestRMAService_Receive_RejectsMissingOriginalCost(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
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
				return []*inventory.StockMovement{{
					Base:          model.Base{ID: 21},
					ItemID:        100,
					SrcLocationID: 10,
					DstLocationID: 20,
					State:         inventory.MovementStateDone,
				}}, nil
			},
			CRUDMock: dao.CRUDMock[inventory.StockMovement]{CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 50
				return movement, nil
			}},
		},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, date)
	helper.AssertError(t, err, true, ErrRMACost)
}

func TestRMAService_Receive_RejectsMissingOrganization(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
		return &RMA{Base: model.Base{ID: 1}, State: StateConfirmed}, nil
	}), RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	_, err := svc.Receive(ctx, 1, 5, time.Now())
	helper.AssertError(t, err, true, ErrRMAOrganization)
}

func TestRMAService_Receive_PropagatesMoveLookupError(t *testing.T) {
	ctx := context.Background()
	moveErr := errors.New("movements down")
	svc := NewTestRMAService(
		rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
			return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), OriginOrderID: helper.Ptr(uint64(9)), State: StateConfirmed}, nil
		}),
		RMALineDAOMock{ListByRMAFunc: func(_ context.Context, _ uint64) ([]*RMALine, error) {
			return []*RMALine{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 2, Disposition: DispositionRestock}}, nil
		}},
		OriginOrderLookupMock{},
		inventory.StockMovementDAOMock{
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*inventory.StockMovement, error) {
				return nil, moveErr
			},
		},
		StockLocationLookupMock{},
		StockLayerLookupMock{},
		ReturnValuerMock{},
		accounting.InvoiceDAOMock{},
		CreditNoteEngineMock{},
		sequence.DAOMock{},
	)

	_, err := svc.Receive(ctx, 1, 5, time.Now())
	helper.AssertError(t, err, true, moveErr)
}

func TestRMAService_Refund_RejectsMissingOrganization(t *testing.T) {
	ctx := context.Background()
	svc := NewTestRMAService(rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
		return &RMA{Base: model.Base{ID: 1}, State: StateReceived}, nil
	}), RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, accounting.InvoiceDAOMock{}, CreditNoteEngineMock{}, sequence.DAOMock{})

	_, err := svc.Refund(ctx, 1, 5, time.Now(), "REF")
	helper.AssertError(t, err, true, ErrRMAOrganization)
}

func TestRMAService_Refund_PropagatesInvoiceLookupError(t *testing.T) {
	ctx := context.Background()
	invoiceErr := errors.New("invoices down")
	invoices := accounting.InvoiceDAOMock{
		FindBySaleOrderFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return nil, invoiceErr
		},
	}
	svc := NewTestRMAService(rmaDAOMock(func(_ context.Context, _ uint64) (*RMA, error) {
		return &RMA{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), OriginOrderID: helper.Ptr(uint64(9)), Type: TypeCustomerReturn, State: StateReceived}, nil
	}), RMALineDAOMock{}, OriginOrderLookupMock{}, inventory.StockMovementDAOMock{}, StockLocationLookupMock{}, StockLayerLookupMock{}, ReturnValuerMock{}, invoices, CreditNoteEngineMock{}, sequence.DAOMock{})

	_, err := svc.Refund(ctx, 1, 5, time.Now(), "REF")
	helper.AssertError(t, err, true, invoiceErr)
}
