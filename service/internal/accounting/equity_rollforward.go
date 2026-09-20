package accounting

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

type EquityMovement struct {
	AccountID    uint64  `json:"account_id"`
	AccountCode  string  `json:"account_code"`
	AccountName  string  `json:"account_name"`
	AccountType  string  `json:"account_type"`
	MovementType string  `json:"movement_type"`
	Amount       float64 `json:"amount"`
}

type EquityRollforward struct {
	OpeningBalance       []EquityMovement `json:"opening_balance"`
	CapitalContributions []EquityMovement `json:"capital_contributions"`
	NetIncome            float64          `json:"net_income"`
	Dividends            []EquityMovement `json:"dividends"`
	OtherChanges         []EquityMovement `json:"other_changes"`
	ClosingBalance       []EquityMovement `json:"closing_balance"`
	PeriodID             uint64           `json:"period_id"`
	Balanced             bool             `json:"balanced"`
}

type EquityDAO interface {
	EquityBalancesByPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]EquityMovement, error)
	OpeningEquityBalance(ctx context.Context, organizationID uint64, dateStart time.Time) ([]EquityMovement, error)
	NetIncomeForPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) (float64, error)
}

type equityDAO struct {
	db *gorm.DB
}

func NewEquityDAO(db *gorm.DB) EquityDAO {
	return equityDAO{db: db}
}

