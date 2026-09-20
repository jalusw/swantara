package sales

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestSaleOrderService_Create(t *testing.T) {
	tests := []struct {
		name        string
		taxes       dao.CRUDMock[reference.Tax]
		orders      SaleOrderDAOMock
		leads       crm.ProspectDAOMock
		stages      dao.CRUDMock[reference.PipelineStage]
		contacts    contacts.ContactDAOMock
		warehouses  inventory.WarehouseDAOMock
		price_books products.PriceBookDAOMock
		productSvc  *products.ProductService
		order       *SaleOrder
		lines       []*SaleOrderLine
		wantErr     error
		wantName    string
		wantState   string
	}{
		{
			name: "computes amounts and sets draft state",
			taxes: dao.CRUDMock[reference.Tax]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
					return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0)}, nil
				},
			},
			orders: SaleOrderDAOMock{
				CreateWithLinesFunc: func(_ context.Context, order *SaleOrder, lines []*SaleOrderLine) (*SaleOrder, error) {
					order.ID = 1
					return order, nil
				},
			},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:     []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, DiscountPct: 10, TaxIDs: helper.Int64Array{9}}},
			wantName:  "SO/00001",
			wantState: OrderStateDraft,
		},
		{
			name:   "rejects when opportunity not won",
			taxes:  dao.CRUDMock[reference.Tax]{},
			orders: SaleOrderDAOMock{},
			leads: crm.ProspectDAOMock{
				CRUDMock: dao.CRUDMock[crm.Prospect]{
					FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
						return &crm.Prospect{Base: model.Base{ID: 3}, Type: crm.ProspectKindOpportunity, StageID: helper.Ptr(uint64(8))}, nil
					},
				},
			},
			stages: dao.CRUDMock[reference.PipelineStage]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
					return &reference.PipelineStage{Base: model.Base{ID: 8}, IsWon: false}, nil
				},
			},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderLeadNotWon,
		},
		{
			name:   "skips CRM validation when nil",
			taxes:  dao.CRUDMock[reference.Tax]{},
			orders: SaleOrderDAOMock{},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:     []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantState: OrderStateDraft,
		},
		{
			name:   "rejects lead from another organization",
			taxes:  dao.CRUDMock[reference.Tax]{},
			orders: SaleOrderDAOMock{},
			leads: crm.ProspectDAOMock{
				CRUDMock: dao.CRUDMock[crm.Prospect]{
					FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
						return &crm.Prospect{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(20)), Type: crm.ProspectKindOpportunity, StageID: helper.Ptr(uint64(8))}, nil
					},
				},
			},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderLeadOrganization,
		},
		{
			name:   "rejects contact from another organization",
			taxes:  dao.CRUDMock[reference.Tax]{},
			orders: SaleOrderDAOMock{},
			contacts: contacts.ContactDAOMock{
				CRUDMock: dao.CRUDMock[contacts.Contact]{
					FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
						return &contacts.Contact{Base: model.Base{ID: 5}, OrganizationID: helper.Ptr(uint64(20))}, nil
					},
				},
			},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderContactOrganization,
		},
		{
			name:   "rejects warehouse from another organization",
			taxes:  dao.CRUDMock[reference.Tax]{},
			orders: SaleOrderDAOMock{},
			warehouses: inventory.WarehouseDAOMock{
				CRUDMock: dao.CRUDMock[reference.Warehouse]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
						return &reference.Warehouse{Base: model.Base{ID: 4}, OrganizationID: helper.Ptr(uint64(20))}, nil
					},
				},
			},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderWarehouseOrganization,
		},
		{
			name:   "rejects orgless warehouse",
			taxes:  dao.CRUDMock[reference.Tax]{},
			orders: SaleOrderDAOMock{},
			warehouses: inventory.WarehouseDAOMock{
				CRUDMock: dao.CRUDMock[reference.Warehouse]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
						return &reference.Warehouse{Base: model.Base{ID: 4}}, nil
					},
				},
			},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderWarehouseOrganization,
		},
		{
			name:   "rejects price_book from another organization",
			taxes:  dao.CRUDMock[reference.Tax]{},
			orders: SaleOrderDAOMock{},
			price_books: products.PriceBookDAOMock{
				CRUDMock: dao.CRUDMock[products.PriceBook]{
					FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
						return &products.PriceBook{Base: model.Base{ID: 2}, OrganizationID: helper.Ptr(uint64(20))}, nil
					},
				},
			},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderPriceBookOrganization,
		},
		{
			name:   "rejects empty lines",
			taxes:  dao.CRUDMock[reference.Tax]{},
			orders: SaleOrderDAOMock{},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   nil,
			wantErr: ErrOrderNoLines,
		},
		{
			name:   "rejects variant from another organization",
			taxes:  dao.CRUDMock[reference.Tax]{},
			orders: SaleOrderDAOMock{},
			productSvc: func() *products.ProductService {
				svc := products.NewProductService(
					products.ItemDAOMock{
						CRUDMock: dao.CRUDMock[products.Item]{
							FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
								return &products.Item{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(20)), ListPrice: 100}, nil
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
					products.PriceBookDAOMock{
						CRUDMock: dao.CRUDMock[products.PriceBook]{
							FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
								return &products.PriceBook{Base: model.Base{ID: 2}, Name: "Retail", CurrencyCode: helper.Ptr("IDR")}, nil
							},
						},
					},
					products.PriceRuleDAOMock{},
				)
				return &svc
			}(),
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderVariantOrganization,
		},
		{
			name: "rejects tax from another organization",
			taxes: dao.CRUDMock[reference.Tax]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
					return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, OrganizationID: helper.Ptr(uint64(20)), Amount: helper.Ptr(10.0)}, nil
				},
			},
			orders: SaleOrderDAOMock{},
			order: &SaleOrder{
				OrganizationID: helper.Ptr(uint64(1)),
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, TaxIDs: helper.Int64Array{9}}},
			wantErr: ErrOrderTaxOrganization,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			taxes := tt.taxes
			if taxes.FindFunc == nil {
				taxes.FindFunc = func(_ context.Context, _ uint64) (*reference.Tax, error) {
					return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0)}, nil
				}
			}
			orders := tt.orders
			if orders.CreateWithLinesFunc == nil && tt.wantErr == nil {
				orders.CreateWithLinesFunc = func(_ context.Context, order *SaleOrder, lines []*SaleOrderLine) (*SaleOrder, error) {
					order.ID = 1
					return order, nil
				}
			}
			productSvc := testPriceService(100)
			if tt.productSvc != nil {
				productSvc = *tt.productSvc
			}
			svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, taxes, productSvc)
			if tt.leads.FindFunc != nil {
				svc.leads = tt.leads
			}
			if tt.stages.FindFunc != nil {
				svc.stages = tt.stages
			}
			if tt.contacts.FindFunc != nil {
				svc.contacts = tt.contacts
			}
			if tt.warehouses.FindFunc != nil {
				svc.warehouses = tt.warehouses
			}
			if tt.price_books.FindFunc != nil {
				svc.price_books = tt.price_books
			}
			order, err := svc.Create(ctx, tt.order, tt.lines)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if tt.wantName != "" && (order.Name == nil || *order.Name != tt.wantName) {
				t.Errorf("name = %v, want %s", order.Name, tt.wantName)
			}
			if tt.wantState != "" && order.State != tt.wantState {
				t.Errorf("state = %s, want %s", order.State, tt.wantState)
			}
		})
	}
}

