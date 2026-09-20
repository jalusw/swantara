package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

const supplierID = uint64(10)
const warehouseID = uint64(20)
const supplierLocationID = uint64(30)
const internalLocationID = uint64(40)
const shipmentID = uint64(50)
const moveID = uint64(60)

func TestPurchaseOrderService_Create(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock) (PurchaseOrderService, *PurchaseOrder)
		lines  []*PurchaseOrderLine
		assert func(t *testing.T, order *PurchaseOrder, err error)
	}{
		{
			name: "prices from supplier catalog and sets draft",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock) (PurchaseOrderService, *PurchaseOrder) {
				priced := 0
				orders.CreateWithLinesFunc = func(_ context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error) {
					for _, line := range lines {
						if line.UnitPrice > 0 {
							priced++
						}
					}
					return order, nil
				}
				svc.offers = OfferEngineMock{
					BestOfferForSupplierFunc: func(_ context.Context, variantID, supplierID uint64, _ amount.Amount, _ time.Time) (*products.SupplierProduct, error) {
						return &products.SupplierProduct{ItemID: variantID, SupplierID: supplierID, Price: helper.Ptr(1500.0)}, nil
					},
				}
				svc.suppliers = SupplierProfileLookupMock{
					FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
						return &contacts.SupplierProfile{Active: true}, nil
					},
				}
				return svc, &PurchaseOrder{
					OrganizationID: helper.Ptr(uint64(1)),
					SupplierID:     supplierID,
					WarehouseID:    helper.Ptr(warehouseID),
					CurrencyCode:   helper.Ptr("IDR"),
				}
			},
			lines: []*PurchaseOrderLine{{QtyOrdered: 3, ItemID: helper.Ptr(uint64(200))}},
			assert: func(t *testing.T, order *PurchaseOrder, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order.State != PurchaseOrderStateDraft {
					t.Errorf("state = %s, want draft", order.State)
				}
				if order.AmountTotal != 4500 {
					t.Errorf("amount_total = %v, want 4500", order.AmountTotal)
				}
				if order.Name == nil || *order.Name != "PO/00001" {
					t.Errorf("name = %v, want PO/00001", order.Name)
				}
			},
		},
		{
			name: "rejects without lines",
			setup: func(svc PurchaseOrderService, _ *PurchaseOrderDAOMock) (PurchaseOrderService, *PurchaseOrder) {
				return svc, &PurchaseOrder{
					OrganizationID: helper.Ptr(uint64(1)),
					SupplierID:     supplierID,
				}
			},
			lines: nil,
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderNoLines)
			},
		},
		{
			name: "rejects inactive supplier",
			setup: func(svc PurchaseOrderService, _ *PurchaseOrderDAOMock) (PurchaseOrderService, *PurchaseOrder) {
				svc.suppliers = SupplierProfileLookupMock{
					FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
						return &contacts.SupplierProfile{Active: false}, nil
					},
				}
				return svc, &PurchaseOrder{
					OrganizationID: helper.Ptr(uint64(1)),
					SupplierID:     supplierID,
				}
			},
			lines: []*PurchaseOrderLine{{QtyOrdered: 1, ItemID: helper.Ptr(uint64(200)), UnitPrice: 10}},
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderVendorNotSupplier)
			},
		},
		{
			name: "rejects without offer",
			setup: func(svc PurchaseOrderService, _ *PurchaseOrderDAOMock) (PurchaseOrderService, *PurchaseOrder) {
				svc.offers = OfferEngineMock{
					BestOfferForSupplierFunc: func(_ context.Context, _ uint64, _ uint64, _ amount.Amount, _ time.Time) (*products.SupplierProduct, error) {
						return nil, ErrPurchaseOrderNoOffer
					},
				}
				svc.suppliers = SupplierProfileLookupMock{
					FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
						return &contacts.SupplierProfile{Active: true}, nil
					},
				}
				return svc, &PurchaseOrder{
					OrganizationID: helper.Ptr(uint64(1)),
					SupplierID:     supplierID,
				}
			},
			lines: []*PurchaseOrderLine{{QtyOrdered: 1, ItemID: helper.Ptr(uint64(200))}},
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderNoOffer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, orders, _, _, _, _, _, _, _ := testPurchaseOrderService()
			svc, order := tt.setup(svc, orders)
			created, err := svc.Create(ctx, order, tt.lines)
			tt.assert(t, created, err)
		})
	}
}

