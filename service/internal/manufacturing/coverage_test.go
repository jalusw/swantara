package manufacturing

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

func TestManufacturingDelegations(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("recipe service", func(t *testing.T) {
		svc := NewRecipeService(products.ItemVariantDAOMock{}, RecipeDAOMock{}, RecipeLineDAOMock{})
		if _, err := svc.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := svc.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if err := svc.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if _, err := svc.ListLines(ctx, 1); err != nil {
			t.Errorf("ListLines = %v", err)
		}
	})

	t.Run("productionOrder service", func(t *testing.T) {
		svc := NewTestProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, sequence.DAOMock{})
		if _, err := svc.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := svc.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := svc.ListComponents(ctx, 1); err != nil {
			t.Errorf("ListComponents = %v", err)
		}
	})

	t.Run("planning service", func(t *testing.T) {
		ledger := inventory.NewLedgerService(inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.TransactionerMock{})
		reorder := inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, ledger, inventory.ItemResolverMock{})
		svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, PlannedSupplyDAOMock{}, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, reorder, ledger, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, nil, nil, nil)
		if _, err := svc.ListRuns(ctx, q); err != nil {
			t.Errorf("ListRuns = %v", err)
		}
		if _, err := svc.FindRun(ctx, 1); err != nil {
			t.Errorf("FindRun = %v", err)
		}
		if _, err := svc.ListDemands(ctx, 1); err != nil {
			t.Errorf("ListDemands = %v", err)
		}
		if _, err := svc.ListPlanned(ctx, 1); err != nil {
			t.Errorf("ListPlanned = %v", err)
		}
		if _, err := svc.CreateForecast(ctx, &DemandPlan{}); err != nil {
			t.Errorf("CreateForecast = %v", err)
		}
		if _, err := svc.ListForecasts(ctx, q); err != nil {
			t.Errorf("ListForecasts = %v", err)
		}
	})

	t.Run("production service", func(t *testing.T) {
		svc := NewTestProductionService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, sequence.DAOMock{})
		if _, err := svc.ListShopTasksByMO(ctx, 1); err != nil {
			t.Errorf("ListShopTasksByMO = %v", err)
		}
	})
}
