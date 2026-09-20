package accounting

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type BudgetLineRequest struct {
	AccountID     uint64  `json:"account_id"`
	DimensionID   *uint64 `json:"dimension_id"`
	PlannedAmount float64 `json:"planned_amount"`
}

type CreateBudgetRequest struct {
	OrganizationID uint64
	Name           string
	DateStart      time.Time
	DateEnd        time.Time
	Lines          []BudgetLineRequest
}

type BudgetVarianceLine struct {
	AccountID       uint64  `json:"account_id"`
	DimensionID     *uint64 `json:"dimension_id"`
	PlannedAmount   float64 `json:"planned_amount"`
	PracticalAmount float64 `json:"practical_amount"`
	Variance        float64 `json:"variance"`
}

type BudgetService struct {
	budgets BudgetDAO
	lines   BudgetLineDAO
	query   BudgetQueryDAO
	tx      db.Transactioner
	now     func() time.Time
}

func NewBudgetService(budgets BudgetDAO, lines BudgetLineDAO, query BudgetQueryDAO, tx db.Transactioner) BudgetService {
	return BudgetService{budgets: budgets, lines: lines, query: query, tx: tx, now: time.Now}
}

func (s BudgetService) Create(ctx context.Context, request CreateBudgetRequest) (*Budget, error) {
	if len(request.Lines) == 0 {
		return nil, ErrBudgetNoLines
	}
	if request.DateEnd.Before(request.DateStart) {
		return nil, ErrBudgetInvalidDates
	}

	lines := make([]*BudgetLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &BudgetLine{
			AccountID:     line.AccountID,
			DimensionID:   line.DimensionID,
			PlannedAmount: line.PlannedAmount,
		}
	}
	budget := &Budget{
		OrganizationID: helper.Ptr(request.OrganizationID),
		Name:           helper.Ptr(request.Name),
		DateStart:      &request.DateStart,
		DateEnd:        &request.DateEnd,
		State:          BudgetStateDraft,
	}

	var created *Budget
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		created, err = s.budgets.CreateWithLinesTx(ctx, tx, budget, lines)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s BudgetService) Find(ctx context.Context, id uint64) (*Budget, error) {
	return s.budgets.Find(ctx, id)
}

func (s BudgetService) List(ctx context.Context, q *query.Query) (*query.Page[Budget], error) {
	return s.budgets.List(ctx, q)
}

func (s BudgetService) ListLines(ctx context.Context, budgetID uint64) ([]*BudgetLine, error) {
	return s.lines.ListByBudget(ctx, budgetID)
}