func TestSaleOrderService_Send(t *testing.T) {
	tests := []struct {
		name      string
		order     *SaleOrder
		wantErr   error
		wantState string
	}{
		{
			name:      "transitions to sent",
			order:     &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateDraft},
			wantState: OrderStateSent,
		},
		{
			name:    "rejects when confirmed",
			order:   &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed},
			wantErr: ErrOrderState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return tt.order, nil },
			}}
			svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

			updated, err := svc.Send(ctx, 1)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if tt.wantState != "" && updated.State != tt.wantState {
				t.Errorf("state = %s, want %s", updated.State, tt.wantState)
			}
		})
	}
}

func TestSaleOrderService_UpdateDraft(t *testing.T) {
	tests := []struct {
		name    string
		order   *SaleOrder
		lines   []*SaleOrderLine
		wantErr error
	}{
		{
			name:    "rejects missing warehouse",
			order:   &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateDraft},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderWarehouse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return tt.order, nil },
			}}
			svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

			_, err := svc.UpdateDraft(ctx, 1, &SaleOrder{ContactID: 5, PriceBookID: helper.Ptr(uint64(2))}, tt.lines)
			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
		})
	}
}

func TestSaleOrderService_Confirm(t *testing.T) {
	OrganizationID := uint64(1)
	tests := []struct {
		name              string
		order             *SaleOrder
		orders            SaleOrderDAOMock
		lines             SaleOrderLineDAOMock
		quants            inventory.StockBalanceDAOMock
		shipments         inventory.ShipmentDAOMock
		movements         inventory.StockMovementDAOMock
		reservations      inventory.StockHoldDAOMock
		wantErr           error
		wantState         string
		wantShipment      bool
		wantShipmentType  string
		wantShipmentState string
	}{
		{
			name: "reserves stock and creates shipment",
			order: &SaleOrder{
				Base: model.Base{ID: 1}, OrganizationID: &OrganizationID,
				Name: helper.Ptr("SO/00001"), State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4)),
			},
			orders: SaleOrderDAOMock{
				CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{
							Base: model.Base{ID: 1}, OrganizationID: &OrganizationID,
							Name: helper.Ptr("SO/00001"), State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4)),
						}, nil
					},
				},
				CreateWithLinesFunc: func(_ context.Context, o *SaleOrder, _ []*SaleOrderLine) (*SaleOrder, error) { return o, nil },
			},
			lines: SaleOrderLineDAOMock{
				ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
					return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
				},
			},
			quants: inventory.StockBalanceDAOMock{
				FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
					return &inventory.StockBalance{Base: model.Base{ID: 3}, OrganizationID: &OrganizationID, Quantity: 10}, nil
				},
			},
			shipments: inventory.ShipmentDAOMock{
				CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
					shipment.ID = 1
					for i := range movements {
						movements[i].ID = uint64(20 + i)
					}
					return shipment, nil
				},
			},
			movements: inventory.StockMovementDAOMock{
				CRUDMock: dao.CRUDMock[inventory.StockMovement]{
					UpdateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
						return movement, nil
					},
				},
			},
			reservations: inventory.StockHoldDAOMock{
				ReserveFunc: func(_ context.Context, quantID uint64, moveID *uint64, qty float64) (*inventory.StockHold, error) {
					return &inventory.StockHold{Base: model.Base{ID: 1}, BalanceID: quantID, MovementID: moveID, Qty: qty}, nil
				},
			},
			wantState:         OrderStateConfirmed,
			wantShipment:      true,
			wantShipmentType:  inventory.ShipmentTypeOutgoing,
			wantShipmentState: inventory.ShipmentStateAssigned,
		},
		{
			name: "rejects insufficient stock",
			order: &SaleOrder{
				Base: model.Base{ID: 1}, OrganizationID: &OrganizationID,
				Name: helper.Ptr("SO/00001"), State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4)),
			},
			orders: SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
					return &SaleOrder{
						Base: model.Base{ID: 1}, OrganizationID: &OrganizationID,
						Name: helper.Ptr("SO/00001"), State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4)),
					}, nil
				},
			}},
			lines: SaleOrderLineDAOMock{
				ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
					return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
				},
			},
			quants: inventory.StockBalanceDAOMock{
				FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
					return &inventory.StockBalance{Base: model.Base{ID: 3}, OrganizationID: &OrganizationID, Quantity: 3}, nil
				},
			},
			shipments: inventory.ShipmentDAOMock{
				CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
					return shipment, nil
				},
			},
			wantErr: ErrOrderStockUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			svc := testSaleOrderService(tt.orders, tt.lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			svc.quants = tt.quants
			svc.shipments = tt.shipments
			svc.movements = tt.movements
			svc.reservations = tt.reservations
			svc.reservationSvc = inventory.NewHoldService(tt.reservations, tt.quants)

			updated, err := svc.Confirm(ctx, 1)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if tt.wantState != "" && updated.State != tt.wantState {
				t.Errorf("state = %s, want %s", updated.State, tt.wantState)
			}
		})
	}
}

