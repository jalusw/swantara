package reference

import (
	"context"
	"math"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type FxRateService struct {
	rates      dao.Base[FxRate]
	currencies dao.Base[Currency]
}

func NewFxRateService(rates dao.Base[FxRate], currencies dao.Base[Currency]) FxRateService {
	return FxRateService{rates: rates, currencies: currencies}
}

func (s FxRateService) List(ctx context.Context, q *query.Query) (*query.Page[FxRate], error) {
	return s.rates.List(ctx, q)
}

func (s FxRateService) Find(ctx context.Context, id uint64) (*FxRate, error) {
	return s.rates.Find(ctx, id)
}

func (s FxRateService) Delete(ctx context.Context, id uint64) error {
	return s.rates.Delete(ctx, id)
}

func (s FxRateService) ListCurrencies(ctx context.Context, q *query.Query) (*query.Page[Currency], error) {
	return s.currencies.List(ctx, q)
}

func (s FxRateService) SearchCurrency(ctx context.Context, field string, value any) (*Currency, error) {
	return s.currencies.Search(ctx, field, value)
}

func (s FxRateService) Create(ctx context.Context, rate *FxRate) (*FxRate, error) {
	if rate.Rate <= 0 {
		return nil, amount.ErrInvalidRate
	}
	if rate.RateType == "" {
		rate.RateType = string(amount.RateSpot)
	}
	if !validRateType(rate.RateType) {
		return nil, amount.ErrInvalidRate
	}
	if rate.ValidFrom.IsZero() {
		return nil, ErrRateValidFromMissing
	}

	if err := s.validateCurrency(ctx, rate.CurrencyCode); err != nil {
		return nil, err
	}

	return s.rates.Create(ctx, rate)
}

func (s FxRateService) Update(ctx context.Context, rate *FxRate) (*FxRate, error) {
	if rate.Rate <= 0 {
		return nil, amount.ErrInvalidRate
	}
	if rate.RateType == "" {
		rate.RateType = string(amount.RateSpot)
	}
	if !validRateType(rate.RateType) {
		return nil, amount.ErrInvalidRate
	}
	if rate.ValidFrom.IsZero() {
		return nil, ErrRateValidFromMissing
	}

	if err := s.validateCurrency(ctx, rate.CurrencyCode); err != nil {
		return nil, err
	}

	return s.rates.Update(ctx, rate)
}

func (s FxRateService) validateCurrency(ctx context.Context, code string) error {
	currency, err := s.currencies.Search(ctx, "code", code)
	if err != nil {
		return err
	}
	if currency == nil {
		return ErrCurrencyNotFound
	}
	return nil
}

func validRateType(value string) bool {
	switch amount.RateType(value) {
	case amount.RateSpot, amount.RateAverage, amount.RateClosing:
		return true
	default:
		return false
	}
}

type UnitService struct {
	categories dao.Base[UnitGroup]
	units      dao.Base[Unit]
}

func NewUnitService(categories dao.Base[UnitGroup], units dao.Base[Unit]) UnitService {
	return UnitService{categories: categories, units: units}
}

func (s UnitService) ListUnits(ctx context.Context, q *query.Query) (*query.Page[Unit], error) {
	return s.units.List(ctx, q)
}

func (s UnitService) FindUnit(ctx context.Context, id uint64) (*Unit, error) {
	return s.units.Find(ctx, id)
}

func (s UnitService) DeleteUnit(ctx context.Context, id uint64) error {
	return s.units.Delete(ctx, id)
}

func (s UnitService) ListCategories(ctx context.Context, q *query.Query) (*query.Page[UnitGroup], error) {
	return s.categories.List(ctx, q)
}

func (s UnitService) FindCategory(ctx context.Context, id uint64) (*UnitGroup, error) {
	return s.categories.Find(ctx, id)
}

func (s UnitService) CreateCategory(ctx context.Context, category *UnitGroup) (*UnitGroup, error) {
	return s.categories.Create(ctx, category)
}

func (s UnitService) UpdateCategory(ctx context.Context, category *UnitGroup) (*UnitGroup, error) {
	return s.categories.Update(ctx, category)
}

func (s UnitService) DeleteCategory(ctx context.Context, id uint64) error {
	return s.categories.Delete(ctx, id)
}

func (s UnitService) CreateUnit(ctx context.Context, unit *Unit) (*Unit, error) {
	if err := s.validateUnit(ctx, unit, 0); err != nil {
		return nil, err
	}
	return s.units.Create(ctx, unit)
}

func (s UnitService) UpdateUnit(ctx context.Context, unit *Unit) (*Unit, error) {
	if err := s.validateUnit(ctx, unit, unit.ID); err != nil {
		return nil, err
	}
	return s.units.Update(ctx, unit)
}

func (s UnitService) validateUnit(ctx context.Context, unit *Unit, excludeID uint64) error {
	if unit.Factor <= 0 {
		return ErrInvalidFactor
	}

	category, err := s.categories.Find(ctx, unit.CategoryID)
	if err != nil {
		return err
	}
	if category == nil {
		return ErrUnitGroupNotFound
	}

	existing, err := s.units.Search(ctx, "name", unit.Name)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != excludeID && existing.CategoryID == unit.CategoryID {
		return ErrUnitNameTaken
	}
	return nil
}

func (s UnitService) Convert(ctx context.Context, qty amount.Amount, fromID, toID uint64) (amount.Amount, error) {
	from, err := s.units.Find(ctx, fromID)
	if err != nil {
		return amount.Amount{}, err
	}
	if from == nil {
		return amount.Amount{}, ErrUnitNotFound
	}

	to, err := s.units.Find(ctx, toID)
	if err != nil {
		return amount.Amount{}, err
	}
	if to == nil {
		return amount.Amount{}, ErrUnitNotFound
	}

	if from.CategoryID != to.CategoryID {
		return amount.Amount{}, ErrUnitGroupMismatch
	}
	if from.Factor <= 0 || to.Factor <= 0 {
		return amount.Amount{}, ErrInvalidFactor
	}

	base := qty.Mul(amount.FromFloat64(from.Factor))
	converted, err := base.Div(amount.FromFloat64(to.Factor))
	if err != nil {
		return amount.Amount{}, err
	}
	return converted.Round(helper.DecimalPlaces(to.Rounding)), nil
}

type PaymentTermService struct {
	terms PaymentTermDAO
}

func NewPaymentTermService(terms PaymentTermDAO) PaymentTermService {
	return PaymentTermService{terms: terms}
}

func (s PaymentTermService) List(ctx context.Context, q *query.Query) (*query.Page[PaymentTerm], error) {
	return s.terms.List(ctx, q)
}

func (s PaymentTermService) Find(ctx context.Context, id uint64) (*PaymentTerm, error) {
	return s.terms.Find(ctx, id)
}

func (s PaymentTermService) ListLines(ctx context.Context, termID uint64) ([]*PaymentTermLine, error) {
	return s.terms.ListLines(ctx, termID)
}

func (s PaymentTermService) Create(ctx context.Context, term *PaymentTerm, lines []*PaymentTermLine) (*PaymentTerm, error) {
	if err := validateTermLines(lines); err != nil {
		return nil, err
	}
	if term.OrganizationID == 0 {
		return nil, ErrTermOrganizationRequired
	}
	if err := s.validateUniqueName(ctx, 0, term.OrganizationID, term.Name); err != nil {
		return nil, err
	}

	created, err := s.terms.Create(ctx, term)
	if err != nil {
		return nil, err
	}
	if err := s.terms.ReplaceLines(ctx, created.ID, lines); err != nil {
		return nil, err
	}
	return created, nil
}

func (s PaymentTermService) Update(ctx context.Context, termID uint64, term *PaymentTerm, lines []*PaymentTermLine) (*PaymentTerm, error) {
	if err := validateTermLines(lines); err != nil {
		return nil, err
	}

	existing, err := s.terms.Find(ctx, termID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrTermNotFound
	}
	if term.OrganizationID != 0 && existing.OrganizationID != term.OrganizationID {
		return nil, ErrTermOrganizationMismatch
	}
	if err := s.validateUniqueName(ctx, termID, existing.OrganizationID, term.Name); err != nil {
		return nil, err
	}

	existing.Name = term.Name
	existing.Note = term.Note
	existing.Code = term.Code
	existing.IsActive = term.IsActive
	existing.TemplateKey = term.TemplateKey

	updated, err := s.terms.Update(ctx, existing)
	if err != nil {
		return nil, err
	}
	if err := s.terms.ReplaceLines(ctx, termID, lines); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s PaymentTermService) validateUniqueName(ctx context.Context, excludeID uint64, organizationID uint64, name string) error {
	page, err := s.terms.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "organization_id", Operator: query.Equal, Value: organizationID},
		{Field: "name", Operator: query.Equal, Value: name},
	}})
	if err != nil {
		return err
	}
	for _, item := range page.Items {
		if item.ID != excludeID {
			return ErrDuplicateTermName
		}
	}
	return nil
}

