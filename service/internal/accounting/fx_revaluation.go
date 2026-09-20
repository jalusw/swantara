package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type FxPosition struct {
	AccountID      uint64        `json:"account_id"`
	AccountType    string        `json:"account_type"`
	CurrencyCode   string        `json:"currency_code"`
	ForeignBalance amount.Amount `json:"foreign_balance"`
	BaseBalance    amount.Amount `json:"base_balance"`
}

type FxRevaluationDAO interface {
	OpenForeignPositions(ctx context.Context, organizationID uint64, baseCurrency string) ([]FxPosition, error)
}

type fxRevaluationDAO struct {
	db *gorm.DB
}

func NewFxRevaluationDAO(db *gorm.DB) FxRevaluationDAO {
	return fxRevaluationDAO{db: db}
}

func (d fxRevaluationDAO) OpenForeignPositions(ctx context.Context, organizationID uint64, baseCurrency string) ([]FxPosition, error) {
	type row struct {
		AccountID    uint64
		AccountType  string
		CurrencyCode string
		Foreign      float64
		Base         float64
	}
	var rows []row
	err := d.db.WithContext(ctx).
		Table("journal_lines AS l").
		Joins("JOIN journal_entrys AS m ON m.id = l.entry_id").
		Joins("JOIN accounts AS a ON a.id = l.account_id").
		Where("m.organization_id = ?", organizationID).
		Where("m.state = ?", EntryStatePosted).
		Where("l.reconciled = ?", false).
		Where("l.currency_code IS NOT NULL AND l.currency_code <> ?", baseCurrency).
		Where("m.deleted_at IS NULL AND l.deleted_at IS NULL AND a.deleted_at IS NULL").
		Where("a.type IN (?,?,?,?,?,?)", AccountTypeCash, AccountTypeBank, AccountTypeReceivable, AccountTypePayable, AccountTypeLiability, AccountTypeTax).
		Select(`
			l.account_id,
			a.type AS account_type,
			l.currency_code,
			COALESCE(SUM(l.amount_currency),0) AS foreign,
			COALESCE(SUM(l.debit - l.credit),0) AS base
		`).
		Group("l.account_id, a.type, l.currency_code").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	positions := make([]FxPosition, 0, len(rows))
	for _, r := range rows {
		positions = append(positions, FxPosition{
			AccountID:      r.AccountID,
			AccountType:    r.AccountType,
			CurrencyCode:   r.CurrencyCode,
			ForeignBalance: amount.FromFloat64(r.Foreign),
			BaseBalance:    amount.FromFloat64(r.Base),
		})
	}
	return positions, nil
}

type FxRevaluationService struct {
	positions   FxRevaluationDAO
	poster      PostingService
	rates       CurrencyRateResolver
	fxAccounts  FxAccountResolver
	journaler   ClosingJournalResolver
	movements   JournalEntryDAO
	orgCurrency func(ctx context.Context, organizationID uint64) (string, error)
}

func NewFxRevaluationService(
	positions FxRevaluationDAO,
	poster PostingService,
	rates CurrencyRateResolver,
	fxAccounts FxAccountResolver,
	journaler ClosingJournalResolver,
	orgCurrency func(ctx context.Context, organizationID uint64) (string, error),
) FxRevaluationService {
	return FxRevaluationService{positions: positions, poster: poster, rates: rates, fxAccounts: fxAccounts, journaler: journaler, orgCurrency: orgCurrency}
}

func (s FxRevaluationService) WithMoves(movements JournalEntryDAO) FxRevaluationService {
	s.movements = movements
	return s
}

func fxRevaluationRef(date time.Time, currencyCode string) string {
	return fmt.Sprintf("FXREVAL/%s/%s", date.Format("20060102"), currencyCode)
}

