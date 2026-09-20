package manufacturing

import (
	"context"
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
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func testProductionService(
	orders ProductionOrderDAOMock,
	components ConsumedMaterialDAOMock,
	workOrders ShopTaskDAOMock,
	operations ProductionStepDAOMock,
	workCenters dao.CRUDMock[reference.WorkCenter],
	locations inventory.StockLocationDAOMock,
	movements inventory.StockMovementDAOMock,
	layers inventory.CostLayerDAOMock,
	resolver inventory.ItemResolverMock,
	poster inventory.PosterMock,
	lines accounting.JournalLineDAOMock,
	sequences sequence.DAOMock,
) ProductionService {
	return NewTestProductionService(orders, components, workOrders, operations, workCenters, locations, movements, layers, resolver, poster, lines, sequences)
}

func productionLocationMock() inventory.StockLocationDAOMock {
	return inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 30}, OrganizationID: helper.Ptr(uint64(1)), Usage: "production"}, nil
			},
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field != "usage" {
					return &query.Page[reference.StockLocation]{}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 30}, OrganizationID: helper.Ptr(uint64(1)), Usage: "production"}}}, nil
			},
		},
	}
}

func inProgressOrderMock(organizationID, src, dst uint64) ProductionOrderDAOMock {
	return ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) {
				return &ProductionOrder{
					Base:              model.Base{ID: 1},
					OrganizationID:    &organizationID,
					ItemID:            200,
					RecipeID:          helper.Ptr(uint64(1)),
					QtyToProduce:      10,
					QtyProduced:       0,
					SrcLocationID:     &src,
					DstLocationID:     &dst,
					State:             ProductionOrderStateInProgress,
					DatePlannedStart:  helper.Ptr(time.Now().Add(24 * time.Hour)),
					DatePlannedFinish: helper.Ptr(time.Now().Add(72 * time.Hour)),
				}, nil
			},
		},
	}
}

func TestProductionService_Start_MovesPlannedToInProgress(t *testing.T) {
	ctx := context.Background()
	order := &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStatePlanned}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
			UpdateFunc: func(_ context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	svc := testProductionService(orders, ConsumedMaterialDAOMock{}, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, moSequenceMock())

	started, err := svc.Start(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if started.State != ProductionOrderStateInProgress {
		t.Errorf("state = %s, want in_progress", started.State)
	}
	if started.DateStart == nil {
		t.Error("date_start is nil, want set")
	}
}

func TestProductionService_Start_RejectsNonPlanned(t *testing.T) {
	ctx := context.Background()
	order := &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateDraft}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
		},
	}
	svc := testProductionService(orders, ConsumedMaterialDAOMock{}, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, moSequenceMock())

	_, err := svc.Start(ctx, 1)
	helper.AssertError(t, err, true, ErrProductionOrderState)
}

func TestProductionService_GenerateShopTasks_CreatesFromOperations(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	orders := inProgressOrderMock(organizationID, 10, 20)
	var created []*ShopTask
	workOrders := ShopTaskDAOMock{
		CRUDMock: dao.CRUDMock[ShopTask]{
			CreateFunc: func(_ context.Context, workOrder *ShopTask) (*ShopTask, error) {
				workOrder.ID = uint64(len(created) + 1)
				created = append(created, workOrder)
				return workOrder, nil
			},
		},
	}
	operations := ProductionStepDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*ProductionStep, error) {
			return []*ProductionStep{
				{Base: model.Base{ID: 11}, Sequence: 10, SetupMinutes: 15, TimeMinutes: 45, WorkCenterID: helper.Ptr(uint64(7))},
				{Base: model.Base{ID: 12}, Sequence: 20, SetupMinutes: 5, TimeMinutes: 25, WorkCenterID: helper.Ptr(uint64(8))},
			}, nil
		},
	}
	workCenters := dao.CRUDMock[reference.WorkCenter]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.WorkCenter, error) {
			return &reference.WorkCenter{Base: model.Base{ID: 7}, Name: "Assembly"}, nil
		},
	}
	svc := testProductionService(orders, ConsumedMaterialDAOMock{}, workOrders, operations, workCenters, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, DefaultShopTaskSequenceMock())

	generated, err := svc.GenerateShopTasks(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(generated) != 2 {
		t.Fatalf("shop tasks = %d, want 2", len(generated))
	}
	if generated[0].Sequence != 10 || generated[0].PlannedMinutes != 60 {
		t.Errorf("first shop task sequence/minutes = %d/%v, want 10/60", generated[0].Sequence, generated[0].PlannedMinutes)
	}
	if generated[0].State != ShopTaskStatePlanned {
		t.Errorf("state = %s, want planned", generated[0].State)
	}
	if generated[1].PlannedMinutes != 30 {
		t.Errorf("second shop task minutes = %v, want 30", generated[1].PlannedMinutes)
	}
}

