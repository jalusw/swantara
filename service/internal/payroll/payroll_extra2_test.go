package payroll

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func payrollTestService(rules dao.CRUDMock[reference.SalaryRule], attendances AttendanceDAOMock) PayrollService {
	return NewPayrollService(PayrollRunDAOMock{}, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, attendances, rules, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequence.NewSequenceService(sequence.DAOMock{}), inventory.TransactionerMock{})
}

func TestSalaryRuleService_Read(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("lists and finds", func(t *testing.T) {
		svc := NewSalaryRuleService(dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Account]{})
		if _, err := svc.List(ctx, q); err != nil {
			t.Errorf("list = %v", err)
		}
		if _, err := svc.Find(ctx, 1); err != nil {
			t.Errorf("find = %v", err)
		}
	})

	t.Run("deletes with lookup", func(t *testing.T) {
		svc := NewSalaryRuleService(dao.CRUDMock[reference.SalaryRule]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
				return &reference.SalaryRule{}, nil
			},
		}, dao.CRUDMock[reference.Account]{})
		if err := svc.Delete(ctx, 1); helper.AssertError(t, err, false, nil) {
			return
		}

		missing := NewSalaryRuleService(dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Account]{})
		helper.AssertError(t, missing.Delete(ctx, 1), true, ErrRuleNotFound)

		dbErr := errors.New("db down")
		failing := NewSalaryRuleService(dao.CRUDMock[reference.SalaryRule]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) { return nil, dbErr },
		}, dao.CRUDMock[reference.Account]{})
		helper.AssertError(t, failing.Delete(ctx, 1), true, dbErr)
	})
}

