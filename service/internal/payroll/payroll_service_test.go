package payroll

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func payrollSequenceMock() sequence.DAOMock {
	return sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "PR/00001"}, nil
		},
	}
}

func TestSalaryRuleService_Create_RejectsUnknownComputeType(t *testing.T) {
	ctx := context.Background()
	svc := NewSalaryRuleService(dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Account]{})

	debit := uint64(1)
	credit := uint64(2)
	_, err := svc.Create(ctx, &reference.SalaryRule{
		Code: "SAL", Name: "Salary", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("magic"),
		AccountDebitID: &debit, AccountCreditID: &credit,
	})
	if helper.AssertError(t, err, true, ErrRuleComputeType) {
		return
	}
}

func TestSalaryRuleService_Create_RejectsUnknownAccount(t *testing.T) {
	ctx := context.Background()
	accounts := dao.CRUDMock[reference.Account]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Account, error) {
			return nil, nil
		},
	}
	svc := NewSalaryRuleService(dao.CRUDMock[reference.SalaryRule]{}, accounts)

	debit := uint64(1)
	credit := uint64(2)
	_, err := svc.Create(ctx, &reference.SalaryRule{
		Code: "SAL", Name: "Salary", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed"),
		AccountDebitID: &debit, AccountCreditID: &credit,
	})
	if helper.AssertError(t, err, true, ErrRuleAccount) {
		return
	}
}

func TestPayrollService_CreateRun_ComputesPayslips(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	employees := EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*Employee, error) {
				return &Employee{Base: model.Base{ID: 1}, Active: true, OrganizationID: &organizationID}, nil
			},
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Employee], error) {
				return &query.Page[Employee]{Items: []*Employee{{Base: model.Base{ID: 1}, Active: true, OrganizationID: &organizationID}}}, nil
			},
		},
	}
	contracts := EmploymentContractDAOMock{
		FindActiveByEmployeeFunc: func(_ context.Context, _ uint64) (*EmploymentContract, error) {
			return &EmploymentContract{Base: model.Base{ID: 1}, Wage: 8000000}, nil
		},
	}
	rules := dao.CRUDMock[reference.SalaryRule]{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.SalaryRule], error) {
			amount := 1000000.0
			items := []*reference.SalaryRule{
				{Base: model.Base{ID: 1}, Code: "BAS", Name: "Basic Salary", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed")},
				{Base: model.Base{ID: 2}, Code: "ALL", Name: "Allowance", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed"), Amount: &amount},
			}
			if len(q.Filters) > 0 && q.Filters[0].Operator == query.IsNull {
				items = []*reference.SalaryRule{{Base: model.Base{ID: 3}, Code: "TAX", Name: "Tax", Category: helper.Ptr("deduction"), ComputeType: helper.Ptr("percent"), Amount: helper.Ptr(5.0)}}
			}
			return &query.Page[reference.SalaryRule]{Items: items}, nil
		},
	}
	var persistedPayslips []*Payslip
	runs := PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[PayrollRun]{},
		CreateWithPayslipsTxFunc: func(_ context.Context, _ *gorm.DB, run *PayrollRun, payslips []*Payslip, _ [][]*PayslipLine) (*PayrollRun, error) {
			run.ID = 1
			persistedPayslips = payslips
			return run, nil
		},
	}
	svc := NewPayrollService(runs, PayslipDAOMock{}, PayslipLineDAOMock{}, employees, contracts, AttendanceDAOMock{}, rules, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequence.NewSequenceService(payrollSequenceMock()), inventory.TransactionerMock{})

	run, err := svc.CreateRun(ctx, CreateRunRequest{
		OrganizationID: organizationID,
		PeriodStart:    time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:      time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
	})
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if run.State != RunStateDraft || run.Name == nil || *run.Name != "PR/00001" {
		t.Fatalf("unexpected run state/name: %+v", run)
	}
	if len(persistedPayslips) != 1 {
		t.Fatalf("expected 1 payslip, got %d", len(persistedPayslips))
	}
	payslip := persistedPayslips[0]
	if payslip.Gross != 9000000 {
		t.Fatalf("expected gross 9000000, got %v", payslip.Gross)
	}
	if payslip.Net != 8550000 {
		t.Fatalf("expected net 8550000, got %v", payslip.Net)
	}
}

