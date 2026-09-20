package returns

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type RMADAO interface {
	dao.CRUD[RMA]
	CreateWithLinesTx(ctx context.Context, tx *gorm.DB, rma *RMA, lines []*RMALine) (*RMA, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, rma *RMA) (*RMA, error)
}

type rmaDAO struct {
	dao.Base[RMA]
	db *gorm.DB
}

func NewRMADAO(db *gorm.DB) RMADAO {
	return rmaDAO{Base: dao.NewBase[RMA](db), db: db}
}

func (d rmaDAO) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, rma *RMA, lines []*RMALine) (*RMA, error) {
	if err := tx.WithContext(ctx).Create(rma).Error; err != nil {
		return nil, err
	}
	for _, line := range lines {
		line.RMAID = rma.ID
		if err := tx.WithContext(ctx).Create(line).Error; err != nil {
			return nil, err
		}
	}
	return rma, nil
}

func (d rmaDAO) UpdateTx(ctx context.Context, tx *gorm.DB, rma *RMA) (*RMA, error) {
	if err := tx.WithContext(ctx).Save(rma).Error; err != nil {
		return nil, err
	}
	return rma, nil
}

type RMALineDAO interface {
	dao.CRUD[RMALine]
	ListByRMA(ctx context.Context, rmaID uint64) ([]*RMALine, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, line *RMALine) (*RMALine, error)
}

type rmaLineDAO struct {
	dao.Base[RMALine]
	db *gorm.DB
}

func NewRMALineDAO(db *gorm.DB) RMALineDAO {
	return rmaLineDAO{Base: dao.NewBase[RMALine](db), db: db}
}

func (d rmaLineDAO) ListByRMA(ctx context.Context, rmaID uint64) ([]*RMALine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "rma_id", Operator: query.Equal, Value: rmaID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d rmaLineDAO) UpdateTx(ctx context.Context, tx *gorm.DB, line *RMALine) (*RMALine, error) {
	if err := tx.WithContext(ctx).Save(line).Error; err != nil {
		return nil, err
	}
	return line, nil
}
