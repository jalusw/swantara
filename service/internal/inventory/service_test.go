package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestWarehouseService_CreateWarehouse_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	warehouses := WarehouseDAOMock{
		CRUDMock: dao.CRUDMock[reference.Warehouse]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Warehouse, error) {
				return &reference.Warehouse{Base: model.Base{ID: 5}, Code: helper.Ptr("WH-01")}, nil
			},
		},
	}

	tests := []struct {
		name       string
		warehouses WarehouseDAOMock
		warehouse  *reference.Warehouse
		wantErr    error
	}{
		{
			name:      "rejects empty name",
			warehouse: &reference.Warehouse{Name: ""},
			wantErr:   ErrWarehouseNameRequired,
		},
		{
			name:       "rejects duplicate code",
			warehouses: warehouses,
			warehouse:  &reference.Warehouse{Name: "Main", Code: helper.Ptr("WH-01")},
			wantErr:    ErrWarehouseCodeTaken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewWarehouseService(tt.warehouses, StockLocationDAOMock{})

			_, err := svc.CreateWarehouse(ctx, tt.warehouse)

			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestWarehouseService_CreateLocation_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		location *reference.StockLocation
		wantErr  error
	}{
		{
			name:     "rejects invalid usage",
			location: &reference.StockLocation{Name: "Shelf A", Usage: "void"},
			wantErr:  ErrInvalidLocationType,
		},
		{
			name:     "rejects unknown parent",
			location: &reference.StockLocation{Name: "Bin", Usage: "internal", ParentID: helper.Ptr(uint64(99))},
			wantErr:  ErrLocationParent,
		},
		{
			name:     "rejects unknown warehouse",
			location: &reference.StockLocation{Name: "Bin", Usage: "internal", WarehouseID: helper.Ptr(uint64(42))},
			wantErr:  ErrWarehouseNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewWarehouseService(WarehouseDAOMock{}, StockLocationDAOMock{})

			_, err := svc.CreateLocation(ctx, tt.location)

			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestLedgerService_InternalTransfer_CreatesThenApplies(t *testing.T) {
	ctx := context.Background()
	applied := false
	movements := StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[StockMovement]{CreateFunc: func(_ context.Context, movement *StockMovement) (*StockMovement, error) {
			movement.ID = 7
			return movement, nil
		}},
		FindForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, movementID uint64) (*StockMovement, error) {
			return &StockMovement{Base: model.Base{ID: movementID}, State: MovementStateConfirmed}, nil
		},
		ApplyTxFunc: func(_ context.Context, _ *gorm.DB, movement *StockMovement) (*StockMovement, error) {
			applied = movement.ID == 7
			movement.State = MovementStateDone
			return movement, nil
		},
	}
	svc := NewLedgerService(movements, StockBalanceDAOMock{}, StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{}, nil
		}},
	}, TransactionerMock{})

	done, err := svc.InternalTransfer(ctx, &StockMovement{ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 20}, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !applied || done.State != MovementStateDone {
		t.Errorf("applied = %v, state = %v, want true/done", applied, done.State)
	}
}

func TestLedgerService_CreateMovement_RejectsZeroProduct(t *testing.T) {
	ctx := context.Background()
	svc := NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{})

	_, err := svc.CreateMovement(ctx, &StockMovement{Qty: 5, SrcLocationID: 10, DstLocationID: 20})
	if helper.AssertError(t, err, true, ErrMovementItem) {
		return
	}
}

