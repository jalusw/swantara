package accounting

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type CashFlowLine struct {
	AccountID       uint64  `json:"account_id"`
	AccountCode     string  `json:"account_code"`
	Name            string  `json:"name"`
	Amount          float64 `json:"amount"`
	CashFlowSection string  `json:"cash_flow_section"`
	OriginType      string  `json:"origin_type"`
}

type CashFlowStatement struct {
	Operating   []CashFlowLine `json:"operating"`
	Investing   []CashFlowLine `json:"investing"`
	Financing   []CashFlowLine `json:"financing"`
	NetChange   float64        `json:"net_change"`
	OpeningCash float64        `json:"opening_cash"`
	ClosingCash float64        `json:"closing_cash"`
	Balanced    bool           `json:"balanced"`
}

type CashFlowDAO interface {
	GetOpeningCash(ctx context.Context, organizationID uint64, dateStart time.Time) (float64, error)
	GenerateByPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]CashFlowLine, error)
}

type cashFlowDAO struct {
	db *gorm.DB
}

func NewCashFlowDAO(db *gorm.DB) CashFlowDAO {
	return cashFlowDAO{db: db}
}

func (d cashFlowDAO) GetOpeningCash(ctx context.Context, organizationID uint64, dateStart time.Time) (float64, error) {
	var cash float64
	err := d.db.WithContext(ctx).
		Table("journal_lines AS l").
		Joins("JOIN journal_entrys AS m ON m.id = l.entry_id").
		Joins("JOIN accounts AS a ON a.id = l.account_id").
		Where("m.organization_id = ?", organizationID).
		Where("m.date < ?", dateStart).
		Where("m.state = ?", EntryStatePosted).
		Where("m.deleted_at IS NULL AND l.deleted_at IS NULL AND a.deleted_at IS NULL").
		Where("a.type IN (?, ?)", AccountTypeCash, AccountTypeBank).
		Select("COALESCE(SUM(l.debit - l.credit), 0)").
		Scan(&cash).Error
	if err != nil {
		return 0, err
	}
	return cash, nil
}

func (d cashFlowDAO) GenerateByPeriod(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) ([]CashFlowLine, error) {
	type row struct {
		AccountID   uint64
		AccountCode string
		Name        string
		Amount      float64
		OriginType  string
		AccountType string
	}

	var rows []row
	err := d.db.WithContext(ctx).
		Table("journal_lines AS l").
		Joins("JOIN journal_entrys AS m ON m.id = l.entry_id").
		Joins("JOIN accounts AS a ON a.id = l.account_id").
		Where("m.organization_id = ?", organizationID).
		Where("m.date >= ? AND m.date <= ?", dateStart, dateEnd).
		Where("m.state = ?", EntryStatePosted).
		Where("m.deleted_at IS NULL AND l.deleted_at IS NULL AND a.deleted_at IS NULL").
		Where("a.type IN (?, ?)", AccountTypeCash, AccountTypeBank).
		Select(`
			l.account_id,
			a.code AS account_code,
			COALESCE(l.name, '') AS name,
			COALESCE(l.debit, 0) - COALESCE(l.credit, 0) AS amount,
			m.origin_type,
			a.type AS account_type
		`).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	lines := make([]CashFlowLine, len(rows))
	for i, r := range rows {
		section := classifyCashFlowByOrigin(r.OriginType, r.AccountType)
		lines[i] = CashFlowLine{
			AccountID:       r.AccountID,
			AccountCode:     r.AccountCode,
			Name:            r.Name,
			Amount:          r.Amount,
			CashFlowSection: section,
			OriginType:      r.OriginType,
		}
	}
	return lines, nil
}

const (
	CashFlowOperating = "operating"
	CashFlowInvesting = "investing"
	CashFlowFinancing = "financing"
)

func classifyCashFlowByOrigin(originType string, _ string) string {
	if isInvestingOrigin(originType) {
		return CashFlowInvesting
	}
	if isFinancingOrigin(originType) {
		return CashFlowFinancing
	}
	return CashFlowOperating
}

func isInvestingOrigin(originType string) bool {
	switch originType {
	case OriginTypeInboundCost, OriginTypeAsset, OriginTypeAssetDisposal, OriginTypeStockCount, OriginTypeProductionOrder:
		return true
	default:
		return false
	}
}

func isFinancingOrigin(originType string) bool {
	switch originType {
	case OriginTypeFinancing, OriginTypeCapitalContribution, OriginTypeDividend, OriginTypeLoan, OriginTypeEquity, OriginTypeWithholding:
		return true
	default:
		return false
	}
}

type CashFlowService struct {
	cashFlow CashFlowDAO
}

func NewCashFlowService(cashFlow CashFlowDAO) CashFlowService {
	return CashFlowService{cashFlow: cashFlow}
}

func (s CashFlowService) Generate(ctx context.Context, organizationID uint64, dateStart, dateEnd time.Time) (*CashFlowStatement, error) {
	openingCash, err := s.cashFlow.GetOpeningCash(ctx, organizationID, dateStart)
	if err != nil {
		return nil, err
	}

	lines, err := s.cashFlow.GenerateByPeriod(ctx, organizationID, dateStart, dateEnd)
	if err != nil {
		return nil, err
	}

	cf := CashFlowStatement{
		OpeningCash: openingCash,
	}
	netChange := 0.0
	for _, line := range lines {
		switch line.CashFlowSection {
		case CashFlowOperating:
			cf.Operating = append(cf.Operating, line)
		case CashFlowInvesting:
			cf.Investing = append(cf.Investing, line)
		case CashFlowFinancing:
			cf.Financing = append(cf.Financing, line)
		}
		netChange += line.Amount
	}
	cf.NetChange = netChange
	cf.ClosingCash = openingCash + netChange
	cf.Balanced = absFloat(cf.ClosingCash-(cf.OpeningCash+cf.NetChange)) <= 0.01
	return &cf, nil
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