func (s PaymentTermService) ValidateTermInOrganization(ctx context.Context, termID uint64, organizationID uint64) error {
	term, err := s.terms.Find(ctx, termID)
	if err != nil {
		return err
	}
	if term == nil || term.OrganizationID != organizationID {
		return ErrPaymentTermNotInOrg
	}
	return nil
}

func SeedDefaultPaymentTerms(ctx context.Context, terms PaymentTermDAO, organizationID uint64) error {
	defs := []struct {
		Name        string
		Note        string
		Code        string
		TemplateKey string
		Lines       []*PaymentTermLine
	}{
		{
			Name: "Immediate Payment", Note: "Due immediately on receipt of invoice", Code: "immediate", TemplateKey: "immediate",
			Lines: []*PaymentTermLine{{Sequence: 10, ValueType: "balance", Value: 100, DaysAfter: 0}},
		},
		{
			Name: "Net 15", Note: "Due within 15 days", Code: "net_15", TemplateKey: "net_15",
			Lines: []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100, DaysAfter: 15}},
		},
		{
			Name: "Net 30", Note: "Due within 30 days", Code: "net_30", TemplateKey: "net_30",
			Lines: []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100, DaysAfter: 30}},
		},
		{
			Name: "Net 45", Note: "Due within 45 days", Code: "net_45", TemplateKey: "net_45",
			Lines: []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100, DaysAfter: 45}},
		},
		{
			Name: "Net 60", Note: "Due within 60 days", Code: "net_60", TemplateKey: "net_60",
			Lines: []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100, DaysAfter: 60}},
		},
		{
			Name: "2/10 Net 30", Note: "2% discount if paid within 10 days, net 30", Code: "2_10_net30", TemplateKey: "2_10_net30",
			Lines: []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100, DaysAfter: 30, DiscountPct: func() *float64 { v := 2.0; return &v }(), DiscountDays: func() *int { v := 10; return &v }()}},
		},
		{
			Name: "30% Advance, Balance Net 30", Note: "30% upfront, balance in 30 days", Code: "30_adv_balance", TemplateKey: "30_adv_balance",
			Lines: []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 30, DaysAfter: 0}, {Sequence: 20, ValueType: "balance", Value: 0, DaysAfter: 30}},
		},
		{
			Name: "50/50", Note: "Half upfront, half on delivery", Code: "50_50", TemplateKey: "50_50",
			Lines: []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 50, DaysAfter: 0}, {Sequence: 20, ValueType: "percent", Value: 50, DaysAfter: 30}},
		},
	}
	svc := PaymentTermService{terms: terms}
	for _, def := range defs {
		page, err := terms.List(ctx, &query.Query{Filters: []query.Filter{
			{Field: "organization_id", Operator: query.Equal, Value: organizationID},
			{Field: "name", Operator: query.Equal, Value: def.Name},
		}})
		if err != nil {
			return err
		}
		if len(page.Items) > 0 {
			continue
		}
		term := &PaymentTerm{
			OrganizationID: organizationID,
			Name:           def.Name,
			Note:           &def.Note,
			Code:           &def.Code,
			TemplateKey:    &def.TemplateKey,
			IsActive:       true,
		}
		if _, err := svc.Create(ctx, term, def.Lines); err != nil {
			return err
		}
	}
	return nil
}

