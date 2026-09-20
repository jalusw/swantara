package accounting

import (
	"context"
	"time"
)

type EquityDAOMock struct {
	EquityBalancesByPeriodFunc func(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]EquityMovement, error)
	OpeningEquityBalanceFunc   func(ctx context.Context, organizationID uint64, dateStart time.Time) ([]EquityMovement, error)
	NetIncomeForPeriodFunc     func(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) (float64, error)
}

func (m EquityDAOMock) EquityBalancesByPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]EquityMovement, error) {
	if m.EquityBalancesByPeriodFunc != nil {
		return m.EquityBalancesByPeriodFunc(ctx, organizationID, dateStart, dateEnd)
	}
	return nil, nil
}

func (m EquityDAOMock) OpeningEquityBalance(ctx context.Context, organizationID uint64, dateStart time.Time) ([]EquityMovement, error) {
	if m.OpeningEquityBalanceFunc != nil {
		return m.OpeningEquityBalanceFunc(ctx, organizationID, dateStart)
	}
	return nil, nil
}

func (m EquityDAOMock) NetIncomeForPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) (float64, error) {
	if m.NetIncomeForPeriodFunc != nil {
		return m.NetIncomeForPeriodFunc(ctx, organizationID, dateStart, dateEnd)
	}
	return 0, nil
}