func TestLedgerService_RebuildQuants_ZeroesStaleAndCreatesMissing(t *testing.T) {
	ctx := context.Background()
	movements := StockMovementDAOMock{
		LedgerTotalsFunc: func(_ context.Context, _ *uint64) ([]LedgerTotal, error) {
			return []LedgerTotal{
				{ItemID: 100, LocationID: 10, BatchID: nil, Total: 50},
				{ItemID: 100, LocationID: 20, BatchID: helper.Ptr(uint64(7)), Total: 12},
			}, nil
		},
	}
	quants := StockBalanceDAOMock{
		ListAllFunc: func(_ context.Context, _ *uint64) ([]*StockBalance, error) {
			return []*StockBalance{
				{Base: model.Base{ID: 1}, ItemID: 100, LocationID: 10, Quantity: 99},
				{Base: model.Base{ID: 2}, ItemID: 200, LocationID: 10, Quantity: 5},
			}, nil
		},
		FindByKeyFunc: func(_ context.Context, itemID, locationID uint64, batchID *uint64) (*StockBalance, error) {
			if itemID == 100 && locationID == 10 {
				return &StockBalance{Base: model.Base{ID: 1}, ItemID: 100, LocationID: 10}, nil
			}
			return nil, nil
		},
	}
	updated := []*StockBalance{}
	quants.UpdateFunc = func(_ context.Context, quant *StockBalance) (*StockBalance, error) {
		updated = append(updated, quant)
		return quant, nil
	}
	created := []*StockBalance{}
	quants.CreateFunc = func(_ context.Context, quant *StockBalance) (*StockBalance, error) {
		created = append(created, quant)
		return quant, nil
	}
	svc := NewLedgerService(movements, quants, StockLocationDAOMock{}, TransactionerMock{})

	if err := svc.RebuildBalances(ctx, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := map[uint64]float64{}
	for _, quant := range updated {
		found[quant.ItemID] = quant.Quantity
	}
	if found[100] != 50 {
		t.Errorf("item 100 qty = %v, want 50", found[100])
	}
	if found[200] != 0 {
		t.Errorf("item 200 qty = %v, want 0 (stale quant zeroed)", found[200])
	}
	if len(created) != 1 || created[0].ItemID != 100 || created[0].LocationID != 20 || created[0].Quantity != 12 {
		t.Errorf("created quants = %+v, want item 100 at location 20 qty 12", created)
	}
}

func daoCRUDFind[E any](fn func(ctx context.Context, id uint64) (*E, error)) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{FindFunc: fn}
}

func TestWarehouseService_UpdateWarehouse_Succeeds(t *testing.T) {
	ctx := context.Background()
	updated := false
	warehouses := WarehouseDAOMock{
		CRUDMock: dao.CRUDMock[reference.Warehouse]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Warehouse, error) {
				return nil, nil
			},
			UpdateFunc: func(_ context.Context, warehouse *reference.Warehouse) (*reference.Warehouse, error) {
				updated = true
				return warehouse, nil
			},
		},
	}
	svc := NewWarehouseService(warehouses, StockLocationDAOMock{})

	got, err := svc.UpdateWarehouse(ctx, &reference.Warehouse{Base: model.Base{ID: 1}, Name: "Main", Code: helper.Ptr("WH-01")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated || got.ID != 1 {
		t.Errorf("updated = %v, id = %v, want true/1", updated, got.ID)
	}
}

func TestWarehouseService_UpdateWarehouse_RejectsCodeTaken(t *testing.T) {
	ctx := context.Background()
	warehouses := WarehouseDAOMock{
		CRUDMock: dao.CRUDMock[reference.Warehouse]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Warehouse, error) {
				return &reference.Warehouse{Base: model.Base{ID: 5}, Code: helper.Ptr("WH-01")}, nil
			},
		},
	}
	svc := NewWarehouseService(warehouses, StockLocationDAOMock{})

	_, err := svc.UpdateWarehouse(ctx, &reference.Warehouse{Base: model.Base{ID: 1}, Name: "Main", Code: helper.Ptr("WH-01")})

	helper.AssertError(t, err, true, ErrWarehouseCodeTaken)
}

func TestWarehouseService_UpdateLocation_Succeeds(t *testing.T) {
	ctx := context.Background()
	updated := false
	warehouses := WarehouseDAOMock{
		CRUDMock: dao.CRUDMock[reference.Warehouse]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
				return &reference.Warehouse{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(9))}, nil
			},
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, id uint64) (*reference.StockLocation, error) {
				if id == 7 {
					return &reference.StockLocation{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(9))}, nil
				}
				return nil, nil
			},
			UpdateFunc: func(_ context.Context, location *reference.StockLocation) (*reference.StockLocation, error) {
				updated = true
				return location, nil
			},
		},
	}
	svc := NewWarehouseService(warehouses, locations)

	got, err := svc.UpdateLocation(ctx, &reference.StockLocation{
		Base:           model.Base{ID: 3},
		Name:           "Bin",
		Usage:          "internal",
		OrganizationID: helper.Ptr(uint64(9)),
		WarehouseID:    helper.Ptr(uint64(42)),
		ParentID:       helper.Ptr(uint64(7)),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated || got.ID != 3 {
		t.Errorf("updated = %v, id = %v, want true/3", updated, got.ID)
	}
}

func TestWarehouseService_UpdateLocation_RejectsMismatchedOrganizations(t *testing.T) {
	ctx := context.Background()
	warehouses := WarehouseDAOMock{
		CRUDMock: dao.CRUDMock[reference.Warehouse]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
				return &reference.Warehouse{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(99))}, nil
			},
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, id uint64) (*reference.StockLocation, error) {
				if id == 7 {
					return &reference.StockLocation{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99))}, nil
				}
				return nil, nil
			},
		},
	}
	svc := NewWarehouseService(warehouses, locations)
	location := &reference.StockLocation{
		Base:           model.Base{ID: 3},
		Name:           "Bin",
		Usage:          "internal",
		OrganizationID: helper.Ptr(uint64(9)),
		WarehouseID:    helper.Ptr(uint64(42)),
		ParentID:       helper.Ptr(uint64(7)),
	}

	_, err := svc.UpdateLocation(ctx, location)

	helper.AssertError(t, err, true, ErrWarehouseMismatch)
}

