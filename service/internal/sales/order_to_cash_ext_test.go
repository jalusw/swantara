package sales

import (
	"context"
	"errors"
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

func TestSaleOrderService_Deliver_Ext(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			"propagates find error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return nil, errors.New("db down")
					},
				}}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"rejects when not confirmed",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateDraft}, nil
					},
				}}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, ErrOrderState) {
					return
				}
			},
		},
		{
			"rejects missing name",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed}, nil
					},
				}}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, ErrOrderNotFound) {
					return
				}
			},
		},
		{
			"rejects no lines",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, ErrOrderNoLines) {
					return
				}
			},
		},
		{
			"propagates list by order error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return nil, errors.New("db down")
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"propagates shipment list error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				shipments := inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
				svc.shipments = shipments

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"propagates movement list error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				shipments := inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateAssigned}}}, nil
						},
					},
				}
				movements := inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return nil, errors.New("db down")
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
				svc.shipments = shipments
				svc.movements = movements

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"skips done and cancelled movements",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				shipments := inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateAssigned}}}, nil
						},
					},
				}
				movements := inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{
							{Base: model.Base{ID: 21}, ItemID: 100, Qty: 3, State: inventory.MovementStateDone},
							{Base: model.Base{ID: 22}, ItemID: 100, Qty: 3, State: inventory.MovementStateCancelled},
						}, nil
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
				svc.shipments = shipments
				svc.movements = movements

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, ErrOrderNothingToDeliver) {
					return
				}
			},
		},
		{
			"propagates ship error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				shipments := inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateAssigned}}}, nil
						},
					},
				}
				movements := inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, Qty: 3, State: inventory.MovementStateAssigned}}, nil
					},
				}
				ship := ShipEngineMock{
					ShipFunc: func(_ context.Context, _ uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
						return nil, errors.New("ship failed")
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
				svc.shipments = shipments
				svc.movements = movements
				svc.ship = ship

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"propagates release error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				shipments := inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateAssigned}}}, nil
						},
					},
				}
				movements := inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, Qty: 3, State: inventory.MovementStateAssigned}}, nil
					},
				}
				reservations := inventory.StockHoldDAOMock{
					ReleaseByMovementFunc: func(_ context.Context, _ uint64) error {
						return errors.New("release failed")
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
				svc.shipments = shipments
				svc.movements = movements
				svc.reservations = reservations

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"propagates shipment update error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				shipments := inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateAssigned}}}, nil
						},
						UpdateFunc: func(_ context.Context, _ *inventory.Shipment) (*inventory.Shipment, error) {
							return nil, errors.New("update failed")
						},
					},
				}
				movements := inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, Qty: 3, State: inventory.MovementStateAssigned}}, nil
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
				svc.shipments = shipments
				svc.movements = movements

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"propagates line update error",
			func(t *testing.T) {
				ctx := context.Background()
				order := &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				lines := SaleOrderLineDAOMock{
					CRUDMock: dao.CRUDMock[SaleOrderLine]{
						UpdateFunc: func(_ context.Context, _ *SaleOrderLine) (*SaleOrderLine, error) {
							return nil, errors.New("update failed")
						},
					},
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				shipments := inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateAssigned}}}, nil
						},
					},
				}
				movements := inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, Qty: 3, State: inventory.MovementStateAssigned}}, nil
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
				svc.shipments = shipments
				svc.movements = movements

				_, err := svc.Deliver(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"skips line without item",
			func(t *testing.T) {
				ctx := context.Background()
				order := &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				updatedLines := make([]*SaleOrderLine, 0)
				lines := SaleOrderLineDAOMock{
					CRUDMock: dao.CRUDMock[SaleOrderLine]{
						UpdateFunc: func(_ context.Context, line *SaleOrderLine) (*SaleOrderLine, error) {
							updatedLines = append(updatedLines, line)
							return line, nil
						},
					},
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{
							{Base: model.Base{ID: 6}, QtyOrdered: 5, QtyDelivered: 0},
							{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, QtyDelivered: 0},
						}, nil
					},
				}
				shipments := inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, State: inventory.ShipmentStateAssigned}}}, nil
						},
					},
				}
				movements := inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, Qty: 3, State: inventory.MovementStateAssigned}}, nil
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
				svc.shipments = shipments
				svc.movements = movements

				updated, err := svc.Deliver(ctx, 1, 20, time.Now())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(updatedLines) != 1 || updatedLines[0].ID != 7 {
					t.Errorf("updated lines = %+v, want only the item line", updatedLines)
				}
				if updated.DeliveryStatus != DeliveryStatusPartial {
					t.Errorf("delivery_status = %s, want partial", updated.DeliveryStatus)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSaleOrderService_CreateInvoice_Ext(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			"propagates find error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return nil, errors.New("db down")
					},
				}}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.CreateInvoice(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"rejects missing organization",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.CreateInvoice(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, ErrOrderNotFound) {
					return
				}
			},
		},
		{
			"propagates list by order error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1))}, nil
					},
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return nil, errors.New("db down")
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.CreateInvoice(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"rejects no lines",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1))}, nil
					},
				}}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.CreateInvoice(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, ErrOrderNoLines) {
					return
				}
			},
		},
		{
			"skips fully invoiced lines",
			func(t *testing.T) {
				ctx := context.Background()
				order := &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: helper.Ptr("SO/00001")}
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{
							{Base: model.Base{ID: 7}, QtyOrdered: 5, QtyDelivered: 3, QtyInvoiced: 3},
							{Base: model.Base{ID: 8}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, QtyDelivered: 3, QtyInvoiced: 3},
						}, nil
					},
				}
				var createdRequest accounting.CreateInvoiceRequest
				invoices := InvoiceEngineMock{
					CreateFunc: func(_ context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
						createdRequest = request
						return &accounting.Invoice{Base: model.Base{ID: 30}, ContactID: request.ContactID}, nil
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testIncomeProductService(4100))
				svc.invoices = invoices

				_, err := svc.CreateInvoice(ctx, 1, 20, time.Now())
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(createdRequest.Lines) != 0 {
					t.Errorf("lines = %+v, want no invoicable lines", createdRequest.Lines)
				}
			},
		},
		{
			"propagates income account error",
			func(t *testing.T) {
				ctx := context.Background()
				order := &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: helper.Ptr("SO/00001")}
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, QtyDelivered: 3}}, nil
					},
				}
				productSvc := products.NewProductService(
					products.ItemDAOMock{
						CRUDMock: dao.CRUDMock[products.Item]{
							FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
								return &products.Item{Base: model.Base{ID: 1}, ListPrice: 100}, nil
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
					dao.CRUDMock[reference.ItemCategory]{},
					products.PriceBookDAOMock{},
					products.PriceRuleDAOMock{},
				)
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, productSvc)

				_, err := svc.CreateInvoice(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"propagates create error",
			func(t *testing.T) {
				ctx := context.Background()
				order := &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: helper.Ptr("SO/00001")}
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, QtyDelivered: 3}}, nil
					},
				}
				invoices := InvoiceEngineMock{
					CreateFunc: func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
						return nil, errors.New("create failed")
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testIncomeProductService(4100))
				svc.invoices = invoices

				_, err := svc.CreateInvoice(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"propagates line update error",
			func(t *testing.T) {
				ctx := context.Background()
				order := &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: helper.Ptr("SO/00001")}
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				lines := SaleOrderLineDAOMock{
					CRUDMock: dao.CRUDMock[SaleOrderLine]{
						UpdateFunc: func(_ context.Context, _ *SaleOrderLine) (*SaleOrderLine, error) {
							return nil, errors.New("update failed")
						},
					},
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, QtyDelivered: 3}}, nil
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testIncomeProductService(4100))

				_, err := svc.CreateInvoice(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"propagates order update error",
			func(t *testing.T) {
				ctx := context.Background()
				order := &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: helper.Ptr("SO/00001")}
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, _ *SaleOrder) (*SaleOrder, error) {
						return nil, errors.New("update failed")
					},
				}}
				lines := SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5, QtyDelivered: 3}}, nil
					},
				}
				svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testIncomeProductService(4100))

				_, err := svc.CreateInvoice(ctx, 1, 20, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestSaleOrderService_CollectPayment_Ext(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			"propagates find error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return nil, errors.New("db down")
					},
				}}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.CollectPayment(ctx, 1, 20, 200, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
		{
			"rejects missing organization",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

				_, err := svc.CollectPayment(ctx, 1, 20, 200, time.Now())
				if helper.AssertError(t, err, true, ErrOrderNotFound) {
					return
				}
			},
		},
		{
			"propagates list open error",
			func(t *testing.T) {
				ctx := context.Background()
				orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(1)), Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				openInvoices := InvoiceLookupMock{
					ListOpenByContactFunc: func(_ context.Context, _ uint64) ([]*accounting.Invoice, error) {
						return nil, errors.New("db down")
					},
				}
				svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
				svc.openInvoices = openInvoices

				_, err := svc.CollectPayment(ctx, 1, 20, 200, time.Now())
				if helper.AssertError(t, err, true, nil) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