func TestSaleOrderService_Cancel(t *testing.T) {
	OrganizationID := uint64(1)
	tests := []struct {
		name      string
		order     *SaleOrder
		wantErr   error
		wantState string
	}{
		{
			name: "releases reservations and cancels shipment",
			order: &SaleOrder{
				Base: model.Base{ID: 1}, OrganizationID: &OrganizationID,
				Name: helper.Ptr("SO/00001"), State: OrderStateConfirmed,
			},
			wantState: OrderStateCancelled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return tt.order, nil },
			}}
			shipments := inventory.ShipmentDAOMock{
				CRUDMock: dao.CRUDMock[inventory.Shipment]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
						return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{{Base: model.Base{ID: 1}, OrganizationID: &OrganizationID, State: inventory.ShipmentStateConfirmed}}}, nil
					},
					FindFunc: func(_ context.Context, _ uint64) (*inventory.Shipment, error) {
						return &inventory.Shipment{Base: model.Base{ID: 1}, OrganizationID: &OrganizationID, State: inventory.ShipmentStateConfirmed}, nil
					},
				},
			}
			movements := inventory.StockMovementDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
					return []*inventory.StockMovement{{Base: model.Base{ID: 21}, ItemID: 100, State: inventory.MovementStateAssigned}}, nil
				},
			}
			released := make([]uint64, 0)
			reservations := inventory.StockHoldDAOMock{
				ReleaseByMovementFunc: func(_ context.Context, moveID uint64) error {
					released = append(released, moveID)
					return nil
				},
			}

			svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			svc.shipments = shipments
			svc.movements = movements
			svc.reservations = reservations

			updated, err := svc.Cancel(ctx, 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if updated.State != tt.wantState {
				t.Errorf("state = %s, want %s", updated.State, tt.wantState)
			}
			if len(released) != 1 || released[0] != 21 {
				t.Errorf("released = %v, want [21]", released)
			}
		})
	}
}

