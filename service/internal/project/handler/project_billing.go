package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/project"
)

type BillTimeMaterialRequest struct {
	JournalID uint64            `json:"journal_id" validate:"required,gt=0"`
	Date      string            `json:"date" validate:"required"`
	TaxIDs    helper.Int64Array `json:"tax_ids"`
}

type BillTimeMaterialResponse struct {
	InvoiceID   uint64  `json:"invoice_id"`
	Name        *string `json:"name"`
	AmountTotal float64 `json:"amount_total"`
}

// @Summary Bill time and material
// @Description Generates a customer invoice from a project's unbilled billable timesheets, valued at the project's billable rate and posted to the given journal and date. Only open time-and-material projects with a billable rate can be billed, and each billed timesheet is recorded to prevent double billing.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path integer true "Project ID"
// @Param body body BillTimeMaterialRequest true "Billing details"
// @Success 201 {object} BillTimeMaterialResponse "Invoice generated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Project not found"
// @Failure 422 {object} httpx.ErrorResponse "Nothing to bill or invalid state"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{id}/bill [post]
func (h ProjectHandler) BillTimeMaterial(c fiber.Ctx) error {
	var request BillTimeMaterialRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDateStr(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
	}
	organizationID, ok := httpx.ResolveOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
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
	invoice, err := h.svc.BillTimeMaterial(c, project.BillTimeMaterialRequest{
		OrganizationID: organizationID,
		ProjectID:      projectRecord.ID,
		JournalID:      request.JournalID,
		Date:           date,
		TaxIDs:         request.TaxIDs,
	})
	if err != nil {
		return writeProjectError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Invoice generated successfully.", BillTimeMaterialResponse{
		InvoiceID:   invoice.ID,
		Name:        invoice.Name,
		AmountTotal: invoice.AmountTotal.Float64(),
	})
}
