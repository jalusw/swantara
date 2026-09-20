package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type PdcInstrumentDAO interface {
	dao.CRUD[PdcInstrument]
	UpdateTx(ctx context.Context, tx *gorm.DB, instrument *PdcInstrument) (*PdcInstrument, error)
	ListDue(ctx context.Context, asOf *time.Time) ([]*PdcInstrument, error)
}

type pdcInstrumentDAO struct {
	dao.Base[PdcInstrument]
	db *gorm.DB
}

func NewPdcInstrumentDAO(db *gorm.DB) PdcInstrumentDAO {
	return pdcInstrumentDAO{Base: dao.NewBase[PdcInstrument](db), db: db}
}

func (d pdcInstrumentDAO) UpdateTx(ctx context.Context, tx *gorm.DB, instrument *PdcInstrument) (*PdcInstrument, error) {
	if err := tx.WithContext(ctx).Save(instrument).Error; err != nil {
		return nil, err
	}
	return instrument, nil
}

func (d pdcInstrumentDAO) ListDue(ctx context.Context, asOf *time.Time) ([]*PdcInstrument, error) {
	date := time.Now().UTC()
	if asOf != nil {
		date = *asOf
	}
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "state", Operator: query.In, Value: []string{PdcStateHeld, PdcStateDeposited}},
		{Field: "due_date", Operator: query.LessEqual, Value: date},
	}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