func TestLedgerService_OnHand_ReturnsQuantQuantity(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(9))}, nil
		},
	}}
	quants := StockBalanceDAOMock{FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
		return &StockBalance{Quantity: 25}, nil
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{})

	got, err := svc.OnHand(ctx, helper.Ptr(uint64(9)), 100, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 25 {
		t.Errorf("OnHand() = %v, want 25", got)
	}
}

func TestLedgerService_OnHand_RejectsLocationOutsideOrganization(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(9))}, nil
		},
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, locations, TransactionerMock{})

	_, err := svc.OnHand(ctx, helper.Ptr(uint64(99)), 100, 10)

	helper.AssertError(t, err, true, ErrLocationNotFound)
}

func TestLedgerService_OnHand_ReturnsZeroWhenNoQuant(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{}, nil
		},
	}}
	quants := StockBalanceDAOMock{FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
		return nil, nil
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{})

	got, err := svc.OnHand(ctx, nil, 100, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("OnHand() = %v, want 0", got)
	}
}

func TestLedgerService_OnHandByProduct_SumsAllQuants(t *testing.T) {
	ctx := context.Background()
	quants := StockBalanceDAOMock{ListByItemFunc: func(_ context.Context, _ uint64) ([]*StockBalance, error) {
		return []*StockBalance{{Quantity: 10}, {Quantity: 5}}, nil
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, StockLocationDAOMock{}, TransactionerMock{})

	got, err := svc.OnHandByProduct(ctx, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 15 {
		t.Errorf("OnHandByProduct() = %v, want 15", got)
	}
}

func TestLedgerService_OnHandByWarehouse_GroupsByWarehouse(t *testing.T) {
	ctx := context.Background()
	quants := StockBalanceDAOMock{ListByItemFunc: func(_ context.Context, _ uint64) ([]*StockBalance, error) {
		return []*StockBalance{{LocationID: 10, Quantity: 10}, {LocationID: 11, Quantity: 5}}, nil
	}}
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, id uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{Base: model.Base{ID: id}, WarehouseID: helper.Ptr(uint64(1))}, nil
		},
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{})

	got, err := svc.OnHandByWarehouse(ctx, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[1] != 15 {
		t.Errorf("OnHandByWarehouse()[1] = %v, want 15", got[1])
	}
}

func TestLedgerService_AvailableToPromise_SubtractsReserved(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{}, nil
		},
	}}
	quants := StockBalanceDAOMock{FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
		return &StockBalance{Quantity: 20, ReservedQty: 5}, nil
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{})

	got, err := svc.AvailableToPromise(ctx, nil, 100, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 15 {
		t.Errorf("AvailableToPromise() = %v, want 15", got)
	}
}

func TestWarehouseService_CreateWarehouse_Succeeds(t *testing.T) {
	ctx := context.Background()
	created := false
	warehouses := WarehouseDAOMock{
		CRUDMock: dao.CRUDMock[reference.Warehouse]{
			CreateFunc: func(_ context.Context, warehouse *reference.Warehouse) (*reference.Warehouse, error) {
				created = true
				return warehouse, nil
			},
		},
	}
	svc := NewWarehouseService(warehouses, StockLocationDAOMock{})

	warehouse := &reference.Warehouse{Name: "Main", Code: helper.Ptr("WH-01")}
	got, err := svc.CreateWarehouse(ctx, warehouse)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created || got.Name != "Main" {
		t.Errorf("created = %v, name = %s, want true/Main", created, got.Name)
	}
}

