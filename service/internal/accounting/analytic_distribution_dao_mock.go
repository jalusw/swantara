package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type DimensionDistributionDAOMock struct {
	dao.CRUDMock[DimensionDistribution]
	ListByLineFunc          func(ctx context.Context, lineID uint64) ([]*DimensionDistribution, error)
	CreateDistributionsFunc func(ctx context.Context, tx *gorm.DB, dists []*DimensionDistribution) error
}

func (m DimensionDistributionDAOMock) ListByLine(ctx context.Context, lineID uint64) ([]*DimensionDistribution, error) {
	if m.ListByLineFunc != nil {
		return m.ListByLineFunc(ctx, lineID)
	}
	return nil, nil
}

func (m DimensionDistributionDAOMock) CreateDistributions(ctx context.Context, tx *gorm.DB, dists []*DimensionDistribution) error {
	if m.CreateDistributionsFunc != nil {
		return m.CreateDistributionsFunc(ctx, tx, dists)
	}
	return nil
}
