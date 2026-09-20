package sequence

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrSequenceNotFound = errors.New("sequence not found")

type Reservation struct {
	Value  int64
	Number string
}

type DAO interface {
	Reserve(ctx context.Context, organizationID uint64, code string, now time.Time) (*Reservation, error)
	ReserveTx(ctx context.Context, tx *gorm.DB, organizationID uint64, code string, now time.Time) (*Reservation, error)
}

type DocumentSequenceDAO interface {
	List(ctx context.Context, q *query.Query) (*query.Page[DocumentSequence], error)
	Create(ctx context.Context, entity *DocumentSequence) (*DocumentSequence, error)
	Update(ctx context.Context, entity *DocumentSequence) (*DocumentSequence, error)
}

type sequenceDAO struct {
	db *gorm.DB
}

func NewDAO(db *gorm.DB) DAO {
	return sequenceDAO{db: db}
}

func (s sequenceDAO) Reserve(ctx context.Context, organizationID uint64, code string, now time.Time) (*Reservation, error) {
	var reservation *Reservation
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		reservation, err = s.ReserveTx(ctx, tx, organizationID, code, now)
		return err
	})
	if err != nil {
		return nil, err
	}
	return reservation, nil
}

func (s sequenceDAO) ReserveTx(ctx context.Context, tx *gorm.DB, organizationID uint64, code string, now time.Time) (*Reservation, error) {
	var seq DocumentSequence
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("organization_id = ?", organizationID).
		Where("code = ?", code).
		First(&seq).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSequenceNotFound
		}
		return nil, err
	}

	if shouldReset(seq.ResetPeriod, seq.UpdatedAt, now) {
		seq.NextNumber = 1
	}

	value := seq.NextNumber
	seq.NextNumber++
	seq.UpdatedAt = now

	err = tx.WithContext(ctx).Model(&DocumentSequence{}).
		Where("id = ?", seq.ID).
		Updates(map[string]any{
			"next_number": seq.NextNumber,
			"updated_at":  now,
		}).Error
	if err != nil {
		return nil, err
	}

	return &Reservation{Value: value, Number: seq.Format(value)}, nil
}
