package amount

import (
	"context"
	"time"
)

type staticRateSource map[string]Amount

func (m staticRateSource) Rate(_ context.Context, currencyCode string, _ uint64, _ RateType, _ time.Time) (Amount, error) {
	rate, ok := m[currencyCode]
	if !ok {
		return Zero(), nil
	}
	return rate, nil
}

type failingRateSource struct {
	err error
}

func (s failingRateSource) Rate(_ context.Context, _ string, _ uint64, _ RateType, _ time.Time) (Amount, error) {
	return Amount{}, s.err
}

type dateRatedSource map[string]map[string]Amount

func (m dateRatedSource) Rate(_ context.Context, currencyCode string, _ uint64, _ RateType, date time.Time) (Amount, error) {
	rate, ok := m[currencyCode][date.Format("2006-01-02")]
	if !ok {
		return Amount{}, ErrInvalidRate
	}
	return rate, nil
}
