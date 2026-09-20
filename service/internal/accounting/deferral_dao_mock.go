package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type DeferredScheduleDAOMock struct {
	dao.CRUDMock[DeferredSchedule]
	ListRunningFunc func(ctx context.Context, organizationID *uint64) ([]*DeferredSchedule, error)
	CreateTxFunc    func(ctx context.Context, tx *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error)
	UpdateTxFunc    func(ctx context.Context, tx *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error)
}

func (m DeferredScheduleDAOMock) ListRunning(ctx context.Context, organizationID *uint64) ([]*DeferredSchedule, error) {
	if m.ListRunningFunc != nil {
		return m.ListRunningFunc(ctx, organizationID)
	}
	return []*DeferredSchedule{}, nil
}

func (m DeferredScheduleDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, schedule)
	}
	return schedule, nil
}

func (m DeferredScheduleDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, schedule *DeferredSchedule) (*DeferredSchedule, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, schedule)
	}
	return schedule, nil
}

type DeferredScheduleLineDAOMock struct {
	dao.CRUDMock[DeferredScheduleLine]
	ListByScheduleFunc func(ctx context.Context, scheduleID uint64) ([]*DeferredScheduleLine, error)
	ListDueFunc        func(ctx context.Context, scheduleID uint64, asOf time.Time) ([]*DeferredScheduleLine, error)
	CreateTxFunc       func(ctx context.Context, tx *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error)
	UpdateTxFunc       func(ctx context.Context, tx *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error)
}

func (m DeferredScheduleLineDAOMock) ListBySchedule(ctx context.Context, scheduleID uint64) ([]*DeferredScheduleLine, error) {
	if m.ListByScheduleFunc != nil {
		return m.ListByScheduleFunc(ctx, scheduleID)
	}
	return []*DeferredScheduleLine{}, nil
}

func (m DeferredScheduleLineDAOMock) ListDue(ctx context.Context, scheduleID uint64, asOf time.Time) ([]*DeferredScheduleLine, error) {
	if m.ListDueFunc != nil {
		return m.ListDueFunc(ctx, scheduleID, asOf)
	}
	return []*DeferredScheduleLine{}, nil
}

func (m DeferredScheduleLineDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, line)
	}
	return line, nil
}

func (m DeferredScheduleLineDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, line *DeferredScheduleLine) (*DeferredScheduleLine, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, line)
	}
	return line, nil
}