func (s BudgetService) Variance(ctx context.Context, budgetID uint64) ([]BudgetVarianceLine, error) {
	budget, err := s.budgets.Find(ctx, budgetID)
	if err != nil {
		return nil, err
	}
	if budget == nil {
		return nil, ErrBudgetNotFound
	}

	lines, err := s.lines.ListByBudget(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	result := make([]BudgetVarianceLine, len(lines))
	for i, line := range lines {
		practical, err := s.query.SumPracticalByPeriod(ctx, *budget.OrganizationID, line.AccountID, *budget.DateStart, *budget.DateEnd)
		if err != nil {
			return nil, err
		}
		planned := amount.FromFloat64(line.PlannedAmount)
		actual := amount.FromFloat64(practical)
		result[i] = BudgetVarianceLine{
			AccountID:       line.AccountID,
			DimensionID:     line.DimensionID,
			PlannedAmount:   line.PlannedAmount,
			PracticalAmount: practical,
			Variance:        actual.Sub(planned).Float64(),
		}
	}
	return result, nil
}

type TaxRuleResolver struct {
	positions   TaxRuleDAO
	taxMaps     TaxRuleTaxMapDAO
	accountMaps TaxRuleAccountMapDAO
	tx          db.Transactioner
}

func NewTaxRuleResolver(positions TaxRuleDAO, taxMaps TaxRuleTaxMapDAO, accountMaps TaxRuleAccountMapDAO, tx db.Transactioner) TaxRuleResolver {
	return TaxRuleResolver{positions: positions, taxMaps: taxMaps, accountMaps: accountMaps, tx: tx}
}

type CreateTaxRuleRequest struct {
	OrganizationID uint64
	Name           string
	CountryCode    string
	AutoApply      bool
	TaxMaps        []*TaxRuleTaxMap
	AccountMaps    []*TaxRuleAccountMap
}

func (r TaxRuleResolver) Find(ctx context.Context, id uint64) (*TaxRule, error) {
	return r.positions.Find(ctx, id)
}

func (r TaxRuleResolver) List(ctx context.Context, q *query.Query) (*query.Page[TaxRule], error) {
	return r.positions.List(ctx, q)
}

func (r TaxRuleResolver) ListTaxMaps(ctx context.Context, positionID uint64) ([]*TaxRuleTaxMap, error) {
	return r.taxMaps.ListByPosition(ctx, positionID)
}

func (r TaxRuleResolver) ListAccountMaps(ctx context.Context, positionID uint64) ([]*TaxRuleAccountMap, error) {
	return r.accountMaps.ListByPosition(ctx, positionID)
}

func (r TaxRuleResolver) Create(ctx context.Context, request CreateTaxRuleRequest) (*TaxRule, error) {
	position := &TaxRule{
		OrganizationID: helper.Ptr(request.OrganizationID),
		Name:           helper.Ptr(request.Name),
		CountryCode:    helper.Ptr(request.CountryCode),
		AutoApply:      request.AutoApply,
		Active:         true,
	}
	var created *TaxRule
	err := r.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		created, err = r.positions.CreateWithMapsTx(ctx, tx, position, request.TaxMaps, request.AccountMaps)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

type ResolveResult struct {
	TaxAccount *uint64
	Account    *uint64
}

func (r TaxRuleResolver) Resolve(ctx context.Context, positionID uint64, srcTaxID *uint64, srcAccountID uint64) (*ResolveResult, error) {
	position, err := r.positions.Find(ctx, positionID)
	if err != nil {
		return nil, err
	}
	if position == nil {
		return nil, ErrTaxRuleNotFound
	}

	result := &ResolveResult{Account: &srcAccountID}
	if srcTaxID != nil {
		maps, err := r.taxMaps.ListByPosition(ctx, positionID)
		if err != nil {
			return nil, err
		}
		for _, taxMap := range maps {
			if taxMap.SrcTaxID == *srcTaxID {
				result.TaxAccount = taxMap.DestTaxID
				break
			}
		}
	}

	accountMaps, err := r.accountMaps.ListByPosition(ctx, positionID)
	if err != nil {
		return nil, err
	}
	for _, accountMap := range accountMaps {
		if accountMap.SrcAccountID == srcAccountID {
			result.Account = &accountMap.DestAccountID
			break
		}
	}
	return result, nil
}

type WithholdingService struct {
	withholdings WithholdingTaxDAO
	poster       Poster
	tx           db.Transactioner
}

func NewWithholdingService(withholdings WithholdingTaxDAO, poster Poster, tx db.Transactioner) WithholdingService {
	return WithholdingService{withholdings: withholdings, poster: poster, tx: tx}
}

type WithholdRequest struct {
	OrganizationID uint64
	JournalID      uint64
	ContactID      uint64
	Amount         float64
	Date           time.Time
	Ref            string
	Scope          string
	WHTID          uint64
	BankAccountID  uint64
	PayableID      uint64
	Description    string
}

func (s WithholdingService) Find(ctx context.Context, id uint64) (*WithholdingTax, error) {
	return s.withholdings.Find(ctx, id)
}

func (s WithholdingService) List(ctx context.Context, q *query.Query) (*query.Page[WithholdingTax], error) {
	return s.withholdings.List(ctx, q)
}

func (s WithholdingService) Create(ctx context.Context, tax *WithholdingTax) (*WithholdingTax, error) {
	tax.Active = true
	return s.withholdings.Create(ctx, tax)
}

func (s WithholdingService) Withhold(ctx context.Context, request WithholdRequest) (*JournalEntry, error) {
	wht, err := s.withholdings.Find(ctx, request.WHTID)
	if err != nil {
		return nil, err
	}
	if wht == nil {
		return nil, ErrWithholdingNotFound
	}
	if wht.AccountID == nil {
		return nil, ErrWithholdingNoAccount
	}
	if wht.Scope != request.Scope {
		return nil, ErrWithholdingScope
	}

	total := amount.FromFloat64(request.Amount)
	rate, err := amount.FromFloat64(wht.RatePct).Div(amount.FromInt64(100))
	if err != nil {
		return nil, err
	}
	withheld := total.Mul(rate).Round(2)
	net := total.Sub(withheld)

	lines := []PostingLine{
		{AccountID: request.PayableID, Debit: total, Name: "Payable"},
		{AccountID: *wht.AccountID, Credit: withheld, Name: "Withholding Tax"},
		{AccountID: request.BankAccountID, Credit: net, Name: "Bank"},
	}
	post := PostRequest{
		OrganizationID: request.OrganizationID,
		JournalID:      request.JournalID,
		Date:           request.Date,
		Ref:            request.Ref,
		OriginType:     OriginTypeWithholding,
		OriginID:       request.WHTID,
		Description:    request.Description,
		Lines:          lines,
	}
	return s.poster.Post(ctx, post)
}

type TaxReturnService struct {
	returns      TaxReturnDAO
	periods      TaxPeriodDAO
	invoiceTaxes InvoiceTaxDAO
	tx           db.Transactioner
	poster       Poster
	journals     dao.CRUD[reference.Journal]
	accounts     AccountLookup
}

func NewTaxReturnService(returns TaxReturnDAO, periods TaxPeriodDAO, invoiceTaxes InvoiceTaxDAO, tx db.Transactioner) TaxReturnService {
	return TaxReturnService{returns: returns, periods: periods, invoiceTaxes: invoiceTaxes, tx: tx}
}

func (s TaxReturnService) WithSettlement(poster Poster, journals dao.CRUD[reference.Journal], accounts AccountLookup) TaxReturnService {
	s.poster = poster
	s.journals = journals
	s.accounts = accounts
	return s
}

type CreateTaxReturnRequest struct {
	OrganizationID uint64
	PeriodID       uint64
	Type           string
}

func (s TaxReturnService) Create(ctx context.Context, request CreateTaxReturnRequest) (*TaxReturn, error) {
	period, err := s.periods.Find(ctx, request.PeriodID)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, ErrPeriodNotFound
	}
	if period.OrganizationID != request.OrganizationID {
		return nil, ErrPeriodNotFound
	}

	existing, err := s.returns.FindByPeriod(ctx, request.PeriodID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrTaxReturnExists
	}

	outputTax, err := s.invoiceTaxes.SumTaxByPeriod(ctx, request.OrganizationID, InvoiceTypeCustomerInvoice, *period.DateStart, *period.DateEnd)
	if err != nil {
		return nil, err
	}
	inputTax, err := s.invoiceTaxes.SumTaxByPeriod(ctx, request.OrganizationID, InvoiceTypeSupplierBill, *period.DateStart, *period.DateEnd)
	if err != nil {
		return nil, err
	}

	taxReturn := &TaxReturn{
		OrganizationID: helper.Ptr(request.OrganizationID),
		PeriodID:       request.PeriodID,
		Type:           request.Type,
		OutputTax:      outputTax,
		InputTax:       inputTax,
		NetPayable:     amount.FromFloat64(outputTax).Sub(amount.FromFloat64(inputTax)).Float64(),
		State:          TaxReturnStateDraft,
	}
	return s.returns.Create(ctx, taxReturn)
}

func (s TaxReturnService) Find(ctx context.Context, id uint64) (*TaxReturn, error) {
	return s.returns.Find(ctx, id)
}

func (s TaxReturnService) List(ctx context.Context, q *query.Query) (*query.Page[TaxReturn], error) {
	return s.returns.List(ctx, q)
}

func (s TaxReturnService) File(ctx context.Context, taxReturnID uint64) (*TaxReturn, error) {
	taxReturn, err := s.returns.Find(ctx, taxReturnID)
	if err != nil {
		return nil, err
	}
	if taxReturn == nil {
		return nil, ErrTaxReturnNotFound
	}
	if taxReturn.State != TaxReturnStateDraft {
		return nil, ErrTaxReturnNotDraft
	}

	now := time.Now().UTC()
	taxReturn.State = TaxReturnStateFiled
	taxReturn.FiledAt = &now
	return s.returns.Update(ctx, taxReturn)
}

func (s TaxReturnService) ExportFiling(ctx context.Context, taxReturnID, organizationID uint64) (string, error) {
	taxReturn, err := s.returns.Find(ctx, taxReturnID)
	if err != nil {
		return "", err
	}
	if taxReturn == nil || taxReturn.OrganizationID == nil || *taxReturn.OrganizationID != organizationID {
		return "", ErrTaxReturnNotFound
	}
	if taxReturn.State != TaxReturnStateFiled && taxReturn.State != TaxReturnStatePaid {
		return "", ErrTaxReturnNotFiled
	}

	period, err := s.periods.Find(ctx, taxReturn.PeriodID)
	if err != nil {
		return "", err
	}
	periodName := ""
	if period != nil && period.Name != "" {
		periodName = period.Name
	}
	filedAt := ""
	if taxReturn.FiledAt != nil {
		filedAt = taxReturn.FiledAt.UTC().Format("2006-01-02")
	}
	rows := [][]string{
		{"tax_return_id", "organization_id", "period_id", "period", "type", "state", "filed_at", "output_tax", "input_tax", "net_payable"},
		{
			strconv.FormatUint(taxReturn.ID, 10),
			strconv.FormatUint(organizationID, 10),
			strconv.FormatUint(taxReturn.PeriodID, 10),
			periodName,
			taxReturn.Type,
			taxReturn.State,
			filedAt,
			strconv.FormatFloat(taxReturn.OutputTax, 'f', 4, 64),
			strconv.FormatFloat(taxReturn.InputTax, 'f', 4, 64),
			strconv.FormatFloat(taxReturn.NetPayable, 'f', 4, 64),
		},
	}
	var content strings.Builder
	writer := csv.NewWriter(&content)
	if err := writer.WriteAll(rows); err != nil {
		return "", err
	}
	writer.Flush()
	return content.String(), writer.Error()
}

func (s TaxReturnService) Pay(ctx context.Context, taxReturnID uint64) (*TaxReturn, error) {
	taxReturn, err := s.returns.Find(ctx, taxReturnID)
	if err != nil {
		return nil, err
	}
	if taxReturn == nil {
		return nil, ErrTaxReturnNotFound
	}
	if taxReturn.State != TaxReturnStateFiled {
		return nil, ErrTaxReturnNotFiled
	}

	if s.poster != nil && taxReturn.OrganizationID != nil && taxReturn.NetPayable != 0 {
		if err := s.postTaxPayment(ctx, taxReturn); err != nil {
			return nil, err
		}
	}

	taxReturn.State = TaxReturnStatePaid
	return s.returns.Update(ctx, taxReturn)
}

func (s TaxReturnService) postTaxPayment(ctx context.Context, taxReturn *TaxReturn) error {
	organizationID := *taxReturn.OrganizationID
	payable, err := s.taxPayableAccount(ctx, organizationID)
	if err != nil {
		return err
	}
	journal, bankAccount, err := s.taxBankJournal(ctx, organizationID)
	if err != nil {
		return err
	}
	net := amount.FromFloat64(taxReturn.NetPayable).Round(4).Abs()
	lines := []PostingLine{
		{AccountID: payable, Name: "Tax Payable", Debit: net},
		{AccountID: bankAccount, Name: "Bank", Credit: net},
	}
	if taxReturn.NetPayable < 0 {
		lines = []PostingLine{
			{AccountID: bankAccount, Name: "Bank", Debit: net},
			{AccountID: payable, Name: "Tax Refund", Credit: net},
		}
	}
	return s.tx.Run(ctx, func(tx *gorm.DB) error {
		_, err := s.poster.PostTx(ctx, tx, PostRequest{
			OrganizationID: organizationID,
			JournalID:      journal.ID,
			Date:           time.Now().UTC(),
			Ref:            fmt.Sprintf("TAXPAY/%d", taxReturn.ID),
			OriginType:     OriginTypeTaxPayment,
			OriginID:       taxReturn.ID,
			Description:    "Tax return payment",
			Lines:          lines,
		})
		return err
	})
}

func (s TaxReturnService) taxPayableAccount(ctx context.Context, organizationID uint64) (uint64, error) {
	if s.accounts == nil {
		return 0, ErrNoPayableAccount
	}
	page, err := s.accounts.List(ctx, &query.Query{Filters: []query.Filter{{Field: "type", Operator: query.Equal, Value: AccountTypeTax}}})
	if err != nil {
		return 0, err
	}
	for _, account := range page.Items {
		if account.OrganizationID == organizationID && account.Active {
			return account.ID, nil
		}
	}
	return 0, ErrNoPayableAccount
}

func (s TaxReturnService) taxBankJournal(ctx context.Context, organizationID uint64) (*reference.Journal, uint64, error) {
	if s.journals == nil {
		return nil, 0, ErrNoBankAccount
	}
	page, err := s.journals.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "organization_id", Operator: query.Equal, Value: organizationID},
		{Field: "type", Operator: query.In, Value: []string{"bank", "cash"}},
	}})
	if err != nil {
		return nil, 0, err
	}
	for _, journal := range page.Items {
		bankAccount := journal.DefaultAccountID
		if bankAccount == nil {
			bankAccount = journal.BankAccountID
		}
		if bankAccount == nil {
			continue
		}
		return journal, *bankAccount, nil
	}
	return nil, 0, ErrNoBankAccount
}

func (s TaxReturnService) Draft(ctx context.Context, taxReturnID uint64) (*TaxReturn, error) {
	taxReturn, err := s.returns.Find(ctx, taxReturnID)
	if err != nil {
		return nil, err
	}
	if taxReturn == nil {
		return nil, ErrTaxReturnNotFound
	}
	if taxReturn.State == TaxReturnStatePaid {
		return nil, ErrTaxReturnPaid
	}

	taxReturn.State = TaxReturnStateDraft
	return s.returns.Update(ctx, taxReturn)
}
