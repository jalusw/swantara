package manufacturing

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
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func receiveTestMocks() (OutsideProcessingOrderDAOMock, ProductionOrderDAOMock, ConsumedMaterialDAOMock, procurement.PurchaseOrderDAOMock, procurement.PurchaseOrderLineDAOMock, inventory.StockMovementDAOMock, inventory.CostLayerDAOMock, inventory.ItemResolverMock, inventory.PosterMock) {
	organizationID := uint64(1)
	dst := uint64(20)
	src := uint64(10)
	order := &OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, SupplierID: 7, PurchaseOrderID: helper.Ptr(uint64(55)), State: OutsideProcessingStateSent}
	outsideOrders := OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
			FindFunc:   func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) { return order, nil },
			UpdateFunc: func(_ context.Context, sub *OutsideProcessingOrder) (*OutsideProcessingOrder, error) { return sub, nil },
		},
	}
	productionOrders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) {
				return &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 200, QtyToProduce: 10, SrcLocationID: &src, DstLocationID: &dst, State: ProductionOrderStatePlanned}, nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	components := ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) {
			return []*ConsumedMaterial{{Base: model.Base{ID: 5}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 10}}, nil
		},
	}
	poDAO := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return &procurement.PurchaseOrder{Base: model.Base{ID: 55}, OrganizationID: &organizationID, AmountUntaxed: 120}, nil
			},
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrderLine]{
			UpdateFunc: func(_ context.Context, l *procurement.PurchaseOrderLine) (*procurement.PurchaseOrderLine, error) {
				return l, nil
			},
		},
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: model.Base{ID: 9}, OrderID: 55, QtyOrdered: 10}}, nil
		},
	}
	movements := inventory.StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[inventory.StockMovement]{
			CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 200
				movement.State = inventory.MovementStateDraft
				return movement, nil
			},
			FindFunc: func(_ context.Context, moveID uint64) (*inventory.StockMovement, error) {
				org := uint64(1)
				return &inventory.StockMovement{Base: model.Base{ID: moveID}, ItemID: 200, Qty: 10, SrcLocationID: src, DstLocationID: dst, State: inventory.MovementStateDraft, OrganizationID: &org}, nil
			},
		},
		ApplyTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
			movement.State = inventory.MovementStateDone
			return movement, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, layer *inventory.CostLayer) (*inventory.CostLayer, error) {
			layer.ID = 90
			return layer, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, layer *inventory.CostLayer) (*inventory.CostLayer, error) {
			return layer, nil
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 400}, StandardCost: 5, Tracking: "none"}, nil
		},
	}
	poster := inventory.PosterMock{
		PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	return outsideOrders, productionOrders, components, poDAO, poLines, movements, layers, resolver, poster
}

