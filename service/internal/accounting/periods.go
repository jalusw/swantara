package accounting

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

const (
	TaxPeriodStateOpen   = "open"
	TaxPeriodStateClosed = "closed"
	TaxPeriodStateLocked = "locked"
)

const (
	TaxPeriodTypeStandard   = "standard"
	TaxPeriodTypeAdjustment = "adjustment"
)

type TaxPeriod struct {
	model.Base
	OrganizationID uint64     `gorm:"not null" json:"organization_id"`
	TaxYearID      uint64     `gorm:"not null" json:"tax_year_id"`
	Name           string     `gorm:"not null" json:"name"`
	DateStart      *time.Time `gorm:"type:date" json:"date_start"`
	DateEnd        *time.Time `gorm:"type:date" json:"date_end"`
	State          string     `gorm:"type:text" json:"state"`
	PeriodType     string     `gorm:"type:text;default:standard" json:"period_type"`
}

func (TaxPeriod) TableName() string {
	return "tax_periods"
}

type TaxPeriodDAO interface {
	dao.CRUD[TaxPeriod]
	ListByOrganization(ctx context.Context, organizationID uint64) ([]*TaxPeriod, error)
	FindByDate(ctx context.Context, organizationID uint64, date time.Time) (*TaxPeriod, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, period *TaxPeriod) (*TaxPeriod, error)
}

type taxPeriodDAO struct {
	dao.Base[TaxPeriod]
	db *gorm.DB
}

func NewTaxPeriodDAO(db *gorm.DB) TaxPeriodDAO {
	return taxPeriodDAO{Base: dao.NewBase[TaxPeriod](db), db: db}
}

func (d taxPeriodDAO) ListByOrganization(ctx context.Context, organizationID uint64) ([]*TaxPeriod, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d taxPeriodDAO) FindByDate(ctx context.Context, organizationID uint64, date time.Time) (*TaxPeriod, error) {
	var period TaxPeriod
	err := d.db.WithContext(ctx).
		Where("organization_id = ? AND date_start <= ? AND date_end >= ? AND deleted_at IS NULL", organizationID, date, date).
		Order("date_start ASC").
		First(&period).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &period, nil
}

func (d taxPeriodDAO) UpdateTx(ctx context.Context, tx *gorm.DB, period *TaxPeriod) (*TaxPeriod, error) {
	if err := tx.WithContext(ctx).Save(period).Error; err != nil {
		return nil, err
	}
	return period, nil
}

type PeriodLookup interface {
	AssertOpen(ctx context.Context, organizationID uint64, date time.Time) error
}

type TaxPeriodService struct {
	periods    TaxPeriodDAO
	years      dao.CRUD[reference.TaxYear]
	closeGuard PeriodCloseDAO
}

func NewTaxPeriodService(periods TaxPeriodDAO, years dao.CRUD[reference.TaxYear]) TaxPeriodService {
	return TaxPeriodService{periods: periods, years: years}
}

func (s TaxPeriodService) WithCloseGuard(guard PeriodCloseDAO) TaxPeriodService {
	s.closeGuard = guard
	return s
}

func (s TaxPeriodService) List(ctx context.Context, q *query.Query) (*query.Page[TaxPeriod], error) {
	return s.periods.List(ctx, q)
}

func (s TaxPeriodService) Find(ctx context.Context, id uint64) (*TaxPeriod, error) {
	return s.periods.Find(ctx, id)
}

func (s TaxPeriodService) Create(ctx context.Context, period *TaxPeriod) (*TaxPeriod, error) {
	if period.State == "" {
		period.State = TaxPeriodStateOpen
	}
	if period.PeriodType == "" {
		period.PeriodType = TaxPeriodTypeStandard
	}
	if !validPeriodState(period.State) {
		return nil, ErrInvalidPeriodState
	}
	if period.PeriodType != TaxPeriodTypeStandard && period.PeriodType != TaxPeriodTypeAdjustment {
		return nil, ErrInvalidPeriod
	}
	if period.DateStart == nil || period.DateEnd == nil || period.DateEnd.Before(*period.DateStart) {
		return nil, ErrInvalidPeriod
	}

	year, err := s.years.Find(ctx, period.TaxYearID)
	if err != nil {
		return nil, err
	}
	if year == nil || year.OrganizationID == nil || *year.OrganizationID != period.OrganizationID {
		return nil, reference.ErrTaxYearNotFound
	}

	return s.periods.Create(ctx, period)
}