func TestSaleOrderService_Done(t *testing.T) {
	tests := []struct {
		name    string
		order   *SaleOrder
		wantErr error
	}{
		{
			name:    "rejects when not confirmed",
			order:   &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateDraft},
			wantErr: ErrOrderState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return tt.order, nil },
			}}
			svc := testSaleOrderService(orders, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

			_, err := svc.Done(ctx, 1)
			helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
		})
	}
}

func TestSaleOrderService_RecomputeStatuses(t *testing.T) {
	OrganizationID := uint64(1)
	tests := []struct {
		name         string
		lines        []*SaleOrderLine
		wantDelivery string
		wantInvoice  string
	}{
		{
			name: "from triplet",
			lines: []*SaleOrderLine{
				{OrderID: 1, QtyOrdered: 10, QtyDelivered: 10, QtyInvoiced: 10},
				{OrderID: 1, QtyOrdered: 5, QtyDelivered: 2, QtyInvoiced: 0},
			},
			wantDelivery: DeliveryStatusPartial,
			wantInvoice:  InvoiceStatusToInvoice,
		},
		{
			name: "all done",
			lines: []*SaleOrderLine{
				{OrderID: 1, QtyOrdered: 10, QtyDelivered: 10, QtyInvoiced: 10},
			},
			wantDelivery: DeliveryStatusDone,
			wantInvoice:  InvoiceStatusInvoiced,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			order := &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &OrganizationID, State: OrderStateConfirmed}
			orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
			}}
			lines := SaleOrderLineDAOMock{
				ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
					return tt.lines, nil
				},
			}
			svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

			updated, err := svc.RecomputeStatuses(ctx, 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if updated.DeliveryStatus != tt.wantDelivery {
				t.Errorf("delivery_status = %s, want %s", updated.DeliveryStatus, tt.wantDelivery)
			}
			if updated.InvoiceStatus != tt.wantInvoice {
				t.Errorf("invoice_status = %s, want %s", updated.InvoiceStatus, tt.wantInvoice)
			}
		})
	}
}

