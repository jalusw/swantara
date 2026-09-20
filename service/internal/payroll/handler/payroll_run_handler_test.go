package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func samplePayrollRun() *payroll.PayrollRun {
	name := "PR/00001"
	return &payroll.PayrollRun{Base: model.Base{ID: 2}, OrganizationID: 1, Name: &name, PeriodStart: timeNow(), PeriodEnd: timeNow().AddDate(0, 0, 5), State: "draft"}
}

func payrollRunTestSvc(runs payroll.PayrollRunDAOMock, payslips payroll.PayslipDAOMock, lines payroll.PayslipLineDAOMock, rules dao.CRUD[reference.SalaryRule], journals dao.CRUD[reference.Journal]) payroll.PayrollService {
	return payroll.NewPayrollService(runs, payslips, lines, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.AttendanceDAOMock{}, rules, journals, inventory.PosterMock{}, sequence.NewSequenceService(sequence.DAOMock{}), inventory.TransactionerMock{})
}

func payrollRunHandlerApp(t *testing.T, runs payroll.PayrollRunDAOMock, payslips payroll.PayslipDAOMock, lines payroll.PayslipLineDAOMock, svc payroll.PayrollService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewPayrollRunHandler(svc)
	handler.Register(app, guards)
	return app
}

func TestPayrollRunHandler_List_ReturnsRuns(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.PayrollRun], error) {
				return &query.Page[payroll.PayrollRun]{Items: []*payroll.PayrollRun{samplePayrollRun()}, Count: 1}, nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPayrollRunHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_List_RejectsMissingTenant(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestPayrollRunHandler_List_ReturnsServerError(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.PayrollRun], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Get_ReturnsRunWithPayslips(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return samplePayrollRun(), nil
			},
		},
	}
	payslips := payroll.PayslipDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Payslip]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Payslip, error) {
				return &payroll.Payslip{}, nil
			},
		},
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*payroll.Payslip, error) {
			return []*payroll.Payslip{
				{Base: model.Base{ID: 8}, RunID: 2, EmployeeID: 7, ContractID: 4, Gross: 5000000, Net: 4000000},
			}, nil
		},
	}
	lines := payroll.PayslipLineDAOMock{
		ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*payroll.PayslipLine, error) {
			return []*payroll.PayslipLine{
				{Base: model.Base{ID: 11}, PayslipID: 8, Code: "BASIC", Name: "Basic Salary", Category: "earning", Amount: 5000000},
			}, nil
		},
	}
	svc := payrollRunTestSvc(runs, payslips, lines, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payslips, lines, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Get_ReturnsNotFound(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return nil, nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Get_ReturnsNotFoundForOtherTenant(t *testing.T) {
	run := samplePayrollRun()
	run.OrganizationID = 2
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return run, nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Get_ReturnsServerError(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Get_ReturnsPayslipServerError(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return samplePayrollRun(), nil
			},
		},
	}
	payslips := payroll.PayslipDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*payroll.Payslip, error) {
			return nil, errors.New("db down")
		},
	}
	svc := payrollRunTestSvc(runs, payslips, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payslips, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Get_ReturnsLineServerError(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return samplePayrollRun(), nil
			},
		},
	}
	payslips := payroll.PayslipDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*payroll.Payslip, error) {
			return []*payroll.Payslip{{Base: model.Base{ID: 8}, RunID: 2, EmployeeID: 7, ContractID: 4}}, nil
		},
	}
	lines := payroll.PayslipLineDAOMock{
		ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*payroll.PayslipLine, error) {
			return nil, errors.New("db down")
		},
	}
	svc := payrollRunTestSvc(runs, payslips, lines, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payslips, lines, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payroll-runs/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Confirm_ConfirmsRun(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return samplePayrollRun(), nil
			},
		},
	}
	payslips := payroll.PayslipDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*payroll.Payslip, error) {
			return []*payroll.Payslip{{Base: model.Base{ID: 8}, RunID: 2, EmployeeID: 7, ContractID: 4, Net: 4000000}}, nil
		},
	}
	lines := payroll.PayslipLineDAOMock{
		ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*payroll.PayslipLine, error) {
			return []*payroll.PayslipLine{
				{Base: model.Base{ID: 11}, PayslipID: 8, RuleID: 5, Code: "BASIC", Name: "Basic Salary", Category: "earning", Amount: 5000000},
			}, nil
		},
	}
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			debit := uint64(1)
			credit := uint64(2)
			return &reference.SalaryRule{Base: model.Base{ID: 5}, Code: "BASIC", Category: helper.Ptr("earning"), AccountDebitID: &debit, AccountCreditID: &credit}, nil
		},
	}
	svc := payrollRunTestSvc(runs, payslips, lines, rules, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payslips, lines, svc, passthroughGuards())

	body := `{"journal_id":1,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/confirm", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Confirm_RejectsInvalidID(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/abc/confirm", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Confirm_RejectsValidation(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":0,"date":""}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/confirm", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Confirm_RejectsInvalidDate(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"date":"bad"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/confirm", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Confirm_MapsNotFound(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return nil, nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/confirm", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Confirm_MapsInvalidState(t *testing.T) {
	run := samplePayrollRun()
	run.State = "confirmed"
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return run, nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/confirm", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Confirm_MapsNoPayslips(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return samplePayrollRun(), nil
			},
		},
	}
	payslips := payroll.PayslipDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*payroll.Payslip, error) {
			return []*payroll.Payslip{}, nil
		},
	}
	svc := payrollRunTestSvc(runs, payslips, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payslips, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/confirm", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Pay_PaysRun(t *testing.T) {
	run := samplePayrollRun()
	run.State = "confirmed"
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return run, nil
			},
		},
	}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			account := uint64(3)
			return &reference.Journal{Base: model.Base{ID: 1}, DefaultAccountID: &account}, nil
		},
	}
	payslips := payroll.PayslipDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*payroll.Payslip, error) {
			return []*payroll.Payslip{{Base: model.Base{ID: 8}, RunID: 2, EmployeeID: 7, ContractID: 4, Net: 4000000}}, nil
		},
	}
	svc := payrollRunTestSvc(runs, payslips, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, journals)
	app := payrollRunHandlerApp(t, runs, payslips, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"net_payable_account_id":2,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Pay_RejectsInvalidID(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/abc/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Pay_RejectsValidation(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":0,"date":""}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Pay_RejectsInvalidDate(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"date":"bad"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Pay_MapsNotFound(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return nil, nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"net_payable_account_id":2,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Pay_MapsInvalidState(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return samplePayrollRun(), nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"net_payable_account_id":2,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Pay_MapsNoBankAccount(t *testing.T) {
	run := samplePayrollRun()
	run.State = "confirmed"
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return run, nil
			},
		},
	}
	journals := dao.CRUDMock[reference.Journal]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, journals)
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"journal_id":1,"net_payable_account_id":2,"date":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/pay", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Close_ClosesRun(t *testing.T) {
	run := samplePayrollRun()
	run.State = "paid"
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return run, nil
			},
			UpdateFunc: func(_ context.Context, updated *payroll.PayrollRun) (*payroll.PayrollRun, error) {
				return updated, nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/close", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Close_RejectsInvalidID(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/abc/close", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Close_MapsNotFound(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return nil, nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/close", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Close_MapsInvalidState(t *testing.T) {
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return samplePayrollRun(), nil
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/close", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Close_ReturnsServerError(t *testing.T) {
	run := samplePayrollRun()
	run.State = "paid"
	runs := payroll.PayrollRunDAOMock{
		CRUDMock: dao.CRUDMock[payroll.PayrollRun]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.PayrollRun, error) {
				return run, nil
			},
			UpdateFunc: func(_ context.Context, _ *payroll.PayrollRun) (*payroll.PayrollRun, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := payrollRunTestSvc(runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, runs, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/payroll-runs/2/close", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Create_RejectsValidation(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"organization_id":0,"period_start":"","period_end":""}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Create_RejectsInvalidPeriodStart(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"organization_id":1,"period_start":"bad","period_end":"2026-08-10"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayrollRunHandler_Create_RejectsInvalidPeriodEnd(t *testing.T) {
	svc := payrollRunTestSvc(payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{})
	app := payrollRunHandlerApp(t, payroll.PayrollRunDAOMock{}, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, svc, passthroughGuards())

	body := `{"organization_id":1,"period_start":"2026-08-05","period_end":"bad"}`
	resp, err := doRequest(app, http.MethodPost, "/payroll-runs", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}