func TestPayrollService_CreateRun_RejectsNoPayslips(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	employees := EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[Employee]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[Employee], error) {
				return &query.Page[Employee]{Items: []*Employee{}}, nil
			},
		},
	}
	svc := NewPayrollService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, employees, EmploymentContractDAOMock{}, AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequence.NewSequenceService(payrollSequenceMock()), inventory.TransactionerMock{})

	_, err := svc.CreateRun(ctx, CreateRunRequest{
		OrganizationID: organizationID,
		PeriodStart:    time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:      time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
	})
	if helper.AssertError(t, err, true, ErrRunNoPayslips) {
		return
	}
}

func TestPayrollService_Confirm_PostsSingleBalancedMove(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	run := &PayrollRun{Base: model.Base{ID: 1}, OrganizationID: organizationID, Name: helper.Ptr("PR/00001"), State: RunStateDraft}
	runs := PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
				return run, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, updated *PayrollRun) (*PayrollRun, error) {
			run = updated
			return updated, nil
		},
	}
	salaryExpense := uint64(6100)
	salaryPayable := uint64(2200)
	wht := uint64(2180)
	earningDebit := salaryExpense
	earningCredit := salaryPayable
	deductionDebit := salaryPayable
	deductionCredit := wht
	earningRule := &reference.SalaryRule{Base: model.Base{ID: 1}, Code: "BAS", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed"), AccountDebitID: &earningDebit, AccountCreditID: &earningCredit}
	deductionRule := &reference.SalaryRule{Base: model.Base{ID: 2}, Code: "TAX", Category: helper.Ptr("deduction"), ComputeType: helper.Ptr("percent"), AccountDebitID: &deductionDebit, AccountCreditID: &deductionCredit}
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, id uint64) (*reference.SalaryRule, error) {
			if id == 1 {
				return earningRule, nil
			}
			return deductionRule, nil
		},
	}
	payslips := PayslipDAOMock{
		CRUDMock: dao.CRUDMock[Payslip]{},
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
			return []*Payslip{{Base: model.Base{ID: 1}, RunID: 1, EmployeeID: 1, Gross: 9000000, Net: 8550000, State: PayslipStateDraft}}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, payslip *Payslip) (*Payslip, error) {
			return payslip, nil
		},
	}
	lines := PayslipLineDAOMock{
		ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*PayslipLine, error) {
			return []*PayslipLine{
				{RuleID: 1, Code: "BAS", Category: "earning", Amount: 8000000},
				{RuleID: 2, Code: "TAX", Category: "deduction", Amount: 450000},
			}, nil
		},
	}
	var posted accounting.PostRequest
	poster := inventory.PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = request
			return &accounting.JournalEntry{Base: model.Base{ID: 5}}, nil
		},
	}
	svc := NewPayrollService(runs, payslips, lines, EmployeeDAOMock{}, EmploymentContractDAOMock{}, AttendanceDAOMock{}, rules, dao.CRUDMock[reference.Journal]{}, poster, sequence.NewSequenceService(payrollSequenceMock()), inventory.TransactionerMock{})

	confirmed, err := svc.Confirm(ctx, 1, 9, time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC))
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if confirmed.State != RunStateConfirmed {
		t.Fatalf("expected confirmed run, got %s", confirmed.State)
	}
	if len(posted.Lines) != 4 {
		t.Fatalf("expected 4 posting lines, got %d", len(posted.Lines))
	}
	debits := amount.Zero()
	credits := amount.Zero()
	for _, line := range posted.Lines {
		debits = debits.Add(line.Debit)
		credits = credits.Add(line.Credit)
	}
	if !amount.IsBalanced(debits, credits, 4) {
		t.Fatalf("expected balanced entry, debits=%v credits=%v", debits, credits)
	}
}

