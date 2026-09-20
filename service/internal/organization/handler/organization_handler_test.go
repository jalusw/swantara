package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/organization"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func organizationHandlerTest(
	t *testing.T,
	orgs dao.CRUDMock[reference.Organization],
	svc organization.Service,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	h := NewOrganizationHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func organizationTestSvc(
	orgs dao.CRUDMock[reference.Organization],
	currencies handlerCurrencySearchMock,
	members handlerMemberProvisionerMock,
) organization.Service {
	return organization.NewOrganizationService(orgs, currencies, members)
}

func organizationHandlerTestNoCaller(
	t *testing.T,
	orgs dao.CRUDMock[reference.Organization],
	svc organization.Service,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewOrganizationHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func sampleOrganization() *reference.Organization {
	return &reference.Organization{
		Base:              model.Base{ID: 1},
		Name:              "Acme",
		BaseCurrency:      "IDR",
		Timezone:          "Asia/Makassar",
		TaxYearStartMonth: 1,
	}
}

func TestOrganizationHandler_List_ReturnsOrganizations(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Organization], error) {
			return &query.Page[reference.Organization]{Items: []*reference.Organization{sampleOrganization()}, Count: 1}, nil
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOrganizationHandler_List_RejectsInvalidQuery(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Organization], error) {
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}, Count: 0}, nil
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_List_ReturnsServerError(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Organization], error) {
			return nil, errors.New("db down")
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOrganizationHandler_List_ExportsCSV(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Organization], error) {
			return &query.Page[reference.Organization]{Items: []*reference.Organization{sampleOrganization()}, Count: 1}, nil
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOrganizationHandler_Get_ReturnsOrganization(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return sampleOrganization(), nil
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOrganizationHandler_Get_ReturnsNotFound(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, nil
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOrganizationHandler_Get_ReturnsServerError(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, errors.New("db down")
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOrganizationHandler_Get_RejectsInvalidID(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodGet, "/organizations/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Create_ReturnsUnauthorized(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTestNoCaller(t, orgs, svc)

	body := `{"name":"Acme","base_currency":"IDR"}`
	resp, err := doRequest(app, http.MethodPost, "/organizations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestOrganizationHandler_Create_CreatesOrganization(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		CreateFunc: func(_ context.Context, org *reference.Organization) (*reference.Organization, error) {
			org.ID = 1
			return org, nil
		},
	}
	currencies := handlerCurrencySearchMock{
		search: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return &reference.Currency{Code: "IDR"}, nil
		},
	}
	members := handlerMemberProvisionerMock{
		provision: func(_ context.Context, _, _ uint64) error {
			return nil
		},
	}
	svc := organizationTestSvc(orgs, currencies, members)
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"Acme","base_currency":"IDR","timezone":"Asia/Makassar","tax_year_start_month":1}`
	resp, err := doRequest(app, http.MethodPost, "/organizations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestOrganizationHandler_Create_RejectsValidation(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"","base_currency":"IDR"}`
	resp, err := doRequest(app, http.MethodPost, "/organizations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Create_MapsServiceErrors(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		CreateFunc: func(_ context.Context, org *reference.Organization) (*reference.Organization, error) {
			return nil, organization.ErrNameRequired
		},
	}
	currencies := handlerCurrencySearchMock{
		search: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return &reference.Currency{Code: "IDR"}, nil
		},
	}
	members := handlerMemberProvisionerMock{}
	svc := organizationTestSvc(orgs, currencies, members)
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"","base_currency":"IDR"}`
	resp, err := doRequest(app, http.MethodPost, "/organizations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Create_RejectsUnknownBaseCurrency(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	currencies := handlerCurrencySearchMock{
		search: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return nil, nil
		},
	}
	svc := organizationTestSvc(orgs, currencies, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"Acme","base_currency":"XXX"}`
	resp, err := doRequest(app, http.MethodPost, "/organizations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Create_RejectsParentCycle(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	currencies := handlerCurrencySearchMock{
		search: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return &reference.Currency{Code: "IDR"}, nil
		},
	}
	svc := organizationTestSvc(orgs, currencies, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"Acme","base_currency":"IDR","parent_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/organizations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Create_RejectsParentNotFound(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, nil
		},
	}
	currencies := handlerCurrencySearchMock{
		search: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return &reference.Currency{Code: "IDR"}, nil
		},
	}
	svc := organizationTestSvc(orgs, currencies, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"Acme","base_currency":"IDR","parent_id":99}`
	resp, err := doRequest(app, http.MethodPost, "/organizations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Create_ReturnsServerError(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	currencies := handlerCurrencySearchMock{
		search: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return nil, errors.New("currency lookup failed")
		},
	}
	svc := organizationTestSvc(orgs, currencies, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"Acme","base_currency":"IDR"}`
	resp, err := doRequest(app, http.MethodPost, "/organizations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOrganizationHandler_Update_ReturnsNotFound(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, nil
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/organizations/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOrganizationHandler_Update_ReturnsServerError(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, errors.New("db down")
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/organizations/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOrganizationHandler_Update_UpdatesOrganization(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return sampleOrganization(), nil
		},
		UpdateFunc: func(_ context.Context, org *reference.Organization) (*reference.Organization, error) {
			return org, nil
		},
	}
	currencies := handlerCurrencySearchMock{
		search: func(_ context.Context, _ string, _ any) (*reference.Currency, error) {
			return &reference.Currency{Code: "USD"}, nil
		},
	}
	svc := organizationTestSvc(orgs, currencies, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"Acme Corp","base_currency":"USD"}`
	resp, err := doRequest(app, http.MethodPut, "/organizations/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOrganizationHandler_Create_RejectsBlankName(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"   ","base_currency":"IDR"}`
	resp, err := doRequest(app, http.MethodPost, "/organizations/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Update_RejectsValidation(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/organizations/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Update_RejectsInvalidID(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/organizations/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Delete_DeletesOrganization(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return sampleOrganization(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return nil
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodDelete, "/organizations/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestOrganizationHandler_Delete_RejectsInvalidID(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodDelete, "/organizations/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOrganizationHandler_Delete_ReturnsServerError(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, errors.New("db down")
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodDelete, "/organizations/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOrganizationHandler_Delete_ReturnsNotFound(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, nil
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodDelete, "/organizations/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOrganizationHandler_Delete_ReturnsConflict(t *testing.T) {
	orgs := dao.CRUDMock[reference.Organization]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return sampleOrganization(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("referenced")
		},
	}
	svc := organizationTestSvc(orgs, handlerCurrencySearchMock{}, handlerMemberProvisionerMock{})
	app := organizationHandlerTest(t, orgs, svc)

	resp, err := doRequest(app, http.MethodDelete, "/organizations/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}

func doRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}
