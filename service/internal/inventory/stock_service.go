package inventory

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type StockService struct {
	movements StockMovementDAO
	shipments ShipmentDAO
	quants    StockBalanceDAO
}

func NewStockService(movements StockMovementDAO, shipments ShipmentDAO, quants StockBalanceDAO) StockService {
	return StockService{movements: movements, shipments: shipments, quants: quants}
}

func (s StockService) ListMovements(ctx context.Context, q *query.Query) (*query.Page[StockMovement], error) {
	return s.movements.List(ctx, q)
}

func (s StockService) FindMovement(ctx context.Context, id uint64) (*StockMovement, error) {
	return s.movements.Find(ctx, id)
}

func (s StockService) DeleteMovement(ctx context.Context, id uint64) error {
	return s.movements.Delete(ctx, id)
}

func (s StockService) ListShipments(ctx context.Context, q *query.Query) (*query.Page[Shipment], error) {
	return s.shipments.List(ctx, q)
}

func (s StockService) FindShipment(ctx context.Context, id uint64) (*Shipment, error) {
	return s.shipments.Find(ctx, id)
}

func (s StockService) ListBalances(ctx context.Context, q *query.Query) (*query.Page[StockBalance], error) {
	return s.quants.List(ctx, q)
}
