package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"gorm.io/gorm"
)

type GeneralLedgerLine struct {
	LineID       uint64        `json:"line_id"`
	EntryID      uint64        `json:"entry_id"`
	Date         time.Time     `json:"date"`
	Ref          *string       `json:"ref"`
	Name         *string       `json:"name"`
	AccountID    uint64        `json:"account_id"`
	ContactID    *uint64       `json:"contact_id"`
	Debit        amount.Amount `json:"debit"`
	Credit       amount.Amount `json:"credit"`
	Balance      amount.Amount `json:"balance"`
	RunningTotal amount.Amount `json:"running_total"`
}

type GeneralLedgerDAO interface {
	ListByAccount(ctx context.Context, organizationID, accountID uint64, dateStart, dateEnd time.Time) ([]GeneralLedgerLine, error)
}

type generalLedgerDAO struct {
	db *gorm.DB
}

func NewGeneralLedgerDAO(db *gorm.DB) GeneralLedgerDAO {
	return generalLedgerDAO{db: db}
}

func (d generalLedgerDAO) ListByAccount(ctx context.Context, organizationID, accountID uint64, dateStart, dateEnd time.Time) ([]GeneralLedgerLine, error) {
	type row struct {
		LineID    uint64
		EntryID   uint64
		Date      time.Time
		Ref       *string
		Name      *string
		AccountID uint64
		ContactID *uint64
		Debit     float64
		Credit    float64
	}
	var rows []row
	err := d.db.WithContext(ctx).
		Table("journal_lines AS l").
		Joins("JOIN journal_entrys AS m ON m.id = l.entry_id").
		Where("m.organization_id = ? AND m.state = ?", organizationID, EntryStatePosted).
		Where("l.account_id = ?", accountID).
		Where("m.date >= ? AND m.date <= ?", dateStart, dateEnd).
		Where("m.deleted_at IS NULL AND l.deleted_at IS NULL").
		Order("m.date ASC, m.id ASC, l.id ASC").
		Select("l.id AS line_id, l.entry_id, m.date, m.ref, l.name, l.account_id, l.contact_id, l.debit, l.credit").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	lines := make([]GeneralLedgerLine, len(rows))
	running := amount.Zero()
	for i, r := range rows {
		debit := amount.FromFloat64(r.Debit)
		credit := amount.FromFloat64(r.Credit)
		balance := debit.Sub(credit)
		running = running.Add(balance)
		lines[i] = GeneralLedgerLine{
			LineID:       r.LineID,
			EntryID:      r.EntryID,
			Date:         r.Date,
			Ref:          r.Ref,
			Name:         r.Name,
			AccountID:    r.AccountID,
			ContactID:    r.ContactID,
			Debit:        debit,
			Credit:       credit,
			Balance:      balance,
			RunningTotal: running,
		}
	}
	return lines, nil
}

type GeneralLedgerService struct {
	ledger GeneralLedgerDAO
}

func NewGeneralLedgerService(ledger GeneralLedgerDAO) GeneralLedgerService {
	return GeneralLedgerService{ledger: ledger}
}

func (s GeneralLedgerService) ListByAccount(ctx context.Context, organizationID, accountID uint64, dateStart, dateEnd time.Time) ([]GeneralLedgerLine, error) {
	if accountID == 0 {
		return nil, ErrInvalidLine
	}
	return s.ledger.ListByAccount(ctx, organizationID, accountID, dateStart, dateEnd)
}
