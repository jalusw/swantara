package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func sampleSalaryRule() *reference.SalaryRule {
	organizationID := uint64(1)
	category := "earning"
	computeType := "fixed"
	amount := 5000000.0
	debit := uint64(1)
	credit := uint64(2)
	return &reference.SalaryRule{
		Base:            model.Base{ID: 5, CreatedAt: timeNow(), UpdatedAt: timeNow()},
		OrganizationID:  &organizationID,
		Code:            "BASIC",
		Name:            "Basic Salary",
		Category:        &category,
		ComputeType:     &computeType,
		Amount:          &amount,
		AccountDebitID:  &debit,
		AccountCreditID: &credit,
	}
}

func salaryRuleHandlerApp(t *testing.T, svc payroll.SalaryRuleService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewSalaryRuleHandler(svc)
	handler.Register(app, guards)
	return app
}

func salaryRuleHandlerWithAccounts(accounts dao.CRUD[reference.Account]) (dao.CRUDMock[reference.SalaryRule], payroll.SalaryRuleService) {
	rules := dao.CRUDMock[reference.SalaryRule]{}
	return rules, payroll.NewSalaryRuleService(rules, accounts)
}

func TestSalaryRuleHandler_List_ReturnsRules(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
			return &query.Page[reference.SalaryRule]{
				Items: []*reference.SalaryRule{sampleSalaryRule()},
				Count: 1,
			}, nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/salary-rules/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_List_RejectsInvalidQuery(t *testing.T) {
	_, svc := salaryRuleHandlerWithAccounts(dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/salary-rules/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_List_RejectsMissingTenant(t *testing.T) {
	_, svc := salaryRuleHandlerWithAccounts(dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/salary-rules/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_List_ReturnsServerError(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalaryRule], error) {
			return nil, errors.New("db down")
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/salary-rules/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Get_ReturnsRule(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return sampleSalaryRule(), nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/salary-rules/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Get_RejectsInvalidID(t *testing.T) {
	_, svc := salaryRuleHandlerWithAccounts(dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/salary-rules/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Get_ReturnsNotFound(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return nil, nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/salary-rules/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Get_ReturnsNotFoundForOtherTenant(t *testing.T) {
	organizationID := uint64(2)
	rule := sampleSalaryRule()
	rule.OrganizationID = &organizationID
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return rule, nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/salary-rules/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Get_ReturnsServerError(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return nil, errors.New("db down")
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/salary-rules/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Create_RejectsValidation(t *testing.T) {
	_, svc := salaryRuleHandlerWithAccounts(dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	body := `{"code":"","name":"","category":"","compute_type":"","account_debit_id":0,"account_credit_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/salary-rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Create_RejectsMissingTenant(t *testing.T) {
	_, svc := salaryRuleHandlerWithAccounts(dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, noTenantGuards())

	body := `{"code":"BASIC","name":"Basic Salary","category":"earning","compute_type":"fixed","account_debit_id":1,"account_credit_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/salary-rules", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Update_UpdatesRule(t *testing.T) {
	accounts := dao.CRUDMock[reference.Account]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Account, error) {
			return &reference.Account{Base: model.Base{ID: 1}}, nil
		},
	}
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return sampleSalaryRule(), nil
		},
		UpdateFunc: func(_ context.Context, rule *reference.SalaryRule) (*reference.SalaryRule, error) {
			return rule, nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, accounts)
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	body := `{"code":"BASIC","name":"Basic Salary","category":"earning","compute_type":"fixed","amount":5000000,"account_debit_id":1,"account_credit_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/salary-rules/5", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Update_RejectsInvalidID(t *testing.T) {
	_, svc := salaryRuleHandlerWithAccounts(dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	body := `{"code":"BASIC","name":"Basic Salary","category":"earning","compute_type":"fixed","account_debit_id":1,"account_credit_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/salary-rules/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Update_ReturnsNotFound(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return nil, nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	body := `{"code":"BASIC","name":"Basic Salary","category":"earning","compute_type":"fixed","account_debit_id":1,"account_credit_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/salary-rules/5", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Update_ReturnsNotFoundForOtherTenant(t *testing.T) {
	organizationID := uint64(2)
	rule := sampleSalaryRule()
	rule.OrganizationID = &organizationID
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return rule, nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	body := `{"code":"BASIC","name":"Basic Salary","category":"earning","compute_type":"fixed","account_debit_id":1,"account_credit_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/salary-rules/5", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Update_RejectsValidation(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return sampleSalaryRule(), nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	body := `{"code":"","name":"","category":"","compute_type":"","account_debit_id":0,"account_credit_id":0}`
	resp, err := doRequest(app, http.MethodPut, "/salary-rules/5", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Update_ReturnsServerError(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return nil, errors.New("db down")
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	body := `{"code":"BASIC","name":"Basic Salary","category":"earning","compute_type":"fixed","account_debit_id":1,"account_credit_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/salary-rules/5", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	accounts := dao.CRUDMock[reference.Account]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Account, error) {
			return &reference.Account{Base: model.Base{ID: 1}}, nil
		},
	}
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return sampleSalaryRule(), nil
		},
		UpdateFunc: func(_ context.Context, _ *reference.SalaryRule) (*reference.SalaryRule, error) {
			return nil, errors.New("db down")
		},
	}
	svc := payroll.NewSalaryRuleService(rules, accounts)
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	body := `{"code":"BASIC","name":"Basic Salary","category":"earning","compute_type":"fixed","amount":5000000,"account_debit_id":1,"account_credit_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/salary-rules/5", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Delete_DeletesRule(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return sampleSalaryRule(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/salary-rules/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Delete_RejectsInvalidID(t *testing.T) {
	_, svc := salaryRuleHandlerWithAccounts(dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/salary-rules/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Delete_ReturnsNotFound(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return nil, nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/salary-rules/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Delete_ReturnsNotFoundForOtherTenant(t *testing.T) {
	organizationID := uint64(2)
	rule := sampleSalaryRule()
	rule.OrganizationID = &organizationID
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return rule, nil
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/salary-rules/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return nil, errors.New("db down")
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/salary-rules/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSalaryRuleHandler_Delete_ReturnsServerError(t *testing.T) {
	rules := dao.CRUDMock[reference.SalaryRule]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SalaryRule, error) {
			return sampleSalaryRule(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		},
	}
	svc := payroll.NewSalaryRuleService(rules, dao.CRUDMock[reference.Account]{})
	app := salaryRuleHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/salary-rules/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
