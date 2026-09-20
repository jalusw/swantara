package sales

import (
	"context"
	"errors"
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

var errAny = errors.New("any error")

func TestSaleOrderService_Create_Ext(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)

	tests := []struct {
		name      string
		setup     func(svc *SaleOrderService)
		order     *SaleOrder
		lines     []*SaleOrderLine
		wantErr   error
		wantState string
		wantField func(t *testing.T, order *SaleOrder)
	}{
		{
			name:    "rejects missing organization",
			order:   &SaleOrder{ContactID: 5},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderContact,
		},
		{
			name: "rejects missing warehouse",
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderWarehouse,
		},
		{
			name: "rejects missing price_book",
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderPriceBook,
		},
		{
			name: "accepts nil CRM lead",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{
					CreateWithLinesFunc: func(_ context.Context, order *SaleOrder, _ []*SaleOrderLine) (*SaleOrder, error) {
						return order, nil
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines: []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantField: func(t *testing.T, order *SaleOrder) {
				if order.ProspectID != nil {
					t.Errorf("prospect_id = %v, want nil", *order.ProspectID)
				}
			},
		},
		{
			name: "rejects zero quantity",
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 0}},
			wantErr: ErrOrderQty,
		},
		{
			name: "rejects discount out of range",
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, DiscountPct: 150}},
			wantErr: ErrOrderDiscount,
		},
		{
			name: "rejects variant missing",
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{QtyOrdered: 2}},
			wantErr: ErrOrderVariant,
		},
		{
			name: "rejects contact not found",
			setup: func(svc *SaleOrderService) {
				svc.contacts = contacts.ContactDAOMock{}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderContact,
		},
		{
			name: "rejects warehouse not found",
			setup: func(svc *SaleOrderService) {
				svc.warehouses = inventory.WarehouseDAOMock{}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderWarehouse,
		},
		{
			name: "rejects price_book not found",
			setup: func(svc *SaleOrderService) {
				svc.price_books = products.PriceBookDAOMock{}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderPriceBook,
		},
		{
			name: "rejects lead not found",
			setup: func(svc *SaleOrderService) {
				svc.leads = crm.ProspectDAOMock{}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderLead,
		},
		{
			name: "rejects lead not opportunity",
			setup: func(svc *SaleOrderService) {
				svc.leads = crm.ProspectDAOMock{
					CRUDMock: dao.CRUDMock[crm.Prospect]{
						FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
							return &crm.Prospect{Base: model.Base{ID: 3}, Type: crm.ProspectKindLead, StageID: helper.Ptr(uint64(8))}, nil
						},
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderLeadNotWon,
		},
		{
			name: "rejects lead stage not found",
			setup: func(svc *SaleOrderService) {
				svc.stages = dao.CRUDMock[reference.PipelineStage]{}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderLeadNotWon,
		},
		{
			name: "rejects stage from another organization",
			setup: func(svc *SaleOrderService) {
				svc.leads = crm.ProspectDAOMock{
					CRUDMock: dao.CRUDMock[crm.Prospect]{
						FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
							return &crm.Prospect{Base: model.Base{ID: 3}, OrganizationID: &organizationID, Type: crm.ProspectKindOpportunity, StageID: helper.Ptr(uint64(8))}, nil
						},
					},
				}
				svc.stages = dao.CRUDMock[reference.PipelineStage]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
						return &reference.PipelineStage{Base: model.Base{ID: 8}, OrganizationID: helper.Ptr(uint64(20)), IsWon: true}, nil
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderLeadNotWon,
		},
		{
			name: "rejects tax not found",
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, TaxIDs: helper.Int64Array{9}}},
			wantErr: ErrOrderTax,
		},
		{
			name: "rejects tax invalid scope",
			setup: func(svc *SaleOrderService) {
				svc.taxes = dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopePurchase, Amount: helper.Ptr(10.0)}, nil
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, TaxIDs: helper.Int64Array{9}}},
			wantErr: ErrOrderTaxInvalid,
		},
		{
			name: "propagates resolve price error",
			setup: func(svc *SaleOrderService) {
				svc.productSvc = products.NewProductService(
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
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "propagates sequence error",
			setup: func(svc *SaleOrderService) {
				svc.sequences = sequence.NewSequenceService(SequenceDAOMock{
					ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
						return nil, errors.New("sequence down")
					},
				})
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "propagates create with lines error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{
					CreateWithLinesFunc: func(_ context.Context, _ *SaleOrder, _ []*SaleOrderLine) (*SaleOrder, error) {
						return nil, errors.New("insert failed")
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "uses provided order date",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{
					CreateWithLinesFunc: func(_ context.Context, order *SaleOrder, _ []*SaleOrderLine) (*SaleOrder, error) {
						return order, nil
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
				OrderDate:      helper.Ptr(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)),
			},
			lines:     []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantState: OrderStateDraft,
		},
		{
			name: "propagates contact error",
			setup: func(svc *SaleOrderService) {
				svc.contacts = contacts.ContactDAOMock{
					CRUDMock: dao.CRUDMock[contacts.Contact]{
						FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
							return nil, errors.New("db down")
						},
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "propagates warehouse error",
			setup: func(svc *SaleOrderService) {
				svc.warehouses = inventory.WarehouseDAOMock{
					CRUDMock: dao.CRUDMock[reference.Warehouse]{
						FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
							return nil, errors.New("db down")
						},
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "propagates price_book error",
			setup: func(svc *SaleOrderService) {
				svc.price_books = products.PriceBookDAOMock{
					CRUDMock: dao.CRUDMock[products.PriceBook]{
						FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
							return nil, errors.New("db down")
						},
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "propagates lead error",
			setup: func(svc *SaleOrderService) {
				svc.leads = crm.ProspectDAOMock{
					CRUDMock: dao.CRUDMock[crm.Prospect]{
						FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
							return nil, errors.New("db down")
						},
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "propagates stage error",
			setup: func(svc *SaleOrderService) {
				svc.stages = dao.CRUDMock[reference.PipelineStage]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
						return nil, errors.New("db down")
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "propagates tax find error",
			setup: func(svc *SaleOrderService) {
				svc.taxes = dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return nil, errors.New("db down")
					},
				}
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, TaxIDs: helper.Int64Array{9}}},
			wantErr: errAny,
		},
		{
			name: "propagates variant organization error",
			setup: func(svc *SaleOrderService) {
				svc.productSvc = products.NewProductService(
					products.ItemDAOMock{},
					products.ItemVariantDAOMock{},
					dao.CRUDMock[reference.ItemCategory]{},
					products.PriceBookDAOMock{},
					products.PriceRuleDAOMock{},
				)
			},
			order: &SaleOrder{
				OrganizationID: &organizationID,
				ContactID:      5,
				PriceBookID:    helper.Ptr(uint64(2)),
				ProspectID:     helper.Ptr(uint64(3)),
				WarehouseID:    helper.Ptr(uint64(4)),
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testSaleOrderService(SaleOrderDAOMock{}, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			if tt.setup != nil {
				tt.setup(&svc)
			}

			order, err := svc.Create(ctx, tt.order, tt.lines)
			if tt.wantErr == errAny {
				if helper.AssertError(t, err, true, nil) {
					return
				}
			} else if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if tt.wantErr != nil {
				return
			}
			if tt.wantState != "" && order.State != tt.wantState {
				t.Errorf("state = %s, want %s", order.State, tt.wantState)
			}
			if tt.wantField != nil {
				tt.wantField(t, order)
			}
		})
	}
}

func TestSaleOrderService_UpdateDraft_Ext(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)

	tests := []struct {
		name    string
		setup   func(svc *SaleOrderService)
		order   *SaleOrder
		lines   []*SaleOrderLine
		wantErr error
		verify  func(t *testing.T, svc SaleOrderService)
	}{
		{
			name: "propagates find error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return nil, errors.New("db down") },
				}}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "rejects not found",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return nil, nil },
				}}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderNotFound,
		},
		{
			name: "rejects non-draft state",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateSent}, nil
					},
				}}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderState,
		},
		{
			name: "rejects empty lines",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateDraft}, nil
					},
				}}
			},
			lines:   nil,
			wantErr: ErrOrderNoLines,
		},
		{
			name: "rejects missing price_book",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateDraft}, nil
					},
				}}
			},
			order:   &SaleOrder{ContactID: 5, WarehouseID: helper.Ptr(uint64(4))},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderPriceBook,
		},
		{
			name: "rejects contact organization mismatch",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
				svc.contacts = contacts.ContactDAOMock{
					CRUDMock: dao.CRUDMock[contacts.Contact]{
						FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
							return &contacts.Contact{Base: model.Base{ID: 5}, OrganizationID: helper.Ptr(uint64(20))}, nil
						},
					},
				}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderContactOrganization,
		},
		{
			name: "rejects warehouse organization mismatch",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
				svc.warehouses = inventory.WarehouseDAOMock{
					CRUDMock: dao.CRUDMock[reference.Warehouse]{
						FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
							return &reference.Warehouse{Base: model.Base{ID: 4}, OrganizationID: helper.Ptr(uint64(20))}, nil
						},
					},
				}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderWarehouseOrganization,
		},
		{
			name: "rejects price_book organization mismatch",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
				svc.price_books = products.PriceBookDAOMock{
					CRUDMock: dao.CRUDMock[products.PriceBook]{
						FindFunc: func(_ context.Context, _ uint64) (*products.PriceBook, error) {
							return &products.PriceBook{Base: model.Base{ID: 2}, OrganizationID: helper.Ptr(uint64(20))}, nil
						},
					},
				}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderPriceBookOrganization,
		},
		{
			name: "rejects variant missing",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
			},
			lines:   []*SaleOrderLine{{QtyOrdered: 2}},
			wantErr: ErrOrderVariant,
		},
		{
			name: "rejects zero quantity",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 0}},
			wantErr: ErrOrderQty,
		},
		{
			name: "rejects discount out of range",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, DiscountPct: -1}},
			wantErr: ErrOrderDiscount,
		},
		{
			name: "replaces lines and updates",
			setup: func(svc *SaleOrderService) {
				existing := &SaleOrder{
					Base: model.Base{ID: 1}, OrganizationID: &organizationID,
					Name: helper.Ptr("SO/00001"), CurrencyCode: helper.Ptr("IDR"), ProspectID: helper.Ptr(uint64(3)),
					State: OrderStateDraft,
				}
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return existing, nil },
					UpdateFunc: func(_ context.Context, order *SaleOrder) (*SaleOrder, error) {
						return order, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ReplaceLinesFunc: func(_ context.Context, _ uint64, _ []*SaleOrderLine) error {
						return nil
					},
				}
				svc.taxes = dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0)}, nil
					},
				}
			},
			lines: []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, TaxIDs: helper.Int64Array{9}}},
			verify: func(t *testing.T, svc SaleOrderService) {
				order, err := svc.UpdateDraft(ctx, 1, &SaleOrder{
					ContactID: 5, PriceBookID: helper.Ptr(uint64(2)), WarehouseID: helper.Ptr(uint64(4)),
					OrderDate: helper.Ptr(time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)),
				}, []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, TaxIDs: helper.Int64Array{9}}})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if order.ID != 1 || order.Name == nil || *order.Name != "SO/00001" || order.State != OrderStateDraft {
					t.Errorf("order = %+v, want preserved draft order", order)
				}
				if order.AmountUntaxed != 200 || order.AmountTax != 20 || order.AmountTotal != 220 {
					t.Errorf("amounts = %v/%v/%v, want 200/20/220", order.AmountUntaxed, order.AmountTax, order.AmountTotal)
				}
			},
		},
		{
			name: "propagates replace lines error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ReplaceLinesFunc: func(_ context.Context, _ uint64, _ []*SaleOrderLine) error {
						return errors.New("replace failed")
					},
				}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "propagates update error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
					UpdateFunc: func(_ context.Context, _ *SaleOrder) (*SaleOrder, error) {
						return nil, errors.New("update failed")
					},
				}}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "rejects variant organization mismatch",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
				svc.productSvc = products.NewProductService(
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
					products.PriceBookDAOMock{},
					products.PriceRuleDAOMock{},
				)
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: ErrOrderVariantOrganization,
		},
		{
			name: "propagates resolve price error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
				svc.productSvc = products.NewProductService(
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
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}},
			wantErr: errAny,
		},
		{
			name: "propagates tax find error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateDraft}, nil
					},
				}}
				svc.taxes = dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return nil, errors.New("db down")
					},
				}
			},
			lines:   []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2, TaxIDs: helper.Int64Array{9}}},
			wantErr: errAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testSaleOrderService(SaleOrderDAOMock{}, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			if tt.setup != nil {
				tt.setup(&svc)
			}

			if tt.verify != nil {
				tt.verify(t, svc)
				return
			}

			order := tt.order
			if order == nil {
				order = &SaleOrder{ContactID: 5, PriceBookID: helper.Ptr(uint64(2)), WarehouseID: helper.Ptr(uint64(4))}
			}
			_, err := svc.UpdateDraft(ctx, 1, order, tt.lines)
			if tt.wantErr == errAny {
				helper.AssertError(t, err, true, nil)
			} else {
				helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			}
		})
	}
}

