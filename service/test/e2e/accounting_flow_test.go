//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestInvoicePostAndPayLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	invoiceID := createDraftInvoice(t, e, f)

	f.authed(t, http.MethodPost, "/invoices/"+itoa(invoiceID)+"/post").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("invoice").
		Object().
		Value("state").
		String().
		Equal("posted")

	payment := f.authed(t, http.MethodPost, "/payments").
		WithJSON(map[string]any{
			"contact_id":    f.contactID,
			"journal_id":    f.journalBankID,
			"amount":        1110000,
			"invoice_ids":   []uint64{invoiceID},
			"allow_advance": true,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	payment.NotEmpty()
}

func TestTaxPeriodLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	period := f.authed(t, http.MethodPost, "/tax-periods").
		WithJSON(map[string]any{
			"name":       "E2E " + gofakeit.Word(),
			"date_start": "2026-01-01",
			"date_end":   "2026-12-31",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("period").
		Object()

	periodID := uint64(period.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/tax-periods/"+itoa(periodID)+"/close").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/tax-periods/"+itoa(periodID)+"/open").
		Expect().
		Status(http.StatusOK)
}

func TestBudgetLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	accountID := firstAccountID(t, newExpect(t), f)

	budget := f.authed(t, http.MethodPost, "/budgets").
		WithJSON(map[string]any{
			"name":       "E2E " + gofakeit.Word(),
			"date_start": "2026-01-01",
			"date_end":   "2026-12-31",
			"lines": []map[string]any{
				{"account_id": accountID, "amount": 10000000},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("budget").
		Object()

	budgetID := uint64(budget.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/budgets/"+itoa(budgetID)+"/variance").
		Expect().
		Status(http.StatusOK)
}

func TestBankStatementLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	statement := f.authed(t, http.MethodPost, "/bank-statements").
		WithJSON(map[string]any{
			"journal_id":    f.journalBankID,
			"balance_start": 0,
			"balance_end":   5000000,
			"lines": []map[string]any{
				{"description": "E2E deposit", "amount": 5000000},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("bank_statement").
		Object()

	statementID := uint64(statement.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/bank-statements/"+itoa(statementID)+"/match").
		Expect().
		Status(http.StatusOK)
}

func TestAccountingReportsE2E(t *testing.T) {
	f := newFlowFixture(t)

	for _, path := range []string{
		"/reports/trial-balance",
		"/reports/cash-flow",
		"/reports/integrity",
		"/invoices/aging",
	} {
		f.authed(t, http.MethodGet, path).
			Expect().
			Status(http.StatusOK)
	}
}
