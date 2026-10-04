//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestReferenceCatalogGapsE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	carrier := e.POST("/api/v1/carriers/").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("carrier").
		Object()

	carrierID := uint64(carrier.Value("id").Number().Raw())

	e.GET("/api/v1/carriers/").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		Expect().
		Status(http.StatusOK)

	e.GET("/api/v1/carriers/"+itoa(carrierID)).
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		Expect().
		Status(http.StatusOK)

	e.PUT("/api/v1/carriers/"+itoa(carrierID)).
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusOK)

	group := e.POST("/api/v1/unit-groups/").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("category").
		Object()

	groupID := uint64(group.Value("id").Number().Raw())

	e.GET("/api/v1/unit-groups/").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		Expect().
		Status(http.StatusOK)

	unitName := "E2E-U-" + gofakeit.Word()
	unit := e.POST("/api/v1/units/").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"category_id": groupID,
			"name":        unitName,
			"factor":      1,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("unit").
		Object()

	unitID := uint64(unit.Value("id").Number().Raw())

	e.GET("/api/v1/units/").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		Expect().
		Status(http.StatusOK)

	e.POST("/api/v1/units/convert").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"from_id": unitID,
			"to_id":   unitID,
			"qty":     "2",
		}).
		Expect().
		Status(http.StatusOK)

	term := f.authed(t, http.MethodPost, "/payment-terms/").
		WithJSON(map[string]any{
			"name": "E2E " + gofakeit.Word(),
			"lines": []map[string]any{
				{"sequence": 1, "value_type": "percent", "value": 100, "days_after": 30},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("payment_term").
		Object()

	termID := uint64(term.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/payment-terms/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/payment-terms/"+itoa(termID)).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/payment-terms/"+itoa(termID)+"/splits").
		WithJSON(map[string]any{"total": "1000.00", "date": "2026-02-15"}).
		Expect().
		Status(http.StatusOK)

	dim := f.authed(t, http.MethodPost, "/dimensions/").
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("account").
		Object()

	dimID := uint64(dim.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/dimensions/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/dimensions/"+itoa(dimID)).
		Expect().
		Status(http.StatusOK)
}

func TestPayrollCatalogGapsE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	position := f.authed(t, http.MethodPost, "/job-positions/").
		WithJSON(map[string]any{"name": "E2E " + gofakeit.JobTitle()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	positionID := uint64(position.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/job-positions/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/job-positions/"+itoa(positionID)).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPut, "/job-positions/"+itoa(positionID)).
		WithJSON(map[string]any{"name": "E2E " + gofakeit.JobTitle()}).
		Expect().
		Status(http.StatusOK)

	debitAccount, creditAccount := listFirstTwoIDs(t, e, f.orgPath, "/accounts", f.tokens.AccessToken, "accounts")

	rule := f.authed(t, http.MethodPost, "/salary-rules/").
		WithJSON(map[string]any{
			"code":              "E2E-" + gofakeit.Word(),
			"name":              "E2E " + gofakeit.Word(),
			"category":          "earning",
			"compute_type":      "fixed",
			"amount":            1000000,
			"account_debit_id":  debitAccount,
			"account_credit_id": creditAccount,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	ruleID := uint64(rule.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/salary-rules/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/salary-rules/"+itoa(ruleID)).
		Expect().
		Status(http.StatusOK)
}

func TestServiceMaintenanceAndPOSGapsE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	equipmentID := createEquipment(t, e, f)

	plan := f.authed(t, http.MethodPost, "/maintenance-plans/").
		WithJSON(map[string]any{
			"equipment_id":  equipmentID,
			"name":          "E2E " + gofakeit.Word(),
			"interval_days": 30,
			"next_due":      "2026-01-01",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("maintenance_plan").
		Object()

	planID := uint64(plan.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/maintenance-plans/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/maintenance-plans/"+itoa(planID)).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/maintenance-plans/generate-orders").
		WithJSON(map[string]any{}).
		Expect().
		Status(http.StatusOK)

	accountID := firstAccountID(t, e, f)

	method := "e2e-" + gofakeit.UUID()

	f.authed(t, http.MethodPut, "/pos/payment-accounts/"+method).
		WithJSON(map[string]any{"account_id": accountID}).
		Expect().
		Status(http.StatusCreated)

	f.authed(t, http.MethodGet, "/pos/payment-accounts/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("payment_accounts").
		Array()

	f.authed(t, http.MethodDelete, "/pos/payment-accounts/"+method).
		Expect().
		Status(http.StatusNoContent)
}

func TestProcurementCatalogGapsE2E(t *testing.T) {
	f := newFlowFixture(t)

	rate := f.authed(t, http.MethodPost, "/currency-rates/").
		WithJSON(map[string]any{
			"from_currency": "USD",
			"to_currency":   "IDR",
			"rate":          16000,
			"rate_date":     "2026-06-01",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("rate").
		Object()

	rateID := uint64(rate.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/currency-rates/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/currency-rates/"+itoa(rateID)).
		Expect().
		Status(http.StatusOK)

	center := f.authed(t, http.MethodPost, "/cost-centers/").
		WithJSON(map[string]any{
			"name": "E2E " + gofakeit.Word(),
			"code": "CC-" + gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("cost_center").
		Object()

	centerID := uint64(center.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/cost-centers/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/cost-centers/"+itoa(centerID)).
		Expect().
		Status(http.StatusOK)
}

func TestAuditAndIntegrationEventsGapsE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	_ = e

	logs := f.authed(t, http.MethodGet, "/audit-logs/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("audit_logs").
		Array()

	if logs.Length().Raw() > 0 {
		firstID := uint64(logs.Element(0).Object().Value("id").Number().Raw())
		f.authed(t, http.MethodGet, "/audit-logs/"+itoa(firstID)).
			Expect().
			Status(http.StatusOK)
	}

	events := f.authed(t, http.MethodGet, "/integration-events/").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("integration_events").
		Array()

	if events.Length().Raw() > 0 {
		firstID := uint64(events.Element(0).Object().Value("id").Number().Raw())
		f.authed(t, http.MethodGet, "/integration-events/"+itoa(firstID)).
			Expect().
			Status(http.StatusOK)
	}
}
