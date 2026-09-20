package accounting

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type AccountBalance struct {
	AccountID uint64  `json:"account_id"`
	Debit     float64 `json:"debit"`
	Credit    float64 `json:"credit"`
}

type AccountBalanceDAO interface {
	ListBalancesByPeriod(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountBalance, error)
}

type accountBalanceDAO struct {
	db *gorm.DB
}

func NewAccountBalanceDAO(db *gorm.DB) AccountBalanceDAO {
	return accountBalanceDAO{db: db}
}

func (d accountBalanceDAO) ListBalancesByPeriod(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountBalance, error) {
	rows := []AccountBalance{}
	if err := d.db.WithContext(ctx).
		Model(&JournalLine{}).
		Joins("JOIN journal_entrys ON journal_entrys.id = journal_lines.entry_id").
		Where("journal_entrys.organization_id = ? AND journal_entrys.state = ?", organizationID, EntryStatePosted).
		Where("journal_entrys.date >= ? AND journal_entrys.date <= ?", start, end).
		Select("journal_lines.account_id AS account_id, COALESCE(SUM(journal_lines.debit), 0) AS debit, COALESCE(SUM(journal_lines.credit), 0) AS credit").
		Group("journal_lines.account_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
