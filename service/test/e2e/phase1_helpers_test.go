//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
)

func postBalancedMove(t *testing.T, e *httpexpect.Expect, f flowContext, amount float64) uint64 {
	t.Helper()

	debitAccount, creditAccount := listFirstTwoIDs(t, e, f.orgPath, "/accounts", f.tokens.AccessToken, "accounts")

	movement := e.POST(f.orgPath+"/account-movements").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"journal_id": f.journalGeneralID,
			"date":       "2026-06-01",
			"lines": []map[string]any{
				{"account_id": debitAccount, "name": "debit line", "debit": amount},
				{"account_id": creditAccount, "name": "credit line", "credit": amount},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("movement").
		Object()

	return uint64(movement.Value("id").Number().Raw())
}

func moveLineIDs(t *testing.T, e *httpexpect.Expect, f flowContext, moveID uint64) (debitLineID, creditLineID uint64) {
	t.Helper()

	lines := e.GET(f.orgPath+"/account-movements/"+itoa(moveID)).
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("lines").
		Array()

	for _, item := range lines.Iter() {
		line := item.Object()
		if line.Value("debit").Number().Raw() > 0 && debitLineID == 0 {
			debitLineID = uint64(line.Value("id").Number().Raw())
		}
		if line.Value("credit").Number().Raw() > 0 && creditLineID == 0 {
			creditLineID = uint64(line.Value("id").Number().Raw())
		}
	}

	if debitLineID == 0 || creditLineID == 0 {
		t.Fatalf("expected debit and credit lines on movement %d", moveID)
	}

	return debitLineID, creditLineID
}

func createPostedSupplierBill(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	accountID := firstIncomeAccountID(t, e, f)

	bill := e.POST(f.orgPath+"/supplier-bills").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"journal_id": f.journalPurchaseID,
			"contact_id": f.contactID,
			"draft":      true,
			"lines": []map[string]any{
				{
					"description": gofakeit.ProductName(),
					"qty":         2,
					"unit_price":  600000,
					"account_id":  accountID,
					"tax_ids":     []uint64{f.taxInputID},
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

	billID := uint64(bill.Value("id").Number().Raw())

	e.POST(f.orgPath+"/supplier-bills/"+itoa(billID)+"/post").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		Expect().
		Status(http.StatusOK)

	return billID
}

func ensureContactBankAccount(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	accounts := e.GET(f.orgPath+"/contacts/"+itoa(f.contactID)+"/bank-accounts").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("bank_accounts").
		Array()

	if accounts.Length().Raw() > 0 {
		return uint64(accounts.Element(0).Object().Value("id").Number().Raw())
	}

	created := e.POST(f.orgPath+"/contacts/"+itoa(f.contactID)+"/bank-accounts").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"account_number": gofakeit.UUID(),
			"bank_name":      "E2E Bank",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("bank_account").
		Object()

	return uint64(created.Value("id").Number().Raw())
}

func createTaxPeriod(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	period := e.POST(f.orgPath+"/tax-periods").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
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

	return uint64(period.Value("id").Number().Raw())
}

func firstShipmentID(t *testing.T, e *httpexpect.Expect, f flowContext) uint64 {
	t.Helper()

	return listFirstID(t, e, f.orgPath, "/stock-shipments", f.tokens.AccessToken, "shipments")
}
