package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/commission"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, uint64(1))
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

func commissionHandlerTest(
	t *testing.T,
	plans commission.CommissionPlanDAOMock,
	rules commission.CommissionRuleDAOMock,
	assigns commission.CommissionAssignmentDAOMock,
	entries commission.CommissionEntryDAOMock,
	svc commission.CommissionService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewCommissionHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func commissionHandlerTestWithTenant(
	t *testing.T,
	plans commission.CommissionPlanDAOMock,
	rules commission.CommissionRuleDAOMock,
	assigns commission.CommissionAssignmentDAOMock,
	entries commission.CommissionEntryDAOMock,
	svc commission.CommissionService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(1))
		return c.Next()
	})
	h := NewCommissionHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func commissionTestSvc(
	plans commission.CommissionPlanDAOMock,
	rules commission.CommissionRuleDAOMock,
	assigns commission.CommissionAssignmentDAOMock,
	entries commission.CommissionEntryDAOMock,
) commission.CommissionService {
	return commission.NewCommissionService(
		plans, rules, assigns, entries,
		commissionPosterMock{},
		accounting.InvoiceDAOMock{},
		accounting.InvoiceLineDAOMock{},
		accounting.PaymentAllocationDAOMock{},
		inventory.ItemResolverMock{},
		commissionTxMock{},
	)
}

type commissionPosterMock struct{}

