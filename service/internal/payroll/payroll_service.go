package payroll

import (
	"context"
	"sort"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type SalaryRuleService struct {
	rules    dao.CRUD[reference.SalaryRule]
	accounts dao.CRUD[reference.Account]
}

func NewSalaryRuleService(
	rules dao.CRUD[reference.SalaryRule],
	accounts dao.CRUD[reference.Account],
) SalaryRuleService {
	return SalaryRuleService{rules: rules, accounts: accounts}
}

func (s SalaryRuleService) List(ctx context.Context, q *query.Query) (*query.Page[reference.SalaryRule], error) {
	return s.rules.List(ctx, q)
}

func (s SalaryRuleService) Find(ctx context.Context, id uint64) (*reference.SalaryRule, error) {
	return s.rules.Find(ctx, id)
}

func (s SalaryRuleService) Create(ctx context.Context, rule *reference.SalaryRule) (*reference.SalaryRule, error) {
	if err := s.validate(ctx, rule); err != nil {
		return nil, err
	}
	return s.rules.Create(ctx, rule)
}

func (s SalaryRuleService) Update(ctx context.Context, rule *reference.SalaryRule) (*reference.SalaryRule, error) {
	existing, err := s.rules.Find(ctx, rule.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrRuleNotFound
	}
	if err := s.validate(ctx, rule); err != nil {
		return nil, err
	}
	return s.rules.Update(ctx, rule)
}

func (s SalaryRuleService) Delete(ctx context.Context, id uint64) error {
	existing, err := s.rules.Find(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrRuleNotFound
	}
	return s.rules.Delete(ctx, id)
}

func (s SalaryRuleService) validate(ctx context.Context, rule *reference.SalaryRule) error {
	if rule.Code == "" {
		return ErrRuleCode
	}
	category := ""
	if rule.Category != nil {
		category = *rule.Category
	}
	switch category {
	case RuleCategoryEarning, RuleCategoryDeduction:
	default:
		return ErrRuleCategory
	}
	computeType := ""
	if rule.ComputeType != nil {
		computeType = *rule.ComputeType
	}
	switch computeType {
	case ComputeTypeFixed, ComputeTypePercent:
	default:
		return ErrRuleComputeType
	}
	if rule.AccountDebitID == nil || rule.AccountCreditID == nil {
		return ErrRuleAccounts
	}
	for _, accountID := range []uint64{*rule.AccountDebitID, *rule.AccountCreditID} {
		account, err := s.accounts.Find(ctx, accountID)
		if err != nil {
			return err
		}
		if account == nil {
			return ErrRuleAccount
		}
	}
	return nil
}

type CreateRunRequest struct {
	OrganizationID uint64
	PeriodStart    time.Time
	PeriodEnd      time.Time
}

type PayrollService struct {
	runs        PayrollRunDAO
	payslips    PayslipDAO
	lines       PayslipLineDAO
	employees   EmployeeDAO
	contracts   EmploymentContractDAO
	attendances AttendanceDAO
	rules       dao.CRUD[reference.SalaryRule]
	journals    dao.CRUD[reference.Journal]
	poster      accounting.Poster
	sequences   sequence.Service
	tx          db.Transactioner
}

func NewPayrollService(
	runs PayrollRunDAO,
	payslips PayslipDAO,
	lines PayslipLineDAO,
	employees EmployeeDAO,
	contracts EmploymentContractDAO,
	attendances AttendanceDAO,
	rules dao.CRUD[reference.SalaryRule],
	journals dao.CRUD[reference.Journal],
	poster accounting.Poster,
	sequences sequence.Service,
	tx db.Transactioner,
) PayrollService {
	return PayrollService{
		runs:        runs,
		payslips:    payslips,
		lines:       lines,
		employees:   employees,
		contracts:   contracts,
		attendances: attendances,
		rules:       rules,
		journals:    journals,
		poster:      poster,
		sequences:   sequences,
		tx:          tx,
	}
}

func (s PayrollService) CreateRun(ctx context.Context, request CreateRunRequest) (*PayrollRun, error) {
	if request.PeriodStart.After(request.PeriodEnd) {
		return nil, ErrRunPeriod
	}
	name, err := s.sequences.Next(ctx, request.OrganizationID, SequencePayrollRunCode)
	if err != nil {
		return nil, ErrRunSequence
	}

	page, err := s.employees.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: request.OrganizationID}}})
	if err != nil {
		return nil, err
	}
	employees := page.Items

	payslips := make([]*Payslip, 0)
	allLines := make([][]*PayslipLine, 0)
	seen := false
	for _, employee := range employees {
		if !employee.Active || employee.OrganizationID == nil || *employee.OrganizationID != request.OrganizationID {
			continue
		}
		contract, err := s.contracts.FindActiveByEmployee(ctx, employee.ID)
		if err != nil {
			return nil, err
		}
		if contract == nil {
			continue
		}
		payslip, lines, err := s.computePayslip(ctx, employee, contract, request.PeriodStart, request.PeriodEnd)
		if err != nil {
			return nil, err
		}
		seen = true
		payslips = append(payslips, payslip)
		allLines = append(allLines, lines)
	}
	if !seen {
		return nil, ErrRunNoPayslips
	}

	run := &PayrollRun{
		OrganizationID: request.OrganizationID,
		Name:           &name,
		PeriodStart:    request.PeriodStart,
		PeriodEnd:      request.PeriodEnd,
		State:          RunStateDraft,
	}
	var created *PayrollRun
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		created, err = s.runs.CreateWithPayslipsTx(ctx, tx, run, payslips, allLines)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s PayrollService) ListRuns(ctx context.Context, q *query.Query) (*query.Page[PayrollRun], error) {
	return s.runs.List(ctx, q)
}