func TestSaleOrderService_Send_Ext(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(svc *SaleOrderService)
		wantErr error
	}{
		{
			name: "rejects not found",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return nil, nil },
				}}
			},
			wantErr: ErrOrderNotFound,
		},
		{
			name: "propagates find error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return nil, errors.New("db down") },
				}}
			},
			wantErr: errAny,
		},
		{
			name: "propagates update error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateDraft}, nil
					},
					UpdateFunc: func(_ context.Context, _ *SaleOrder) (*SaleOrder, error) {
						return nil, errors.New("update failed")
					},
				}}
			},
			wantErr: errAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testSaleOrderService(SaleOrderDAOMock{}, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			if tt.setup != nil {
				tt.setup(&svc)
			}

			_, err := svc.Send(ctx, 1)
			if tt.wantErr == errAny {
				helper.AssertError(t, err, true, nil)
			} else {
				helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			}
		})
	}
}

func TestSaleOrderService_Confirm_Ext(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)

	tests := []struct {
		name    string
		setup   func(svc *SaleOrderService)
		wantErr error
		verify  func(t *testing.T, svc SaleOrderService)
	}{
		{
			name: "rejects not found",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return nil, nil },
				}}
			},
			wantErr: ErrOrderNotFound,
		},
		{
			name: "rejects when cancelled",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateCancelled}, nil
					},
				}}
			},
			wantErr: ErrOrderState,
		},
		{
			name: "rejects missing warehouse",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateSent}, nil
					},
				}}
			},
			wantErr: ErrOrderWarehouse,
		},
		{
			name: "propagates list by order error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4))}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return nil, errors.New("db down")
					},
				}
			},
			wantErr: errAny,
		},
		{
			name: "rejects no lines",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4))}, nil
					},
				}}
			},
			wantErr: ErrOrderNoLines,
		},
		{
			name: "rejects missing stock location",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4))}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, Usage: "receipt"}}}, nil
						},
					},
				}
			},
			wantErr: ErrOrderLocation,
		},
		{
			name: "rejects missing customer location",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4))}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if len(q.Filters) > 0 && q.Filters[0].Field == "usage" {
								return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{}}, nil
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, OrganizationID: &organizationID, Usage: "internal"}}}, nil
						},
					},
				}
			},
			wantErr: ErrOrderCustomerLocation,
		},
		{
			name: "propagates find by key error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4))}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				svc.quants = inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return nil, errors.New("db down")
					},
				}
			},
			wantErr: errAny,
		},
		{
			name: "propagates create with movements error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4))}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				svc.quants = inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, OrganizationID: &organizationID, Quantity: 10}, nil
					},
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CreateWithMovementsFunc: func(_ context.Context, _ *inventory.Shipment, _ []*inventory.StockMovement) (*inventory.Shipment, error) {
						return nil, errors.New("insert failed")
					},
				}
			},
			wantErr: errAny,
		},
		{
			name: "rolls back and cancels on reserve error",
			setup: func(svc *SaleOrderService) {
				order := &SaleOrder{
					Base: model.Base{ID: 1}, OrganizationID: &organizationID,
					Name: helper.Ptr("SO/00001"), State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4)),
				}
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{
							{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5},
							{Base: model.Base{ID: 8}, ItemID: helper.Ptr(uint64(101)), QtyOrdered: 5},
						}, nil
					},
				}
				svc.quants = inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, OrganizationID: &organizationID, Quantity: 10}, nil
					},
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
						shipment.ID = 1
						for i := range movements {
							movements[i].ID = uint64(20 + i)
						}
						return shipment, nil
					},
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						FindFunc: func(_ context.Context, _ uint64) (*inventory.Shipment, error) {
							return &inventory.Shipment{Base: model.Base{ID: 1}, State: inventory.ShipmentStateConfirmed}, nil
						},
					},
				}
				svc.movements = inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{
							{Base: model.Base{ID: 20}, ItemID: 100, Qty: 5, State: inventory.MovementStateAssigned},
							{Base: model.Base{ID: 21}, ItemID: 101, Qty: 5, State: inventory.MovementStateAssigned},
						}, nil
					},
				}
				reserveCalls := 0
				released := make([]uint64, 0)
				svc.reservations = inventory.StockHoldDAOMock{
					ReserveFunc: func(_ context.Context, _ uint64, moveID *uint64, _ float64) (*inventory.StockHold, error) {
						reserveCalls++
						if reserveCalls == 2 {
							return nil, errors.New("reserve failed")
						}
						return &inventory.StockHold{Base: model.Base{ID: uint64(reserveCalls)}, BalanceID: 3, MovementID: moveID, Qty: 5}, nil
					},
					ReleaseByMovementFunc: func(_ context.Context, moveID uint64) error {
						released = append(released, moveID)
						return nil
					},
				}
				svc.reservationSvc = inventory.NewHoldService(svc.reservations, svc.quants)
			},
			wantErr: errAny,
		},
		{
			name: "rolls back on movement update error",
			setup: func(svc *SaleOrderService) {
				order := &SaleOrder{
					Base: model.Base{ID: 1}, OrganizationID: &organizationID,
					Name: helper.Ptr("SO/00001"), State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4)),
				}
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				svc.quants = inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, OrganizationID: &organizationID, Quantity: 10}, nil
					},
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
						shipment.ID = 1
						for i := range movements {
							movements[i].ID = uint64(20 + i)
						}
						return shipment, nil
					},
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						FindFunc: func(_ context.Context, _ uint64) (*inventory.Shipment, error) {
							return &inventory.Shipment{Base: model.Base{ID: 1}, State: inventory.ShipmentStateConfirmed}, nil
						},
					},
				}
				svc.movements = inventory.StockMovementDAOMock{
					CRUDMock: dao.CRUDMock[inventory.StockMovement]{
						UpdateFunc: func(_ context.Context, _ *inventory.StockMovement) (*inventory.StockMovement, error) {
							return nil, errors.New("update failed")
						},
					},
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{Base: model.Base{ID: 20}, ItemID: 100, Qty: 5, State: inventory.MovementStateAssigned}}, nil
					},
				}
				released := make([]uint64, 0)
				svc.reservations = inventory.StockHoldDAOMock{
					ReserveFunc: func(_ context.Context, _ uint64, moveID *uint64, _ float64) (*inventory.StockHold, error) {
						return &inventory.StockHold{Base: model.Base{ID: 1}, BalanceID: 3, MovementID: moveID, Qty: 5}, nil
					},
					ReleaseByMovementFunc: func(_ context.Context, moveID uint64) error {
						released = append(released, moveID)
						return nil
					},
				}
				svc.reservationSvc = inventory.NewHoldService(svc.reservations, svc.quants)
			},
			wantErr: errAny,
		},
		{
			name: "rolls back on shipment update error",
			setup: func(svc *SaleOrderService) {
				order := &SaleOrder{
					Base: model.Base{ID: 1}, OrganizationID: &organizationID,
					Name: helper.Ptr("SO/00001"), State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4)),
				}
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				svc.quants = inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, OrganizationID: &organizationID, Quantity: 10}, nil
					},
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
						shipment.ID = 1
						for i := range movements {
							movements[i].ID = uint64(20 + i)
						}
						return shipment, nil
					},
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						FindFunc: func(_ context.Context, _ uint64) (*inventory.Shipment, error) {
							return &inventory.Shipment{Base: model.Base{ID: 1}, State: inventory.ShipmentStateConfirmed}, nil
						},
						UpdateFunc: func(_ context.Context, _ *inventory.Shipment) (*inventory.Shipment, error) {
							return nil, errors.New("update failed")
						},
					},
				}
				svc.movements = inventory.StockMovementDAOMock{
					ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
						return []*inventory.StockMovement{{Base: model.Base{ID: 20}, ItemID: 100, Qty: 5, State: inventory.MovementStateAssigned}}, nil
					},
				}
				released := make([]uint64, 0)
				svc.reservations = inventory.StockHoldDAOMock{
					ReserveFunc: func(_ context.Context, _ uint64, moveID *uint64, _ float64) (*inventory.StockHold, error) {
						return &inventory.StockHold{Base: model.Base{ID: 1}, BalanceID: 3, MovementID: moveID, Qty: 5}, nil
					},
					ReleaseByMovementFunc: func(_ context.Context, moveID uint64) error {
						released = append(released, moveID)
						return nil
					},
				}
				svc.reservationSvc = inventory.NewHoldService(svc.reservations, svc.quants)
			},
			wantErr: errAny,
		},
		{
			name: "skips lines without item",
			setup: func(svc *SaleOrderService) {
				order := &SaleOrder{
					Base: model.Base{ID: 1}, OrganizationID: &organizationID,
					Name: helper.Ptr("SO/00001"), State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4)),
				}
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return order, nil },
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{
							{Base: model.Base{ID: 7}, QtyOrdered: 5},
							{Base: model.Base{ID: 8}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5},
						}, nil
					},
				}
				svc.quants = inventory.StockBalanceDAOMock{
					FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
						return &inventory.StockBalance{Base: model.Base{ID: 3}, OrganizationID: &organizationID, Quantity: 10}, nil
					},
				}
				svc.shipments = inventory.ShipmentDAOMock{
					CreateWithMovementsFunc: func(_ context.Context, shipment *inventory.Shipment, movements []*inventory.StockMovement) (*inventory.Shipment, error) {
						shipment.ID = 1
						for i := range movements {
							movements[i].ID = uint64(20 + i)
						}
						return shipment, nil
					},
				}
				svc.reservations = inventory.StockHoldDAOMock{
					ReserveFunc: func(_ context.Context, _ uint64, moveID *uint64, _ float64) (*inventory.StockHold, error) {
						return &inventory.StockHold{Base: model.Base{ID: 1}, BalanceID: 3, MovementID: moveID, Qty: 5}, nil
					},
				}
				svc.reservationSvc = inventory.NewHoldService(svc.reservations, svc.quants)
			},
			verify: func(t *testing.T, svc SaleOrderService) {
				updated, err := svc.Confirm(ctx, 1)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updated.State != OrderStateConfirmed {
					t.Errorf("state = %s, want confirmed", updated.State)
				}
			},
		},
		{
			name: "propagates stock location list error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4))}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return nil, errors.New("db down")
						},
					},
				}
			},
			wantErr: errAny,
		},
		{
			name: "propagates customer location list error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateSent, WarehouseID: helper.Ptr(uint64(4))}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 5}}, nil
					},
				}
				svc.locations = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
							if len(q.Filters) > 0 && q.Filters[0].Field == "usage" {
								return nil, errors.New("db down")
							}
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, Usage: "internal"}}}, nil
						},
					},
				}
			},
			wantErr: errAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testSaleOrderService(SaleOrderDAOMock{}, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			if tt.setup != nil {
				tt.setup(&svc)
			}

			if tt.verify != nil {
				tt.verify(t, svc)
				return
			}

			_, err := svc.Confirm(ctx, 1)
			if tt.wantErr == errAny {
				helper.AssertError(t, err, true, nil)
			} else {
				helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			}
		})
	}
}

