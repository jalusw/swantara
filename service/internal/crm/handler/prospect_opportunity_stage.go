package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type AdvanceStageRequest struct {
	StageID uint64 `json:"stage_id" validate:"required,gt=0"`
}

// @Summary Advance opportunity stage
// @Description Moves an open opportunity to another pipeline stage, syncing the opportunity's probability from the target stage and closing the opportunity when the target stage is flagged as won. Closed opportunities and records not typed as opportunities are rejected with a conflict, and the target stage must exist.
// @Tags CRM Opportunities
// @Accept json
// @Produce json
// @Param id path integer true "CRM opportunity ID"
// @Param body body AdvanceStageRequest true "Target stage"
// @Success 200 {object} AdvanceStageResponseEnvelope "Opportunity advanced successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Opportunity not found"
// @Failure 409 {object} httpx.ErrorResponse "Opportunity is closed or not promotable"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown stage"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/opportunities/{id}/advance-stage [post]
func (h OpportunityHandler) AdvanceStage(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid opportunity id provided.", nil)
	}

	var request AdvanceStageRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm opportunity lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to advance opportunity.", err)
	}
	if prospect == nil || prospect.Type != crm.ProspectKindOpportunity || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Opportunity not found.")
	}

	advanced, err := h.svc.AdvanceStage(c, prospectID, request.StageID)
	if err != nil {
		return writeProspectError(c, err)
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to advance opportunity.", err)
	}

	return httpx.CreateSuccessResponse(c, "Opportunity advanced successfully.", AdvanceStageResponse{
		Opportunity: newProspectResponse(advanced, stages),
	})
}

// @Summary Win opportunity
// @Description Marks an open opportunity as won, moving it to the configured won stage, setting its probability to 100, clearing any lost reason, and recording a closed timestamp. Closed or non-opportunity records are rejected with a conflict, and a won stage must be configured.
// @Tags CRM Opportunities
// @Accept json
// @Produce json
// @Param id path integer true "CRM opportunity ID"
// @Success 200 {object} WinOpportunityResponseEnvelope "Opportunity won successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Opportunity not found"
// @Failure 409 {object} httpx.ErrorResponse "Opportunity is closed or not an opportunity"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/opportunities/{id}/win [post]
func (h OpportunityHandler) Win(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid opportunity id provided.", nil)
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm opportunity lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to win opportunity.", err)
	}
	if prospect == nil || prospect.Type != crm.ProspectKindOpportunity || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Opportunity not found.")
	}

	won, err := h.svc.Win(c, prospectID)
	if err != nil {
		return writeProspectError(c, err)
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to win opportunity.", err)
	}

	return httpx.CreateSuccessResponse(c, "Opportunity won successfully.", WinOpportunityResponse{
		Opportunity: newProspectResponse(won, stages),
	})
}

type LoseOpportunityRequest struct {
	LostReason string `json:"lost_reason" validate:"required"`
}

// @Summary Lose opportunity
// @Description Marks an open opportunity as lost, storing a required non-empty lost reason, zeroing the probability, and recording a closed timestamp. Closed or non-opportunity records are rejected with a conflict, and a lost reason must be supplied.
// @Tags CRM Opportunities
// @Accept json
// @Produce json
// @Param id path integer true "CRM opportunity ID"
// @Param body body LoseOpportunityRequest true "Lost reason"
// @Success 200 {object} LoseOpportunityResponseEnvelope "Opportunity lost successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Opportunity not found"
// @Failure 409 {object} httpx.ErrorResponse "Opportunity is closed or not an opportunity"
// @Failure 422 {object} httpx.ErrorResponse "Lost reason required"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/opportunities/{id}/lose [post]
func (h OpportunityHandler) Lose(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid opportunity id provided.", nil)
	}

	var request LoseOpportunityRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm opportunity lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to lose opportunity.", err)
	}
	if prospect == nil || prospect.Type != crm.ProspectKindOpportunity || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Opportunity not found.")
	}

	lost, err := h.svc.Lose(c, prospectID, request.LostReason)
	if err != nil {
		return writeProspectError(c, err)
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to lose opportunity.", err)
	}

	return httpx.CreateSuccessResponse(c, "Opportunity lost successfully.", LoseOpportunityResponse{
		Opportunity: newProspectResponse(lost, stages),
	})
}
