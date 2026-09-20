package accounting

import (
	"context"
)

type FxRevaluationDAOMock struct {
	OpenForeignPositionsFunc func(ctx context.Context, organizationID uint64, baseCurrency string) ([]FxPosition, error)
}

func (m FxRevaluationDAOMock) OpenForeignPositions(ctx context.Context, organizationID uint64, baseCurrency string) ([]FxPosition, error) {
	if m.OpenForeignPositionsFunc != nil {
		return m.OpenForeignPositionsFunc(ctx, organizationID, baseCurrency)
	}
	return nil, nil
}