func TestPurchaseOrderService_UpdateDraft(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock) PurchaseOrderService
		input  *PurchaseOrder
		lines  []*PurchaseOrderLine
		assert func(t *testing.T, order *PurchaseOrder, err error)
	}{
		{
			name: "replaces lines and recomputes totals",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock) PurchaseOrderService {
				order := draftOrder()
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc:   func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *PurchaseOrder) (*PurchaseOrder, error) { return o, nil },
				}
				lines.ReplaceLinesFunc = func(_ context.Context, _ uint64, _ []*PurchaseOrderLine) error { return nil }
				return svc
			},
			input: &PurchaseOrder{
				Base:           model.Base{ID: 1},
				OrganizationID: helper.Ptr(uint64(1)),
				SupplierID:     supplierID,
				WarehouseID:    helper.Ptr(warehouseID),
				CurrencyCode:   helper.Ptr("IDR"),
			},
			lines: []*PurchaseOrderLine{{QtyOrdered: 5, ItemID: helper.Ptr(uint64(200)), UnitPrice: 2000}},
			assert: func(t *testing.T, order *PurchaseOrder, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order.State != PurchaseOrderStateDraft {
					t.Errorf("state = %s, want draft", order.State)
				}
				if order.Name == nil || *order.Name != "PO/00001" {
					t.Errorf("name = %v, want PO/00001 preserved", order.Name)
				}
				if order.AmountTotal != 10000 {
					t.Errorf("amount_total = %v, want 10000", order.AmountTotal)
				}
			},
		},
		{
			name: "rejects non-draft",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, _ *PurchaseOrderLineDAOMock) PurchaseOrderService {
				order := draftOrder()
				order.State = PurchaseOrderStateSent
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				return svc
			},
			input: &PurchaseOrder{
				Base:        model.Base{ID: 1},
				SupplierID:  supplierID,
				WarehouseID: helper.Ptr(warehouseID),
			},
			lines: []*PurchaseOrderLine{{QtyOrdered: 1, ItemID: helper.Ptr(uint64(200)), UnitPrice: 10}},
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderState)
			},
		},
		{
			name: "rejects without lines",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, _ *PurchaseOrderLineDAOMock) PurchaseOrderService {
				order := draftOrder()
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				return svc
			},
			input: &PurchaseOrder{
				Base:        model.Base{ID: 1},
				SupplierID:  supplierID,
				WarehouseID: helper.Ptr(warehouseID),
			},
			lines: nil,
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderNoLines)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, orders, lines, _, _, _, _, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc, orders, lines)
			updated, err := svc.UpdateDraft(ctx, tt.input, tt.lines)
			tt.assert(t, updated, err)
		})
	}
}

func TestPurchaseOrderService_Confirm(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, approvals *ApprovalEngineMock) PurchaseOrderService
		assert func(t *testing.T, order *PurchaseOrder, err error)
	}{
		{
			name: "below threshold transitions to sent",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, _ *ApprovalEngineMock) PurchaseOrderService {
				order := draftOrder()
				order.AmountTotal = 1000
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc:   func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *PurchaseOrder) (*PurchaseOrder, error) { return o, nil },
				}
				return svc
			},
			assert: func(t *testing.T, order *PurchaseOrder, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order.State != PurchaseOrderStateSent {
					t.Errorf("state = %s, want sent", order.State)
				}
			},
		},
		{
			name: "above threshold requires approval",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, approvals *ApprovalEngineMock) PurchaseOrderService {
				order := draftOrder()
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				svc.configs = NewSystemConfigSource(ConfigLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
						return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
							{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Key: configApprovalThreshold, Value: []byte(`1000`)},
							{Base: model.Base{ID: 2}, OrganizationID: helper.Ptr(uint64(1)), Key: configApproverIDs, Value: []byte(`[7]`)},
						}}, nil
					},
				})
				approvals.CreateFunc = func(_ context.Context, _ uint64, _ string, _ uint64, _ uint64, _ []uint64) (*crosscutting.ApprovalRequest, error) {
					return &crosscutting.ApprovalRequest{}, nil
				}
				return svc
			},
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderApprovalPending)
			},
		},
		{
			name: "above threshold requires approvers",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, _ *ApprovalEngineMock) PurchaseOrderService {
				order := draftOrder()
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				svc.configs = NewSystemConfigSource(ConfigLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
						return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
							{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Key: configApprovalThreshold, Value: []byte(`1000`)},
						}}, nil
					},
				})
				return svc
			},
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderApprovalRequired)
			},
		},
		{
			name: "above threshold approved transitions to sent",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, approvals *ApprovalEngineMock) PurchaseOrderService {
				order := draftOrder()
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc:   func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *PurchaseOrder) (*PurchaseOrder, error) { return o, nil },
				}
				svc.configs = NewSystemConfigSource(ConfigLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
						return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{
							{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Key: configApprovalThreshold, Value: []byte(`1000`)},
						}}, nil
					},
				})
				approvals.IsApprovedFunc = func(_ context.Context, _ string, _ uint64) (bool, error) { return true, nil }
				return svc
			},
			assert: func(t *testing.T, order *PurchaseOrder, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order.State != PurchaseOrderStateSent {
					t.Errorf("state = %s, want sent", order.State)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, orders, _, approvals, _, _, _, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc, orders, approvals)
			confirmed, err := svc.Confirm(ctx, 1, 9)
			tt.assert(t, confirmed, err)
		})
	}
}

