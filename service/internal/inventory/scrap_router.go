package inventory

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

var (
	ErrScrapLocation = errors.New("no scrap location found for this organization")
)

type Scrapper interface {
	Scrap(ctx context.Context, movementID uint64, journalID, expenseAccountID uint64, date time.Time) (*CostLayer, error)
}

type ScrapRouter struct {
	shipments ShipmentDAO
	movements StockMovementDAO
	locations StockLocationDAO
	valuation Scrapper
}

func NewScrapRouter(shipments ShipmentDAO, movements StockMovementDAO, locations StockLocationDAO, valuation Scrapper) ScrapRouter {
	return ScrapRouter{
		shipments: shipments,
		movements: movements,
		locations: locations,
		valuation: valuation,
	}
}

func (s ScrapRouter) RouteToScrap(ctx context.Context, organizationID, shipmentID uint64, itemID uint64, journalID, expenseAccountID uint64, date time.Time) error {
	shipment, err := s.shipments.Find(ctx, shipmentID)
	if err != nil {
		return err
	}
	if shipment == nil {
		return ErrMovementNotFound
	}
	if shipment.OrganizationID == nil || *shipment.OrganizationID != organizationID {
		return ErrOrganizationMissing
	}
	if expenseAccountID == 0 {
		return ErrScrapAccount
	}

	movements, err := s.movements.ListByShipment(ctx, shipmentID)
	if err != nil {
		return err
	}
	var received *StockMovement
	for _, movement := range movements {
		if movement.ItemID == itemID && movement.State == MovementStateDone {
			received = movement
			break
		}
	}
	if received == nil {
		return ErrMovementNotFound
	}

	scrapLocation, err := s.findScrapLocation(ctx, organizationID)
	if err != nil {
		return err
	}
	if scrapLocation == nil {
		return ErrScrapLocation
	}

	scrapMove := &StockMovement{
		OrganizationID: &organizationID,
		ItemID:         itemID,
		Qty:            received.Qty,
		BatchID:        received.BatchID,
		SrcLocationID:  received.DstLocationID,
		DstLocationID:  scrapLocation.ID,
		State:          MovementStateConfirmed,
		OriginType:     helper.Ptr(accounting.OriginTypeQualityScrap),
		OriginID:       &shipmentID,
		ScheduledDate:  helper.Ptr(date),
	}
	if _, err := s.movements.Create(ctx, scrapMove); err != nil {
		return err
	}

	if _, err := s.valuation.Scrap(ctx, scrapMove.ID, journalID, expenseAccountID, date); err != nil {
		return err
	}
	return nil
}

func (s ScrapRouter) findScrapLocation(ctx context.Context, organizationID uint64) (*reference.StockLocation, error) {
	page, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	for _, location := range page.Items {
		if location.Usage == "scrap" {
			return location, nil
		}
	}
	return nil, nil
}
