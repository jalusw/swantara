package reporting

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type KpiAccountSummary struct {
	OrganizationID uint64     `gorm:"primaryKey" json:"organization_id"`
	TaxPeriodID    uint64     `gorm:"primaryKey" json:"tax_period_id"`
	AccountID      uint64     `gorm:"primaryKey" json:"account_id"`
	OpeningDebit   float64    `gorm:"type:numeric(18,4)" json:"opening_debit"`
	OpeningCredit  float64    `gorm:"type:numeric(18,4)" json:"opening_credit"`
	PeriodDebit    float64    `gorm:"type:numeric(18,4)" json:"period_debit"`
	PeriodCredit   float64    `gorm:"type:numeric(18,4)" json:"period_credit"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (KpiAccountSummary) TableName() string {
	return "kpi_account_summaries"
}

type AccountPeriodBalance struct {
	AccountID uint64
	Debit     float64
	Credit    float64
}

type KpiSummaryDAO interface {
	OpeningBalances(ctx context.Context, organizationID uint64, before time.Time) ([]AccountPeriodBalance, error)
	PeriodActivity(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountPeriodBalance, error)
	ReplacePeriod(ctx context.Context, organizationID, periodID uint64, rows []KpiAccountSummary) error
	HasPeriod(ctx context.Context, organizationID, periodID uint64) (bool, error)
	TrialBalance(ctx context.Context, organizationID, periodID uint64) ([]TrialBalanceRow, error)
	ListPeriod(ctx context.Context, organizationID, periodID uint64) ([]KpiAccountSummary, error)
}

type kpiSummaryDAO struct {
	db *gorm.DB
}

func NewKpiSummaryDAO(db *gorm.DB) KpiSummaryDAO {
	return kpiSummaryDAO{db: db}
}

func (d kpiSummaryDAO) OpeningBalances(ctx context.Context, organizationID uint64, before time.Time) ([]AccountPeriodBalance, error) {
	rows := []AccountPeriodBalance{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT l.account_id, COALESCE(SUM(l.debit), 0) AS debit, COALESCE(SUM(l.credit), 0) AS credit
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date < ? AND m.deleted_at IS NULL
		GROUP BY l.account_id`, organizationID, before).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d kpiSummaryDAO) PeriodActivity(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountPeriodBalance, error) {
	rows := []AccountPeriodBalance{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT l.account_id, COALESCE(SUM(l.debit), 0) AS debit, COALESCE(SUM(l.credit), 0) AS credit
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND m.deleted_at IS NULL
		GROUP BY l.account_id`, organizationID, start, end).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d kpiSummaryDAO) ReplacePeriod(ctx context.Context, organizationID, periodID uint64, rows []KpiAccountSummary) error {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("organization_id = ? AND tax_period_id = ?", organizationID, periodID).Delete(&KpiAccountSummary{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(rows) > 0 {
		if err := tx.Create(&rows).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (d kpiSummaryDAO) HasPeriod(ctx context.Context, organizationID, periodID uint64) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&KpiAccountSummary{}).
		Where("organization_id = ? AND tax_period_id = ?", organizationID, periodID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d kpiSummaryDAO) TrialBalance(ctx context.Context, organizationID, periodID uint64) ([]TrialBalanceRow, error) {
	rows := []TrialBalanceRow{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT s.account_id,
		       a.code,
		       a.name,
		       a.type AS account_type,
		       s.opening_debit,
		       s.opening_credit,
		       s.period_debit,
		       s.period_credit,
		       s.opening_debit + s.period_debit AS closing_debit,
		       s.opening_credit + s.period_credit AS closing_credit
		FROM kpi_account_summaries s
		JOIN accounts a ON a.id = s.account_id AND a.deleted_at IS NULL
		WHERE s.organization_id = ? AND s.tax_period_id = ?
		ORDER BY a.code`, organizationID, periodID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d kpiSummaryDAO) ListPeriod(ctx context.Context, organizationID, periodID uint64) ([]KpiAccountSummary, error) {
	rows := []KpiAccountSummary{}
	err := d.db.WithContext(ctx).
		Where("organization_id = ? AND tax_period_id = ?", organizationID, periodID).
		Order("account_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
