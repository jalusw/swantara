package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type ReplenishmentCandidatesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ReplenishmentCandidatesResponse `json:"data"`
}
type ReplenishmentCandidatesResponse struct {
	Candidates []inventory.ReplenishmentCandidate `json:"candidates"`
}

// @Summary Get replenishment candidates
// @Description Computes replenishment candidates from the active reorder rules of the caller's organization. For each rule whose on-hand stock at its location has fallen below the configured minimum, a candidate is produced with a recommended reorder quantity of the maximum minus on-hand rounded up to the rule's quantity multiple. Results can be returned as JSON or exported as a CSV file when the Accept header requests CSV.
// @Tags Reorder Rules
// @Accept json
// @Produce json
// @Success 200 {object} ReplenishmentCandidatesResponseEnvelope "Replenishment candidates retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /reorder-rules/candidates [get]
func (h ReorderHandler) Candidates(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	candidates, err := h.svc.Candidates(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("replenishment candidates failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to compute replenishment candidates.", err)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "reorder-candidates.csv", candidates)
	}

	return httpx.CreateSuccessResponse(c, "Replenishment candidates retrieved successfully.", ReplenishmentCandidatesResponse{
		Candidates: candidates,
	})
}
