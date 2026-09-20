package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

type DimensionDistribution struct {
	model.Base
	JournalLineID uint64        `gorm:"not null" json:"journal_line_id"`
	DimensionID   uint64        `gorm:"not null" json:"dimension_id"`
	Percent       float64       `gorm:"type:numeric(5,2)" json:"percent"`
	Amount        amount.Amount `gorm:"type:numeric(18,4)" json:"amount"`
}

func (DimensionDistribution) TableName() string {
	return "dimension_distributions"
}

type DimensionDistributionDAO interface {
	dao.CRUD[DimensionDistribution]
	ListByLine(ctx context.Context, lineID uint64) ([]*DimensionDistribution, error)
	CreateDistributions(ctx context.Context, tx *gorm.DB, dists []*DimensionDistribution) error
}

type dimensionDistributionDAO struct {
	dao.Base[DimensionDistribution]
	db *gorm.DB
}

func NewDimensionDistributionDAO(db *gorm.DB) DimensionDistributionDAO {
	return dimensionDistributionDAO{Base: dao.NewBase[DimensionDistribution](db), db: db}
}

func (d dimensionDistributionDAO) ListByLine(ctx context.Context, lineID uint64) ([]*DimensionDistribution, error) {
	var items []DimensionDistribution
	if err := d.db.WithContext(ctx).Where("journal_line_id = ?", lineID).Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]*DimensionDistribution, len(items))
	for i := range items {
		out[i] = &items[i]
	}
	return out, nil
}

func (d dimensionDistributionDAO) CreateDistributions(ctx context.Context, tx *gorm.DB, dists []*DimensionDistribution) error {
	if len(dists) == 0 {
		return nil
	}
	return tx.WithContext(ctx).CreateInBatches(dists, 100).Error
}