func (s TaxPeriodService) Close(ctx context.Context, periodID uint64) (*TaxPeriod, error) {
	period, err := s.periods.Find(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, ErrPeriodNotFound
	}
	if period.State == TaxPeriodStateLocked {
		return nil, ErrPeriodLocked
	}
	if period.State == TaxPeriodStateClosed {
		return nil, ErrPeriodAlreadyClosed
	}
	if s.closeGuard != nil {
		entry, err := s.closeGuard.FindByPeriod(ctx, periodID)
		if err != nil {
			return nil, err
		}
		if entry == nil {
			return nil, ErrCloseRequiresEntries
		}
	}

	period.State = TaxPeriodStateClosed
	return s.periods.Update(ctx, period)
}

func (s TaxPeriodService) Lock(ctx context.Context, periodID uint64) (*TaxPeriod, error) {
	period, err := s.periods.Find(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, ErrPeriodNotFound
	}
	if period.State != TaxPeriodStateClosed {
		return nil, ErrPeriodNotClosed
	}

	period.State = TaxPeriodStateLocked
	return s.periods.Update(ctx, period)
}

func (s TaxPeriodService) Open(ctx context.Context, periodID uint64) (*TaxPeriod, error) {
	period, err := s.periods.Find(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, ErrPeriodNotFound
	}
	if period.State == TaxPeriodStateLocked {
		return nil, ErrPeriodLocked
	}

	period.State = TaxPeriodStateOpen
	return s.periods.Update(ctx, period)
}

func (s TaxPeriodService) AssertOpen(ctx context.Context, organizationID uint64, date time.Time) error {
	period, err := s.periods.FindByDate(ctx, organizationID, date)
	if err != nil {
		return err
	}
	if period == nil {
		return ErrPeriodNotFound
	}
	if period.State == TaxPeriodStateLocked {
		return ErrPeriodLocked
	}
	if period.State == TaxPeriodStateClosed {
		return ErrPeriodNotClosed
	}
	if period.State != TaxPeriodStateOpen {
		return ErrPeriodLocked
	}
	if year, err := s.years.Find(ctx, period.TaxYearID); err == nil && year != nil && year.State != nil && *year.State == TaxPeriodStateLocked {
		return ErrPeriodLocked
	}
	return nil
}

func (s TaxPeriodService) LockTaxYear(ctx context.Context, taxYearID uint64) error {
	year, err := s.years.Find(ctx, taxYearID)
	if err != nil {
		return err
	}
	if year == nil {
		return reference.ErrTaxYearNotFound
	}
	if year.OrganizationID == nil {
		return reference.ErrTaxYearNotFound
	}
	periods, err := s.periods.ListByOrganization(ctx, *year.OrganizationID)
	if err != nil {
		return err
	}
	for _, period := range periods {
		if period.TaxYearID != taxYearID {
			continue
		}
		if period.State == TaxPeriodStateLocked {
			continue
		}
		if period.State != TaxPeriodStateClosed {
			return ErrPeriodNotClosed
		}
		period.State = TaxPeriodStateLocked
		if _, err := s.periods.Update(ctx, period); err != nil {
			return err
		}
	}
	locked := TaxPeriodStateLocked
	year.State = &locked
	if _, err := s.years.Update(ctx, year); err != nil {
		return err
	}
	return nil
}

func validPeriodState(state string) bool {
	switch state {
	case TaxPeriodStateOpen, TaxPeriodStateClosed, TaxPeriodStateLocked:
		return true
	default:
		return false
	}
}
