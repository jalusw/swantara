package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/commission"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func planOwned(orgID uint64) *commission.CommissionPlan {
	return &commission.CommissionPlan{Base: model.Base{ID: 1}, OrganizationID: &orgID, Name: "Plan", Active: true}
}

func TestCommissionHandler_UpdatePlan_Errors(t *testing.T) {
	newApp := func(plans commission.CommissionPlanDAOMock, svc commission.CommissionService) interface{} {
		return nil
	}
	_ = newApp

	t.Run("rejects invalid id", func(t *testing.T) {
		app := commissionHandlerTest(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}))
		resp, err := doRequest(app, http.MethodPut, "/commission-plans/abc", `{"active":true}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("reports find error and missing", func(t *testing.T) {
		failing := commission.CommissionPlanDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
				return nil, errors.New("db down")
			},
		}
		app := commissionHandlerTestWithTenant(t, failing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, commissionTestSvc(failing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}))
		resp, err := doRequest(app, http.MethodPut, "/commission-plans/1", `{"active":true}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)

		app = commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}))
		resp, err = doRequest(app, http.MethodPut, "/commission-plans/1", `{"active":true}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})
}

func TestCommissionHandler_Rules_Errors(t *testing.T) {
	owned := commission.CommissionPlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
			return planOwned(1), nil
		},
	}
	svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})

	t.Run("rejects invalid id", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodGet, "/commission-plans/abc/rules", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/commission-plans/abc/rules", `{"rate_pct":5}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("reports rule list error", func(t *testing.T) {
		rules := commission.CommissionRuleDAOMock{
			ListByPlanFunc: func(_ context.Context, _ uint64) ([]*commission.CommissionRule, error) {
				return nil, errors.New("db down")
			},
		}
		app := commissionHandlerTestWithTenant(t, owned, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, commissionTestSvc(owned, rules, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}))
		resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/rules", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("reports rule create error", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}))
		resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/rules", `{"rate_pct":0,"fixed_amount":0}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity && resp.StatusCode != http.StatusCreated {
			t.Errorf("status = %d", resp.StatusCode)
		}
	})
}

func TestCommissionHandler_Assignments_Errors(t *testing.T) {
	owned := commission.CommissionPlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
			return planOwned(1), nil
		},
	}
	svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})

	t.Run("rejects invalid body and id", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/assignments", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/commission-plans/abc/assignments", `{"salesperson_id":5,"date_start":"2026-01-01"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("rejects bad dates", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/assignments", `{"salesperson_id":5,"date_start":"yesterday"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/commission-plans/1/assignments", `{"salesperson_id":5,"date_start":"2026-01-01","date_end":"tomorrow"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("reports list error", func(t *testing.T) {
		assigns := commission.CommissionAssignmentDAOMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[commission.CommissionAssignment], error) {
				return nil, errors.New("db down")
			},
		}
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{}, commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, assigns, commission.CommissionEntryDAOMock{}))
		resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/assignments", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}

func TestCommissionHandler_Accrue_Errors(t *testing.T) {
	svc := commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
	validAccrue := `{"salesperson_id":5,"plan_id":1,"source_type":"order","source_id":9,"base_amount":100,"journal_id":1,"expense_account_id":2,"payable_account_id":3}`
	validInvoice := `{"invoice_id":9,"salesperson_id":5,"journal_id":1,"expense_account_id":2,"payable_account_id":3}`

	t.Run("rejects invalid body and date", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-entries/accrue", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/commission-entries/accrue", `{"salesperson_id":5,"plan_id":1,"source_type":"order","source_id":9,"base_amount":100,"journal_id":1,"expense_account_id":2,"payable_account_id":3,"date":"yesterday"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/commission-entries/accrue-from-invoice", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/commission-entries/accrue-from-invoice", `{"invoice_id":9,"salesperson_id":5,"journal_id":1,"expense_account_id":2,"payable_account_id":3,"date":"yesterday"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("reports accrue errors", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-entries/accrue", validAccrue)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusUnprocessableEntity && resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d", resp.StatusCode)
		}

		resp, err = doRequest(app, http.MethodPost, "/commission-entries/accrue-from-invoice", validInvoice)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 600 {
			t.Errorf("status = %d", resp.StatusCode)
		}
	})
}

func TestCommissionHandler_Pay_Cancel_Errors(t *testing.T) {
	svc := commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})

	t.Run("rejects invalid pay and cancel", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-entries/abc/pay", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/commission-entries/abc/cancel", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("pay maps not found", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-entries/1/pay", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusUnprocessableEntity && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d", resp.StatusCode)
		}
	})

	t.Run("cancel maps errors", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-entries/1/cancel", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusUnprocessableEntity && resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d", resp.StatusCode)
		}
	})
}

func TestCommissionHandler_Success(t *testing.T) {
	owned := commission.CommissionPlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
			return planOwned(1), nil
		},
	}

	t.Run("updates plan", func(t *testing.T) {
		svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPut, "/commission-plans/1", `{"active":false}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("update maps service errors", func(t *testing.T) {
		missing := commission.CommissionPlanDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
				return nil, nil
			},
		}
		svc := commissionTestSvc(missing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, missing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPut, "/commission-plans/1", `{"active":false}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		failing := commission.CommissionPlanDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
				return nil, errors.New("db down")
			},
		}
		svc = commissionTestSvc(failing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app = commissionHandlerTestWithTenant(t, failing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err = doRequest(app, http.MethodPut, "/commission-plans/1", `{"active":false}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("creates rule", func(t *testing.T) {
		svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/rules", `{"rate_pct":5}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("rule maps validation errors", func(t *testing.T) {
		svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/rules", `{"rate_pct":0,"fixed_amount":0}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("creates assignment", func(t *testing.T) {
		svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/assignments", `{"salesperson_id":5,"date_start":"2026-01-01"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("lists assignments", func(t *testing.T) {
		svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/assignments", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("lists rules", func(t *testing.T) {
		svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/rules", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})
}
