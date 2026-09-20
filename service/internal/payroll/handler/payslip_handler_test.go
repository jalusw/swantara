package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func payslipHandlerApp(t *testing.T, payslips payroll.PayslipDAOMock, lines payroll.PayslipLineDAOMock, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	svc := payroll.NewPayrollService(payroll.PayrollRunDAOMock{}, payslips, lines, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.AttendanceDAOMock{}, dao.CRUDMock[reference.SalaryRule]{}, dao.CRUDMock[reference.Journal]{}, inventory.PosterMock{}, sequence.NewSequenceService(sequence.DAOMock{}), inventory.TransactionerMock{})
	handler := NewPayslipHandler(svc)
	handler.Register(app, guards)
	return app
}

func samplePayslip() *payroll.Payslip {
	return &payroll.Payslip{Base: model.Base{ID: 8}, RunID: 2, EmployeeID: 7, ContractID: 4, Gross: 5000000, Net: 4000000}
}

func samplePayslipLines() []*payroll.PayslipLine {
	return []*payroll.PayslipLine{
		{Base: model.Base{ID: 11}, PayslipID: 8, Code: "BASIC", Name: "Basic Salary", Category: "earning", Amount: 5000000},
	}
}

func TestPayslipHandler_List_ReturnsPayslips(t *testing.T) {
	payslips := payroll.PayslipDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Payslip]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Payslip], error) {
				return &query.Page[payroll.Payslip]{Items: []*payroll.Payslip{samplePayslip()}, Count: 1}, nil
			},
		},
	}
	lines := payroll.PayslipLineDAOMock{
		ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*payroll.PayslipLine, error) {
			return samplePayslipLines(), nil
		},
	}
	app := payslipHandlerApp(t, payslips, lines, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payslips/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPayslipHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := payslipHandlerApp(t, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payslips/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayslipHandler_List_ReturnsServerError(t *testing.T) {
	payslips := payroll.PayslipDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Payslip]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Payslip], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := payslipHandlerApp(t, payslips, payroll.PayslipLineDAOMock{}, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payslips/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPayslipHandler_List_ReturnsLineServerError(t *testing.T) {
	payslips := payroll.PayslipDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Payslip]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Payslip], error) {
				return &query.Page[payroll.Payslip]{Items: []*payroll.Payslip{samplePayslip()}, Count: 1}, nil
			},
		},
	}
	lines := payroll.PayslipLineDAOMock{
		ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*payroll.PayslipLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := payslipHandlerApp(t, payslips, lines, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payslips/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPayslipHandler_Get_RejectsInvalidID(t *testing.T) {
	app := payslipHandlerApp(t, payroll.PayslipDAOMock{}, payroll.PayslipLineDAOMock{}, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payslips/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPayslipHandler_Get_ReturnsNotFound(t *testing.T) {
	payslips := payroll.PayslipDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Payslip]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Payslip, error) {
				return nil, nil
			},
		},
	}
	app := payslipHandlerApp(t, payslips, payroll.PayslipLineDAOMock{}, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payslips/8", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPayslipHandler_Get_ReturnsServerError(t *testing.T) {
	payslips := payroll.PayslipDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Payslip]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Payslip, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := payslipHandlerApp(t, payslips, payroll.PayslipLineDAOMock{}, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payslips/8", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPayslipHandler_Get_ReturnsLineServerError(t *testing.T) {
	payslips := payroll.PayslipDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Payslip]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Payslip, error) {
				return samplePayslip(), nil
			},
		},
	}
	lines := payroll.PayslipLineDAOMock{
		ListByPayslipFunc: func(_ context.Context, _ uint64) ([]*payroll.PayslipLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := payslipHandlerApp(t, payslips, lines, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/payslips/8", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
