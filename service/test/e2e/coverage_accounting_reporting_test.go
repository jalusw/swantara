//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestJournalEntriesAndLedgerGapsE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	debitAccount, creditAccount := listFirstTwoIDs(t, e, f.orgPath, "/accounts", f.tokens.AccessToken, "accounts")

	entry := f.authed(t, http.MethodPost, "/journal-entries/").
		WithJSON(map[string]any{
			"journal_id":  f.journalGeneralID,
			"date":        "2026-01-15",
			"description": "E2E gaps",
			"lines": []map[string]any{
				{"account_id": debitAccount, "name": "debit", "debit": 100000},
				{"account_id": creditAccount, "name": "credit", "credit": 100000},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("entry").
		Object()

	entryID := uint64(entry.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/journal-entries/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/journal-entries/"+itoa(entryID)).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/journal-entries/"+itoa(entryID)+"/reverse").
		WithJSON(map[string]any{
			"journal_id":  f.journalGeneralID,
			"description": "E2E reverse gaps",
		}).
		Expect().
		Status(http.StatusCreated)

	f.authed(t, http.MethodGet, "/general-ledger/").
		WithQuery("account_id", itoa(debitAccount)).
		WithQuery("from", "2026-01-01").
		WithQuery("to", "2026-12-31").
		Expect().
		Status(http.StatusOK)
}

func TestPDCInstrumentsGapsE2E(t *testing.T) {
	f := newFlowFixture(t)

	instrument := f.authed(t, http.MethodPost, "/pdc-instruments/").
		WithJSON(map[string]any{
			"contact_id": f.contactID,
			"direction":  "inbound",
			"amount":     500000,
			"due_date":   "2026-04-01",
			"journal_id": f.journalBankID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	instrumentID := uint64(instrument.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/pdc-instruments/"+itoa(instrumentID)+"/deposit").
		Expect().
		Status(http.StatusOK)
}

func TestReportingCloseGapsE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	periodID := createTaxPeriod(t, e, f)
	debitAccount, creditAccount := listFirstTwoIDs(t, e, f.orgPath, "/accounts", f.tokens.AccessToken, "accounts")

	accrual := f.authed(t, http.MethodPost, "/accruals/").
		WithJSON(map[string]any{
			"period_id": periodID,
			"name":      "E2E " + gofakeit.Word(),
			"lines": []map[string]any{
				{"account_id": debitAccount, "debit": 100000},
				{"account_id": creditAccount, "credit": 100000},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	accrual.NotEmpty()

	f.authed(t, http.MethodGet, "/accruals/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/accruals/reverse-due").
		WithJSON(map[string]any{"as_of": "2026-03-31"}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/fx-revaluations/").
		WithJSON(map[string]any{"period_id": periodID}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/fx-revaluations/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/period-close/").
		WithJSON(map[string]any{"period_id": periodID}).
		Expect().
		Status(http.StatusOK)
}

func TestInterorgRulesConsolidationGapsE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	rule := f.authed(t, http.MethodPost, "/interorganization-rules/").
		WithJSON(map[string]any{"auto_mirror": true}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("interorganization_rule").
		Object()

	ruleID := uint64(rule.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/interorganization-rules/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPut, "/interorganization-rules/"+itoa(ruleID)).
		WithJSON(map[string]any{"auto_mirror": false}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/interorganization-transactions/").
		Expect().
		Status(http.StatusOK)

	periodID := createTaxPeriod(t, e, f)

	run := f.authed(t, http.MethodPost, "/consolidation-runs/").
		WithJSON(map[string]any{
			"period_id":          periodID,
			"reporting_currency": "USD",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("consolidation_run").
		Object()

	runID := uint64(run.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/consolidation-runs/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/consolidation-runs/"+itoa(runID)).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/consolidation-runs/"+itoa(runID)+"/run").
		WithJSON(map[string]any{}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodDelete, "/interorganization-rules/"+itoa(ruleID)).
		Expect().
		Status(http.StatusNoContent)
}

func TestInventoryNewRoutesGapsE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	f.buildStock(t, e, 20, 600000)

	f.authed(t, http.MethodGet, "/shipments/").
		Expect().
		Status(http.StatusOK)

	hold := f.authed(t, http.MethodPost, "/stock-holds/").
		WithJSON(map[string]any{
			"item_id":     f.variantID,
			"location_id": f.stockLocationID,
			"qty":         "5",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("reservation").
		Object()

	holdID := uint64(hold.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/stock-holds/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodDelete, "/stock-holds/"+itoa(holdID)).
		Expect().
		Status(http.StatusNoContent)

	f.authed(t, http.MethodGet, "/batches/").
		Expect().
		Status(http.StatusOK)

	count := f.authed(t, http.MethodPost, "/stock-counts/").
		WithJSON(map[string]any{
			"name":        "E2E " + gofakeit.Word(),
			"location_id": f.stockLocationID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "counted_qty": 10},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("count").
		Object()

	countID := uint64(count.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/stock-counts/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/stock-counts/"+itoa(countID)).
		Expect().
		Status(http.StatusOK)

	secondWarehouse := e.POST(f.orgPath + "/warehouses").
		WithHeader("Authorization", authHeader(f.tokens.AccessToken)).
		WithJSON(map[string]any{
			"name": "E2E " + gofakeit.Word(),
			"code": "WH-" + gofakeit.UUID(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("warehouse").
		Object()
	dstWarehouseID := uint64(secondWarehouse.Value("id").Number().Raw())

	transitAccount := firstAccountID(t, e, f)

	transfer := f.authed(t, http.MethodPost, "/warehouse-transfers/").
		WithJSON(map[string]any{
			"src_warehouse_id": f.warehouseID,
			"dst_warehouse_id": dstWarehouseID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 4},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("warehouse_transfer").
		Object()

	transferID := uint64(transfer.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/warehouse-transfers/"+itoa(transferID)+"/send").
		WithJSON(map[string]any{
			"journal_id":         f.journalGeneralID,
			"transit_account_id": transitAccount,
		}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/warehouse-transfers/"+itoa(transferID)+"/receive").
		WithJSON(map[string]any{
			"journal_id":         f.journalGeneralID,
			"transit_account_id": transitAccount,
		}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/inbound-costs/").
		Expect().
		Status(http.StatusOK)
}

func TestManufacturingProcurementNewRoutesGapsE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	f.authed(t, http.MethodGet, "/production-orders/").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/purchase-requests/").
		Expect().
		Status(http.StatusOK)

	requesterID := defaultUserID(t, f.tokens.AccessToken)

	requisition := f.authed(t, http.MethodPost, "/purchase-requests/").
		WithJSON(map[string]any{
			"requester_id": requesterID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 5},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("request").
		Object()

	requisitionID := uint64(requisition.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/purchase-requests/"+itoa(requisitionID)+"/confirm").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/purchase-requests/"+itoa(requisitionID)+"/approve").
		Expect().
		Status(http.StatusOK)

	quoteRequest := f.authed(t, http.MethodPost, "/supplier-quote-requests/").
		WithJSON(map[string]any{
			"requester_id": requesterID,
			"supplier_id":  f.contactID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 10},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("quoteRequest").
		Object()

	quoteRequest.NotEmpty()

	agreement := f.authed(t, http.MethodPost, "/supply-agreements/").
		WithJSON(map[string]any{
			"supplier_id": f.contactID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 100, "unit_price": 600000},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("agreement").
		Object()

	agreementID := uint64(agreement.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/supply-agreements/"+itoa(agreementID)+"/activate").
		Expect().
		Status(http.StatusOK)

	_ = e
}