func TestPayrollService_Summarize(t *testing.T) {
	ctx := context.Background()
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	t.Run("sums closed records and skips open", func(t *testing.T) {
		checkOut := time.Date(2026, 8, 1, 17, 0, 0, 0, time.UTC)
		svc := payrollTestService(dao.CRUDMock[reference.SalaryRule]{}, AttendanceDAOMock{
			ListByEmployeeAndDateRangeFunc: func(_ context.Context, _ uint64, _, _ string) ([]*Attendance, error) {
				return []*Attendance{
					{WorkedHours: 8, OvertimeHours: 1, LateMinutes: 5, EarlyDepartureMinutes: 2, CheckOut: &checkOut},
					{WorkedHours: 4},
				}, nil
			},
		})
		got, err := svc.summarizeAttendance(ctx, 7, from, to)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.TotalWorkedHours != 8 || got.TotalDays != 1 || got.TotalLateMinutes != 5 {
			t.Errorf("summary = %+v", got)
		}
	})

	t.Run("propagates list error", func(t *testing.T) {
		dbErr := errors.New("db down")
		svc := payrollTestService(dao.CRUDMock[reference.SalaryRule]{}, AttendanceDAOMock{
			ListByEmployeeAndDateRangeFunc: func(_ context.Context, _ uint64, _, _ string) ([]*Attendance, error) {
				return nil, dbErr
			},
		})
		_, err := svc.summarizeAttendance(ctx, 7, from, to)
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestPayrollService_EvaluateRule(t *testing.T) {
	contract := &EmploymentContract{Wage: 22000}
	summary := &AttendanceSummary{TotalWorkedHours: 170, TotalOvertimeHours: 10, TotalLateMinutes: 5, TotalDays: 22}

	newSvc := func() PayrollService {
		return payrollTestService(dao.CRUDMock[reference.SalaryRule]{}, AttendanceDAOMock{})
	}

	t.Run("fixed codes", func(t *testing.T) {
		fixed := func(code string, amt *float64) *reference.SalaryRule {
			ct := ComputeTypeFixed
			return &reference.SalaryRule{Code: code, ComputeType: &ct, Amount: amt}
		}
		cases := []struct {
			name string
			rule *reference.SalaryRule
			want float64
		}{
			{name: "bas from wage", rule: fixed("BAS", nil), want: 22000},
			{name: "bas from amount", rule: fixed("BAS", helper.Ptr(20000.0)), want: 20000},
			{name: "ot with hours", rule: fixed("OT", nil), want: 10 * (22000 / 8 / 22.0) * 1.5},
			{name: "reg", rule: fixed("REG", nil), want: 160 * (22000 / 8 / 22.0)},
			{name: "work", rule: fixed("WORK", nil), want: 170},
			{name: "days", rule: fixed("DAYS", nil), want: 22},
			{name: "late", rule: fixed("LATE", nil), want: 5},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				got, err := newSvc().evaluateRule(tt.rule, contract, amount.Zero(), summary)
				if helper.AssertError(t, err, false, nil) {
					return
				}
				if got.Float64() != tt.want {
					t.Errorf("value = %v, want %v", got.Float64(), tt.want)
				}
			})
		}
	})

	t.Run("ot without hours and negative regulars", func(t *testing.T) {
		ct := ComputeTypeFixed
		empty := &AttendanceSummary{}
		got, err := newSvc().evaluateRule(&reference.SalaryRule{Code: "OT", ComputeType: &ct}, contract, amount.Zero(), empty)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if !got.IsZero() {
			t.Errorf("ot = %v, want 0", got)
		}

		negative := &AttendanceSummary{TotalWorkedHours: 2, TotalOvertimeHours: 5}
		got, err = newSvc().evaluateRule(&reference.SalaryRule{Code: "REG", ComputeType: &ct}, contract, amount.Zero(), negative)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if !got.IsZero() {
			t.Errorf("reg = %v, want 0", got)
		}
	})

	t.Run("percent and unknown", func(t *testing.T) {
		ct := ComputeTypePercent
		got, err := newSvc().evaluateRule(&reference.SalaryRule{Code: "TAX", ComputeType: &ct, Amount: helper.Ptr(10.0)}, contract, amount.FromFloat64(1000), summary)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.Float64() != 100 {
			t.Errorf("percent = %v, want 100", got.Float64())
		}

		bad := "mystery"
		_, err = newSvc().evaluateRule(&reference.SalaryRule{Code: "X", ComputeType: &bad}, contract, amount.Zero(), summary)
		helper.AssertError(t, err, true, ErrRuleComputeType)

		_, err = newSvc().evaluateRule(&reference.SalaryRule{Code: "X"}, contract, amount.Zero(), summary)
		helper.AssertError(t, err, true, ErrRuleComputeType)
	})
}

func TestPayrollService_ComputePayslip_Errors(t *testing.T) {
	ctx := context.Background()
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	dbErr := errors.New("db down")

	emp := func() *Employee { return &Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10))} }
	contract := &EmploymentContract{Base: model.Base{ID: 3}, Wage: 22000}

	t.Run("rejects employee without org", func(t *testing.T) {
		svc := payrollTestService(dao.CRUDMock[reference.SalaryRule]{}, AttendanceDAOMock{})
		_, _, err := svc.computePayslip(ctx, &Employee{Base: model.Base{ID: 7}}, contract, from, to)
		helper.AssertError(t, err, true, ErrRunNoPayslips)
	})

	t.Run("propagates attendance error", func(t *testing.T) {
		svc := payrollTestService(dao.CRUDMock[reference.SalaryRule]{}, AttendanceDAOMock{
			ListByEmployeeAndDateRangeFunc: func(_ context.Context, _ uint64, _, _ string) ([]*Attendance, error) {
				return nil, dbErr
			},
		})
		_, _, err := svc.computePayslip(ctx, emp(), contract, from, to)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("propagates rules error", func(t *testing.T) {
		rules := dao.CRUDMock[reference.SalaryRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
				return nil, dbErr
			},
		}
		svc := payrollTestService(rules, AttendanceDAOMock{})
		_, _, err := svc.computePayslip(ctx, emp(), contract, from, to)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("propagates rule eval error", func(t *testing.T) {
		bad := "mystery"
		rules := dao.CRUDMock[reference.SalaryRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
				return &query.Page[reference.SalaryRule]{Items: []*reference.SalaryRule{{Code: "X", ComputeType: &bad}}, Count: 1}, nil
			},
		}
		svc := payrollTestService(rules, AttendanceDAOMock{})
		_, _, err := svc.computePayslip(ctx, emp(), contract, from, to)
		helper.AssertError(t, err, true, ErrRuleComputeType)
	})
}

func TestPayrollService_Close_Bank(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(runs PayrollRunDAOMock, journals dao.CRUDMock[reference.Journal]) PayrollService {
		return NewPayrollService(runs, PayslipDAOMock{}, PayslipLineDAOMock{}, EmployeeDAOMock{}, EmploymentContractDAOMock{}, AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, journals, inventory.PosterMock{}, sequence.NewSequenceService(sequence.DAOMock{}), inventory.TransactionerMock{})
	}

	t.Run("close propagates lookup error", func(t *testing.T) {
		svc := newSvc(PayrollRunDAOMock{
			CRUDMock: dao.CRUDMock[PayrollRun]{
				FindFunc: func(_ context.Context, _ uint64) (*PayrollRun, error) { return nil, dbErr },
			},
		}, dao.CRUDMock[reference.Journal]{})
		_, err := svc.Close(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("bank account branches", func(t *testing.T) {
		svc := newSvc(PayrollRunDAOMock{}, dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) { return nil, dbErr },
		})
		_, err := svc.bankAccount(ctx, 1)
		helper.AssertError(t, err, true, dbErr)

		missing := newSvc(PayrollRunDAOMock{}, dao.CRUDMock[reference.Journal]{})
		_, err = missing.bankAccount(ctx, 1)
		helper.AssertError(t, err, true, ErrRunNotFound)
	})
}
