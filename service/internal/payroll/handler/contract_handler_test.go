package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func contractHandlerTestSvc(employees payroll.EmployeeDAOMock, contracts payroll.EmploymentContractDAOMock) payroll.HRService {
	return payroll.NewHRService(employees, contracts, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
}

func contractHandlerApp(t *testing.T, contracts payroll.EmploymentContractDAOMock, svc payroll.HRService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewContractHandler(svc)
	handler.Register(app, guards)
	return app
}

func TestContractHandler_List_ReturnsContracts(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.EmploymentContract], error) {
				return &query.Page[payroll.EmploymentContract]{Items: []*payroll.EmploymentContract{
					{Base: model.Base{ID: 4}, EmployeeID: 7, DateStart: timeNow(), Wage: 5000000, WageType: "monthly", CurrencyCode: "IDR", State: "active"},
				}, Count: 1}, nil
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/contracts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContractHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/contracts/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_List_ReturnsUnresolvedOrganization(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/contracts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_List_ReturnsServerError(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.EmploymentContract], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/contracts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContractHandler_Get_ReturnsContract(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return &payroll.EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, DateStart: timeNow(), Wage: 5000000, WageType: "monthly", CurrencyCode: "IDR", State: "active"}, nil
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/contracts/4", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContractHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/contracts/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Get_ReturnsUnresolvedOrganization(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/contracts/4", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Get_ReturnsNotFound(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return nil, nil
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/contracts/4", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContractHandler_Get_ReturnsServerError(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/contracts/4", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContractHandler_Create_CreatesContract(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		},
	}
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			CreateFunc: func(_ context.Context, contract *payroll.EmploymentContract) (*payroll.EmploymentContract, error) {
				contract.ID = 4
				return contract, nil
			},
		},
	}
	svc := contractHandlerTestSvc(employees, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	body := `{"employee_id":7,"date_start":"2026-08-01","wage":5000000}`
	resp, err := doRequest(app, http.MethodPost, "/contracts", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestContractHandler_Create_RejectsValidation(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":0,"wage":0}`
	resp, err := doRequest(app, http.MethodPost, "/contracts", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Create_ReturnsUnresolvedOrganization(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, noTenantGuards())

	body := `{"employee_id":7,"wage":5000000}`
	resp, err := doRequest(app, http.MethodPost, "/contracts", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Create_RejectsInvalidDateStart(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"wage":5000000,"date_start":"bad"}`
	resp, err := doRequest(app, http.MethodPost, "/contracts", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Create_RejectsInvalidDateEnd(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"wage":5000000,"date_end":"bad"}`
	resp, err := doRequest(app, http.MethodPost, "/contracts", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Create_MapsMissingEmployee(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return nil, nil
			},
		},
	}
	svc := contractHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"wage":5000000}`
	resp, err := doRequest(app, http.MethodPost, "/contracts", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Create_ReturnsServerError(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		},
	}
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			CreateFunc: func(_ context.Context, _ *payroll.EmploymentContract) (*payroll.EmploymentContract, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := contractHandlerTestSvc(employees, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	body := `{"employee_id":7,"wage":5000000}`
	resp, err := doRequest(app, http.MethodPost, "/contracts", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContractHandler_Update_UpdatesContract(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return &payroll.EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, DateStart: timeNow(), Wage: 5000000, State: "active"}, nil
			},
			UpdateFunc: func(_ context.Context, contract *payroll.EmploymentContract) (*payroll.EmploymentContract, error) {
				return contract, nil
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	body := `{"date_start":"2026-08-01","wage":6000000}`
	resp, err := doRequest(app, http.MethodPut, "/contracts/4", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContractHandler_Update_RejectsInvalidID(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"wage":6000000}`
	resp, err := doRequest(app, http.MethodPut, "/contracts/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Update_ReturnsUnresolvedOrganization(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, noTenantGuards())

	body := `{"wage":6000000}`
	resp, err := doRequest(app, http.MethodPut, "/contracts/4", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Update_ReturnsNotFound(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return nil, nil
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	body := `{"wage":6000000}`
	resp, err := doRequest(app, http.MethodPut, "/contracts/4", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContractHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	body := `{"wage":6000000}`
	resp, err := doRequest(app, http.MethodPut, "/contracts/4", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContractHandler_Update_RejectsValidation(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return &payroll.EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, DateStart: timeNow()}, nil
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	body := `{"wage":0}`
	resp, err := doRequest(app, http.MethodPut, "/contracts/4", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Update_RejectsInvalidDateStart(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return &payroll.EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, DateStart: timeNow()}, nil
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	body := `{"wage":6000000,"date_start":"bad"}`
	resp, err := doRequest(app, http.MethodPut, "/contracts/4", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Update_RejectsInvalidDateEnd(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return &payroll.EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, DateStart: timeNow()}, nil
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	body := `{"wage":6000000,"date_end":"bad"}`
	resp, err := doRequest(app, http.MethodPut, "/contracts/4", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return &payroll.EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, DateStart: timeNow(), Wage: 5000000}, nil
			},
			UpdateFunc: func(_ context.Context, _ *payroll.EmploymentContract) (*payroll.EmploymentContract, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	body := `{"wage":6000000}`
	resp, err := doRequest(app, http.MethodPut, "/contracts/4", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestContractHandler_Terminate_RejectsInvalidID(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/contracts/abc/terminate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Terminate_ReturnsUnresolvedOrganization(t *testing.T) {
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{})
	app := contractHandlerApp(t, payroll.EmploymentContractDAOMock{}, svc, noTenantGuards())

	resp, err := doRequest(app, http.MethodPost, "/contracts/4/terminate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestContractHandler_Terminate_ReturnsNotFound(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return nil, nil
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/contracts/4/terminate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestContractHandler_Terminate_ReturnsServerErrorOnFind(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := contractHandlerTestSvc(payroll.EmployeeDAOMock{}, contracts)
	app := contractHandlerApp(t, contracts, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/contracts/4/terminate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
