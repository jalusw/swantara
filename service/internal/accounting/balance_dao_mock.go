package accounting

import (
	"context"
	"time"
)

type AccountBalanceDAOMock struct {
	ListBalancesByPeriodFunc func(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountBalance, error)
}

func (m AccountBalanceDAOMock) ListBalancesByPeriod(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountBalance, error) {
	if m.ListBalancesByPeriodFunc != nil {
		return m.ListBalancesByPeriodFunc(ctx, organizationID, start, end)
	}
	return nil, nil
}