func (s PayrollService) FindRun(ctx context.Context, id uint64) (*PayrollRun, error) {
	return s.runs.Find(ctx, id)
}

func (s PayrollService) ListPayslipsByRun(ctx context.Context, runID uint64) ([]*Payslip, error) {
	return s.payslips.ListByRun(ctx, runID)
}

func (s PayrollService) ListPayslips(ctx context.Context, q *query.Query) (*query.Page[Payslip], error) {
	return s.payslips.List(ctx, q)
}

func (s PayrollService) FindPayslip(ctx context.Context, id uint64) (*Payslip, error) {
	return s.payslips.Find(ctx, id)
}

func (s PayrollService) ListPayslipLines(ctx context.Context, payslipID uint64) ([]*PayslipLine, error) {
	return s.lines.ListByPayslip(ctx, payslipID)
}

func (s PayrollService) computePayslip(ctx context.Context, employee *Employee, contract *EmploymentContract, periodStart, periodEnd time.Time) (*Payslip, []*PayslipLine, error) {
	if employee.OrganizationID == nil {
		return nil, nil, ErrRunNoPayslips
	}
	organizationID := *employee.OrganizationID

	attendanceSummary, err := s.summarizeAttendance(ctx, employee.ID, periodStart, periodEnd)
	if err != nil {
		return nil, nil, err
	}

	owned, err := s.rules.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, nil, err
	}
	global, err := s.rules.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.IsNull}}})
	if err != nil {
		return nil, nil, err
	}
	sorted := make([]*reference.SalaryRule, 0, len(owned.Items)+len(global.Items))
	sorted = append(sorted, owned.Items...)
	sorted = append(sorted, global.Items...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })

	gross := amount.Zero()
	deductions := amount.Zero()
	lines := make([]*PayslipLine, 0, len(sorted))
	for _, rule := range sorted {
		value, err := s.evaluateRule(rule, contract, gross, attendanceSummary)
		if err != nil {
			return nil, nil, err
		}
		category := ""
		if rule.Category != nil {
			category = *rule.Category
		}
		line := &PayslipLine{
			RuleID:   rule.ID,
			Code:     rule.Code,
			Name:     rule.Name,
			Category: category,
			Amount:   value.Float64(),
		}
		lines = append(lines, line)
		if category == RuleCategoryDeduction {
			deductions = deductions.Add(value)
		} else {
			gross = gross.Add(value)
		}
	}

	net := gross.Sub(deductions)
	return &Payslip{
		EmployeeID: employee.ID,
		ContractID: contract.ID,
		Gross:      gross.Round(4).Float64(),
		Net:        net.Round(4).Float64(),
		State:      PayslipStateDraft,
	}, lines, nil
}