func TestPurchaseOrderService_Receive(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, quality *QualityEngineMock, receive *ReceiveEngineMock) PurchaseOrderService
		assert func(t *testing.T, order *PurchaseOrder, err error)
	}{
		{
			name: "receives movements and increments qty",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, quality *QualityEngineMock, receive *ReceiveEngineMock) PurchaseOrderService {
				order := draftOrder()
				order.State = PurchaseOrderStateSent
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc:   func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *PurchaseOrder) (*PurchaseOrder, error) { return o, nil },
				}
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{purchaseLine(5)}, nil
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{}, nil
						},
					},
					CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, _ []*inventory.StockMovement) (*inventory.Shipment, error) {
						shipment.ID = shipmentID
						return shipment, nil
					},
				}
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if q.Filters[0].Field == "usage" {
								return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: supplierLocationID}, Usage: "supplier"}}}, nil
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: internalLocationID}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
						},
					},
				}
				svc.movements = inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{Base: model.Base{ID: moveID}, ItemID: 200, Qty: 5, State: inventory.MovementStateConfirmed}}, nil
					},
				}
				receive.ReceiveFunc = func(_ context.Context, _ uint64, _ amount.Amount, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
					return &inventory.CostLayer{}, nil
				}
				quality.TriggerChecksFunc = func(_ context.Context, _ uint64, _ uint64, _ []uint64) (int, error) {
					return 1, nil
				}
				lines.CRUDMock = dao.CRUDMock[PurchaseOrderLine]{
					UpdateFunc: func(_ context.Context, line *PurchaseOrderLine) (*PurchaseOrderLine, error) {
						return line, nil
					},
				}
				return svc
			},
			assert: func(t *testing.T, order *PurchaseOrder, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order.State != PurchaseOrderStateConfirmed {
					t.Errorf("state = %s, want confirmed", order.State)
				}
				if order.ReceiptStatus != PurchaseOrderReceiptStatusDone {
					t.Errorf("receipt_status = %s, want done", order.ReceiptStatus)
				}
			},
		},
		{
			name: "rejects when nothing to receive",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, _ *QualityEngineMock, _ *ReceiveEngineMock) PurchaseOrderService {
				order := draftOrder()
				order.State = PurchaseOrderStateConfirmed
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{purchaseLine(5)}, nil
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: shipmentID}, Type: inventory.ShipmentTypeIncoming}}}, nil
						},
					},
				}
				svc.movements = inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{Base: model.Base{ID: moveID}, ItemID: 200, Qty: 5, State: inventory.MovementStateDone}}, nil
					},
				}
				return svc
			},
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderNothingToReceive)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, orders, lines, _, quality, receive, _, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc, orders, lines, quality, receive)
			updated, err := svc.Receive(ctx, 1, 90, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
			tt.assert(t, updated, err)
		})
	}
}