func (commissionPosterMock) Post(context.Context, accounting.PostRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

func (commissionPosterMock) PostTx(context.Context, *gorm.DB, accounting.PostRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

func (commissionPosterMock) Reverse(context.Context, accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

func (commissionPosterMock) ReverseTx(context.Context, *gorm.DB, accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

type commissionTxMock struct{}

func (commissionTxMock) Run(_ context.Context, fn func(tx *gorm.DB) error) error {
	return fn(nil)
}

func commissionPlanFixture() *commission.CommissionPlan {
	return &commission.CommissionPlan{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(1)),
		Name:           "Sales Plan",
		Basis:          commission.BasisRevenue,
		Active:         true,
	}
}

func commissionRuleFixture() *commission.CommissionRule {
	return &commission.CommissionRule{Base: model.Base{ID: 1}, PlanID: 1, RatePct: 5}
}

func commissionAssignmentFixture() *commission.CommissionAssignment {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return &commission.CommissionAssignment{
		Base: model.Base{ID: 1}, PlanID: 1, SalespersonID: 5, DateStart: &start,
	}
}

func commissionEntryFixture() *commission.CommissionEntry {
	return &commission.CommissionEntry{
		Base: model.Base{ID: 1}, SalespersonID: 5, PlanID: 1,
		SourceType: "invoice", SourceID: 7, BaseAmount: 1000, CommissionAmount: 50,
		State: commission.EntryStateConfirmed,
	}
}

func TestCommissionHandler_ListPlans_ReturnsPlans(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionPlan], error) {
		return &query.Page[commission.CommissionPlan]{Items: []*commission.CommissionPlan{commissionPlanFixture()}, Count: 1}, nil
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestCommissionHandler_ListPlans_RejectsInvalidQuery(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestCommissionHandler_ListPlans_ReturnsServerError(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionPlan], error) {
		return nil, errors.New("db down")
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestCommissionHandler_GetPlan_ReturnsPlan(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestCommissionHandler_GetPlan_ReturnsNotFound(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return nil, nil
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestCommissionHandler_GetPlan_ReturnsNotFoundForForeignTenant(t *testing.T) {
	foreign := commissionPlanFixture()
	foreign.OrganizationID = helper.Ptr(uint64(2))
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return foreign, nil
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTestWithTenant(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestCommissionHandler_GetPlan_ReturnsServerError(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return nil, errors.New("db down")
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestCommissionHandler_CreatePlan_CreatesPlan(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{CreateFunc: func(_ context.Context, plan *commission.CommissionPlan) (*commission.CommissionPlan, error) {
		plan.ID = 9
		return plan, nil
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTestWithTenant(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-plans/", `{"name":"Sales Plan","basis":"revenue"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestCommissionHandler_CreatePlan_RejectsInvalidBody(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-plans/", `{"basis":"bogus"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestCommissionHandler_CreatePlan_PropagatesServiceError(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{CreateFunc: func(_ context.Context, _ *commission.CommissionPlan) (*commission.CommissionPlan, error) {
		return nil, commission.ErrPlanInvalidBasis
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-plans/", `{"name":"Sales Plan","basis":"revenue"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestCommissionHandler_UpdatePlan_UpdatesPlan(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
			return commissionPlanFixture(), nil
		},
		UpdateFunc: func(_ context.Context, plan *commission.CommissionPlan) (*commission.CommissionPlan, error) {
			return plan, nil
		},
	}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPut, "/commission-plans/1", `{"active":false}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestCommissionHandler_UpdatePlan_ReturnsNotFound(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return nil, nil
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPut, "/commission-plans/1", `{"active":false}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestCommissionHandler_ListRules_ReturnsRules(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	rules := commission.CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*commission.CommissionRule, error) {
		return []*commission.CommissionRule{commissionRuleFixture()}, nil
	}}
	svc := commissionTestSvc(plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/rules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestCommissionHandler_ListRules_ReturnsNotFound(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return nil, nil
	}}
	rules := commission.CommissionRuleDAOMock{}
	svc := commissionTestSvc(plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/rules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestCommissionHandler_ListRules_ReturnsServerError(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	rules := commission.CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*commission.CommissionRule, error) {
		return nil, errors.New("db down")
	}}
	svc := commissionTestSvc(plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/rules", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestCommissionHandler_CreateRule_CreatesRule(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	rules := commission.CommissionRuleDAOMock{CreateFunc: func(_ context.Context, rule *commission.CommissionRule) (*commission.CommissionRule, error) {
		rule.ID = 9
		return rule, nil
	}}
	svc := commissionTestSvc(plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/rules", `{"rate_pct":5}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestCommissionHandler_CreateRule_ReturnsNotFound(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return nil, nil
	}}
	rules := commission.CommissionRuleDAOMock{}
	svc := commissionTestSvc(plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/rules", `{"rate_pct":5}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestCommissionHandler_CreateRule_PropagatesServiceError(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	rules := commission.CommissionRuleDAOMock{CreateFunc: func(_ context.Context, _ *commission.CommissionRule) (*commission.CommissionRule, error) {
		return nil, commission.ErrRuleInvalid
	}}
	svc := commissionTestSvc(plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/rules", `{"rate_pct":5}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestCommissionHandler_ListAssignments_ReturnsAssignments(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	assigns := commission.CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionAssignment], error) {
		return &query.Page[commission.CommissionAssignment]{Items: []*commission.CommissionAssignment{commissionAssignmentFixture()}}, nil
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/assignments", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestCommissionHandler_ListAssignments_ReturnsServerError(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	assigns := commission.CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionAssignment], error) {
		return nil, errors.New("db down")
	}}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/assignments", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestCommissionHandler_CreateAssignment_CreatesAssignment(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	assigns := commission.CommissionAssignmentDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionAssignment], error) {
			return &query.Page[commission.CommissionAssignment]{Items: []*commission.CommissionAssignment{}}, nil
		},
		CreateFunc: func(_ context.Context, assignment *commission.CommissionAssignment) (*commission.CommissionAssignment, error) {
			assignment.ID = 9
			return assignment, nil
		},
	}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/assignments", `{"salesperson_id":5,"date_start":"2026-01-01"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestCommissionHandler_CreateAssignment_RejectsInvalidDate(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	assigns := commission.CommissionAssignmentDAOMock{}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{})
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/assignments", `{"salesperson_id":5,"date_start":"bogus"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestCommissionHandler_ListEntries_ReturnsEntries(t *testing.T) {
	entries := commission.CommissionEntryDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionEntry], error) {
		return &query.Page[commission.CommissionEntry]{Items: []*commission.CommissionEntry{commissionEntryFixture()}, Count: 1}, nil
	}}
	svc := commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries)
	app := commissionHandlerTest(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-entries/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestCommissionHandler_ListEntries_ReturnsServerError(t *testing.T) {
	entries := commission.CommissionEntryDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionEntry], error) {
		return nil, errors.New("db down")
	}}
	svc := commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries)
	app := commissionHandlerTest(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries, svc)

	resp, err := doRequest(app, http.MethodGet, "/commission-entries/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestCommissionHandler_Accrue_AccruesEntry(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	rules := commission.CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*commission.CommissionRule, error) {
		return []*commission.CommissionRule{commissionRuleFixture()}, nil
	}}
	assigns := commission.CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionAssignment], error) {
		return &query.Page[commission.CommissionAssignment]{Items: []*commission.CommissionAssignment{commissionAssignmentFixture()}}, nil
	}}
	entries := commission.CommissionEntryDAOMock{CreateTxFunc: func(_ context.Context, _ *gorm.DB, entry *commission.CommissionEntry) (*commission.CommissionEntry, error) {
		entry.ID = 9
		return entry, nil
	}}
	svc := commissionTestSvc(plans, rules, assigns, entries)
	app := commissionHandlerTest(t, plans, rules, assigns, entries, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-entries/accrue", `{
		"salesperson_id":5,"plan_id":1,"source_type":"invoice","source_id":7,"base_amount":1000,
		"journal_id":3,"expense_account_id":200,"payable_account_id":300,"date":"2026-02-01"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestCommissionHandler_Accrue_RejectsInvalidDate(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{}
	entries := commission.CommissionEntryDAOMock{}
	svc := commissionTestSvc(plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries)
	app := commissionHandlerTest(t, plans, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-entries/accrue", `{
		"salesperson_id":5,"plan_id":1,"source_type":"invoice","source_id":7,"base_amount":1000,
		"journal_id":3,"expense_account_id":200,"payable_account_id":300,"date":"bogus"
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestCommissionHandler_AccrueFromInvoice_AccruesEntry(t *testing.T) {
	plans := commission.CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
		return commissionPlanFixture(), nil
	}}
	rules := commission.CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*commission.CommissionRule, error) {
		return []*commission.CommissionRule{commissionRuleFixture()}, nil
	}}
	assigns := commission.CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionAssignment], error) {
		return &query.Page[commission.CommissionAssignment]{Items: []*commission.CommissionAssignment{commissionAssignmentFixture()}}, nil
	}}
	entries := commission.CommissionEntryDAOMock{CreateTxFunc: func(_ context.Context, _ *gorm.DB, entry *commission.CommissionEntry) (*commission.CommissionEntry, error) {
		entry.ID = 9
		return entry, nil
	}}
	invoices := accounting.InvoiceDAOMock{CRUDMock: dao.CRUDMock[accounting.Invoice]{FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
		return &accounting.Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1)), AmountUntaxed: amount.FromFloat64(1000)}, nil
	}}}
	svc := commission.NewCommissionService(
		plans, rules, assigns, entries,
		commissionPosterMock{}, invoices, accounting.InvoiceLineDAOMock{}, accounting.PaymentAllocationDAOMock{},
		inventory.ItemResolverMock{}, commissionTxMock{},
	)
	app := commissionHandlerTest(t, plans, rules, assigns, entries, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-entries/accrue-from-invoice", `{
		"invoice_id":7,"salesperson_id":5,"journal_id":3,"expense_account_id":200,"payable_account_id":300
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestCommissionHandler_Pay_PaysEntry(t *testing.T) {
	entries := commission.CommissionEntryDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionEntry, error) {
			return commissionEntryFixture(), nil
		},
		UpdateFunc: func(_ context.Context, entry *commission.CommissionEntry) (*commission.CommissionEntry, error) {
			return entry, nil
		},
	}
	svc := commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries)
	app := commissionHandlerTest(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-entries/1/pay", `{"payslip_id":77}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestCommissionHandler_Pay_PropagatesServiceError(t *testing.T) {
	entries := commission.CommissionEntryDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionEntry, error) {
			return &commission.CommissionEntry{State: commission.EntryStateDraft}, nil
		},
	}
	svc := commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries)
	app := commissionHandlerTest(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-entries/1/pay", `{"payslip_id":77}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestCommissionHandler_Cancel_CancelsEntry(t *testing.T) {
	entries := commission.CommissionEntryDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionEntry, error) {
			return &commission.CommissionEntry{Base: model.Base{ID: 1}, State: commission.EntryStateDraft}, nil
		},
		UpdateFunc: func(_ context.Context, entry *commission.CommissionEntry) (*commission.CommissionEntry, error) {
			return entry, nil
		},
	}
	svc := commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries)
	app := commissionHandlerTest(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-entries/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestCommissionHandler_Cancel_PropagatesServiceError(t *testing.T) {
	entries := commission.CommissionEntryDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionEntry, error) {
			return &commission.CommissionEntry{Base: model.Base{ID: 1}, State: commission.EntryStatePaid}, nil
		},
	}
	svc := commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries)
	app := commissionHandlerTest(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, entries, svc)

	resp, err := doRequest(app, http.MethodPost, "/commission-entries/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}