func (s PayrollService) summarizeAttendance(ctx context.Context, employeeID uint64, periodStart, periodEnd time.Time) (*AttendanceSummary, error) {
	from := periodStart.Format("2006-01-02")
	to := periodEnd.Format("2006-01-02")
	records, err := s.attendances.ListByEmployeeAndDateRange(ctx, employeeID, from, to)
	if err != nil {
		return nil, err
	}
	summary := &AttendanceSummary{}
	for _, r := range records {
		if r.CheckOut != nil {
			summary.TotalWorkedHours += r.WorkedHours
			summary.TotalOvertimeHours += r.OvertimeHours
			summary.TotalLateMinutes += r.LateMinutes
			summary.TotalEarlyMinutes += r.EarlyDepartureMinutes
			summary.TotalDays++
		}
	}
	return summary, nil
}

func (s PayrollService) evaluateRule(rule *reference.SalaryRule, contract *EmploymentContract, gross amount.Amount, attendance *AttendanceSummary) (amount.Amount, error) {
	computeType := ""
	if rule.ComputeType != nil {
		computeType = *rule.ComputeType
	}
	ruleAmount := amount.Zero()
	if rule.Amount != nil {
		ruleAmount = amount.FromFloat64(*rule.Amount)
	}
	switch computeType {
	case ComputeTypeFixed:
		switch rule.Code {
		case "BAS":
			if rule.Amount == nil {
				return amount.FromFloat64(contract.Wage), nil
			}
		case "OT":
			if attendance.TotalOvertimeHours > 0 {
				rate := contract.Wage / DefaultWorkHoursPerDay / 22.0
				return amount.FromFloat64(attendance.TotalOvertimeHours * rate * 1.5), nil
			}
			return amount.Zero(), nil
		case "REG":
			rate := contract.Wage / DefaultWorkHoursPerDay / 22.0
			regularHours := attendance.TotalWorkedHours - attendance.TotalOvertimeHours
			if regularHours < 0 {
				regularHours = 0
			}
			return amount.FromFloat64(regularHours * rate), nil
		case "WORK":
			return amount.FromFloat64(attendance.TotalWorkedHours), nil
		case "DAYS":
			return amount.FromFloat64(float64(attendance.TotalDays)), nil
		case "LATE":
			return amount.FromFloat64(float64(attendance.TotalLateMinutes)), nil
		}
		return ruleAmount, nil
	case ComputeTypePercent:
		percent, err := gross.Mul(ruleAmount).Div(amount.FromFloat64(100))
		if err != nil {
			return amount.Amount{}, err
		}
		return percent, nil
	default:
		return amount.Amount{}, ErrRuleComputeType
	}
}

func (s PayrollService) Confirm(ctx context.Context, runID, journalID uint64, date time.Time) (*PayrollRun, error) {
	run, err := s.runs.Find(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, ErrRunNotFound
	}
	if run.State != RunStateDraft {
		return nil, ErrRunState
	}
	if date.IsZero() {
		date = time.Now().UTC()
	}

	payslips, err := s.payslips.ListByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	if len(payslips) == 0 {
		return nil, ErrRunNoPayslips
	}
	posting, err := s.buildPosting(ctx, runID, payslips)
	if err != nil {
		return nil, err
	}

	name := ""
	if run.Name != nil {
		name = *run.Name
	}
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		_, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
			OrganizationID: run.OrganizationID,
			JournalID:      journalID,
			Date:           date,
			Ref:            name,
			OriginType:     OriginPayrollRun,
			OriginID:       run.ID,
			Description:    "Payroll run",
			Lines:          posting,
		})
		if err != nil {
			return err
		}
		run.State = RunStateConfirmed
		if _, err := s.runs.UpdateTx(ctx, tx, run); err != nil {
			return err
		}
		for _, payslip := range payslips {
			payslip.State = PayslipStatePosted
			if _, err := s.payslips.UpdateTx(ctx, tx, payslip); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return run, nil
}