func TestProductionService_GenerateShopTasks_RejectsMissingWorkCenter(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	orders := inProgressOrderMock(organizationID, 10, 20)
	operations := ProductionStepDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*ProductionStep, error) {
			return []*ProductionStep{{Base: model.Base{ID: 11}, WorkCenterID: nil}}, nil
		},
	}
	svc := testProductionService(orders, ConsumedMaterialDAOMock{}, ShopTaskDAOMock{}, operations, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, moSequenceMock())

	_, err := svc.GenerateShopTasks(ctx, 1)
	helper.AssertError(t, err, true, ErrShopTaskWorkCenter)
}

func TestProductionService_Consume_MovesRawToProductionAndTracks(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)
	dst := uint64(20)
	orders := inProgressOrderMock(organizationID, src, dst)
	component := &ConsumedMaterial{Base: model.Base{ID: 5}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 22, QtyConsumed: 0}
	components := ConsumedMaterialDAOMock{
		CRUDMock: dao.CRUDMock[ConsumedMaterial]{
			FindFunc: func(_ context.Context, _ uint64) (*ConsumedMaterial, error) { return component, nil },
			UpdateFunc: func(_ context.Context, comp *ConsumedMaterial) (*ConsumedMaterial, error) {
				return comp, nil
			},
		},
	}
	movements := inventory.StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[inventory.StockMovement]{
			CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 100
				movement.State = inventory.MovementStateDraft
				return movement, nil
			},
			FindFunc: func(_ context.Context, moveID uint64) (*inventory.StockMovement, error) {
				return &inventory.StockMovement{Base: model.Base{ID: moveID}, ItemID: 300, Qty: 2, SrcLocationID: src, DstLocationID: 30, State: inventory.MovementStateDraft, OrganizationID: &organizationID}, nil
			},
		},
		ApplyTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
			movement.State = inventory.MovementStateDone
			return movement, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: model.Base{ID: 1}, Quantity: 50, UnitCost: helper.Ptr(2.0), RemainingQty: 50, RemainingValue: 100}}, nil
		},
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
			return inventory.ResolvedItem{
				StockAccounts: inventory.StockAccounts{StockValuationAccountID: 400, StockInputAccountID: 401, CogsAccountID: 402, StockOutputAccountID: 403},
				Tracking:      "none",
			}, nil
		},
	}
	var posted []accounting.PostRequest
	poster := inventory.PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = append(posted, request)
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := testProductionService(orders, components, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), movements, layers, resolver, poster, accounting.JournalLineDAOMock{}, moSequenceMock())

	consumed, err := svc.Consume(ctx, 1, 5, 2, 500, 600, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if consumed.QtyConsumed != 2 {
		t.Errorf("qty_consumed = %v, want 2", consumed.QtyConsumed)
	}
	if consumed.StockMovementID == nil || *consumed.StockMovementID != 100 {
		t.Errorf("stock_movement_id = %v, want 100", consumed.StockMovementID)
	}
	if len(posted) != 1 {
		t.Fatalf("postings = %d, want 1", len(posted))
	}
	if posted[0].Lines[0].AccountID != 600 || posted[0].Lines[1].AccountID != 400 {
		t.Errorf("consume lines accounts = %d/%d, want 600/400", posted[0].Lines[0].AccountID, posted[0].Lines[1].AccountID)
	}
}

