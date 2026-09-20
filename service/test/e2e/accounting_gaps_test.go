//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestJournalEntryPostAndReverseE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	moveID := postBalancedMove(t, e, f, 1000000)

	reversed := f.authed(t, http.MethodPost, "/account-movements/"+itoa(moveID)+"/reverse").
		WithJSON(map[string]any{
			"journal_id":  f.journalGeneralID,
			"description": "E2E reversal",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("movement").
		Object()

	reversed.NotEmpty()

	f.authed(t, http.MethodGet, "/account-movements").
		Expect().
		Status(http.StatusOK)
}

func TestManualReconciliationE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	firstMove := postBalancedMove(t, e, f, 500000)
	secondMove := postBalancedMove(t, e, f, 500000)

	debitLine, _ := moveLineIDs(t, e, f, firstMove)
	_, creditLine := moveLineIDs(t, e, f, secondMove)

	f.authed(t, http.MethodPost, "/reconciliations").
		WithJSON(map[string]any{
			"debit_line_id":  debitLine,
			"credit_line_id": creditLine,
			"amount":         500000,
		}).
		Expect().
		Status(http.StatusCreated)
}

func TestInvoiceWriteOffAndCreditNoteE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	invoiceID := createDraftInvoice(t, e, f)

	f.authed(t, http.MethodPost, "/invoices/"+itoa(invoiceID)+"/post").
		Expect().
		Status(http.StatusOK)

	expenseAccount, _ := listFirstTwoIDs(t, e, f.orgPath, "/accounts", f.tokens.AccessToken, "accounts")

	f.authed(t, http.MethodPost, "/invoices/"+itoa(invoiceID)+"/write-off").
		WithJSON(map[string]any{"expense_account_id": expenseAccount}).
		Expect().
		Status(http.StatusOK)

	secondInvoiceID := createDraftInvoice(t, e, f)

	f.authed(t, http.MethodPost, "/invoices/"+itoa(secondInvoiceID)+"/post").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/invoices/"+itoa(secondInvoiceID)+"/credit-note").
		WithJSON(map[string]any{"journal_id": f.journalSaleID}).
		Expect().
		Status(http.StatusCreated)
}

func TestSupplierBillPostAndOutboundPayE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	billID := createPostedSupplierBill(t, e, f)

	f.authed(t, http.MethodGet, "/invoices/"+itoa(billID)).
		Expect().
		Status(http.StatusOK)

	payment := f.authed(t, http.MethodPost, "/payments/outbound").
		WithJSON(map[string]any{
			"contact_id":    f.contactID,
			"journal_id":    f.journalBankID,
			"amount":        1332000,
			"invoice_ids":   []uint64{billID},
			"allow_advance": true,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("payment").
		Object()

	payment.NotEmpty()
}

func TestTaxPeriodLockAndTrialBalanceE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	periodID := createTaxPeriod(t, e, f)

	f.authed(t, http.MethodPost, "/tax-periods/"+itoa(periodID)+"/lock").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/tax-periods/"+itoa(periodID)+"/trial-balance").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/tax-periods/"+itoa(periodID)+"/open").
		Expect().
		Status(http.StatusOK)
}

func TestReminderGenerateE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	invoiceID := createDraftInvoice(t, e, f)

	f.authed(t, http.MethodPost, "/invoices/"+itoa(invoiceID)+"/post").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/reminder/actions/generate").
		WithJSON(map[string]any{}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/reminder/actions").
		Expect().
		Status(http.StatusOK)
}

func TestTaxRuleResolveE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	accountID := firstAccountID(t, e, f)

	position := f.authed(t, http.MethodPost, "/tax-rules").
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("tax_rule").
		Object()

	positionID := uint64(position.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/tax-rules/"+itoa(positionID)+"/resolve").
		WithJSON(map[string]any{"account_id": accountID}).
		Expect().
		Status(http.StatusOK)
}