func TestWarehouseService_CreateWarehouse_WithoutCodeSkipsLookup(t *testing.T) {
	ctx := context.Background()
	searched := false
	warehouses := WarehouseDAOMock{
		CRUDMock: dao.CRUDMock[reference.Warehouse]{
			SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Warehouse, error) {
				searched = true
				return nil, nil
			},
			CreateFunc: func(_ context.Context, warehouse *reference.Warehouse) (*reference.Warehouse, error) {
				return warehouse, nil
			},
		},
	}
	svc := NewWarehouseService(warehouses, StockLocationDAOMock{})

	_, err := svc.CreateWarehouse(ctx, &reference.Warehouse{Name: "Main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if searched {
		t.Error("code lookup must be skipped when the warehouse has no code")
	}
}

func TestWarehouseService_CreateLocation_Succeeds(t *testing.T) {
	ctx := context.Background()
	created := false
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			CreateFunc: func(_ context.Context, location *reference.StockLocation) (*reference.StockLocation, error) {
				created = true
				return location, nil
			},
		},
	}
	svc := NewWarehouseService(WarehouseDAOMock{}, locations)

	location := &reference.StockLocation{Name: "Shelf A", Usage: "internal"}
	got, err := svc.CreateLocation(ctx, location)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created || got.Name != "Shelf A" {
		t.Errorf("created = %v, name = %s, want true/Shelf A", created, got.Name)
	}
}

func TestWarehouseService_UpdateLocation_RejectsParentMismatch(t *testing.T) {
	ctx := context.Background()
	warehouses := WarehouseDAOMock{
		CRUDMock: dao.CRUDMock[reference.Warehouse]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
				return &reference.Warehouse{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(9))}, nil
			},
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, id uint64) (*reference.StockLocation, error) {
				if id == 7 {
					return &reference.StockLocation{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99))}, nil
				}
				return nil, nil
			},
		},
	}
	svc := NewWarehouseService(warehouses, locations)
	location := &reference.StockLocation{
		Base:           model.Base{ID: 3},
		Name:           "Bin",
		Usage:          "internal",
		OrganizationID: helper.Ptr(uint64(9)),
		WarehouseID:    helper.Ptr(uint64(42)),
		ParentID:       helper.Ptr(uint64(7)),
	}

	_, err := svc.UpdateLocation(ctx, location)
	if helper.AssertError(t, err, true, ErrLocationParent) {
		return
	}
}

func TestLedgerService_CreateMovement_RejectsMissingLocations(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, nil
		},
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, locations, TransactionerMock{})

	_, err := svc.CreateMovement(ctx, &StockMovement{ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 20})
	if helper.AssertError(t, err, true, ErrLocationNotFound) {
		return
	}
}

func TestLedgerService_CreateMovement_PropagatesLocationFindError(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, ErrLocationNotFound
		},
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, locations, TransactionerMock{})

	_, err := svc.CreateMovement(ctx, &StockMovement{ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 20})
	if helper.AssertError(t, err, true, ErrLocationNotFound) {
		return
	}
}

func TestLedgerService_OnHand_PropagatesQuantError(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{}, nil
		},
	}}
	quants := StockBalanceDAOMock{FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
		return nil, ErrBalanceNotFound
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{})

	_, err := svc.OnHand(ctx, nil, 100, 10)
	if helper.AssertError(t, err, true, ErrBalanceNotFound) {
		return
	}
}

func TestLedgerService_OnHandByWarehouse_RejectsUnknownLocation(t *testing.T) {
	ctx := context.Background()
	quants := StockBalanceDAOMock{ListByItemFunc: func(_ context.Context, _ uint64) ([]*StockBalance, error) {
		return []*StockBalance{{LocationID: 10, Quantity: 10}}, nil
	}}
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, ErrLocationNotFound
		},
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{})

	_, err := svc.OnHandByWarehouse(ctx, 100)
	if helper.AssertError(t, err, true, ErrLocationNotFound) {
		return
	}
}

