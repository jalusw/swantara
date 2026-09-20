package accounting

import (
	"context"
	"time"
)

type CashFlowDAOMock struct {
	GetOpeningCashFunc   func(ctx context.Context, organizationID uint64, dateStart time.Time) (float64, error)
	GenerateByPeriodFunc func(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]CashFlowLine, error)
}

func (m CashFlowDAOMock) GetOpeningCash(ctx context.Context, organizationID uint64, dateStart time.Time) (float64, error) {
	if m.GetOpeningCashFunc != nil {
		return m.GetOpeningCashFunc(ctx, organizationID, dateStart)
	}
	return 0, nil
}

func (m CashFlowDAOMock) GenerateByPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]CashFlowLine, error) {
	if m.GenerateByPeriodFunc != nil {
		return m.GenerateByPeriodFunc(ctx, organizationID, dateStart, dateEnd)
	}
	return nil, nil
}
