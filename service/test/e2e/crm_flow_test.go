//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestProspectToWonOpportunityE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	lead := f.authed(t, http.MethodPost, "/crm/leads").
		WithJSON(map[string]any{
			"name":             "Acme " + gofakeit.Word(),
			"email":            gofakeit.Email(),
			"source":           "website",
			"expected_revenue": 11100000,
			"probability":      60,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("lead").
		Object()

	prospectID := uint64(lead.Value("id").Number().Raw())
	lead.Value("type").String().Equal("lead")

	opportunity := f.authed(t, http.MethodPost, "/crm/leads/"+itoa(prospectID)+"/promote").
		WithJSON(map[string]any{
			"stage_id":         f.firstOpportunityStageID(t, e),
			"expected_revenue": 11100000,
			"probability":      60,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("lead").
		Object()

	opportunityID := uint64(opportunity.Value("id").Number().Raw())
	opportunity.Value("type").String().Equal("opportunity")

	f.authed(t, http.MethodGet, "/crm/pipeline").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("forecast").
		Object().
		Value("total_expected_revenue").
		Number().
		Ge(11100000)

	midStage := f.otherStageID(t, e, f.firstOpportunityStageID(t, e))

	f.authed(t, http.MethodPost, "/crm/opportunities/"+itoa(opportunityID)+"/advance-stage").
		WithJSON(map[string]any{"stage_id": midStage}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("opportunity").
		Object().
		Value("stage_id").
		Number().
		Equal(midStage)

	closed := f.authed(t, http.MethodPost, "/crm/opportunities/"+itoa(opportunityID)+"/win").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("opportunity").
		Object()

	closed.Value("is_won").Boolean().IsTrue()
	closed.Value("closed_at").NotNull()

	activity := f.authed(t, http.MethodPost, "/crm/activities").
		WithJSON(map[string]any{
			"lead_id": prospectID,
			"type":    "call",
			"summary": "Follow up on closed deal",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("activity").
		Object()

	activityID := uint64(activity.Value("id").Number().Raw())
	activity.Value("done").Boolean().IsFalse()

	f.authed(t, http.MethodPost, "/crm/activities/"+itoa(activityID)+"/done").
		Expect().
		Status(http.StatusOK)
}