func TestLedgerService_OnHandByWarehouse_SkipsLocationsWithoutWarehouse(t *testing.T) {
	ctx := context.Background()
	quants := StockBalanceDAOMock{ListByItemFunc: func(_ context.Context, _ uint64) ([]*StockBalance, error) {
		return []*StockBalance{{LocationID: 10, Quantity: 10}}, nil
	}}
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{Base: model.Base{ID: 10}}, nil
		},
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{})

	got, err := svc.OnHandByWarehouse(ctx, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("by warehouse = %v, want empty map", got)
	}
}

func TestLedgerService_AvailableToPromise_PropagatesQuantError(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{}, nil
		},
	}}
	quants := StockBalanceDAOMock{FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
		return nil, ErrBalanceNotFound
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{})

	_, err := svc.AvailableToPromise(ctx, nil, 100, 10)
	if helper.AssertError(t, err, true, ErrBalanceNotFound) {
		return
	}
}

func TestLedgerService_AvailableToPromise_ReturnsZeroWithoutQuant(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{}, nil
		},
	}}
	quants := StockBalanceDAOMock{FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
		return nil, nil
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{})

	got, err := svc.AvailableToPromise(ctx, nil, 100, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("AvailableToPromise() = %v, want 0", got)
	}
}

func TestLedgerService_RebuildQuants_PropagatesLedgerTotalsError(t *testing.T) {
	ctx := context.Background()
	movements := StockMovementDAOMock{LedgerTotalsFunc: func(_ context.Context, _ *uint64) ([]LedgerTotal, error) {
		return nil, ErrMovementNotFound
	}}
	svc := NewLedgerService(movements, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{})

	err := svc.RebuildBalances(ctx, nil)
	if helper.AssertError(t, err, true, ErrMovementNotFound) {
		return
	}
}

func TestLedgerService_RebuildQuants_PropagatesFindByKeyError(t *testing.T) {
	ctx := context.Background()
	movements := StockMovementDAOMock{LedgerTotalsFunc: func(_ context.Context, _ *uint64) ([]LedgerTotal, error) {
		return []LedgerTotal{{ItemID: 100, LocationID: 10, Total: 50}}, nil
	}}
	quants := StockBalanceDAOMock{
		ListAllFunc: func(_ context.Context, _ *uint64) ([]*StockBalance, error) {
			return []*StockBalance{}, nil
		},
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return nil, ErrBalanceNotFound
		},
	}
	svc := NewLedgerService(movements, quants, StockLocationDAOMock{}, TransactionerMock{})

	err := svc.RebuildBalances(ctx, nil)
	if helper.AssertError(t, err, true, ErrBalanceNotFound) {
		return
	}
}

func TestLedgerService_InternalTransfer_PropagatesCreateMovementError(t *testing.T) {
	ctx := context.Background()
	svc := NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{})

	_, err := svc.InternalTransfer(ctx, &StockMovement{Qty: 5}, time.Now())
	if helper.AssertError(t, err, true, ErrMovementItem) {
		return
	}
}

func TestLedgerService_OnHandByProduct_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	quants := StockBalanceDAOMock{ListByItemFunc: func(_ context.Context, _ uint64) ([]*StockBalance, error) {
		return nil, ErrBalanceNotFound
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, quants, StockLocationDAOMock{}, TransactionerMock{})

	_, err := svc.OnHandByProduct(ctx, 100)
	if helper.AssertError(t, err, true, ErrBalanceNotFound) {
		return
	}
}

func TestWarehouseService_CreateLocation_PropagatesWarehouseFindError(t *testing.T) {
	ctx := context.Background()
	warehouses := WarehouseDAOMock{
		CRUDMock: dao.CRUDMock[reference.Warehouse]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
				return nil, ErrWarehouseNotFound
			},
		},
	}
	svc := NewWarehouseService(warehouses, StockLocationDAOMock{})

	_, err := svc.CreateLocation(ctx, &reference.StockLocation{Name: "Bin", Usage: "internal", WarehouseID: helper.Ptr(uint64(42))})
	if helper.AssertError(t, err, true, ErrWarehouseNotFound) {
		return
	}
}

func TestLedgerService_OnHand_PropagatesLocationFindError(t *testing.T) {
	ctx := context.Background()
	locations := StockLocationDAOMock{CRUDMock: dao.CRUDMock[reference.StockLocation]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, ErrLocationNotFound
		},
	}}
	svc := NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, locations, TransactionerMock{})

	_, err := svc.OnHand(ctx, nil, 100, 10)
	if helper.AssertError(t, err, true, ErrLocationNotFound) {
		return
	}
}