func TestSaleOrderService_Cancel_Ext(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(svc *SaleOrderService)
		wantErr error
	}{
		{
			name: "rejects not found",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return nil, nil },
				}}
			},
			wantErr: ErrOrderNotFound,
		},
		{
			name: "rejects when done",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateDone}, nil
					},
				}}
			},
			wantErr: ErrOrderState,
		},
		{
			name: "rejects missing name",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed}, nil
					},
				}}
			},
			wantErr: ErrOrderNotFound,
		},
		{
			name: "propagates list error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed, Name: helper.Ptr("SO/00001")}, nil
					},
				}}
				svc.shipments = inventory.ShipmentDAOMock{
					CRUDMock: dao.CRUDMock[inventory.Shipment]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
							return nil, errors.New("db down")
						},
					},
				}
			},
			wantErr: errAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testSaleOrderService(SaleOrderDAOMock{}, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			if tt.setup != nil {
				tt.setup(&svc)
			}

			_, err := svc.Cancel(ctx, 1)
			if tt.wantErr == errAny {
				helper.AssertError(t, err, true, nil)
			} else {
				helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			}
		})
	}
}

func TestSaleOrderService_Done_Ext(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setup     func(svc *SaleOrderService)
		wantErr   error
		wantState string
	}{
		{
			name: "rejects not found",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return nil, nil },
				}}
			},
			wantErr: ErrOrderNotFound,
		},
		{
			name: "transitions to done",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed}, nil
					},
				}}
			},
			wantState: OrderStateDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testSaleOrderService(SaleOrderDAOMock{}, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			if tt.setup != nil {
				tt.setup(&svc)
			}

			updated, err := svc.Done(ctx, 1)
			if helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr) {
				return
			}
			if tt.wantState != "" && updated.State != tt.wantState {
				t.Errorf("state = %s, want %s", updated.State, tt.wantState)
			}
		})
	}
}