func TestProductionService_Consume_RejectsInvalidInputs(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)

	tests := []struct {
		name       string
		order      *ProductionOrder
		component  *ConsumedMaterial
		qty        float64
		wipAccount uint64
		wantErr    error
	}{
		{name: "zero quantity", order: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateInProgress, OrganizationID: &organizationID}, component: &ConsumedMaterial{Base: model.Base{ID: 5}, ProductionOrderID: 1, QtyPlanned: 10}, qty: 0, wantErr: ErrConsumeQuantity},
		{name: "missing wip account", order: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateInProgress, OrganizationID: &organizationID}, component: &ConsumedMaterial{Base: model.Base{ID: 5}, ProductionOrderID: 1, QtyPlanned: 10}, qty: 1, wipAccount: 0, wantErr: ErrWIPAccount},
		{name: "wrong state", order: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStatePlanned, OrganizationID: &organizationID}, component: &ConsumedMaterial{Base: model.Base{ID: 5}, ProductionOrderID: 1, QtyPlanned: 10}, qty: 1, wipAccount: 600, wantErr: ErrProductionOrderState},
		{name: "component not on productionOrder", order: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateInProgress, OrganizationID: &organizationID, SrcLocationID: helper.Ptr(uint64(10))}, component: &ConsumedMaterial{Base: model.Base{ID: 5}, ProductionOrderID: 99, QtyPlanned: 10}, qty: 1, wipAccount: 600, wantErr: ErrConsumeComponent},
		{name: "quantity exceeds remaining", order: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateInProgress, OrganizationID: &organizationID, SrcLocationID: helper.Ptr(uint64(10))}, component: &ConsumedMaterial{Base: model.Base{ID: 5}, ProductionOrderID: 1, QtyPlanned: 10, QtyConsumed: 9}, qty: 2, wipAccount: 600, wantErr: ErrConsumeQuantity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders := ProductionOrderDAOMock{
				CRUDMock: dao.CRUDMock[ProductionOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return tt.order, nil },
				},
			}
			components := ConsumedMaterialDAOMock{
				CRUDMock: dao.CRUDMock[ConsumedMaterial]{
					FindFunc: func(_ context.Context, _ uint64) (*ConsumedMaterial, error) { return tt.component, nil },
				},
			}
			svc := testProductionService(orders, components, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, moSequenceMock())

			_, err := svc.Consume(ctx, 1, 5, tt.qty, 500, tt.wipAccount, time.Now())
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProductionService_Produce_ValuesFinishedAtStandard(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	dst := uint64(20)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 200, QtyToProduce: 10, QtyProduced: 0, DstLocationID: &dst, State: ProductionOrderStateInProgress}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
			UpdateFunc: func(_ context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
				return productionOrder, nil
			},
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
				return &inventory.StockMovement{Base: model.Base{ID: moveID}, ItemID: 200, Qty: 4, SrcLocationID: 30, DstLocationID: dst, State: inventory.MovementStateDraft, OrganizationID: &organizationID}, nil
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
			return inventory.ResolvedItem{
				StockAccounts: inventory.StockAccounts{StockValuationAccountID: 400, StockInputAccountID: 401, CogsAccountID: 402, StockOutputAccountID: 403},
				Tracking:      "none",
				StandardCost:  5,
			}, nil
		},
	}
	var posted []accounting.PostRequest
	poster := inventory.PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = append(posted, request)
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := testProductionService(orders, ConsumedMaterialDAOMock{}, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), movements, layers, resolver, poster, accounting.JournalLineDAOMock{}, moSequenceMock())

	produced, err := svc.Produce(ctx, 1, 4, 500, 600, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if produced.QtyProduced != 4 {
		t.Errorf("qty_produced = %v, want 4", produced.QtyProduced)
	}
	if len(posted) != 1 {
		t.Fatalf("postings = %d, want 1", len(posted))
	}
	if posted[0].Lines[0].AccountID != 400 || posted[0].Lines[1].AccountID != 600 {
		t.Errorf("produce lines accounts = %d/%d, want 400/600", posted[0].Lines[0].AccountID, posted[0].Lines[1].AccountID)
	}
	layer := posted[0]
	if got := layer.Lines[0].Debit; !got.Equal(amount.FromFloat64(20)) {
		t.Errorf("produce debit = %v, want 20 (4 x 5 standard)", got)
	}
}

func TestProductionService_Produce_RejectsExceedingRemaining(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	dst := uint64(20)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, QtyToProduce: 10, QtyProduced: 9, DstLocationID: &dst, State: ProductionOrderStateInProgress}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
		},
	}
	svc := testProductionService(orders, ConsumedMaterialDAOMock{}, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, moSequenceMock())

	_, err := svc.Produce(ctx, 1, 2, 500, 600, time.Now())
	helper.AssertError(t, err, true, ErrProduceQuantity)
}

