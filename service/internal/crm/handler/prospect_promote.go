package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type PromoteLeadRequest struct {
	StageID         uint64   `json:"stage_id" validate:"required,gt=0"`
	SalespersonID   *uint64  `json:"salesperson_id"`
	SalesGroupID    *uint64  `json:"sales_group_id"`
	ExpectedRevenue float64  `json:"expected_revenue"`
	Probability     *float64 `json:"probability"`
	Priority        int16    `json:"priority"`
	ExpectedClose   *string  `json:"expected_close"`
}

type PromoteLeadResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PromoteLeadResponse `json:"data"`
}
type PromoteLeadResponse struct {
	Prospect ProspectResponse `json:"prospect"`
}

// @Summary Promote prospect to opportunity
// @Description Qualifies an existing prospect into an opportunity by assigning a required pipeline stage and optional sales context, salesperson or sales team, expected revenue, probability, priority, and expected close date. Promotion is only allowed for records typed as leads, and the target stage must exist. When no probability is supplied the target stage's probability is applied, and validation enforces non-negative revenue and priority and a probability between 0 and 100.
// @Tags CRM Leads
// @Accept json
// @Produce json
// @Param id path integer true "CRM prospect ID"
// @Param body body PromoteLeadRequest true "Promotion details"
// @Success 200 {object} PromoteLeadResponseEnvelope "Prospect promoted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Prospect not found"
// @Failure 409 {object} httpx.ErrorResponse "Prospect is not promotable"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown reference"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/leads/{id}/promote [post]
func (h ProspectHandler) Promote(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid prospect id provided.", nil)
	}

	var request PromoteLeadRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm prospect lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to promote prospect.", err)
	}
	if prospect == nil || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Prospect not found.")
	}

	expectedClose, err := helper.ParseDate(request.ExpectedClose)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expected close date provided.", nil)
	}

	promoted, err := h.svc.Promote(c, prospectID, crm.PromotionInput{
		StageID:         request.StageID,
		SalespersonID:   request.SalespersonID,
		SalesGroupID:    request.SalesGroupID,
		ExpectedRevenue: request.ExpectedRevenue,
		Probability:     request.Probability,
		Priority:        request.Priority,
		ExpectedClose:   expectedClose,
	})
	if err != nil {
		return writeProspectError(c, err)
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to promote prospect.", err)
	}

	return httpx.CreateSuccessResponse(c, "Prospect promoted successfully.", PromoteLeadResponse{
		Prospect: newProspectResponse(promoted, stages),
	})
}
