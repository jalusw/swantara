package sales

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestSaleOrderService_Deliver(t *testing.T) {
	baseOrder := func() *SaleOrder {
		return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: helper.Ptr("SO/00001"), State: OrderStateConfirmed}
	}
	tests := []struct {
		name      string
		orders    SaleOrderDAOMock
		lines     SaleOrderLineDAOMock
		shipments inventory.ShipmentDAOMock
		movements inventory.StockMovementDAOMock
		ship      ShipEngineMock
		res       inventory.StockHoldDAOMock
		wantErr   error
		wantState string
	}{
		{
			name:   "ships movements and increments delivered qty",
			orders: SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return baseOrder(), nil }}},
			lines: SaleOrderLineDAOMock{
				ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
					return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, QtyDelivered: 0}}, nil
				},
				CRUDMock: dao.CRUDMock[SaleOrderLine]{
					UpdateFunc: func(_ context.Context, line *SaleOrderLine) (*SaleOrderLine, error) { return line, nil },
				},
			},
			shipments: inventory.ShipmentDAOMock{
				CRUDMock: dao.CRUDMock[inventory.Shipment]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
						return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateAssigned}}}, nil
					},
				},
			},
			movements: inventory.StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
					return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, Qty: 3, State: inventory.MovementStateAssigned}}, nil
				},
			},
			ship: ShipEngineMock{
				ShipFunc: func(_ context.Context, moveID uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
					return &inventory.CostLayer{MovementID: &moveID}, nil
				},
			},
			res: inventory.StockHoldDAOMock{
				ReleaseByMovementFunc: func(_ context.Context, _ uint64) error { return nil },
			},
			wantState: DeliveryStatusPartial,
		},
		{
			name:   "rejects when nothing to ship",
			orders: SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return baseOrder(), nil }}},
			lines: SaleOrderLineDAOMock{
				ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
					return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
				},
			},
			shipments: inventory.ShipmentDAOMock{
				CRUDMock: dao.CRUDMock[inventory.Shipment]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
						return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateConfirmed}}}, nil
					},
				},
			},
			movements: inventory.StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
					return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, Qty: 3, State: inventory.MovementStateAssigned}}, nil
				},
			},
			wantErr: ErrOrderNothingToDeliver,
		},
		{
			name:   "rejects without shipment",
			orders: SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return baseOrder(), nil }}},
			lines: SaleOrderLineDAOMock{
				ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
					return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
				},
			},
			shipments: inventory.ShipmentDAOMock{
				CRUDMock: dao.CRUDMock[inventory.Shipment]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
						return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{}}, nil
					},
				},
			},
			wantErr: ErrOrderShipmentNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc := testSaleOrderService(tt.orders, tt.lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			svc.shipments = tt.shipments
			svc.movements = tt.movements
			svc.ship = tt.ship
			svc.reservations = tt.res

			updated, err := svc.Deliver(ctx, 1, 20, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if tt.wantState != "" && updated.DeliveryStatus != tt.wantState {
				t.Errorf("delivery_status = %s, want %s", updated.DeliveryStatus, tt.wantState)
			}
		})
	}
}

