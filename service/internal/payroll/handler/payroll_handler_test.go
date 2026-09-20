package handler

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestSalaryRuleHandler_Create_CreatesRule(t *testing.T) {
	accounts := dao.CRUDMock[reference.Account]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Account, error) {
			return &reference.Account{Base: model.Base{ID: 1}}, nil
		},
	}
	rules := dao.CRUDMock[reference.SalaryRule]{
		CreateFunc: func(_ context.Context, rule *reference.SalaryRule) (*reference.SalaryRule, error) {
			rule.ID = 5
			return rule, nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, accounts)
	handler := NewSalaryRuleHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"code":"BASIC","name":"Basic Salary","category":"earning","compute_type":"fixed","amount":5000000,"account_debit_id":1,"account_credit_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/salary-rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Create_RejectsUnknownAccount(t *testing.T) {
	accounts := dao.CRUDMock[reference.Account]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Account, error) {
			return nil, nil
		},
	}
	svc := payroll.NewSalaryRuleService(dao.CRUDMock[reference.SalaryRule]{}, accounts)
	handler := NewSalaryRuleHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"code":"BASIC","name":"Basic Salary","category":"earning","compute_type":"fixed","amount":5000000,"account_debit_id":1,"account_credit_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/salary-rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Create_RejectsInvertedPeriod(t *testing.T) {
	sequences := sequence.NewSequenceService(sequence.DAOMock{})
	svc := payroll.NewPayrollService(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequences, inventory.TransactionerMock{})
	handler := NewPayrollRunHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"organization_id":1,"period_start":"2026-08-10","period_end":"2026-08-05"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Create_CreatesRun(t *testing.T) {
	sequences := sequence.NewSequenceService(sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "PR/00001"}, nil
		},
	})
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Employee], error) {
				return &query.Page[payroll.Employee]{Items: []*payroll.Employee{
					{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1)), Active: true},
				}}, nil
			},
		},
	}
	contracts := payroll.EmploymentContractDAOMock{
		FindActiveByEmployeeFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
			return &payroll.EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, Wage: 5000000}, nil
		},
	}
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.PayrollRun], error) {
				return &query.Page[payroll.PayrollRun]{Items: []*payroll.PayrollRun{}}, nil
			},
		},
		CreateWithPayslipsTxFunc: func(_ context.Context, _ *gorm.DB, run *payroll.PayrollRun, _ []*payroll.Payslip, _ [][]*payroll.PayslipLine) (*payroll.PayrollRun, error) {
			run.ID = 2
			return run, nil
		},
	}
	svc := payroll.NewPayrollService(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, employees, contracts, payroll.AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequences, inventory.TransactionerMock{})
	handler := NewPayrollRunHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"organization_id":1,"period_start":"2026-08-05","period_end":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestPayslipHandler_Get_ReturnsPayslip(t *testing.T) {
	payslips := payroll.PayslipDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Payslip]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Payslip, error) {
				return &payroll.Payslip{Base: model.Base{ID: 8}, RunID: 2, EmployeeID: 7, ContractID: 4, Gross: 5000000, Net: 4000000}, nil
			},
		},
	}
	lines := payroll.PayslipLineDAOMock{
		ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*payroll.PayslipLine, error) {
			return []*payroll.PayslipLine{
				{Base: model.Base{ID: 11}, PayslipID: 8, Code: "BASIC", Name: "Basic Salary", Category: "earning", Amount: 5000000},
			}, nil
		},
	}
	handler := NewPayslipHandler(payroll.NewPayrollService(payroll.PayrollRunDAOMock{}, payslips, lines, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequence.NewSequenceService(sequence.DAOMock{}), inventory.TransactionerMock{}))

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payslips/8", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}
