package quality

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type QualityPointDAOMock struct {
	dao.CRUDMock[reference.QualityPoint]
	ListByItemFunc func(ctx context.Context, itemID uint64) ([]*reference.QualityPoint, error)
}

func (m QualityPointDAOMock) ListByItem(ctx context.Context, itemID uint64) ([]*reference.QualityPoint, error) {
	if m.ListByItemFunc != nil {
		return m.ListByItemFunc(ctx, itemID)
	}
	return []*reference.QualityPoint{}, nil
}

type QualityCheckDAOMock struct {
	dao.CRUDMock[QualityCheck]
	ListByShipmentFunc func(ctx context.Context, shipmentID uint64) ([]*QualityCheck, error)
}

func (m QualityCheckDAOMock) ListByShipment(ctx context.Context, shipmentID uint64) ([]*QualityCheck, error) {
	if m.ListByShipmentFunc != nil {
		return m.ListByShipmentFunc(ctx, shipmentID)
	}
	return []*QualityCheck{}, nil
}

func (m QualityCheckDAOMock) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[QualityCheck], error) {
	return m.List(ctx, q)
}

func (m QualityCheckDAOMock) FindInOrg(ctx context.Context, id, organizationID uint64) (*QualityCheck, error) {
	return m.Find(ctx, id)
}

type QualityAlertDAOMock struct {
	dao.CRUDMock[QualityAlert]
	ListByCheckFunc func(ctx context.Context, checkID uint64) ([]*QualityAlert, error)
}

func (m QualityAlertDAOMock) ListByCheck(ctx context.Context, checkID uint64) ([]*QualityAlert, error) {
	if m.ListByCheckFunc != nil {
		return m.ListByCheckFunc(ctx, checkID)
	}
	return []*QualityAlert{}, nil
}

func (m QualityAlertDAOMock) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[QualityAlert], error) {
	return m.List(ctx, q)
}

func (m QualityAlertDAOMock) FindInOrg(ctx context.Context, id, organizationID uint64) (*QualityAlert, error) {
	return m.Find(ctx, id)
}