func TestSaleOrderService_CreateInvoice(t *testing.T) {
	tests := []struct {
		name       string
		productSvc products.ProductService
		wantErr    error
		wantQty    float64
		wantAcct   uint64
	}{
		{
			name:       "builds lines and increments invoiced qty",
			productSvc: testIncomeProductService(4100),
			wantQty:    3,
			wantAcct:   4100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			order := &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: helper.Ptr("SO/00001"), ContactID: 5}
			orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
			}}
			line := &SaleOrderLine{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), Description: helper.Ptr("Widget"), QtyOrdered: 5, QtyDelivered: 3, QtyInvoiced: 0, UnitPrice: 100, TaxIDs: helper.Int64Array{9}}
			lines := SaleOrderLineDAOMock{
				ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
					return []*SaleOrderLine{line}, nil
				},
			}
			var createdRequest accounting.CreateInvoiceRequest
			invoices := InvoiceEngineMock{
				CreateFunc: func(_ context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
					createdRequest = request
					return &accounting.Invoice{Base: model.Base{ID: 30}, ContactID: request.ContactID}, nil
				},
			}
			svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, tt.productSvc)
			svc.invoices = invoices

			invoice, err := svc.CreateInvoice(ctx, 1, 20, time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC))
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if invoice.ContactID != 5 {
				t.Errorf("invoice contact = %d, want 5", invoice.ContactID)
			}
			if len(createdRequest.Lines) != 1 {
				t.Fatalf("invoice lines = %d, want 1", len(createdRequest.Lines))
			}
			if createdRequest.Lines[0].Qty != tt.wantQty || createdRequest.Lines[0].AccountID != tt.wantAcct {
				t.Errorf("line = qty %v account %d, want %v/%v", createdRequest.Lines[0].Qty, createdRequest.Lines[0].AccountID, tt.wantQty, tt.wantAcct)
			}
			if line.QtyInvoiced != tt.wantQty {
				t.Errorf("qty_invoiced = %v, want %v", line.QtyInvoiced, tt.wantQty)
			}
		})
	}
}

func TestSaleOrderService_CollectPayment(t *testing.T) {
	tests := []struct {
		name         string
		openInvoices InvoiceLookupMock
		wantErr      error
	}{
		{
			name: "allocates open invoices",
			openInvoices: InvoiceLookupMock{
				ListOpenByContactFunc: func(_ context.Context, contactID uint64) ([]*accounting.Invoice, error) {
					return []*accounting.Invoice{
						{Base: model.Base{ID: 11}, ContactID: contactID},
						{Base: model.Base{ID: 12}, ContactID: contactID},
					}, nil
				},
			},
		},
		{
			name: "rejects without open invoices",
			openInvoices: InvoiceLookupMock{
				ListOpenByContactFunc: func(_ context.Context, _ uint64) ([]*accounting.Invoice, error) {
					return []*accounting.Invoice{}, nil
				},
			},
			wantErr: ErrOrderNoInvoices,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			order := &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: helper.Ptr("SO/00001"), ContactID: 5}
			orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
			}}
			var createdRequest accounting.CreatePaymentRequest
			payments := PaymentEngineMock{
				CreateFunc: func(_ context.Context, request accounting.CreatePaymentRequest) (*accounting.Payment, error) {
					createdRequest = request
					return &accounting.Payment{Base: model.Base{ID: 40}, ContactID: request.ContactID, Amount: request.Amount}, nil
				},
			}
			svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			svc.openInvoices = tt.openInvoices
			svc.payments = payments

			payment, err := svc.CollectPayment(ctx, 1, 20, 200, time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC))
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if payment.ContactID != 5 || payment.Amount != 200 {
				t.Errorf("payment = contact %d amount %v, want 5/200", payment.ContactID, payment.Amount)
			}
			if len(createdRequest.InvoiceIDs) != 2 || createdRequest.InvoiceIDs[0] != 11 || createdRequest.InvoiceIDs[1] != 12 {
				t.Errorf("invoice ids = %v, want [11 12]", createdRequest.InvoiceIDs)
			}
		})
	}
}

func testIncomeProductService(incomeAccountID uint64) products.ProductService {
	return products.NewProductService(
		products.ItemDAOMock{
			CRUDMock: dao.CRUDMock[products.Item]{
				FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
					return &products.Item{Base: model.Base{ID: 1}, ListPrice: 100, CategoryID: helper.Ptr(uint64(1))}, nil
				},
			},
		},
		products.ItemVariantDAOMock{
			CRUDMock: dao.CRUDMock[products.ItemVariant]{
				FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
					return &products.ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
				},
			},
		},
		dao.CRUDMock[reference.ItemCategory]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.ItemCategory, error) {
				return &reference.ItemCategory{Base: model.Base{ID: 1}, IncomeAccountID: helper.Ptr(incomeAccountID)}, nil
			},
		},
		products.PriceBookDAOMock{},
		products.PriceRuleDAOMock{},
	)
}
