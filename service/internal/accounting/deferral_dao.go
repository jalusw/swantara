package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type DeferredScheduleDAO interface {
	dao.CRUD[DeferredSchedule]
	ListRunning(ctx context.Context, organizationID *uint64) ([]*DeferredSchedule, error)
	CreateTx(ctx context.Context, tx *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error)
}

type deferredScheduleDAO struct {
	dao.Base[DeferredSchedule]
	db *gorm.DB
}

func NewDeferredScheduleDAO(db *gorm.DB) DeferredScheduleDAO {
	return deferredScheduleDAO{Base: dao.NewBase[DeferredSchedule](db), db: db}
}

func (d deferredScheduleDAO) ListRunning(ctx context.Context, organizationID *uint64) ([]*DeferredSchedule, error) {
	query := d.db.WithContext(ctx).Where("state = ?", DeferredStateRunning)
	if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	}
	var entities []DeferredSchedule
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*DeferredSchedule, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

func (d deferredScheduleDAO) CreateTx(ctx context.Context, tx *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error) {
	if err := tx.WithContext(ctx).Create(schedule).Error; err != nil {
		return nil, err
	}
	return schedule, nil
}

func (d deferredScheduleDAO) UpdateTx(ctx context.Context, tx *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error) {
	if err := tx.WithContext(ctx).Save(schedule).Error; err != nil {
		return nil, err
	}
	return schedule, nil
}

type DeferredScheduleLineDAO interface {
	dao.CRUD[DeferredScheduleLine]
	ListBySchedule(ctx context.Context, scheduleID uint64) ([]*DeferredScheduleLine, error)
	ListDue(ctx context.Context, scheduleID uint64, asOf time.Time) ([]*DeferredScheduleLine, error)
	CreateTx(ctx context.Context, tx *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error)
}

type deferredScheduleLineDAO struct {
	dao.Base[DeferredScheduleLine]
	db *gorm.DB
}

func NewDeferredScheduleLineDAO(db *gorm.DB) DeferredScheduleLineDAO {
	return deferredScheduleLineDAO{Base: dao.NewBase[DeferredScheduleLine](db), db: db}
}

func (d deferredScheduleLineDAO) ListBySchedule(ctx context.Context, scheduleID uint64) ([]*DeferredScheduleLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "schedule_id", Operator: query.Equal, Value: scheduleID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d deferredScheduleLineDAO) ListDue(ctx context.Context, scheduleID uint64, asOf time.Time) ([]*DeferredScheduleLine, error) {
	var entities []DeferredScheduleLine
	if err := d.db.WithContext(ctx).
		Where("schedule_id = ? AND posted = ? AND recognition_date IS NOT NULL AND recognition_date <= ?", scheduleID, false, asOf).
		Order("sequence ASC").
		Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*DeferredScheduleLine, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

func (d deferredScheduleLineDAO) CreateTx(ctx context.Context, tx *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error) {
	if err := tx.WithContext(ctx).Create(line).Error; err != nil {
		return nil, err
	}
	return line, nil
}

func (d deferredScheduleLineDAO) UpdateTx(ctx context.Context, tx *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error) {
	if err := tx.WithContext(ctx).Save(line).Error; err != nil {
		return nil, err
	}
	return line, nil
}