func TestPurchaseOrderService_CreateSupplierBill(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, quality *QualityEngineMock, bills *BillEngineMock) PurchaseOrderService
		override bool
		assert   func(t *testing.T, invoice *accounting.Invoice, err error)
	}{
		{
			name: "uses stock input account and quality gate",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, quality *QualityEngineMock, bills *BillEngineMock) PurchaseOrderService {
				order := draftOrder()
				order.State = PurchaseOrderStateConfirmed
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc:   func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *PurchaseOrder) (*PurchaseOrder, error) { return o, nil },
				}
				line := purchaseLine(5)
				line.QtyReceived = 5
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{line}, nil
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: shipmentID}, Type: inventory.ShipmentTypeIncoming}}}, nil
						},
					},
				}
				quality.HasFailedChecksFunc = func(_ context.Context, _ uint64) (bool, error) { return false, nil }
				bills.CreateSupplierBillFunc = func(_ context.Context, req accounting.CreateSupplierBillRequest) (*accounting.Invoice, error) {
					return &accounting.Invoice{Base: model.Base{ID: 500}, ContactID: req.ContactID}, nil
				}
				lines.CRUDMock = dao.CRUDMock[PurchaseOrderLine]{
					UpdateFunc: func(_ context.Context, l *PurchaseOrderLine) (*PurchaseOrderLine, error) {
						return l, nil
					},
				}
				return svc
			},
			override: false,
			assert: func(t *testing.T, invoice *accounting.Invoice, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if invoice == nil {
					t.Fatal("expected invoice")
				}
			},
		},
		{
			name: "blocks on failed quality checks",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, quality *QualityEngineMock, _ *BillEngineMock) PurchaseOrderService {
				order := draftOrder()
				order.State = PurchaseOrderStateConfirmed
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				line := purchaseLine(5)
				line.QtyReceived = 5
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{line}, nil
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: shipmentID}, Type: inventory.ShipmentTypeIncoming}}}, nil
						},
					},
				}
				quality.HasFailedChecksFunc = func(_ context.Context, _ uint64) (bool, error) { return true, nil }
				return svc
			},
			override: false,
			assert: func(t *testing.T, _ *accounting.Invoice, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderBillQualityBlocked)
			},
		},
		{
			name: "override skips quality gate",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock, quality *QualityEngineMock, bills *BillEngineMock) PurchaseOrderService {
				order := draftOrder()
				order.State = PurchaseOrderStateConfirmed
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				line := purchaseLine(5)
				line.QtyReceived = 5
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{line}, nil
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: shipmentID}, Type: inventory.ShipmentTypeIncoming}}}, nil
						},
					},
				}
				quality.HasFailedChecksFunc = func(_ context.Context, _ uint64) (bool, error) { return true, nil }
				bills.CreateSupplierBillFunc = func(_ context.Context, req accounting.CreateSupplierBillRequest) (*accounting.Invoice, error) {
					return &accounting.Invoice{Base: model.Base{ID: 500}, ContactID: req.ContactID}, nil
				}
				return svc
			},
			override: true,
			assert: func(t *testing.T, invoice *accounting.Invoice, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if invoice == nil {
					t.Fatal("expected invoice despite failed checks")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, orders, lines, _, quality, _, bills, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc, orders, lines, quality, bills)
			invoice, err := svc.CreateSupplierBill(ctx, 1, 90, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC), tt.override)
			tt.assert(t, invoice, err)
		})
	}
}

func TestPurchaseOrderService_PaySupplierBill(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, openInvoices *OpenInvoiceLookupMock, payments *OutboundPaymentEngineMock) PurchaseOrderService
		assert func(t *testing.T, payment *accounting.Payment, err error)
	}{
		{
			name: "allocates open invoices",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, openInvoices *OpenInvoiceLookupMock, payments *OutboundPaymentEngineMock) PurchaseOrderService {
				order := draftOrder()
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				openInvoices.ListOpenByContactFunc = func(_ context.Context, _ uint64) ([]*accounting.Invoice, error) {
					return []*accounting.Invoice{{Base: model.Base{ID: 500}, AmountResidual: amount.FromFloat64(1000000)}}, nil
				}
				payments.CreateOutboundFunc = func(_ context.Context, req accounting.CreatePaymentRequest) (*accounting.Payment, error) {
					return &accounting.Payment{Base: model.Base{ID: 900}, ContactID: req.ContactID, Amount: req.Amount}, nil
				}
				return svc
			},
			assert: func(t *testing.T, payment *accounting.Payment, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if payment == nil {
					t.Fatal("expected payment")
				}
			},
		},
		{
			name: "rejects without open bills",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, openInvoices *OpenInvoiceLookupMock, _ *OutboundPaymentEngineMock) PurchaseOrderService {
				order := draftOrder()
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				openInvoices.ListOpenByContactFunc = func(_ context.Context, _ uint64) ([]*accounting.Invoice, error) {
					return []*accounting.Invoice{}, nil
				}
				return svc
			},
			assert: func(t *testing.T, _ *accounting.Payment, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderNoOpenBills)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, orders, _, _, _, _, _, openInvoices, payments := testPurchaseOrderService()
			svc = tt.setup(svc, orders, openInvoices, payments)
			payment, err := svc.PaySupplierBill(ctx, 1, 90, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
			tt.assert(t, payment, err)
		})
	}
}

