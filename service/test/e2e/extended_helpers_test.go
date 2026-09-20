//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
)

func sscanfOrgPath(orgPath string, orgID *uint64) (int, error) {
	prefix := "/api/v1/organizations/"
	idPart := strings.TrimPrefix(orgPath, prefix)
	var id uint64
	n, err := fmt.Sscanf(idPart, "%d", &id)
	if err != nil {
		return n, err
	}
	*orgID = id
	return n, nil
}

func authHeader(token string) string {
	return "Bearer " + token
}

func (f flowContext) organizationID() uint64 {
	var orgID uint64
	_, err := sscanfOrgPath(f.orgPath, &orgID)
	if err != nil {
		return 0
	}
	return orgID
}

func defaultUserID(t *testing.T, accessToken string) uint64 {
	t.Helper()

	users := newExpect(t).GET("/api/v1/users").
		WithHeader("Authorization", authHeader(accessToken)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("users").
		Array()

	return uint64(users.Element(0).Object().Value("id").Number().Raw())
}

func listFirstID(t *testing.T, e *httpexpect.Expect, orgPath, listPath, token, arrayKey string) uint64 {
	t.Helper()

	arr := e.GET(orgPath+listPath).
		WithHeader("Authorization", authHeader(token)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value(arrayKey).
		Array()

	if arr.Length().Raw() == 0 {
		t.Fatalf("expected at least one %s", arrayKey)
	}

	return uint64(arr.Element(0).Object().Value("id").Number().Raw())
}

func listFirstTwoIDs(t *testing.T, e *httpexpect.Expect, orgPath, listPath, token, arrayKey string) (uint64, uint64) {
	t.Helper()

	arr := e.GET(orgPath+listPath).
		WithHeader("Authorization", authHeader(token)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value(arrayKey).
		Array()

	if arr.Length().Raw() < 2 {
		t.Fatalf("expected at least two %s", arrayKey)
	}

	return uint64(arr.Element(0).Object().Value("id").Number().Raw()),
		uint64(arr.Element(1).Object().Value("id").Number().Raw())
}

func firstAccountID(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	return listFirstID(t, e, f.orgPath, "/accounts", f.tokens.AccessToken, "accounts")
}

func firstIncomeAccountID(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	accounts := e.GET(f.orgPath+"/accounts").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("accounts").
		Array()

	for _, item := range accounts.Iter() {
		obj := item.Object()
		if obj.Value("type").String().Raw() == "income" {
			return uint64(obj.Value("id").Number().Raw())
		}
	}

	return uint64(accounts.Element(0).Object().Value("id").Number().Raw())
}

func createDepartment(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	dept := e.POST(f.orgPath+"/departments").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{"name": "E2E " + gofakeit.JobTitle()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("department").
		Object()

	return uint64(dept.Value("id").Number().Raw())
}

func createEmployee(t *testing.T, e *httpexpect.Expect, f flowContext, departmentID uint64) uint64 {
	t.Helper()

	body := map[string]any{
		"name":            gofakeit.Name(),
		"employee_number": "EMP-" + gofakeit.UUID(),
		"wage":            8000000,
	}
	if departmentID != 0 {
		body["department_id"] = departmentID
	}

	employee := e.POST(f.orgPath+"/employees").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(body).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("employee").
		Object()

	return uint64(employee.Value("id").Number().Raw())
}

func createLeaveType(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	leaveType := e.POST(f.orgPath+"/leave-types").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("leave_type").
		Object()

	return uint64(leaveType.Value("id").Number().Raw())
}

func createProject(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	project := e.POST(f.orgPath+"/projects").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"name":         "E2E " + gofakeit.Company(),
			"contact_id":   f.contactID,
			"billing_type": "fixed",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("project").
		Object()

	return uint64(project.Value("id").Number().Raw())
}

func createExpenseCategory(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	category := e.POST(f.orgPath+"/expense-categories").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"name":               "E2E " + gofakeit.Word(),
			"expense_account_id": firstAccountID(t, e, f),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("category").
		Object()

	return uint64(category.Value("id").Number().Raw())
}

func createAssetCategory(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	category := e.POST(f.orgPath+"/asset-categories").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"name":          "E2E " + gofakeit.Word(),
			"method":        "linear",
			"method_number": 12,
			"method_period": "month",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("asset_category").
		Object()

	return uint64(category.Value("id").Number().Raw())
}

func createSubscriptionPlan(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	plan := e.POST(f.orgPath+"/subscription-plans").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"name":               "E2E " + gofakeit.Word(),
			"recurring_interval": "month",
			"recurring_count":    1,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("plan").
		Object()

	return uint64(plan.Value("id").Number().Raw())
}

func createEquipment(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	equipment := e.POST(f.orgPath+"/equipments").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("equipment").
		Object()

	return uint64(equipment.Value("id").Number().Raw())
}

func createQualityPoint(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	point := e.POST(f.orgPath+"/quality-points").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"name":    "E2E " + gofakeit.Word(),
			"item_id": f.variantID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("point").
		Object()

	return uint64(point.Value("id").Number().Raw())
}

func createDraftInvoice(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	accountID := firstIncomeAccountID(t, e, f)

	invoice := e.POST(f.orgPath+"/invoices").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"journal_id": f.journalSaleID,
			"contact_id": f.contactID,
			"draft":      true,
			"lines": []map[string]any{
				{
					"description": gofakeit.ProductName(),
					"qty":         2,
					"unit_price":  500000,
					"account_id":  accountID,
					"tax_ids":     []uint64{f.taxOutputID},
				},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("invoice").
		Object()

	return uint64(invoice.Value("id").Number().Raw())
}
