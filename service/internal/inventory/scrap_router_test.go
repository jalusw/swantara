package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestScrapRouter_RoutesReceivedQuantityToScrap(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return &Shipment{Base: model.Base{ID: 50}, OrganizationID: &organizationID}, nil
			},
		},
	}
	movements := StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[StockMovement]{
			CreateFunc: func(_ context.Context, movement *StockMovement) (*StockMovement, error) {
				movement.ID = 77
				return movement, nil
			},
		},
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{
				{Base: model.Base{ID: 1}, ItemID: 100, Qty: 8, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDone},
				{Base: model.Base{ID: 2}, ItemID: 100, Qty: 2, SrcLocationID: 30, DstLocationID: 10, State: MovementStateConfirmed},
			}, nil
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field != "organization_id" {
					t.Fatalf("field = %s, want organization_id", q.Filters[0].Field)
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
					{Base: model.Base{ID: 10}, Usage: "internal"},
					{Base: model.Base{ID: 60}, Usage: "scrap"},
				}}, nil
			},
		},
	}
	var scrapped uint64
	valuation := scrapRouterValuationMock{
		scrapFunc: func(_ context.Context, movementID uint64, _ uint64, _ uint64, _ time.Time) (*CostLayer, error) {
			scrapped = movementID
			return &CostLayer{Base: model.Base{ID: 5}}, nil
		},
	}
	router := NewScrapRouter(shipments, movements, locations, valuation)

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scrapped != 77 {
		t.Errorf("scrapped movement = %d, want 77", scrapped)
	}
}

func TestScrapRouter_RejectsWhenNoReceivedMove(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return &Shipment{Base: model.Base{ID: 50}, OrganizationID: &organizationID}, nil
			},
		},
	}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 8, State: MovementStateConfirmed}}, nil
		},
	}
	router := NewScrapRouter(shipments, movements, StockLocationDAOMock{}, scrapRouterValuationMock{})

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Now())
	if helper.AssertError(t, err, true, ErrMovementNotFound) {
		return
	}
}

func TestScrapRouter_RejectsWithoutScrapLocation(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return &Shipment{Base: model.Base{ID: 50}, OrganizationID: &organizationID}, nil
			},
		},
	}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 8, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDone}}, nil
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 10}, Usage: "internal"}}}, nil
			},
		},
	}
	router := NewScrapRouter(shipments, movements, locations, scrapRouterValuationMock{})

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Now())
	if helper.AssertError(t, err, true, ErrScrapLocation) {
		return
	}
}

func TestScrapRouter_RejectsMissingExpenseAccount(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return &Shipment{Base: model.Base{ID: 50}, OrganizationID: &organizationID}, nil
			},
		},
	}
	router := NewScrapRouter(shipments, StockMovementDAOMock{}, StockLocationDAOMock{}, scrapRouterValuationMock{})

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 0, time.Now())
	if helper.AssertError(t, err, true, ErrScrapAccount) {
		return
	}
}

type scrapRouterValuationMock struct {
	scrapFunc func(ctx context.Context, movementID uint64, journalID, expenseAccountID uint64, date time.Time) (*CostLayer, error)
}

func (m scrapRouterValuationMock) Scrap(ctx context.Context, movementID uint64, journalID, expenseAccountID uint64, date time.Time) (*CostLayer, error) {
	if m.scrapFunc != nil {
		return m.scrapFunc(ctx, movementID, journalID, expenseAccountID, date)
	}
	return &CostLayer{}, nil
}

func TestScrapRouter_RejectsUnknownShipment(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return nil, nil
			},
		},
	}
	router := NewScrapRouter(shipments, StockMovementDAOMock{}, StockLocationDAOMock{}, scrapRouterValuationMock{})

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Now())
	if helper.AssertError(t, err, true, ErrMovementNotFound) {
		return
	}
}

func TestScrapRouter_RejectsShipmentFromOtherOrganization(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return &Shipment{Base: model.Base{ID: 50}, OrganizationID: helper.Ptr(uint64(2))}, nil
			},
		},
	}
	router := NewScrapRouter(shipments, StockMovementDAOMock{}, StockLocationDAOMock{}, scrapRouterValuationMock{})

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Now())
	if helper.AssertError(t, err, true, ErrOrganizationMissing) {
		return
	}
}

func TestScrapRouter_RejectsWhenShipmentFindFails(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return nil, ErrMovementNotFound
			},
		},
	}
	router := NewScrapRouter(shipments, StockMovementDAOMock{}, StockLocationDAOMock{}, scrapRouterValuationMock{})

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Now())
	if helper.AssertError(t, err, true, ErrMovementNotFound) {
		return
	}
}

func TestScrapRouter_RejectsWhenListMovesFails(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return &Shipment{Base: model.Base{ID: 50}, OrganizationID: &organizationID}, nil
			},
		},
	}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return nil, ErrMovementNotFound
		},
	}
	router := NewScrapRouter(shipments, movements, StockLocationDAOMock{}, scrapRouterValuationMock{})

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Now())
	if helper.AssertError(t, err, true, ErrMovementNotFound) {
		return
	}
}

func TestScrapRouter_RejectsWhenMoveCreateFails(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return &Shipment{Base: model.Base{ID: 50}, OrganizationID: &organizationID}, nil
			},
		},
	}
	movements := StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[StockMovement]{
			CreateFunc: func(_ context.Context, _ *StockMovement) (*StockMovement, error) {
				return nil, ErrMovementQty
			},
		},
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 8, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDone}}, nil
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 60}, Usage: "scrap"}}}, nil
			},
		},
	}
	router := NewScrapRouter(shipments, movements, locations, scrapRouterValuationMock{})

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Now())
	if helper.AssertError(t, err, true, ErrMovementQty) {
		return
	}
}

func TestScrapRouter_RejectsWhenValuationFails(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return &Shipment{Base: model.Base{ID: 50}, OrganizationID: &organizationID}, nil
			},
		},
	}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 8, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDone}}, nil
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 60}, Usage: "scrap"}}}, nil
			},
		},
	}
	valuation := scrapRouterValuationMock{
		scrapFunc: func(_ context.Context, _ uint64, _ uint64, _ uint64, _ time.Time) (*CostLayer, error) {
			return nil, ErrScrapAccount
		},
	}
	router := NewScrapRouter(shipments, movements, locations, valuation)

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Now())
	if helper.AssertError(t, err, true, ErrScrapAccount) {
		return
	}
}

func TestScrapRouter_RejectsWhenScrapLocationLookupFails(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	shipments := ShipmentDAOMock{
		CRUDMock: dao.CRUDMock[Shipment]{
			FindFunc: func(_ context.Context, _ uint64) (*Shipment, error) {
				return &Shipment{Base: model.Base{ID: 50}, OrganizationID: &organizationID}, nil
			},
		},
	}
	movements := StockMovementDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) {
			return []*StockMovement{{Base: model.Base{ID: 1}, ItemID: 100, Qty: 8, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDone}}, nil
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return nil, ErrLocationNotFound
			},
		},
	}
	router := NewScrapRouter(shipments, movements, locations, scrapRouterValuationMock{})

	err := router.RouteToScrap(ctx, organizationID, 50, 100, 90, 6000, time.Now())
	if helper.AssertError(t, err, true, ErrLocationNotFound) {
		return
	}
}