func TestPurchaseOrderService_CreateFromRequest(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock) PurchaseOrderService
		assert func(t *testing.T, order *PurchaseOrder, err error)
	}{
		{
			name: "rejects when not approved",
			setup: func(svc PurchaseOrderService, _ *PurchaseOrderDAOMock) PurchaseOrderService {
				svc.requisitions = &PurchaseRequestDAOMock{
					CRUDMock: dao.CRUDMock[PurchaseRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) {
							return &PurchaseRequest{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), State: RequestStateConfirmed}, nil
						},
					},
				}
				return svc
			},
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrRequisitionNotConvertible)
			},
		},
		{
			name: "converts approved requisition",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock) PurchaseOrderService {
				svc.requisitions = &PurchaseRequestDAOMock{
					CRUDMock: dao.CRUDMock[PurchaseRequest]{
						FindFunc: func(_ context.Context, _ uint64) (*PurchaseRequest, error) {
							return &PurchaseRequest{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), State: RequestStateApproved}, nil
						},
					},
				}
				svc.linesDAO = &PurchaseRequestLineDAOMock{
					ListByRequestFunc: func(_ context.Context, _ uint64) ([]*PurchaseRequestLine, error) {
						return []*PurchaseRequestLine{{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(200)), Qty: 7}}, nil
					},
				}
				svc.suppliers = SupplierProfileLookupMock{
					FindByContactFunc: func(_ context.Context, _ uint64) (*contacts.SupplierProfile, error) {
						return &contacts.SupplierProfile{Active: true}, nil
					},
				}
				svc.offers = OfferEngineMock{
					BestOfferForSupplierFunc: func(_ context.Context, variantID, supplierID uint64, _ amount.Amount, _ time.Time) (*products.SupplierProduct, error) {
						return &products.SupplierProduct{ItemID: variantID, SupplierID: supplierID, Price: helper.Ptr(500.0)}, nil
					},
				}
				orders.CreateWithLinesFunc = func(_ context.Context, order *PurchaseOrder, lines []*PurchaseOrderLine) (*PurchaseOrder, error) {
					return order, nil
				}
				return svc
			},
			assert: func(t *testing.T, order *PurchaseOrder, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order.State != PurchaseOrderStateDraft {
					t.Errorf("state = %s, want draft", order.State)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, orders, _, _, _, _, _, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc, orders)
			order, err := svc.CreateFromRequest(ctx, 1, &PurchaseOrder{SupplierID: supplierID, WarehouseID: helper.Ptr(warehouseID)})
			tt.assert(t, order, err)
		})
	}
}

func TestPurchaseOrderService_RecordReturn(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock) PurchaseOrderService
		assert func(t *testing.T, err error)
	}{
		{
			name: "increments and recomputes",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock) PurchaseOrderService {
				order := draftOrder()
				order.State = PurchaseOrderStateSent
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc:   func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *PurchaseOrder) (*PurchaseOrder, error) { return o, nil },
				}
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{purchaseLine(5)}, nil
				}
				lines.CRUDMock = dao.CRUDMock[PurchaseOrderLine]{
					UpdateFunc: func(_ context.Context, l *PurchaseOrderLine) (*PurchaseOrderLine, error) {
						return l, nil
					},
				}
				return svc
			},
			assert: func(t *testing.T, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "line not found",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock, lines *PurchaseOrderLineDAOMock) PurchaseOrderService {
				order := draftOrder()
				order.State = PurchaseOrderStateSent
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				lines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*PurchaseOrderLine, error) {
					return []*PurchaseOrderLine{}, nil
				}
				return svc
			},
			assert: func(t *testing.T, err error) {
				if !errors.Is(err, ErrPurchaseOrderLineNotFound) {
					t.Fatalf("error = %v, want %v", err, ErrPurchaseOrderLineNotFound)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, orders, lines, _, _, _, _, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc, orders, lines)
			err := svc.RecordReturn(ctx, 1, 200, 2)
			tt.assert(t, err)
		})
	}
}

