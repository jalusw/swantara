package sequence

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	dao DAO
	now func() time.Time
}

func NewSequenceService(dao DAO) Service {
	return Service{dao: dao, now: time.Now}
}

func (s Service) Next(ctx context.Context, organizationID uint64, code string) (string, error) {
	return s.next(ctx, nil, organizationID, code)
}

func (s Service) NextTx(ctx context.Context, tx *gorm.DB, organizationID uint64, code string) (string, error) {
	return s.next(ctx, tx, organizationID, code)
}

func (s Service) next(ctx context.Context, tx *gorm.DB, organizationID uint64, code string) (string, error) {
	if tx == nil {
		reservation, err := s.dao.Reserve(ctx, organizationID, code, s.now())
		if err != nil {
			return "", err
		}
		return reservation.Number, nil
	}
	reservation, err := s.dao.ReserveTx(ctx, tx, organizationID, code, s.now())
	if err != nil {
		return "", err
	}
	return reservation.Number, nil
}