func TestSaleOrderService_ListLines_Ext(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		setup     func(svc *SaleOrderService)
		wantCount int
		wantQty   float64
	}{
		{
			name: "returns lines",
			setup: func(svc *SaleOrderService) {
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, QtyOrdered: 5}}, nil
					},
				}
			},
			wantCount: 1,
			wantQty:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testSaleOrderService(SaleOrderDAOMock{}, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			if tt.setup != nil {
				tt.setup(&svc)
			}

			items, err := svc.ListLines(ctx, 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != tt.wantCount || items[0].QtyOrdered != tt.wantQty {
				t.Errorf("items = %+v, want %d lines with qty %v", items, tt.wantCount, tt.wantQty)
			}
		})
	}
}

func TestSaleOrderService_RecomputeStatuses_Ext(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(svc *SaleOrderService)
		wantErr error
	}{
		{
			name: "propagates list by order error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return nil, errors.New("db down")
					},
				}
			},
			wantErr: errAny,
		},
		{
			name: "propagates find error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return nil, errors.New("db down")
					},
				}}
			},
			wantErr: errAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testSaleOrderService(SaleOrderDAOMock{}, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			if tt.setup != nil {
				tt.setup(&svc)
			}

			_, err := svc.RecomputeStatuses(ctx, 1)
			if tt.wantErr == errAny {
				helper.AssertError(t, err, true, nil)
			} else {
				helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			}
		})
	}
}