func (s PaymentTermService) Delete(ctx context.Context, termID uint64) error {
	return s.terms.DeleteWithLines(ctx, termID)
}

func (s PaymentTermService) Splits(ctx context.Context, termID uint64, total amount.Amount, date time.Time) ([]PaymentSplit, error) {
	term, err := s.terms.Find(ctx, termID)
	if err != nil {
		return nil, err
	}
	if term == nil {
		return nil, ErrTermNotFound
	}

	lines, err := s.terms.ListLines(ctx, termID)
	if err != nil {
		return nil, err
	}
	if err := validateTermLines(lines); err != nil {
		return nil, err
	}

	splits := make([]PaymentSplit, 0, len(lines))
	var assigned amount.Amount
	for _, line := range lines {
		split := PaymentSplit{
			Sequence:     line.Sequence,
			ValueType:    line.ValueType,
			DaysAfter:    line.DaysAfter,
			DayOfMonth:   line.DayOfMonth,
			DiscountPct:  line.DiscountPct,
			DiscountDays: line.DiscountDays,
			DueDate:      termLineDueDate(line, date),
		}
		switch line.ValueType {
		case "percent":
			split.Amount = total.Mul(amount.FromFloat64(line.Value / 100)).Round(2)
		case "fixed":
			split.Amount = amount.FromFloat64(line.Value).Round(2)
		case "balance":
			split.Amount = total.Sub(assigned).Round(2)
		default:
			return nil, ErrInvalidTermLines
		}
		assigned = assigned.Add(split.Amount)
		splits = append(splits, split)
	}
	return splits, nil
}