func TestPurchaseOrderService_Cancel(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock) PurchaseOrderService
		assert func(t *testing.T, order *PurchaseOrder, err error)
	}{
		{
			name: "transitions draft to cancelled",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock) PurchaseOrderService {
				order := draftOrder()
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc:   func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, o *PurchaseOrder) (*PurchaseOrder, error) { return o, nil },
				}
				return svc
			},
			assert: func(t *testing.T, order *PurchaseOrder, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order.State != PurchaseOrderStateCancelled {
					t.Errorf("state = %s, want cancelled", order.State)
				}
			},
		},
		{
			name: "rejects done state",
			setup: func(svc PurchaseOrderService, orders *PurchaseOrderDAOMock) PurchaseOrderService {
				order := draftOrder()
				order.State = PurchaseOrderStateDone
				orders.CRUDMock = dao.CRUDMock[PurchaseOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*PurchaseOrder, error) { return order, nil },
				}
				return svc
			},
			assert: func(t *testing.T, _ *PurchaseOrder, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderState)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, orders, _, _, _, _, _, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc, orders)
			cancelled, err := svc.Cancel(ctx, 1)
			tt.assert(t, cancelled, err)
		})
	}
}

func TestPurchaseOrderService_RecomputeTotals(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService) PurchaseOrderService
		lines  []*PurchaseOrderLine
		assert func(t *testing.T, order *PurchaseOrder)
	}{
		{
			name: "applies percent tax",
			setup: func(svc PurchaseOrderService) PurchaseOrderService {
				svc.taxes = dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return &reference.Tax{Base: model.Base{ID: 1}, Amount: helper.Ptr(10.0), Type: reference.TaxTypePercent}, nil
					},
				}
				return svc
			},
			lines: []*PurchaseOrderLine{{QtyOrdered: 2, UnitPrice: 100, TaxIDs: helper.Int64Array{1}}},
			assert: func(t *testing.T, order *PurchaseOrder) {
				if order.AmountUntaxed != 200 {
					t.Errorf("amount_untaxed = %v, want 200", order.AmountUntaxed)
				}
				if order.AmountTax != 20 {
					t.Errorf("amount_tax = %v, want 20", order.AmountTax)
				}
				if order.AmountTotal != 220 {
					t.Errorf("amount_total = %v, want 220", order.AmountTotal)
				}
			},
		},
		{
			name: "skips missing tax",
			setup: func(svc PurchaseOrderService) PurchaseOrderService {
				svc.taxes = dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return nil, nil
					},
				}
				return svc
			},
			lines: []*PurchaseOrderLine{{QtyOrdered: 2, UnitPrice: 100, TaxIDs: helper.Int64Array{1}}},
			assert: func(t *testing.T, order *PurchaseOrder) {
				if order.AmountTax != 0 {
					t.Errorf("amount_tax = %v, want 0", order.AmountTax)
				}
			},
		},
		{
			name: "skips tax on find error",
			setup: func(svc PurchaseOrderService) PurchaseOrderService {
				svc.taxes = dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return nil, errors.New("db down")
					},
				}
				return svc
			},
			lines: []*PurchaseOrderLine{{QtyOrdered: 2, UnitPrice: 100, TaxIDs: helper.Int64Array{1}}},
			assert: func(t *testing.T, order *PurchaseOrder) {
				if order.AmountTax != 0 {
					t.Errorf("amount_tax = %v, want 0", order.AmountTax)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc)
			order := draftOrder()
			svc.recomputeTotals(ctx, order, tt.lines)
			tt.assert(t, order)
		})
	}
}

