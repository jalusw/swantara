package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

type PeriodCloseEntry struct {
	model.Base
	OrganizationID uint64     `gorm:"not null" json:"organization_id"`
	PeriodID       uint64     `gorm:"not null" json:"period_id"`
	EntryID        *uint64    `json:"entry_id"`
	ClosingDate    time.Time  `gorm:"type:date;not null" json:"closing_date"`
	State          string     `gorm:"type:text" json:"state"`
	PostedAt       *time.Time `json:"posted_at"`
}

func (PeriodCloseEntry) TableName() string {
	return "period_close_entries"
}

const (
	PeriodCloseStateDraft  = "draft"
	PeriodCloseStatePosted = "posted"
)

type PeriodCloseDAO interface {
	FindByPeriod(ctx context.Context, periodID uint64) (*PeriodCloseEntry, error)
	Create(ctx context.Context, entry *PeriodCloseEntry) (*PeriodCloseEntry, error)
	CreateTx(ctx context.Context, tx *gorm.DB, entry *PeriodCloseEntry) (*PeriodCloseEntry, error)
	Update(ctx context.Context, entry *PeriodCloseEntry) (*PeriodCloseEntry, error)
}

type periodCloseDAO struct {
	db *gorm.DB
}

func NewPeriodCloseDAO(db *gorm.DB) PeriodCloseDAO {
	return periodCloseDAO{db: db}
}