type PaymentSplit struct {
	Sequence     int           `json:"sequence"`
	ValueType    string        `json:"value_type"`
	Amount       amount.Amount `json:"amount"`
	DueDate      time.Time     `json:"due_date"`
	DaysAfter    int           `json:"days_after"`
	DayOfMonth   *int          `json:"day_of_month,omitempty"`
	DiscountPct  *float64      `json:"discount_pct,omitempty"`
	DiscountDays *int          `json:"discount_days,omitempty"`
}

func termLineDueDate(line *PaymentTermLine, date time.Time) time.Time {
	if line.DayOfMonth == nil {
		return date.AddDate(0, 0, line.DaysAfter)
	}
	due := time.Date(date.Year(), date.Month(), *line.DayOfMonth, 0, 0, 0, 0, time.UTC)
	if due.Before(date) {
		due = time.Date(date.Year(), date.Month()+1, *line.DayOfMonth, 0, 0, 0, 0, time.UTC)
	}
	return due
}

func validateTermLines(lines []*PaymentTermLine) error {
	if len(lines) == 0 {
		return ErrInvalidTermLines
	}

	var percentSum float64
	balanceCount := 0
	for _, line := range lines {
		switch line.ValueType {
		case "percent", "fixed", "balance":
		default:
			return ErrInvalidTermLines
		}
		if line.ValueType == "percent" {
			percentSum += line.Value
		}
		if line.ValueType == "balance" {
			balanceCount++
		}
	}
	if balanceCount > 1 {
		return ErrMultipleBalanceLines
	}
	if balanceCount == 1 {
		if percentSum > 100 {
			return ErrInvalidTermLines
		}
		return nil
	}
	if math.Abs(percentSum-100) > 0.0001 {
		return ErrInvalidTermLines
	}
	return nil
}

type DimensionService struct {
	accounts dao.Base[Dimension]
}

func NewDimensionService(accounts dao.Base[Dimension]) DimensionService {
	return DimensionService{accounts: accounts}
}

func (s DimensionService) List(ctx context.Context, q *query.Query) (*query.Page[Dimension], error) {
	return s.accounts.List(ctx, q)
}

func (s DimensionService) Find(ctx context.Context, id uint64) (*Dimension, error) {
	return s.accounts.Find(ctx, id)
}

func (s DimensionService) Delete(ctx context.Context, id uint64) error {
	return s.accounts.Delete(ctx, id)
}

func (s DimensionService) Create(ctx context.Context, account *Dimension) (*Dimension, error) {
	if err := s.validateDimension(ctx, account, 0); err != nil {
		return nil, err
	}
	return s.accounts.Create(ctx, account)
}