func TestSaleOrderService_RecordReturn_Ext(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	itemID := uint64(100)

	tests := []struct {
		name    string
		setup   func(svc *SaleOrderService)
		wantErr error
	}{
		{
			name: "rejects not found",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) { return nil, nil },
				}}
			},
			wantErr: ErrOrderNotFound,
		},
		{
			name: "propagates list by order error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateConfirmed}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return nil, errors.New("db down")
					},
				}
			},
			wantErr: errAny,
		},
		{
			name: "propagates line update error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateConfirmed}, nil
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					CRUDMock: dao.CRUDMock[SaleOrderLine]{
						UpdateFunc: func(_ context.Context, _ *SaleOrderLine) (*SaleOrderLine, error) {
							return nil, errors.New("update failed")
						},
					},
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: &itemID, QtyOrdered: 10}}, nil
					},
				}
			},
			wantErr: errAny,
		},
		{
			name: "propagates order update error",
			setup: func(svc *SaleOrderService) {
				svc.orders = SaleOrderDAOMock{CRUDMock: dao.CRUDMock[SaleOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*SaleOrder, error) {
						return &SaleOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: OrderStateConfirmed}, nil
					},
					UpdateFunc: func(_ context.Context, _ *SaleOrder) (*SaleOrder, error) {
						return nil, errors.New("update failed")
					},
				}}
				svc.lines = SaleOrderLineDAOMock{
					ListByOrderFunc: func(_ context.Context, _ uint64) ([]*SaleOrderLine, error) {
						return []*SaleOrderLine{{Base: model.Base{ID: 7}, ItemID: &itemID, QtyOrdered: 10}}, nil
					},
				}
			},
			wantErr: errAny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testSaleOrderService(SaleOrderDAOMock{}, SaleOrderLineDAOMock{}, dao.CRUDMock[reference.Tax]{}, testPriceService(100))
			if tt.setup != nil {
				tt.setup(&svc)
			}

			err := svc.RecordReturn(ctx, 1, itemID, 4)
			if tt.wantErr == errAny {
				helper.AssertError(t, err, true, nil)
			} else {
				helper.AssertError(t, err, tt.wantErr != nil, tt.wantErr)
			}
		})
	}
}