func (d equityDAO) EquityBalancesByPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]EquityMovement, error) {
	type row struct {
		AccountID   uint64
		AccountCode string
		AccountName string
		AccountType string
		Amount      float64
	}

	var rows []row
	err := d.db.WithContext(ctx).
		Table("journal_lines AS l").
		Joins("JOIN journal_entrys AS m ON m.id = l.entry_id").
		Joins("JOIN accounts AS a ON a.id = l.account_id").
		Where("m.organization_id = ?", organizationID).
		Where("m.state = ?", EntryStatePosted).
		Where("m.date >= ? AND m.date <= ?", dateStart, dateEnd).
		Where("a.type = ?", AccountTypeEquity).
		Where("m.deleted_at IS NULL AND l.deleted_at IS NULL AND a.deleted_at IS NULL").
		Select(`
			l.account_id,
			a.code AS account_code,
			a.name AS account_name,
			a.type AS account_type,
			COALESCE(SUM(l.credit - l.debit), 0) AS amount
		`).
		Group("l.account_id, a.code, a.name, a.type").
		Order("a.code ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	movements := make([]EquityMovement, len(rows))
	for i, r := range rows {
		movements[i] = EquityMovement{
			AccountID:   r.AccountID,
			AccountCode: r.AccountCode,
			AccountName: r.AccountName,
			AccountType: r.AccountType,
			Amount:      r.Amount,
		}
	}
	return movements, nil
}

func (d equityDAO) OpeningEquityBalance(ctx context.Context, organizationID uint64, dateStart time.Time) ([]EquityMovement, error) {
	type row struct {
		AccountID   uint64
		AccountCode string
		AccountName string
		AccountType string
		Amount      float64
	}

	var rows []row
	err := d.db.WithContext(ctx).
		Table("journal_lines AS l").
		Joins("JOIN journal_entrys AS m ON m.id = l.entry_id").
		Joins("JOIN accounts AS a ON a.id = l.account_id").
		Where("m.organization_id = ?", organizationID).
		Where("m.state = ?", EntryStatePosted).
		Where("m.date < ?", dateStart).
		Where("a.type = ?", AccountTypeEquity).
		Where("m.deleted_at IS NULL AND l.deleted_at IS NULL AND a.deleted_at IS NULL").
		Select(`
			l.account_id,
			a.code AS account_code,
			a.name AS account_name,
			a.type AS account_type,
			COALESCE(SUM(l.credit - l.debit), 0) AS amount
		`).
		Group("l.account_id, a.code, a.name, a.type").
		Order("a.code ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	movements := make([]EquityMovement, len(rows))
	for i, r := range rows {
		movements[i] = EquityMovement{
			AccountID:   r.AccountID,
			AccountCode: r.AccountCode,
			AccountName: r.AccountName,
			AccountType: r.AccountType,
			Amount:      r.Amount,
		}
	}
	return movements, nil
}

func (d equityDAO) NetIncomeForPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) (float64, error) {
	var netIncome float64
	err := d.db.WithContext(ctx).
		Table("journal_lines AS l").
		Joins("JOIN journal_entrys AS m ON m.id = l.entry_id").
		Joins("JOIN accounts AS a ON a.id = l.account_id").
		Where("m.organization_id = ?", organizationID).
		Where("m.state = ?", EntryStatePosted).
		Where("m.date >= ? AND m.date <= ?", dateStart, dateEnd).
		Where("a.type IN ?", []string{AccountTypeIncome, AccountTypeExpense, AccountTypeCOGS, AccountTypeTax, AccountTypeDepreciation}).
		Where("m.deleted_at IS NULL AND l.deleted_at IS NULL AND a.deleted_at IS NULL").
		Select(`COALESCE(SUM(l.credit - l.debit), 0)`).
		Scan(&netIncome).Error
	if err != nil {
		return 0, err
	}
	return netIncome, nil
}

type EquityService struct {
	equity  EquityDAO
	periods TaxPeriodDAO
}

func NewEquityService(equity EquityDAO, periods TaxPeriodDAO) EquityService {
	return EquityService{equity: equity, periods: periods}
}

func (s EquityService) Generate(ctx context.Context, organizationID, periodID uint64) (*EquityRollforward, error) {
	period, err := s.periods.Find(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if period == nil || period.OrganizationID != organizationID {
		return nil, ErrPeriodNotFound
	}
	if period.DateStart == nil || period.DateEnd == nil {
		return nil, ErrInvalidPeriod
	}

	opening, err := s.equity.OpeningEquityBalance(ctx, organizationID, *period.DateStart)
	if err != nil {
		return nil, err
	}

	movements, err := s.equity.EquityBalancesByPeriod(ctx, organizationID, *period.DateStart, *period.DateEnd)
	if err != nil {
		return nil, err
	}

	netIncome, err := s.equity.NetIncomeForPeriod(ctx, organizationID, *period.DateStart, *period.DateEnd)
	if err != nil {
		return nil, err
	}

	closing := make([]EquityMovement, 0, len(opening)+len(movements))
	closingMap := make(map[uint64]*EquityMovement)
	for _, m := range opening {
		cp := m
		closingMap[m.AccountID] = &cp
	}
	for _, m := range movements {
		if existing, ok := closingMap[m.AccountID]; ok {
			existing.Amount += m.Amount
		} else {
			cp := m
			closingMap[m.AccountID] = &cp
		}
	}
	for _, v := range closingMap {
		closing = append(closing, *v)
	}

	capital, dividends, other := splitEquityMovements(movements)
	rollforward := &EquityRollforward{
		OpeningBalance:       opening,
		CapitalContributions: capital,
		NetIncome:            netIncome,
		Dividends:            dividends,
		OtherChanges:         other,
		ClosingBalance:       closing,
		PeriodID:             periodID,
	}
	rollforward.Balanced = equityBalanced(opening, movements, closing)
	return rollforward, nil
}

func equityBalanced(opening, movements, closing []EquityMovement) bool {
	openingSum := 0.0
	for _, m := range opening {
		openingSum += m.Amount
	}
	movementSum := 0.0
	for _, m := range movements {
		movementSum += m.Amount
	}
	closingSum := 0.0
	for _, m := range closing {
		closingSum += m.Amount
	}
	diff := closingSum - (openingSum + movementSum)
	if diff < 0 {
		diff = -diff
	}
	return diff <= 0.01
}

func splitEquityMovements(movements []EquityMovement) (capital, dividends, other []EquityMovement) {
	for _, m := range movements {
		name := strings.ToLower(m.AccountName)
		code := strings.ToLower(m.AccountCode)
		switch {
		case strings.Contains(name, "dividend") || strings.Contains(code, "divid"):
			m.MovementType = "dividend"
			dividends = append(dividends, m)
		case strings.Contains(name, "capital") || strings.Contains(name, "contribution") || strings.Contains(code, "capital"):
			m.MovementType = "capital_contribution"
			capital = append(capital, m)
		default:
			m.MovementType = "other"
			other = append(other, m)
		}
	}
	if capital == nil {
		capital = []EquityMovement{}
	}
	if dividends == nil {
		dividends = []EquityMovement{}
	}
	if other == nil {
		other = []EquityMovement{}
	}
	return
}