func (s DimensionService) Update(ctx context.Context, account *Dimension) (*Dimension, error) {
	if err := s.validateDimension(ctx, account, account.ID); err != nil {
		return nil, err
	}
	return s.accounts.Update(ctx, account)
}

func (s DimensionService) validateDimension(ctx context.Context, account *Dimension, excludeID uint64) error {
	if account.Code != nil && *account.Code != "" {
		duplicate, err := s.accounts.Search(ctx, "code", *account.Code)
		if err != nil {
			return err
		}
		if duplicate != nil && duplicate.ID != excludeID && sameOrganization(duplicate.OrganizationID, account.OrganizationID) {
			return ErrDuplicateCode
		}
	}

	if account.ParentID != nil {
		parent, err := s.accounts.Find(ctx, *account.ParentID)
		if err != nil {
			return err
		}
		if parent == nil {
			return ErrParentNotFound
		}
	}

	return nil
}

func sameOrganization(a, b *uint64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

var validAccountTypes = map[string]struct{}{
	"asset": {}, "liability": {}, "equity": {}, "income": {}, "expense": {},
	"receivable": {}, "payable": {}, "bank": {}, "cash": {}, "cogs": {}, "tax": {},
	"current_asset": {}, "fixed_asset": {}, "depreciation": {},
}

type AccountService struct {
	accounts dao.Base[Account]
}

func NewAccountService(accounts dao.Base[Account]) AccountService {
	return AccountService{accounts: accounts}
}

func (s AccountService) List(ctx context.Context, q *query.Query) (*query.Page[Account], error) {
	return s.accounts.List(ctx, q)
}

func (s AccountService) Find(ctx context.Context, id uint64) (*Account, error) {
	return s.accounts.Find(ctx, id)
}

func (s AccountService) Delete(ctx context.Context, id uint64) error {
	return s.accounts.Delete(ctx, id)
}

func (s AccountService) Create(ctx context.Context, account *Account) (*Account, error) {
	if err := s.validateAccount(ctx, account, 0); err != nil {
		return nil, err
	}
	return s.accounts.Create(ctx, account)
}

func (s AccountService) Update(ctx context.Context, account *Account) (*Account, error) {
	if err := s.validateAccount(ctx, account, account.ID); err != nil {
		return nil, err
	}
	return s.accounts.Update(ctx, account)
}

func (s AccountService) validateAccount(ctx context.Context, account *Account, excludeID uint64) error {
	if _, ok := validAccountTypes[account.Type]; !ok {
		return ErrInvalidAccountType
	}

	duplicate, err := s.accounts.Search(ctx, "code", account.Code)
	if err != nil {
		return err
	}
	if duplicate != nil && duplicate.ID != excludeID && duplicate.OrganizationID == account.OrganizationID {
		return ErrDuplicateAccount
	}

	if account.ParentID != nil {
		parent, err := s.accounts.Find(ctx, *account.ParentID)
		if err != nil {
			return err
		}
		if parent == nil || parent.OrganizationID != account.OrganizationID {
			return ErrInvalidParent
		}
	}

	return nil
}

var validJournalTypes = map[string]struct{}{
	"sale": {}, "purchase": {}, "bank": {}, "cash": {}, "general": {},
}

type JournalService struct {
	journals dao.Base[Journal]
	accounts dao.Base[Account]
}

func NewJournalService(journals dao.Base[Journal], accounts dao.Base[Account]) JournalService {
	return JournalService{journals: journals, accounts: accounts}
}

func (s JournalService) List(ctx context.Context, q *query.Query) (*query.Page[Journal], error) {
	return s.journals.List(ctx, q)
}

func (s JournalService) Find(ctx context.Context, id uint64) (*Journal, error) {
	return s.journals.Find(ctx, id)
}

func (s JournalService) Delete(ctx context.Context, id uint64) error {
	return s.journals.Delete(ctx, id)
}

func (s JournalService) Create(ctx context.Context, journal *Journal) (*Journal, error) {
	if err := s.validateJournal(ctx, journal); err != nil {
		return nil, err
	}
	return s.journals.Create(ctx, journal)
}

func (s JournalService) Update(ctx context.Context, journal *Journal) (*Journal, error) {
	if err := s.validateJournal(ctx, journal); err != nil {
		return nil, err
	}
	return s.journals.Update(ctx, journal)
}

func (s JournalService) validateJournal(ctx context.Context, journal *Journal) error {
	if _, ok := validJournalTypes[journal.Type]; !ok {
		return ErrInvalidJournalType
	}

	if journal.DefaultAccountID != nil {
		account, err := s.accounts.Find(ctx, *journal.DefaultAccountID)
		if err != nil {
			return err
		}
		if account == nil || account.OrganizationID != journal.OrganizationID {
			return ErrJournalDefaultAccount
		}
	}

	return nil
}

var validTaxTypes = map[string]struct{}{
	"percent": {}, "fixed": {}, "group": {},
}

var validTaxScopes = map[string]struct{}{
	"sale": {}, "purchase": {}, "none": {},
}

type TaxService struct {
	taxes    dao.Base[Tax]
	accounts dao.Base[Account]
}

func NewTaxService(taxes dao.Base[Tax], accounts dao.Base[Account]) TaxService {
	return TaxService{taxes: taxes, accounts: accounts}
}

func (s TaxService) List(ctx context.Context, q *query.Query) (*query.Page[Tax], error) {
	return s.taxes.List(ctx, q)
}

func (s TaxService) Find(ctx context.Context, id uint64) (*Tax, error) {
	return s.taxes.Find(ctx, id)
}

func (s TaxService) Delete(ctx context.Context, id uint64) error {
	return s.taxes.Delete(ctx, id)
}

func (s TaxService) Create(ctx context.Context, tax *Tax) (*Tax, error) {
	if err := s.validateTax(ctx, tax); err != nil {
		return nil, err
	}
	return s.taxes.Create(ctx, tax)
}

func (s TaxService) Update(ctx context.Context, tax *Tax) (*Tax, error) {
	if err := s.validateTax(ctx, tax); err != nil {
		return nil, err
	}
	return s.taxes.Update(ctx, tax)
}

func (s TaxService) validateTax(ctx context.Context, tax *Tax) error {
	if _, ok := validTaxTypes[tax.Type]; !ok {
		return ErrInvalidTaxType
	}
	if _, ok := validTaxScopes[tax.Scope]; !ok {
		return ErrInvalidTaxScope
	}
	if (tax.Type == "percent" || tax.Type == "fixed") && (tax.Amount == nil || *tax.Amount <= 0) {
		return ErrTaxAmountMissing
	}

	for _, accountID := range []*uint64{tax.TaxAccountID, tax.RefundTaxAccountID} {
		if accountID == nil {
			continue
		}
		account, err := s.accounts.Find(ctx, *accountID)
		if err != nil {
			return err
		}
		if account == nil {
			return ErrInvalidTaxAccount
		}
	}

	return nil
}

type TaxYearService struct {
	years dao.Base[TaxYear]
}

func NewTaxYearService(years dao.Base[TaxYear]) TaxYearService {
	return TaxYearService{years: years}
}

func (s TaxYearService) List(ctx context.Context, q *query.Query) (*query.Page[TaxYear], error) {
	return s.years.List(ctx, q)
}

func (s TaxYearService) Find(ctx context.Context, id uint64) (*TaxYear, error) {
	return s.years.Find(ctx, id)
}

func (s TaxYearService) Delete(ctx context.Context, id uint64) error {
	return s.years.Delete(ctx, id)
}

func (s TaxYearService) Create(ctx context.Context, year *TaxYear) (*TaxYear, error) {
	if err := s.validateTaxYear(year); err != nil {
		return nil, err
	}
	return s.years.Create(ctx, year)
}

func (s TaxYearService) Update(ctx context.Context, year *TaxYear) (*TaxYear, error) {
	if err := s.validateTaxYear(year); err != nil {
		return nil, err
	}
	return s.years.Update(ctx, year)
}

func (s TaxYearService) validateTaxYear(year *TaxYear) error {
	if year.DateStart == nil || year.DateEnd == nil || year.DateEnd.Before(*year.DateStart) {
		return ErrInvalidTaxYear
	}
	return nil
}