func TestPayrollService_Confirm_RejectsWrongState(t *testing.T) {
	ctx := context.Background()
	runs := PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
				return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed}, nil
			},
		},
	}
	svc := NewPayrollService(runs, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequence.NewSequenceService(payrollSequenceMock()), inventory.TransactionerMock{})

	if _, err := svc.Confirm(ctx, 1, 9, time.Now()); !helper.AssertError(t, err, true, ErrRunState) {
		return
	}
}

func TestPayrollService_Pay_PostsNetPayableToBank(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	run := &PayrollRun{Base: model.Base{ID: 1}, OrganizationID: organizationID, Name: helper.Ptr("PR/00001"), State: RunStateConfirmed}
	runs := PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
				return run, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, updated *PayrollRun) (*PayrollRun, error) {
			run = updated
			return updated, nil
		},
	}
	bankAccount := uint64(1110)
	journal := &reference.Journal{Base: model.Base{ID: 9}, DefaultAccountID: &bankAccount}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return journal, nil
		},
	}
	netPayable := uint64(2200)
	earningCredit := netPayable
	earningRule := &reference.SalaryRule{Base: model.Base{ID: 1}, Code: "BAS", Category: helper.Ptr("earning"), ComputeType: helper.Ptr("fixed"), AccountCreditID: &earningCredit}
	rules := dao.CRUDMock[reference.SalaryRule]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
			return &query.Page[reference.SalaryRule]{Items: []*reference.SalaryRule{earningRule}}, nil
		},
	}
	payslips := PayslipDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*Payslip, error) {
			return []*Payslip{{Base: model.Base{ID: 1}, RunID: 1, Net: 4000000}, {Base: model.Base{ID: 2}, RunID: 1, Net: 4550000}}, nil
		},
	}
	var posted accounting.PostRequest
	poster := inventory.PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = request
			return &accounting.JournalEntry{Base: model.Base{ID: 6}}, nil
		},
	}
	svc := NewPayrollService(runs, payslips, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, AttendanceDAOMock{}, rules, journals, poster, sequence.NewSequenceService(payrollSequenceMock()), inventory.TransactionerMock{})

	paid, err := svc.Pay(ctx, 1, 9, netPayable, time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC))
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if paid.State != RunStatePaid {
		t.Fatalf("expected paid run, got %s", paid.State)
	}
	if len(posted.Lines) != 2 {
		t.Fatalf("expected 2 posting lines, got %d", len(posted.Lines))
	}
	total := posted.Lines[0].Debit.Add(posted.Lines[1].Debit).Sub(posted.Lines[0].Credit).Sub(posted.Lines[1].Credit)
	if !total.IsZero() {
		t.Fatalf("expected balanced payment entry, got %v", total)
	}
}

func TestPayrollService_Pay_RejectsMissingBankAccount(t *testing.T) {
	ctx := context.Background()
	runs := PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
				return &PayrollRun{Base: model.Base{ID: 1}, State: RunStateConfirmed}, nil
			},
		},
	}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 9}}, nil
		},
	}
	svc := NewPayrollService(runs, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, journals, inventory.PosterMock{}, sequence.NewSequenceService(payrollSequenceMock()), inventory.TransactionerMock{})

	if _, err := svc.Pay(ctx, 1, 9, 2200, time.Now()); !helper.AssertError(t, err, true, ErrNoBankAccount) {
		return
	}
}

func TestPayrollService_Close(t *testing.T) {
	ctx := context.Background()
	run := &PayrollRun{Base: model.Base{ID: 1}, State: RunStatePaid}
	runs := PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) {
				return run, nil
			},
			UpdateFunc: func(_ context.Context, updated *PayrollRun) (*PayrollRun, error) {
				run = updated
				return updated, nil
			},
		},
	}
	svc := NewPayrollService(runs, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequence.NewSequenceService(payrollSequenceMock()), inventory.TransactionerMock{})

	closed, err := svc.Close(ctx, 1)
	if helper.AssertError(t, err, false, nil) {
		return
	}
	if closed.State != RunStateClosed {
		t.Fatalf("expected closed run, got %s", closed.State)
	}
}