func TestPurchaseOrderService_DestLocation(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService) (PurchaseOrderService, *PurchaseOrder)
		assert func(t *testing.T, location uint64, err error)
	}{
		{
			name: "uses explicit location",
			setup: func(svc PurchaseOrderService) (PurchaseOrderService, *PurchaseOrder) {
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
							return &reference.StockLocation{Base: model.Base{ID: 40}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}, nil
						},
					},
				}
				order := draftOrder()
				order.DestLocationID = helper.Ptr(uint64(40))
				return svc, order
			},
			assert: func(t *testing.T, location uint64, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if location != 40 {
					t.Errorf("location = %d, want 40", location)
				}
			},
		},
		{
			name: "rejects foreign explicit location",
			setup: func(svc PurchaseOrderService) (PurchaseOrderService, *PurchaseOrder) {
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
							return &reference.StockLocation{Base: model.Base{ID: 40}, OrganizationID: helper.Ptr(uint64(2)), Usage: "internal"}, nil
						},
					},
				}
				order := draftOrder()
				order.DestLocationID = helper.Ptr(uint64(40))
				return svc, order
			},
			assert: func(t *testing.T, _ uint64, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderLocation)
			},
		},
		{
			name: "rejects missing explicit location",
			setup: func(svc PurchaseOrderService) (PurchaseOrderService, *PurchaseOrder) {
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
							return nil, nil
						},
					},
				}
				order := draftOrder()
				order.DestLocationID = helper.Ptr(uint64(40))
				return svc, order
			},
			assert: func(t *testing.T, _ uint64, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderLocation)
			},
		},
		{
			name: "falls back to warehouse internal location",
			setup: func(svc PurchaseOrderService) (PurchaseOrderService, *PurchaseOrder) {
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
								{Base: model.Base{ID: 40}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"},
							}}, nil
						},
					},
				}
				return svc, draftOrder()
			},
			assert: func(t *testing.T, location uint64, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if location != 40 {
					t.Errorf("location = %d, want 40", location)
				}
			},
		},
		{
			name: "rejects orgless internal location",
			setup: func(svc PurchaseOrderService) (PurchaseOrderService, *PurchaseOrder) {
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
								{Base: model.Base{ID: 40}, Usage: "internal"},
							}}, nil
						},
					},
				}
				return svc, draftOrder()
			},
			assert: func(t *testing.T, _ uint64, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderLocation)
			},
		},
		{
			name: "rejects without warehouse or location",
			setup: func(svc PurchaseOrderService) (PurchaseOrderService, *PurchaseOrder) {
				order := draftOrder()
				order.WarehouseID = nil
				return svc, order
			},
			assert: func(t *testing.T, _ uint64, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderLocation)
			},
		},
		{
			name: "rejects without internal location",
			setup: func(svc PurchaseOrderService) (PurchaseOrderService, *PurchaseOrder) {
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
								{Base: model.Base{ID: 40}, Usage: "supplier"},
							}}, nil
						},
					},
				}
				return svc, draftOrder()
			},
			assert: func(t *testing.T, _ uint64, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderLocation)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
			svc, order := tt.setup(svc)
			location, err := svc.destLocation(ctx, order)
			tt.assert(t, location, err)
		})
	}
}

func TestPurchaseOrderService_SupplierLocation(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService) PurchaseOrderService
		orgID  *uint64
		assert func(t *testing.T, location uint64, err error)
	}{
		{
			name: "returns matching organization location",
			setup: func(svc PurchaseOrderService) PurchaseOrderService {
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
								{Base: model.Base{ID: 30}, OrganizationID: helper.Ptr(uint64(1)), Usage: "supplier"},
							}}, nil
						},
					},
				}
				return svc
			},
			orgID: draftOrder().OrganizationID,
			assert: func(t *testing.T, location uint64, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if location != 30 {
					t.Errorf("location = %d, want 30", location)
				}
			},
		},
		{
			name: "accepts global location",
			setup: func(svc PurchaseOrderService) PurchaseOrderService {
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
								{Base: model.Base{ID: 30}, Usage: "supplier"},
							}}, nil
						},
					},
				}
				return svc
			},
			orgID: draftOrder().OrganizationID,
			assert: func(t *testing.T, location uint64, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if location != 30 {
					t.Errorf("location = %d, want 30", location)
				}
			},
		},
		{
			name: "rejects when missing",
			setup: func(svc PurchaseOrderService) PurchaseOrderService {
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{}, nil
						},
					},
				}
				return svc
			},
			orgID: draftOrder().OrganizationID,
			assert: func(t *testing.T, _ uint64, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderSupplierLocation)
			},
		},
		{
			name:  "rejects without organization",
			setup: func(svc PurchaseOrderService) PurchaseOrderService { return svc },
			orgID: nil,
			assert: func(t *testing.T, _ uint64, err error) {
				helper.AssertError(t, err, true, ErrPurchaseOrderNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc)
			location, err := svc.supplierLocation(ctx, tt.orgID)
			tt.assert(t, location, err)
		})
	}
}

