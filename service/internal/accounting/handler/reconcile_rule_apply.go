package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ApplyRulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ApplyRulesResponse `json:"data"`
}
type ApplyRulesResponse struct {
	Applied int `json:"applied"`
}

// @Summary Apply reconcile rules
// @Description Finds and applies reconciliation matches for the caller's organization, optionally skipping candidates below a minimum score
// @Tags Reconciliations
// @Accept json
// @Produce json
// @Param min_score query integer false "Minimum candidate score to apply"
// @Success 200 {object} ApplyRulesResponseEnvelope "Reconciliation rules applied successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Reconciliation amount is invalid"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reconcile-rules/apply [post]
func (h ReconcileRuleHandler) Apply(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}
	minScore := 0
	if raw := c.Query("min_score"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid min_score provided.", nil)
		}
		minScore = parsed
	}

	candidates, err := h.engine.FindMatches(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("reconcile rule match failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to find matches.", err)
	}

	applied, err := h.engine.ApplyMatches(c, candidates, minScore)
	if err != nil {
		return writeReconcileRuleError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Reconciliation rules applied successfully.", ApplyRulesResponse{
		Applied: applied,
	})
}

type ReconcileProposalResponse struct {
	RuleID       uint64  `json:"rule_id"`
	RuleName     string  `json:"rule_name"`
	DebitLineID  uint64  `json:"debit_line_id"`
	CreditLineID uint64  `json:"credit_line_id"`
	Score        int     `json:"score"`
	Amount       float64 `json:"amount"`
}

type SuggestRulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data SuggestRulesResponse `json:"data"`
}
type SuggestRulesResponse struct {
	Proposals []ReconcileProposalResponse `json:"proposals"`
}

// @Summary Suggest reconcile matches
// @Description Ranks reconciliation rule matches for the caller's organization by score without applying them, optionally filtering below a minimum score
// @Tags Reconciliations
// @Accept json
// @Produce json
// @Param min_score query integer false "Minimum candidate score to suggest"
// @Success 200 {object} SuggestRulesResponseEnvelope "Reconciliation suggestions retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid min_score"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reconcile-rules/suggest [post]
func (h ReconcileRuleHandler) Suggest(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}
	minScore := 0
	if raw := c.Query("min_score"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid min_score provided.", nil)
		}
		minScore = parsed
	}

	proposals, err := h.engine.SuggestMatches(c, *organizationID, minScore)
	if err != nil {
		httpx.RequestLog(c).Error("reconcile rule suggest failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to suggest matches.", err)
	}
	items := make([]ReconcileProposalResponse, len(proposals))
	for i, proposal := range proposals {
		items[i] = ReconcileProposalResponse{
			RuleID:       proposal.RuleID,
			RuleName:     proposal.RuleName,
			DebitLineID:  proposal.DebitLineID,
			CreditLineID: proposal.CreditLineID,
			Score:        proposal.Score,
			Amount:       proposal.Amount.Float64(),
		}
	}
	return httpx.CreateSuccessResponse(c, "Reconciliation suggestions retrieved successfully.", SuggestRulesResponse{
		Proposals: items,
	})
}
