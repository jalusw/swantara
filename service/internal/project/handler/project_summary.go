package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ProjectSummaryResponse struct {
	ProjectID      uint64  `json:"project_id"`
	PlannedHours   float64 `json:"planned_hours"`
	EffectiveHours float64 `json:"effective_hours"`
	Utilization    float64 `json:"utilization"`
	BilledAmount   float64 `json:"billed_amount"`
	UnbilledAmount float64 `json:"unbilled_amount"`
	BillableAmount float64 `json:"billable_amount"`
	CostAmount     float64 `json:"cost_amount"`
	MarginAmount   float64 `json:"margin_amount"`
}

// @Summary Project summary
// @Description Gets a project's budget versus actual performance, including planned and effective hours with utilization, billed and unbilled billable amounts, cost derived from active employee contracts, and the resulting margin.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path integer true "Project ID"
// @Success 200 {object} ProjectSummaryResponse "Project summary retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Project not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{id}/summary [get]
func (h ProjectHandler) Summary(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	projectRecord, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("project get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve project.", err)
	}
	if projectRecord == nil || !httpx.OwnsTenant(c, helper.Ptr(projectRecord.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Project not found.")
	}
	summary, err := h.svc.Summary(c, projectRecord.ID)
	if err != nil {
		return writeProjectError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Project summary retrieved successfully.", ProjectSummaryResponse{
		ProjectID:      summary.ProjectID,
		PlannedHours:   summary.PlannedHours,
		EffectiveHours: summary.EffectiveHours,
		Utilization:    summary.Utilization,
		BilledAmount:   summary.BilledAmount,
		UnbilledAmount: summary.UnbilledAmount,
		BillableAmount: summary.BillableAmount,
		CostAmount:     summary.CostAmount,
		MarginAmount:   summary.MarginAmount,
	})
}
