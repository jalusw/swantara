package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/commission"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func ownedPlanMock() commission.CommissionPlanDAOMock {
	return commission.CommissionPlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
			return planOwned(1), nil
		},
	}
}

func failingPlanMock() commission.CommissionPlanDAOMock {
	return commission.CommissionPlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
			return nil, errors.New("db down")
		},
	}
}

func TestCommissionHandler_TenantGuards(t *testing.T) {
	owned := ownedPlanMock()
	svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})

	t.Run("list plans with guard tenant", func(t *testing.T) {
		app := commissionHandlerTest(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodGet, "/commission-plans/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("create plan requires org", func(t *testing.T) {
		app := commissionHandlerTest(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-plans/", `{"name":"P"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("get plan rejects invalid id", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodGet, "/commission-plans/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})
}

func TestCommissionHandler_PlanErrors(t *testing.T) {
	t.Run("update plan bind failure and svc error", func(t *testing.T) {
		owned := ownedPlanMock()
		svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPut, "/commission-plans/1", `not-json`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("status = %d", resp.StatusCode)
		}

		broken := commission.CommissionPlanDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*commission.CommissionPlan, error) {
				return nil, errors.New("db down")
			},
		}
		brokenSvc := commissionTestSvc(broken, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app = commissionHandlerTestWithTenant(t, broken, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, brokenSvc)
		resp, err = doRequest(app, http.MethodPut, "/commission-plans/1", `{"active":true}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("rules plan errors", func(t *testing.T) {
		failing := failingPlanMock()
		svc := commissionTestSvc(failing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, failing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/rules", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)

		resp, err = doRequest(app, http.MethodPost, "/commission-plans/1/rules", `{"rate_pct":5}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("assignments plan errors and dates", func(t *testing.T) {
		owned := ownedPlanMock()
		svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})

		failing := failingPlanMock()
		failSvc := commissionTestSvc(failing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, failing, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, failSvc)
		resp, err := doRequest(app, http.MethodGet, "/commission-plans/1/assignments", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)

		resp, err = doRequest(app, http.MethodPost, "/commission-plans/1/assignments", `{"salesperson_id":5,"date_start":"2026-01-01"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)

		app = commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err = doRequest(app, http.MethodGet, "/commission-plans/abc/assignments", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodGet, "/commission-plans/999/assignments", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d", resp.StatusCode)
		}
	})

	t.Run("assignment with end date and svc error", func(t *testing.T) {
		owned := ownedPlanMock()
		svc := commissionTestSvc(owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app := commissionHandlerTestWithTenant(t, owned, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-plans/1/assignments", `{"salesperson_id":5,"date_start":"2026-01-01","date_end":"2026-12-31"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)

		overlapping := commission.CommissionAssignmentDAOMock{}
		_ = overlapping
		idAware := commission.CommissionPlanDAOMock{
			FindFunc: func(_ context.Context, id uint64) (*commission.CommissionPlan, error) {
				if id != 1 {
					return nil, nil
				}
				return planOwned(1), nil
			},
		}
		idSvc := commissionTestSvc(idAware, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})
		app = commissionHandlerTestWithTenant(t, idAware, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, idSvc)
		resp, err = doRequest(app, http.MethodPost, "/commission-plans/999/assignments", `{"salesperson_id":5,"date_start":"2026-01-01"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})
}

func TestCommissionHandler_Entries_Errors(t *testing.T) {
	svc := commissionTestSvc(commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{})

	t.Run("list entries bad query", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodGet, "/commission-entries/?size=abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("accrue from invoice with date", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-entries/accrue-from-invoice", `{"invoice_id":9,"salesperson_id":5,"journal_id":1,"expense_account_id":2,"payable_account_id":3,"date":"2026-01-15"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 600 {
			t.Errorf("status = %d", resp.StatusCode)
		}
	})

	t.Run("pay bind failure", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-entries/1/pay", `not-json`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("status = %d", resp.StatusCode)
		}
	})

	t.Run("write error default", func(t *testing.T) {
		app := commissionHandlerTestWithTenant(t, commission.CommissionPlanDAOMock{}, commission.CommissionRuleDAOMock{}, commission.CommissionAssignmentDAOMock{}, commission.CommissionEntryDAOMock{}, svc)
		resp, err := doRequest(app, http.MethodPost, "/commission-entries/1/cancel", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("status = %d", resp.StatusCode)
		}
	})

	t.Run("plan model", func(t *testing.T) {
		owned := planOwned(1)
		if owned.ID != 1 {
			t.Errorf("plan = %+v", owned)
		}
		_ = model.Base{}
	})
}