func (s PayrollService) buildPosting(ctx context.Context, runID uint64, payslips []*Payslip) ([]accounting.PostingLine, error) {
	debits := map[uint64]amount.Amount{}
	credits := map[uint64]amount.Amount{}
	for _, payslip := range payslips {
		lines, err := s.lines.ListByPayslip(ctx, payslip.ID)
		if err != nil {
			return nil, err
		}
		for _, line := range lines {
			rule, err := s.rules.Find(ctx, line.RuleID)
			if err != nil {
				return nil, err
			}
			if rule == nil {
				return nil, ErrRuleNotFound
			}
			if rule.AccountDebitID == nil || rule.AccountCreditID == nil {
				return nil, ErrRuleAccounts
			}
			value := amount.FromFloat64(line.Amount)
			debits[*rule.AccountDebitID] = debits[*rule.AccountDebitID].Add(value)
			credits[*rule.AccountCreditID] = credits[*rule.AccountCreditID].Add(value)
		}
	}

	posting := make([]accounting.PostingLine, 0, len(debits)+len(credits))
	for accountID, value := range debits {
		if value.IsPositive() {
			posting = append(posting, accounting.PostingLine{AccountID: accountID, Name: "Payroll", Debit: value.Round(4)})
		}
	}
	for accountID, value := range credits {
		if value.IsPositive() {
			posting = append(posting, accounting.PostingLine{AccountID: accountID, Name: "Payroll", Credit: value.Round(4)})
		}
	}
	return posting, nil
}

func (s PayrollService) Pay(ctx context.Context, runID, journalID, netPayableAccountID uint64, date time.Time) (*PayrollRun, error) {
	if netPayableAccountID == 0 {
		return nil, ErrNoNetPayable
	}
	run, err := s.runs.Find(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, ErrRunNotFound
	}
	if run.State != RunStateConfirmed {
		return nil, ErrRunState
	}
	if date.IsZero() {
		date = time.Now().UTC()
	}

	bankAccount, err := s.bankAccount(ctx, journalID)
	if err != nil {
		return nil, err
	}

	payslips, err := s.payslips.ListByRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	total := amount.Zero()
	for _, payslip := range payslips {
		total = total.Add(amount.FromFloat64(payslip.Net))
	}
	if total.IsZero() {
		return nil, ErrNoNetPayable
	}

	name := ""
	if run.Name != nil {
		name = *run.Name
	}
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		_, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
			OrganizationID: run.OrganizationID,
			JournalID:      journalID,
			Date:           date,
			Ref:            name,
			OriginType:     OriginPayrollRun,
			OriginID:       run.ID,
			Description:    "Payroll payment",
			Lines: []accounting.PostingLine{
				{AccountID: netPayableAccountID, Name: "Net Payable", Debit: total.Round(4)},
				{AccountID: bankAccount, Name: "Bank", Credit: total.Round(4)},
			},
		})
		if err != nil {
			return err
		}
		run.State = RunStatePaid
		if _, err := s.runs.UpdateTx(ctx, tx, run); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return run, nil
}

func (s PayrollService) Close(ctx context.Context, runID uint64) (*PayrollRun, error) {
	run, err := s.runs.Find(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, ErrRunNotFound
	}
	if run.State != RunStatePaid {
		return nil, ErrRunState
	}
	run.State = RunStateClosed
	return s.runs.Update(ctx, run)
}

func (s PayrollService) bankAccount(ctx context.Context, journalID uint64) (uint64, error) {
	journal, err := s.journals.Find(ctx, journalID)
	if err != nil {
		return 0, err
	}
	if journal == nil {
		return 0, ErrRunNotFound
	}
	if journal.DefaultAccountID == nil {
		return 0, ErrNoBankAccount
	}
	return *journal.DefaultAccountID, nil
}