func TestWithholdingCreateAndApplyE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	accountID := firstAccountID(t, e, f)
	bankAccountID := ensureContactBankAccount(t, e, f)

	wht := f.authed(t, http.MethodPost, "/withholding-taxes").
		WithJSON(map[string]any{
			"name":       "E2E " + gofakeit.Word(),
			"rate_pct":   2,
			"account_id": accountID,
			"scope":      "purchase",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("withholding_tax").
		Object()

	whtID := uint64(wht.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/withholding-taxes/apply").
		WithJSON(map[string]any{
			"journal_id":         f.journalGeneralID,
			"contact_id":         f.contactID,
			"amount":             1000000,
			"scope":              "purchase",
			"withholding_tax_id": whtID,
			"bank_account_id":    bankAccountID,
			"payable_account_id": accountID,
		}).
		Expect().
		Status(http.StatusCreated)
}

func TestTaxReturnFilePayOpenE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	periodID := createTaxPeriod(t, e, f)

	taxReturn := f.authed(t, http.MethodPost, "/tax-returns").
		WithJSON(map[string]any{"period_id": periodID}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("tax_return").
		Object()

	taxReturnID := uint64(taxReturn.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/tax-returns/"+itoa(taxReturnID)+"/file").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/tax-returns/"+itoa(taxReturnID)+"/pay").
		Expect().
		Status(http.StatusOK)
}

func TestDeferralScheduleAndRecognizeE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	invoiceID := createDraftInvoice(t, e, f)

	f.authed(t, http.MethodPost, "/invoices/"+itoa(invoiceID)+"/post").
		Expect().
		Status(http.StatusOK)

	debitAccount, creditAccount := listFirstTwoIDs(t, e, f.orgPath, "/accounts", f.tokens.AccessToken, "accounts")

	schedule := f.authed(t, http.MethodPost, "/deferrals").
		WithJSON(map[string]any{
			"type":                     "deferred_revenue",
			"source_type":              "invoice",
			"source_id":                invoiceID,
			"total_amount":             1000000,
			"balance_sheet_account_id": debitAccount,
			"pl_account_id":            creditAccount,
			"method":                   "linear",
			"date_start":               "2026-01-01",
			"periods":                  3,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	scheduleID := uint64(schedule.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/deferrals/"+itoa(scheduleID)+"/lines").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/deferrals/recognize").
		WithJSON(map[string]any{"as_of": "2026-06-30"}).
		Expect().
		Status(http.StatusOK)
}

func TestReconcileRulesLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	accountID := firstAccountID(t, e, f)

	rule := f.authed(t, http.MethodPost, "/reconcile-rules").
		WithJSON(map[string]any{
			"name":          "E2E " + gofakeit.Word(),
			"account_id":    accountID,
			"match_contact": true,
			"match_amount":  true,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("rule").
		Object()

	_ = uint64(rule.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/reconcile-rules/suggest").
		WithJSON(map[string]any{}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/reconcile-rules").
		Expect().
		Status(http.StatusOK)
}

func TestPaymentBatchSepaE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	billID := createPostedSupplierBill(t, e, f)

	payment := f.authed(t, http.MethodPost, "/payments/outbound").
		WithJSON(map[string]any{
			"contact_id":    f.contactID,
			"journal_id":    f.journalBankID,
			"amount":        1332000,
			"invoice_ids":   []uint64{billID},
			"allow_advance": true,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("payment").
		Object()

	paymentID := uint64(payment.Value("id").Number().Raw())

	batch := f.authed(t, http.MethodPost, "/payment-batches").
		WithJSON(map[string]any{
			"journal_id":  f.journalBankID,
			"name":        "E2E " + gofakeit.Word(),
			"payment_ids": []uint64{paymentID},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("batch").
		Object()

	batchID := uint64(batch.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/payment-batches/"+itoa(batchID)+"/generate-sepa").
		WithJSON(map[string]any{
			"journal_id":  f.journalBankID,
			"name":        "E2E " + gofakeit.Word(),
			"payment_ids": []uint64{paymentID},
		}).
		Expect().
		Status(http.StatusOK)
}