func TestSaleOrderService_WarehouseStockLocation(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)

	tests := []struct {
		name       string
		setup      func(locations *inventory.StockLocationDAOMock)
		wantID     uint64
		wantErr    bool
		wantErrMsg error
	}{
		{
			name: "prefers organization location",
			setup: func(locs *inventory.StockLocationDAOMock) {
				otherOrgID := uint64(2)
				*locs = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
								{Base: model.Base{ID: 7}, Usage: "internal"},
								{Base: model.Base{ID: 8}, Usage: "internal", OrganizationID: &otherOrgID},
								{Base: model.Base{ID: 9}, Usage: "internal", OrganizationID: &organizationID},
							}}, nil
						},
					},
				}
			},
			wantID: 9,
		},
		{
			name: "rejects global location",
			setup: func(locs *inventory.StockLocationDAOMock) {
				*locs = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
								{Base: model.Base{ID: 7}, Usage: "internal"},
							}}, nil
						},
					},
				}
			},
			wantErr:    true,
			wantErrMsg: ErrOrderLocation,
		},
		{
			name: "rejects foreign organization location",
			setup: func(locs *inventory.StockLocationDAOMock) {
				otherOrgID := uint64(2)
				*locs = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
								{Base: model.Base{ID: 8}, Usage: "internal", OrganizationID: &otherOrgID},
							}}, nil
						},
					},
				}
			},
			wantErr:    true,
			wantErrMsg: ErrOrderLocation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locs := inventory.StockLocationDAOMock{}
			if tt.setup != nil {
				tt.setup(&locs)
			}
			svc := SaleOrderService{locations: locs}

			locationID, err := svc.warehouseStockLocation(ctx, &organizationID, 4)
			if helper.AssertError(t, err, tt.wantErr, tt.wantErrMsg) {
				return
			}
			if locationID != tt.wantID {
				t.Fatalf("expected location %d, got %d", tt.wantID, locationID)
			}
		})
	}
}

func TestSaleOrderService_CustomerLocation(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)

	tests := []struct {
		name    string
		setup   func(locations *inventory.StockLocationDAOMock)
		wantID  uint64
		wantErr bool
	}{
		{
			name: "prefers organization location",
			setup: func(locs *inventory.StockLocationDAOMock) {
				*locs = inventory.StockLocationDAOMock{
					CRUDMock: dao.CRUDMock[reference.StockLocation]{
						ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
							return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
								{Base: model.Base{ID: 11}, Usage: "customer"},
								{Base: model.Base{ID: 12}, Usage: "customer", OrganizationID: &organizationID},
							}}, nil
						},
					},
				}
			},
			wantID: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locs := inventory.StockLocationDAOMock{}
			if tt.setup != nil {
				tt.setup(&locs)
			}
			svc := SaleOrderService{locations: locs}

			locationID, err := svc.customerLocation(ctx, &organizationID)
			if helper.AssertError(t, err, tt.wantErr, nil) {
				return
			}
			if locationID != tt.wantID {
				t.Fatalf("expected location %d, got %d", tt.wantID, locationID)
			}
		})
	}
}