func TestPurchaseOrderService_Threshold(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(svc PurchaseOrderService) PurchaseOrderService
		assert func(t *testing.T, threshold float64)
	}{
		{
			name: "returns zero on config error",
			setup: func(svc PurchaseOrderService) PurchaseOrderService {
				svc.configs = NewSystemConfigSource(ConfigLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
						return nil, errors.New("db down")
					},
				})
				return svc
			},
			assert: func(t *testing.T, threshold float64) {
				if threshold != 0 {
					t.Errorf("threshold = %v, want 0", threshold)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc, _, _, _, _, _, _, _, _ := testPurchaseOrderService()
			svc = tt.setup(svc)
			order := draftOrder()
			threshold := svc.threshold(ctx, order)
			tt.assert(t, threshold)
		})
	}
}

type expenseAccountEngineStub struct{ accountID uint64 }

func (e expenseAccountEngineStub) ResolveExpenseAccount(_ context.Context, _ uint64) (uint64, error) {
	return e.accountID, nil
}

func testExpenseAccountEngine(accountID uint64) ExpenseAccountEngine {
	return expenseAccountEngineStub{accountID: accountID}
}

func draftOrder() *PurchaseOrder {
	return &PurchaseOrder{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(1)),
		Name:           helper.Ptr("PO/00001"),
		SupplierID:     supplierID,
		WarehouseID:    helper.Ptr(warehouseID),
		State:          PurchaseOrderStateDraft,
		AmountTotal:    1000000,
	}
}

func purchaseLine(qty float64) *PurchaseOrderLine {
	return &PurchaseOrderLine{
		Base:        model.Base{ID: 100},
		OrderID:     1,
		ItemID:      helper.Ptr(uint64(200)),
		Description: helper.Ptr("Widget"),
		QtyOrdered:  qty,
		UnitPrice:   1000,
	}
}

func testPurchaseOrderService() (PurchaseOrderService, *PurchaseOrderDAOMock, *PurchaseOrderLineDAOMock, *ApprovalEngineMock, *QualityEngineMock, *ReceiveEngineMock, *BillEngineMock, *OpenInvoiceLookupMock, *OutboundPaymentEngineMock) {
	orders := &PurchaseOrderDAOMock{}
	lines := &PurchaseOrderLineDAOMock{}
	requisitions := &PurchaseRequestDAOMock{}
	requisitionLines := &PurchaseRequestLineDAOMock{}
	agreements := &SupplyAgreementDAOMock{}
	agreementLines := &SupplyAgreementLineDAOMock{}
	creditMemos := &PurchaseCreditMemoDAOMock{}
	debitMemos := &PurchaseDebitMemoDAOMock{}
	batches := &PaymentBatchDAOMock{}
	batchLines := &PaymentBatchLineDAOMock{}
	suppliers := &SupplierProfileLookupMock{}
	approvals := &ApprovalEngineMock{}
	quality := &QualityEngineMock{}
	receive := &ReceiveEngineMock{}
	bills := &BillEngineMock{}
	openInvoices := &OpenInvoiceLookupMock{}
	payments := &OutboundPaymentEngineMock{}

	svc := NewPurchaseOrderService(
		orders,
		lines,
		requisitions,
		requisitionLines,
		agreements,
		agreementLines,
		creditMemos,
		debitMemos,
		batches,
		batchLines,
		sequence.NewSequenceService(sequence.DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
				return &sequence.Reservation{Value: 1, Number: "PO/00001"}, nil
			},
		}),
		NewSystemConfigSource(ConfigLookupMock{}),
		OfferEngineMock{},
		contacts.ContactDAOMock{
			CRUDMock: dao.CRUDMock[contacts.Contact]{
				FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return &contacts.Contact{Base: model.Base{ID: supplierID}}, nil
				},
			},
		},
		suppliers,
		inventory.WarehouseDAOMock{
			CRUDMock: dao.CRUDMock[reference.Warehouse]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
					return &reference.Warehouse{Base: model.Base{ID: warehouseID}, OrganizationID: helper.Ptr(uint64(1))}, nil
				},
			},
		},
		inventory.StockLocationDAOMock{},
		inventory.ShipmentDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.ItemResolverMock{
			ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
				return inventory.ResolvedItem{
					StockAccounts: inventory.StockAccounts{StockInputAccountID: 1310, StockValuationAccountID: 1300},
					Tracking:      "none",
				}, nil
			},
		},
		testExpenseAccountEngine(6000),
		dao.CRUDMock[reference.Tax]{},
		receive,
		approvals,
		bills,
		openInvoices,
		payments,
		quality,
		CurrencyConverterMock{},
	)
	return svc, orders, lines, approvals, quality, receive, bills, openInvoices, payments
}