func (d periodCloseDAO) FindByPeriod(ctx context.Context, periodID uint64) (*PeriodCloseEntry, error) {
	var entry PeriodCloseEntry
	err := d.db.WithContext(ctx).
		Where("period_id = ? AND deleted_at IS NULL", periodID).
		First(&entry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

func (d periodCloseDAO) Create(ctx context.Context, entry *PeriodCloseEntry) (*PeriodCloseEntry, error) {
	if err := d.db.WithContext(ctx).Create(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

func (d periodCloseDAO) CreateTx(ctx context.Context, tx *gorm.DB, entry *PeriodCloseEntry) (*PeriodCloseEntry, error) {
	if err := tx.WithContext(ctx).Create(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

func (d periodCloseDAO) Update(ctx context.Context, entry *PeriodCloseEntry) (*PeriodCloseEntry, error) {
	if err := d.db.WithContext(ctx).Save(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

type PeriodAccountBalanceDAO interface {
	SumByAccountAndPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]PeriodAccountBalance, error)
}

type PeriodAccountBalance struct {
	AccountID   uint64  `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	AccountType string  `json:"account_type"`
	TotalDebit  float64 `json:"total_debit"`
	TotalCredit float64 `json:"total_credit"`
}

type periodAccountBalanceDAO struct {
	db *gorm.DB
}

func NewPeriodAccountBalanceDAO(db *gorm.DB) PeriodAccountBalanceDAO {
	return periodAccountBalanceDAO{db: db}
}

func (d periodAccountBalanceDAO) SumByAccountAndPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]PeriodAccountBalance, error) {
	type row struct {
		AccountID   uint64
		AccountCode string
		AccountName string
		AccountType string
		TotalDebit  float64
		TotalCredit float64
	}

	var rows []row
	err := d.db.WithContext(ctx).
		Table("journal_lines AS l").
		Joins("JOIN journal_entrys AS m ON m.id = l.entry_id").
		Joins("JOIN accounts AS a ON a.id = l.account_id").
		Where("m.organization_id = ? AND m.state = ?", organizationID, EntryStatePosted).
		Where("m.date >= ? AND m.date <= ?", dateStart, dateEnd).
		Where("m.deleted_at IS NULL AND l.deleted_at IS NULL AND a.deleted_at IS NULL").
		Select(`
			l.account_id,
			a.code AS account_code,
			a.name AS account_name,
			a.type AS account_type,
			COALESCE(SUM(l.debit), 0) AS total_debit,
			COALESCE(SUM(l.credit), 0) AS total_credit
		`).
		Group("l.account_id, a.code, a.name, a.type").
		Order("a.code ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	balances := make([]PeriodAccountBalance, len(rows))
	for i, r := range rows {
		balances[i] = PeriodAccountBalance(r)
	}
	return balances, nil
}

type PeriodCloseService struct {
	periods         TaxPeriodDAO
	balances        PeriodAccountBalanceDAO
	movements       JournalEntryDAO
	lines           JournalLineDAO
	periodCloses    PeriodCloseDAO
	poster          PostingService
	journalResolver ClosingJournalResolver
	accounts        AccountFinder
	txer            db.Transactioner
	now             func() time.Time
}

func NewPeriodCloseService(
	periods TaxPeriodDAO,
	balances PeriodAccountBalanceDAO,
	movements JournalEntryDAO,
	lines JournalLineDAO,
	periodCloses PeriodCloseDAO,
	poster PostingService,
) PeriodCloseService {
	return PeriodCloseService{
		periods:      periods,
		balances:     balances,
		movements:    movements,
		lines:        lines,
		periodCloses: periodCloses,
		poster:       poster,
		now:          time.Now,
	}
}

func (s PeriodCloseService) WithJournalResolver(journalResolver ClosingJournalResolver) PeriodCloseService {
	s.journalResolver = journalResolver
	return s
}

func (s PeriodCloseService) WithTransactioner(txer db.Transactioner) PeriodCloseService {
	s.txer = txer
	return s
}

func (s PeriodCloseService) WithAccounts(accounts AccountFinder) PeriodCloseService {
	s.accounts = accounts
	return s
}

func (s PeriodCloseService) Close(ctx context.Context, organizationID, periodID uint64, retainedEarningsAccountID uint64) (*PeriodCloseEntry, error) {
	period, err := s.periods.Find(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if period == nil || period.OrganizationID != organizationID {
		return nil, ErrPeriodNotFound
	}
	if period.State != TaxPeriodStateOpen {
		return nil, ErrPeriodNotOpen
	}
	if period.DateStart == nil || period.DateEnd == nil {
		return nil, ErrInvalidPeriod
	}
	if retainedEarningsAccountID == 0 {
		return nil, ErrInvalidRetainedEarnings
	}
	if s.accounts != nil {
		account, err := s.accounts.Find(ctx, retainedEarningsAccountID)
		if err != nil {
			return nil, err
		}
		if account == nil || account.OrganizationID != organizationID {
			return nil, ErrInvalidRetainedEarnings
		}
		if !account.Active {
			return nil, ErrAccountInactive
		}
		if account.Type != AccountTypeEquity {
			return nil, ErrInvalidRetainedEarnings
		}
	}

	existing, err := s.periodCloses.FindByPeriod(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPeriodAlreadyClosed
	}

	balances, err := s.balances.SumByAccountAndPeriod(ctx, organizationID, *period.DateStart, *period.DateEnd)
	if err != nil {
		return nil, err
	}

	now := s.now().UTC()
	closeLines := make([]PostingLine, 0)
	netIncome := amount.Zero()

	for _, b := range balances {
		if !isIncomeOrExpense(b.AccountType) {
			continue
		}
		debit := amount.FromFloat64(b.TotalDebit)
		credit := amount.FromFloat64(b.TotalCredit)
		balance := debit.Sub(credit)

		if balance.IsZero() {
			continue
		}

		if isIncome(b.AccountType) {
			if balance.IsNegative() {
				closeLines = append(closeLines, PostingLine{
					AccountID: b.AccountID,
					Name:      "Close " + b.AccountName,
					Debit:     balance.Abs(),
				})
			} else {
				closeLines = append(closeLines, PostingLine{
					AccountID: b.AccountID,
					Name:      "Close " + b.AccountName,
					Credit:    balance,
				})
			}
			netIncome = netIncome.Sub(balance)
		} else {
			if balance.IsPositive() {
				closeLines = append(closeLines, PostingLine{
					AccountID: b.AccountID,
					Name:      "Close " + b.AccountName,
					Credit:    balance,
				})
			} else {
				closeLines = append(closeLines, PostingLine{
					AccountID: b.AccountID,
					Name:      "Close " + b.AccountName,
					Debit:     balance.Abs(),
				})
			}
			netIncome = netIncome.Sub(balance)
		}
	}

	if len(closeLines) == 0 {
		if s.txer == nil {
			period.State = TaxPeriodStateClosed
			if _, err := s.periods.Update(ctx, period); err != nil {
				return nil, err
			}
			entry, err := s.periodCloses.Create(ctx, &PeriodCloseEntry{
				OrganizationID: organizationID,
				PeriodID:       periodID,
				ClosingDate:    *period.DateEnd,
				State:          PeriodCloseStatePosted,
				PostedAt:       &now,
			})
			if err != nil {
				return nil, err
			}
			return entry, nil
		}
		var entry *PeriodCloseEntry
		err := s.txer.Run(ctx, func(tx *gorm.DB) error {
			period.State = TaxPeriodStateClosed
			if _, err := s.periods.UpdateTx(ctx, tx, period); err != nil {
				return err
			}
			created, err := s.periodCloses.CreateTx(ctx, tx, &PeriodCloseEntry{
				OrganizationID: organizationID,
				PeriodID:       periodID,
				ClosingDate:    *period.DateEnd,
				State:          PeriodCloseStatePosted,
				PostedAt:       &now,
			})
			if err != nil {
				return err
			}
			entry = created
			return nil
		})
		return entry, err
	}

	if netIncome.IsPositive() {
		closeLines = append(closeLines, PostingLine{
			AccountID: retainedEarningsAccountID,
			Name:      "Net income to retained earnings",
			Credit:    netIncome,
		})
	} else if netIncome.IsNegative() {
		closeLines = append(closeLines, PostingLine{
			AccountID: retainedEarningsAccountID,
			Name:      "Net loss to retained earnings",
			Debit:     netIncome.Abs(),
		})
	}

	if s.journalResolver == nil {
		return nil, ErrClosingJournalMissing
	}
	journalID, rerr := s.journalResolver.ClosingJournalID(ctx, organizationID)
	if rerr != nil {
		return nil, rerr
	}
	if journalID == 0 {
		return nil, ErrClosingJournalMissing
	}
	if s.txer == nil {
		posted, err := s.poster.Post(ctx, PostRequest{
			OrganizationID: organizationID,
			JournalID:      journalID,
			Date:           *period.DateEnd,
			Description:    "Period close " + period.Name,
			Lines:          closeLines,
		})
		if err != nil {
			return nil, err
		}

		closeEntry, err := s.periodCloses.Create(ctx, &PeriodCloseEntry{
			OrganizationID: organizationID,
			PeriodID:       periodID,
			EntryID:        helper.Ptr(posted.ID),
			ClosingDate:    *period.DateEnd,
			State:          PeriodCloseStatePosted,
			PostedAt:       &now,
		})
		if err != nil {
			return nil, err
		}

		period.State = TaxPeriodStateClosed
		if _, err := s.periods.Update(ctx, period); err != nil {
			return nil, err
		}

		return closeEntry, nil
	}

	var result *PeriodCloseEntry
	err = s.txer.Run(ctx, func(tx *gorm.DB) error {
		posted, err := s.poster.PostTx(ctx, tx, PostRequest{
			OrganizationID: organizationID,
			JournalID:      journalID,
			Date:           *period.DateEnd,
			Description:    "Period close " + period.Name,
			Lines:          closeLines,
		})
		if err != nil {
			return err
		}
		closeEntry, err := s.periodCloses.CreateTx(ctx, tx, &PeriodCloseEntry{
			OrganizationID: organizationID,
			PeriodID:       periodID,
			EntryID:        helper.Ptr(posted.ID),
			ClosingDate:    *period.DateEnd,
			State:          PeriodCloseStatePosted,
			PostedAt:       &now,
		})
		if err != nil {
			return err
		}
		period.State = TaxPeriodStateClosed
		if _, err := s.periods.UpdateTx(ctx, tx, period); err != nil {
			return err
		}
		result = closeEntry
		return nil
	})
	return result, err
}

func isIncomeOrExpense(accountType string) bool {
	switch accountType {
	case AccountTypeIncome, AccountTypeExpense, AccountTypeCOGS, AccountTypeTax, AccountTypeDepreciation:
		return true
	default:
		return isIncome(accountType) || isExpense(accountType)
	}
}

func isIncome(accountType string) bool {
	return accountType == AccountTypeIncome
}

func isExpense(accountType string) bool {
	return accountType == AccountTypeExpense || accountType == AccountTypeCOGS || accountType == AccountTypeTax || accountType == AccountTypeDepreciation
}
