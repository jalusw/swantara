package sequence

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type DAOMock struct {
	ReserveFunc func(ctx context.Context, organizationID uint64, code string, now time.Time) (*Reservation, error)
}

func (m DAOMock) Reserve(ctx context.Context, organizationID uint64, code string, now time.Time) (*Reservation, error) {
	if m.ReserveFunc != nil {
		return m.ReserveFunc(ctx, organizationID, code, now)
	}
	return nil, nil
}

func (m DAOMock) ReserveTx(ctx context.Context, tx *gorm.DB, organizationID uint64, code string, now time.Time) (*Reservation, error) {
	return m.Reserve(ctx, organizationID, code, now)
}