func TestProductionService_RecordLabor_PostsWIPAndUpdatesShopTask(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateInProgress}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
		},
	}
	workOrder := &ShopTask{Base: model.Base{ID: 9}, ProductionOrderID: 1, WorkCenterID: 7, State: ShopTaskStateInProgress}
	workOrders := ShopTaskDAOMock{
		CRUDMock: dao.CRUDMock[ShopTask]{
			FindFunc: func(_ context.Context, _ uint64) (*ShopTask, error) { return workOrder, nil },
			UpdateFunc: func(_ context.Context, wo *ShopTask) (*ShopTask, error) {
				return wo, nil
			},
		},
	}
	workCenters := dao.CRUDMock[reference.WorkCenter]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.WorkCenter, error) {
			return &reference.WorkCenter{Base: model.Base{ID: 7}, CostPerHour: helper.Ptr(60.0)}, nil
		},
	}
	var posted []accounting.PostRequest
	poster := inventory.PosterMock{
		PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = append(posted, request)
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := testProductionService(orders, ConsumedMaterialDAOMock{}, workOrders, ProductionStepDAOMock{}, workCenters, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, poster, accounting.JournalLineDAOMock{}, moSequenceMock())

	updated, err := svc.RecordLabor(ctx, 1, 9, 30, 500, 600, 700, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ActualMinutes != 30 {
		t.Errorf("actual_minutes = %v, want 30", updated.ActualMinutes)
	}
	if updated.State != ShopTaskStateDone {
		t.Errorf("state = %s, want done", updated.State)
	}
	if len(posted) != 1 {
		t.Fatalf("postings = %d, want 1", len(posted))
	}
	if posted[0].Lines[0].AccountID != 600 || posted[0].Lines[1].AccountID != 700 {
		t.Errorf("labor lines accounts = %d/%d, want 600/700", posted[0].Lines[0].AccountID, posted[0].Lines[1].AccountID)
	}
	if got := posted[0].Lines[0].Debit; !got.Equal(amount.FromFloat64(30)) {
		t.Errorf("labor debit = %v, want 30 (30 min at 60/hour)", got)
	}
}

func TestProductionService_RecordLabor_RejectsMismatchAndInvalidInput(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateInProgress}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
		},
	}

	tests := []struct {
		name       string
		workOrder  *ShopTask
		minutes    float64
		wipAccount uint64
		laborAcc   uint64
		wantErr    error
	}{
		{name: "zero minutes", workOrder: &ShopTask{Base: model.Base{ID: 9}, ProductionOrderID: 1, WorkCenterID: 7}, minutes: 0, wipAccount: 600, laborAcc: 700, wantErr: ErrShopTaskLabor},
		{name: "missing wip account", workOrder: &ShopTask{Base: model.Base{ID: 9}, ProductionOrderID: 1, WorkCenterID: 7}, minutes: 10, wipAccount: 0, laborAcc: 700, wantErr: ErrWIPAccount},
		{name: "missing labor account", workOrder: &ShopTask{Base: model.Base{ID: 9}, ProductionOrderID: 1, WorkCenterID: 7}, minutes: 10, wipAccount: 600, laborAcc: 0, wantErr: ErrLaborAccount},
		{name: "shop task mismatch", workOrder: &ShopTask{Base: model.Base{ID: 9}, ProductionOrderID: 99, WorkCenterID: 7}, minutes: 10, wipAccount: 600, laborAcc: 700, wantErr: ErrShopTaskMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workOrders := ShopTaskDAOMock{
				CRUDMock: dao.CRUDMock[ShopTask]{
					FindFunc: func(_ context.Context, _ uint64) (*ShopTask, error) { return tt.workOrder, nil },
				},
			}
			workCenters := dao.CRUDMock[reference.WorkCenter]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.WorkCenter, error) {
					return &reference.WorkCenter{Base: model.Base{ID: 7}, CostPerHour: helper.Ptr(60.0)}, nil
				},
			}
			svc := testProductionService(orders, ConsumedMaterialDAOMock{}, workOrders, ProductionStepDAOMock{}, workCenters, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, moSequenceMock())

			_, err := svc.RecordLabor(ctx, 1, 9, tt.minutes, 500, tt.wipAccount, tt.laborAcc, time.Now())
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProductionService_SettleVariance_PostsVarianceAndCloses(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateInProgress}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
			UpdateFunc: func(_ context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	lines := accounting.JournalLineDAOMock{
		BalanceByOriginAndAccountFunc: func(_ context.Context, _ string, _ uint64, _ uint64) (float64, error) {
			return 15, nil
		},
	}
	var posted []accounting.PostRequest
	poster := inventory.PosterMock{
		PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = append(posted, request)
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := testProductionService(orders, ConsumedMaterialDAOMock{}, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, poster, lines, moSequenceMock())

	settled, err := svc.SettleVariance(ctx, 1, 500, 600, 800, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settled.State != ProductionOrderStateDone {
		t.Errorf("state = %s, want done", settled.State)
	}
	if len(posted) != 1 {
		t.Fatalf("postings = %d, want 1", len(posted))
	}
	if posted[0].Lines[0].AccountID != 800 || posted[0].Lines[1].AccountID != 600 {
		t.Errorf("variance lines accounts = %d/%d, want 800/600", posted[0].Lines[0].AccountID, posted[0].Lines[1].AccountID)
	}
}

func TestProductionService_SettleVariance_ZeroBalanceClosesWithoutPosting(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateInProgress}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
			UpdateFunc: func(_ context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	lines := accounting.JournalLineDAOMock{
		BalanceByOriginAndAccountFunc: func(_ context.Context, _ string, _ uint64, _ uint64) (float64, error) {
			return 0, nil
		},
	}
	var posted int
	poster := inventory.PosterMock{
		PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted++
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := testProductionService(orders, ConsumedMaterialDAOMock{}, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, poster, lines, moSequenceMock())

	settled, err := svc.SettleVariance(ctx, 1, 500, 600, 800, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settled.State != ProductionOrderStateDone {
		t.Errorf("state = %s, want done", settled.State)
	}
	if posted != 0 {
		t.Errorf("postings = %d, want 0", posted)
	}
}
