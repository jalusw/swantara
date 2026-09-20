//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestApprovalRequestLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	approval := f.authed(t, http.MethodPost, "/approval-requests").
		WithJSON(map[string]any{
			"entity_type": "sale_order",
			"entity_id":   1,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("approval_request").
		Object()

	approvalID := uint64(approval.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/approval-requests/"+itoa(approvalID)+"/decide").
		WithJSON(map[string]any{"decision": "approve"}).
		Expect().
		Status(http.StatusOK)
}

func TestMessageLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	message := f.authed(t, http.MethodPost, "/messages").
		WithJSON(map[string]any{
			"entity_type": "contact",
			"entity_id":   f.contactID,
			"body":        "E2E " + gofakeit.Sentence(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("message").
		Object()

	_ = uint64(message.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/messages").
		WithQuery("entity_type", "contact").
		WithQuery("entity_id", itoa(f.contactID)).
		Expect().
		Status(http.StatusOK)
}

func TestSystemConfigLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	key := "e2e_" + gofakeit.UUID()

	config := f.authed(t, http.MethodPost, "/system-configs").
		WithJSON(map[string]any{
			"key":   key,
			"value": "42",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("system_config").
		Object()

	configID := uint64(config.Value("id").Number().Raw())

	f.authed(t, http.MethodPut, "/system-configs/"+itoa(configID)).
		WithJSON(map[string]any{
			"key":   key,
			"value": "43",
		}).
		Expect().
		Status(http.StatusOK)
}

func TestReportingKPIsE2E(t *testing.T) {
	f := newFlowFixture(t)

	for _, path := range []string{
		"/kpis/sales",
		"/kpis/pipeline",
		"/kpis/inventory",
		"/kpis/finance",
		"/kpis/procurement",
		"/kpis/manufacturing",
	} {
		f.authed(t, http.MethodGet, path).
			Expect().
			Status(http.StatusOK)
	}
}

func TestFxRateLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	rate := f.authed(t, http.MethodPost, "/fx-rates").
		WithJSON(map[string]any{
			"from_currency": "USD",
			"to_currency":   "IDR",
			"rate":          16000,
			"date":          "2026-06-01",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	rate.NotEmpty()

	f.authed(t, http.MethodPost, "/fx-rates/resolve").
		WithJSON(map[string]any{
			"from_currency": "USD",
			"to_currency":   "IDR",
			"date":          "2026-06-01",
		}).
		Expect().
		Status(http.StatusOK)
}
