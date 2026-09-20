package accounting

import "context"

type TrialBalanceDAOMock struct {
	GenerateByPeriodFunc func(ctx context.Context, organizationID, periodID uint64) ([]TrialBalanceLine, error)
	GenerateAsOfFunc     func(ctx context.Context, organizationID uint64, asOfDate string, includeZero bool) ([]TrialBalanceLine, error)
}

func (m TrialBalanceDAOMock) GenerateByPeriod(ctx context.Context, organizationID, periodID uint64) ([]TrialBalanceLine, error) {
	if m.GenerateByPeriodFunc != nil {
		return m.GenerateByPeriodFunc(ctx, organizationID, periodID)
	}
	return nil, nil
}

func (m TrialBalanceDAOMock) GenerateAsOf(ctx context.Context, organizationID uint64, asOfDate string, includeZero bool) ([]TrialBalanceLine, error) {
	if m.GenerateAsOfFunc != nil {
		return m.GenerateAsOfFunc(ctx, organizationID, asOfDate, includeZero)
	}
	return nil, nil
}
