package accounting

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type PeriodCloseDAOMock struct {
	FindByPeriodFunc func(ctx context.Context, periodID uint64) (*PeriodCloseEntry, error)
	CreateFunc       func(ctx context.Context, entry *PeriodCloseEntry) (*PeriodCloseEntry, error)
	UpdateFunc       func(ctx context.Context, entry *PeriodCloseEntry) (*PeriodCloseEntry, error)
}

func (m PeriodCloseDAOMock) FindByPeriod(ctx context.Context, periodID uint64) (*PeriodCloseEntry, error) {
	if m.FindByPeriodFunc != nil {
		return m.FindByPeriodFunc(ctx, periodID)
	}
	return nil, nil
}

func (m PeriodCloseDAOMock) Create(ctx context.Context, entry *PeriodCloseEntry) (*PeriodCloseEntry, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entry)
	}
	entry.ID = 1
	return entry, nil
}

func (m PeriodCloseDAOMock) CreateTx(ctx context.Context, _ *gorm.DB, entry *PeriodCloseEntry) (*PeriodCloseEntry, error) {
	return m.Create(ctx, entry)
}

func (m PeriodCloseDAOMock) Update(ctx context.Context, entry *PeriodCloseEntry) (*PeriodCloseEntry, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entry)
	}
	return entry, nil
}

type PeriodAccountBalanceDAOMock struct {
	SumByAccountAndPeriodFunc func(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]PeriodAccountBalance, error)
}

func (m PeriodAccountBalanceDAOMock) SumByAccountAndPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]PeriodAccountBalance, error) {
	if m.SumByAccountAndPeriodFunc != nil {
		return m.SumByAccountAndPeriodFunc(ctx, organizationID, dateStart, dateEnd)
	}
	return nil, nil
}