func (s FxRevaluationService) Revalue(ctx context.Context, organizationID uint64, date time.Time) (int, error) {
	baseCurrency, err := s.orgCurrency(ctx, organizationID)
	if err != nil {
		return 0, err
	}
	journalID, err := s.journaler.ClosingJournalID(ctx, organizationID)
	if err != nil {
		return 0, err
	}
	positions, err := s.positions.OpenForeignPositions(ctx, organizationID, baseCurrency)
	if err != nil {
		return 0, err
	}
	posted := 0
	for _, p := range positions {
		ref := fxRevaluationRef(date, p.CurrencyCode)
		if s.movements != nil {
			existing, err := s.movements.Search(ctx, "ref", ref)
			if err != nil {
				return posted, err
			}
			if existing != nil {
				continue
			}
		}
		rate, err := s.rates.Rate(ctx, p.CurrencyCode, organizationID, amount.RateClosing, date)
		if err != nil {
			return posted, ErrRateNotFound
		}
		closing := p.ForeignBalance.Mul(rate).Round(postingPrecision)
		delta := closing.Sub(p.BaseBalance).Round(postingPrecision)
		if delta.IsZero() {
			continue
		}
		var debitAccount, creditAccount uint64
		absDelta := delta.Abs()
		isCreditNormal := p.AccountType == AccountTypePayable || p.AccountType == AccountTypeLiability || p.AccountType == AccountTypeTax
		if isCreditNormal {
			if delta.IsPositive() {
				debitAccount, err = s.fxAccounts.FxLossAccountID(ctx, organizationID)
				if err != nil {
					return posted, err
				}
				creditAccount = p.AccountID
			} else {
				debitAccount = p.AccountID
				creditAccount, err = s.fxAccounts.FxGainAccountID(ctx, organizationID)
				if err != nil {
					return posted, err
				}
			}
		} else {
			if delta.IsPositive() {
				debitAccount = p.AccountID
				creditAccount, err = s.fxAccounts.FxGainAccountID(ctx, organizationID)
				if err != nil {
					return posted, err
				}
			} else {
				debitAccount, err = s.fxAccounts.FxLossAccountID(ctx, organizationID)
				if err != nil {
					return posted, err
				}
				creditAccount = p.AccountID
			}
		}
		_, err = s.poster.Post(ctx, PostRequest{
			OrganizationID: organizationID,
			JournalID:      journalID,
			Date:           date,
			Ref:            ref,
			OriginType:     OriginTypeFxRevaluation,
			Description:    fmt.Sprintf("FX revaluation %s %s", p.CurrencyCode, date.Format("2006-01-02")),
			Lines: []PostingLine{
				{AccountID: debitAccount, Name: "FX Revaluation", Debit: absDelta},
				{AccountID: creditAccount, Name: "FX Gain/Loss (unrealized)", Credit: absDelta},
			},
		})
		if err != nil {
			return posted, err
		}
		posted++
	}
	return posted, nil
}

func (s FxRevaluationService) ReverseRevaluation(ctx context.Context, organizationID uint64, date time.Time) (int, error) {
	if s.movements == nil {
		return 0, nil
	}
	page, err := s.movements.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "organization_id", Operator: query.Equal, Value: organizationID},
		{Field: "origin_type", Operator: query.Equal, Value: OriginTypeFxRevaluation},
		{Field: "date", Operator: query.Equal, Value: date},
	}})
	if err != nil {
		return 0, err
	}
	reversed := 0
	for _, entry := range page.Items {
		if entry.State != EntryStatePosted {
			continue
		}
		existing, err := s.movements.FindByReversed(ctx, entry.ID)
		if err != nil {
			return reversed, err
		}
		if existing != nil {
			continue
		}
		ref := ""
		if entry.Ref != nil {
			ref = *entry.Ref + "/REV"
		}
		_, err = s.poster.Reverse(ctx, ReverseRequest{
			OrganizationID: organizationID,
			JournalID:      entry.JournalID,
			Date:           date,
			Ref:            ref,
			Description:    "FX revaluation reversal",
			EntryID:        entry.ID,
		})
		if err != nil {
			return reversed, err
		}
		reversed++
	}
	return reversed, nil
}
