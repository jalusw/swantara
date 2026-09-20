package procurement

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type CostCenterService struct {
	centers CostCenterDAO
}

func NewCostCenterService(centers CostCenterDAO) CostCenterService {
	return CostCenterService{centers: centers}
}

func (s CostCenterService) List(ctx context.Context, q *query.Query) (*query.Page[CostCenter], error) {
	return s.centers.List(ctx, q)
}

func (s CostCenterService) Find(ctx context.Context, id uint64) (*CostCenter, error) {
	return s.centers.Find(ctx, id)
}

func (s CostCenterService) Create(ctx context.Context, center *CostCenter) (*CostCenter, error) {
	return s.centers.Create(ctx, center)
}

func (s CostCenterService) Update(ctx context.Context, center *CostCenter) (*CostCenter, error) {
	return s.centers.Update(ctx, center)
}

func (s CostCenterService) Delete(ctx context.Context, id uint64) error {
	return s.centers.Delete(ctx, id)
}