func TestOutsideProcessingService_Receive_Failures(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	build := func(mutate func(*OutsideProcessingOrderDAOMock, *ProductionOrderDAOMock, *ConsumedMaterialDAOMock, *procurement.PurchaseOrderDAOMock, *procurement.PurchaseOrderLineDAOMock, *inventory.StockMovementDAOMock, *inventory.CostLayerDAOMock, *inventory.ItemResolverMock, *inventory.PosterMock)) OutsideProcessingService {
		outsideOrders, productionOrders, components, poDAO, poLines, movements, layers, resolver, poster := receiveTestMocks()
		mutate(&outsideOrders, &productionOrders, &components, &poDAO, &poLines, &movements, &layers, &resolver, &poster)
		return testOutsideProcessingService(outsideOrders, productionOrders, components, outsideProcessingRecipeMock(), PurchaseOrderCreatorMock{}, poDAO, poLines, supplierLocationMock(), movements, layers, resolver, poster)
	}
	noop := func(*OutsideProcessingOrderDAOMock, *ProductionOrderDAOMock, *ConsumedMaterialDAOMock, *procurement.PurchaseOrderDAOMock, *procurement.PurchaseOrderLineDAOMock, *inventory.StockMovementDAOMock, *inventory.CostLayerDAOMock, *inventory.ItemResolverMock, *inventory.PosterMock) {
	}

	t.Run("po lookup error and missing", func(t *testing.T) {
		svc := build(func(_ *OutsideProcessingOrderDAOMock, _ *ProductionOrderDAOMock, _ *ConsumedMaterialDAOMock, poDAO *procurement.PurchaseOrderDAOMock, _ *procurement.PurchaseOrderLineDAOMock, _ *inventory.StockMovementDAOMock, _ *inventory.CostLayerDAOMock, _ *inventory.ItemResolverMock, _ *inventory.PosterMock) {
			poDAO.FindFunc = func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) { return nil, dbErr }
		})
		_, err := svc.Receive(ctx, 1, 500, 600, 700, time.Now())
		helper.AssertError(t, err, true, dbErr)

		svc = build(func(_ *OutsideProcessingOrderDAOMock, _ *ProductionOrderDAOMock, _ *ConsumedMaterialDAOMock, poDAO *procurement.PurchaseOrderDAOMock, _ *procurement.PurchaseOrderLineDAOMock, _ *inventory.StockMovementDAOMock, _ *inventory.CostLayerDAOMock, _ *inventory.ItemResolverMock, _ *inventory.PosterMock) {
			poDAO.FindFunc = func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) { return nil, nil }
		})
		_, err = svc.Receive(ctx, 1, 500, 600, 700, time.Now())
		helper.AssertError(t, err, true, ErrOutsideProcessingPurchaseOrder)
	})

	t.Run("po lines error and update error", func(t *testing.T) {
		svc := build(func(_ *OutsideProcessingOrderDAOMock, _ *ProductionOrderDAOMock, _ *ConsumedMaterialDAOMock, _ *procurement.PurchaseOrderDAOMock, poLines *procurement.PurchaseOrderLineDAOMock, _ *inventory.StockMovementDAOMock, _ *inventory.CostLayerDAOMock, _ *inventory.ItemResolverMock, _ *inventory.PosterMock) {
			poLines.ListByOrderFunc = func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) { return nil, dbErr }
		})
		_, err := svc.Receive(ctx, 1, 500, 600, 700, time.Now())
		helper.AssertError(t, err, true, dbErr)

		svc = build(func(_ *OutsideProcessingOrderDAOMock, _ *ProductionOrderDAOMock, _ *ConsumedMaterialDAOMock, _ *procurement.PurchaseOrderDAOMock, poLines *procurement.PurchaseOrderLineDAOMock, _ *inventory.StockMovementDAOMock, _ *inventory.CostLayerDAOMock, _ *inventory.ItemResolverMock, _ *inventory.PosterMock) {
			poLines.UpdateFunc = func(_ context.Context, _ *procurement.PurchaseOrderLine) (*procurement.PurchaseOrderLine, error) {
				return nil, dbErr
			}
		})
		_, err = svc.Receive(ctx, 1, 500, 600, 700, time.Now())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("poster error", func(t *testing.T) {
		svc := build(func(_ *OutsideProcessingOrderDAOMock, _ *ProductionOrderDAOMock, _ *ConsumedMaterialDAOMock, _ *procurement.PurchaseOrderDAOMock, _ *procurement.PurchaseOrderLineDAOMock, _ *inventory.StockMovementDAOMock, _ *inventory.CostLayerDAOMock, _ *inventory.ItemResolverMock, poster *inventory.PosterMock) {
			poster.PostFunc = func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
				return nil, dbErr
			}
		})
		_, err := svc.Receive(ctx, 1, 500, 600, 700, time.Now())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("components and resolver errors", func(t *testing.T) {
		svc := build(func(_ *OutsideProcessingOrderDAOMock, _ *ProductionOrderDAOMock, components *ConsumedMaterialDAOMock, _ *procurement.PurchaseOrderDAOMock, _ *procurement.PurchaseOrderLineDAOMock, _ *inventory.StockMovementDAOMock, _ *inventory.CostLayerDAOMock, _ *inventory.ItemResolverMock, _ *inventory.PosterMock) {
			components.ListByProductionOrderFunc = func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) { return nil, dbErr }
		})
		_, err := svc.Receive(ctx, 1, 500, 600, 700, time.Now())
		helper.AssertError(t, err, true, dbErr)

		svc = build(func(_ *OutsideProcessingOrderDAOMock, _ *ProductionOrderDAOMock, _ *ConsumedMaterialDAOMock, _ *procurement.PurchaseOrderDAOMock, _ *procurement.PurchaseOrderLineDAOMock, _ *inventory.StockMovementDAOMock, _ *inventory.CostLayerDAOMock, resolver *inventory.ItemResolverMock, _ *inventory.PosterMock) {
			resolver.ResolveFunc = func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
				return inventory.ResolvedItem{}, dbErr
			}
		})
		_, err = svc.Receive(ctx, 1, 500, 600, 700, time.Now())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("ledger error", func(t *testing.T) {
		svc := build(func(_ *OutsideProcessingOrderDAOMock, _ *ProductionOrderDAOMock, _ *ConsumedMaterialDAOMock, _ *procurement.PurchaseOrderDAOMock, _ *procurement.PurchaseOrderLineDAOMock, movements *inventory.StockMovementDAOMock, _ *inventory.CostLayerDAOMock, _ *inventory.ItemResolverMock, _ *inventory.PosterMock) {
			movements.CreateFunc = func(_ context.Context, _ *inventory.StockMovement) (*inventory.StockMovement, error) {
				return nil, dbErr
			}
		})
		_, err := svc.Receive(ctx, 1, 500, 600, 700, time.Now())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("productionOrder and order update errors", func(t *testing.T) {
		svc := build(func(_ *OutsideProcessingOrderDAOMock, productionOrders *ProductionOrderDAOMock, _ *ConsumedMaterialDAOMock, _ *procurement.PurchaseOrderDAOMock, _ *procurement.PurchaseOrderLineDAOMock, _ *inventory.StockMovementDAOMock, _ *inventory.CostLayerDAOMock, _ *inventory.ItemResolverMock, _ *inventory.PosterMock) {
			productionOrders.UpdateFunc = func(_ context.Context, _ *ProductionOrder) (*ProductionOrder, error) { return nil, dbErr }
		})
		_, err := svc.Receive(ctx, 1, 500, 600, 700, time.Now())
		helper.AssertError(t, err, true, dbErr)
		_ = noop
	})
}

var _ = reference.StockLocation{}
var _ = query.Query{}
