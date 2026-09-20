package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListReconcileRulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListReconcileRulesResponse `json:"data"`
}
type ListReconcileRulesResponse struct {
	Rules []ReconcileRuleResponse `json:"rules"`
}

// @Summary List reconcile rules
// @Description Lists reconciliation rules with pagination, sorting, and filtering, automatically scoping results to the caller's organization
// @Tags Reconciliations
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. sequence:asc)"
// @Param filter query string false "Filters (repeatable, e.g. active:eq:true)"
// @Success 200 {object} ListReconcileRulesResponseEnvelope "Reconcile rules retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reconcile-rules [get]
func (h ReconcileRuleHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, reconcileRuleQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.engine.ListRules(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("reconcile rule list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve reconcile rules.", err)
	}

	items := make([]ReconcileRuleResponse, len(page.Items))
	for i, rule := range page.Items {
		items[i] = newReconcileRuleResponse(rule)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Reconcile rules retrieved successfully.", ListReconcileRulesResponse{
		Rules: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type CreateReconcileRuleRequest struct {
	OrganizationID  *uint64 `json:"organization_id"`
	Name            string  `json:"name" validate:"required"`
	AccountID       *uint64 `json:"account_id"`
	MatchContact    bool    `json:"match_contact"`
	MatchAmount     bool    `json:"match_amount"`
	MatchRef        bool    `json:"match_ref"`
	AmountTolerance float64 `json:"amount_tolerance" validate:"gte=0"`
	Active          bool    `json:"active"`
	Sequence        int     `json:"sequence" validate:"gte=0"`
}

type CreateReconcileRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateReconcileRuleResponse `json:"data"`
}
type CreateReconcileRuleResponse struct {
	Rule ReconcileRuleResponse `json:"rule"`
}

// @Summary Create reconcile rule
// @Description Creates a new reconciliation rule for the caller's organization
// @Tags Reconciliations
// @Accept json
// @Produce json
// @Param body body CreateReconcileRuleRequest true "Reconcile rule details"
// @Success 201 {object} CreateReconcileRuleResponseEnvelope "Reconcile rule created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid request"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reconcile-rules [post]
func (h ReconcileRuleHandler) Create(c fiber.Ctx) error {
	var request CreateReconcileRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	rule := &accounting.ReconcileRule{
		OrganizationID:  *organizationID,
		Name:            &request.Name,
		AccountID:       request.AccountID,
		MatchContact:    request.MatchContact,
		MatchAmount:     request.MatchAmount,
		MatchRef:        request.MatchRef,
		AmountTolerance: request.AmountTolerance,
		Active:          request.Active,
		Sequence:        request.Sequence,
	}

	created, err := h.engine.CreateRule(c, rule)
	if err != nil {
		httpx.RequestLog(c).Error("reconcile rule create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create reconcile rule.", err)
	}

	return httpx.CreateCreatedResponse(c, "Reconcile rule created successfully.", CreateReconcileRuleResponse{
		Rule: newReconcileRuleResponse(created),
	})
}