func TestSaleOrderService_RecordReturn(t *testing.T) {
	OrganizationID := uint64(1)
	itemID := uint64(100)
	tests := []struct {
		name         string
		itemID       uint64
		qty          float64
		lines        []*SaleOrderLine
		wantErr      error
		wantDelivery string
		wantInvoice  string
	}{
		{
			name:   "increments and recomputes",
			itemID: itemID,
			qty:    4,
			lines: []*SaleOrderLine{
				{OrderID: 1, ItemID: &itemID, QtyOrdered: 10, QtyDelivered: 10, QtyInvoiced: 10},
			},
			wantDelivery: DeliveryStatusPartial,
			wantInvoice:  InvoiceStatusToInvoice,
		},
		{
			name:   "fully returned line is pending",
			itemID: itemID,
			qty:    10,
			lines: []*SaleOrderLine{
				{OrderID: 1, ItemID: &itemID, QtyOrdered: 10, QtyDelivered: 10, QtyInvoiced: 10},
			},
			wantDelivery: DeliveryStatusPending,
			wantInvoice:  InvoiceStatusNo,
		},
		{
			name:    "line not found",
			itemID:  999,
			qty:     1,
			lines:   nil,
			wantErr: ErrLineNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			order := &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &OrganizationID, State: OrderStateConfirmed}
			orders := SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
			}}
			var updatedLine *SaleOrderLine
			lines := SaleOrderLineDAOMock{
				CRUDMock: dao.CRUDMock[SaleOrderLine]{
					UpdateFunc: func(_ context.Context, line *SaleOrderLine) (*SaleOrderLine, error) {
						updatedLine = line
						return line, nil
					},
				},
				ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
					return tt.lines, nil
				},
			}
			svc := testSaleOrderService(orders, lines, dao.CRUDMock[reference.Tax]{}, testPriceService(100))

			err := svc.RecordReturn(ctx, 1, tt.itemID, tt.qty)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if tt.wantDelivery != "" && order.DeliveryStatus != tt.wantDelivery {
				t.Errorf("delivery_status = %s, want %s", order.DeliveryStatus, tt.wantDelivery)
			}
			if tt.wantInvoice != "" && order.InvoiceStatus != tt.wantInvoice {
				t.Errorf("invoice_status = %s, want %s", order.InvoiceStatus, tt.wantInvoice)
			}
			if updatedLine != nil && updatedLine.QtyReturns != tt.qty {
				t.Errorf("qty_returns = %v, want %v", updatedLine.QtyReturns, tt.qty)
			}
		})
	}
}

func testPriceService(listPrice float64) products.ProductService {
	return products.NewProductService(
		products.ItemDAOMock{
			CRUDMock: dao.CRUDMock[products.Item]{
				FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
					return &products.Item{Base: model.Base{ID: 1}, ListPrice: listPrice}, nil
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
		products.PriceBookDAOMock{
			CRUDMock: dao.CRUDMock[products.PriceBook]{
				FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
					return &products.PriceBook{Base: model.Base{ID: 2}, Name: "Retail", CurrencyCode: helper.Ptr("IDR")}, nil
				},
			},
		},
		products.PriceRuleDAOMock{},
	)
}

func testSaleOrderService(orders SaleOrderDAOMock, lines SaleOrderLineDAOMock, taxes dao.CRUDMock[reference.Tax], productSvc products.ProductService) SaleOrderService {
	sequences := sequence.NewSequenceService(SequenceDAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "SO/00001"}, nil
		},
	})
	stages := dao.CRUDMock[reference.PipelineStage]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
			return &reference.PipelineStage{Base: model.Base{ID: 8}, IsWon: true}, nil
		},
	}
	return NewSaleOrderService(
		orders,
		lines,
		sequences,
		productSvc,
		products.PriceBookDAOMock{
			CRUDMock: dao.CRUDMock[products.PriceBook]{
				FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
					return &products.PriceBook{Base: model.Base{ID: 2}, Name: "Retail", CurrencyCode: helper.Ptr("IDR")}, nil
				},
			},
		},
		taxes,
		contacts.ContactDAOMock{
			CRUDMock: dao.CRUDMock[contacts.Contact]{
				FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return &contacts.Contact{Base: model.Base{ID: 5}}, nil
				},
			},
		},
		crm.ProspectDAOMock{
			CRUDMock: dao.CRUDMock[crm.Prospect]{
				FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
					return &crm.Prospect{Base: model.Base{ID: 3}, Type: crm.ProspectKindOpportunity, StageID: helper.Ptr(uint64(8))}, nil
				},
			},
		},
		stages,
		inventory.WarehouseDAOMock{
			CRUDMock: dao.CRUDMock[reference.Warehouse]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
					return &reference.Warehouse{Base: model.Base{ID: 4}, Name: "Main Warehouse", OrganizationID: helper.Ptr(uint64(1))}, nil
				},
			},
		},
		inventory.StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
					if len(q.Filters) > 0 && q.Filters[0].Field == "usage" {
						return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 90}, OrganizationID: helper.Ptr(uint64(1)), Usage: "customer"}}}, nil
					}
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}}}, nil
				},
			},
		},
		inventory.StockBalanceDAOMock{},
		inventory.ShipmentDAOMock{},
		inventory.StockMovementDAOMock{},
		inventory.StockHoldDAOMock{},
		inventory.NewHoldService(inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}),
		ShipEngineMock{},
		InvoiceEngineMock{},
		InvoiceLookupMock{},
		PaymentEngineMock{},
	)
}
