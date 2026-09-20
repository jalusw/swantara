package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateTaxReturnResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateTaxReturnResponse `json:"data"`
}
type UpdateTaxReturnResponse struct {
	TaxReturn TaxReturnResponse `json:"tax_return"`
}

// @Summary File tax return
// @Description Transitions a draft tax return to filed, recording the filed-at timestamp on the return. Only draft returns can be filed.
// @Tags Tax Returns
// @Accept json
// @Produce json
// @Param id path integer true "Tax return ID"
// @Success 200 {object} UpdateTaxReturnResponseEnvelope "Tax return filed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax return not found"
// @Failure 422 {object} httpx.ErrorResponse "Tax return must be in draft state"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-returns/{id}/file [post]
func (h TaxReturnHandler) File(c fiber.Ctx) error {
	taxReturnID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax return id provided.", nil)
	}

	taxReturn, err := h.svc.File(c, taxReturnID)
	if err != nil {
		return writeTaxReturnError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Tax return filed successfully.", UpdateTaxReturnResponse{
		TaxReturn: newTaxReturnResponse(taxReturn),
	})
}

// @Summary Pay tax return
// @Description Transitions a filed tax return to paid, finalizing its net payable as settled. Only filed returns can be paid, and a paid return can no longer be reopened.
// @Tags Tax Returns
// @Accept json
// @Produce json
// @Param id path integer true "Tax return ID"
// @Success 200 {object} UpdateTaxReturnResponseEnvelope "Tax return paid successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax return not found"
// @Failure 422 {object} httpx.ErrorResponse "Tax return must be filed first"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-returns/{id}/pay [post]
func (h TaxReturnHandler) Pay(c fiber.Ctx) error {
	taxReturnID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax return id provided.", nil)
	}

	taxReturn, err := h.svc.Pay(c, taxReturnID)
	if err != nil {
		return writeTaxReturnError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Tax return paid successfully.", UpdateTaxReturnResponse{
		TaxReturn: newTaxReturnResponse(taxReturn),
	})
}

type ExportTaxReturnResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ExportTaxReturnResponse `json:"data"`
}
type ExportTaxReturnResponse struct {
	CSV string `json:"csv"`
}

// @Summary Export tax filing
// @Description Exports a filed or paid tax return as CSV rows for e-filing with the tax office. Draft returns cannot be exported.
// @Tags Tax Returns
// @Accept json
// @Produce json
// @Param id path integer true "Tax return ID"
// @Success 200 {object} ExportTaxReturnResponseEnvelope "Tax filing exported successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax return not found"
// @Failure 422 {object} httpx.ErrorResponse "Tax return must be filed first"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-returns/{id}/export [get]
func (h TaxReturnHandler) Export(c fiber.Ctx) error {
	taxReturnID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax return id provided.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	content, err := h.svc.ExportFiling(c, taxReturnID, *organizationID)
	if err != nil {
		return writeTaxReturnError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Tax filing exported successfully.", ExportTaxReturnResponse{
		CSV: content,
	})
}

// @Summary Reopen tax return
// @Description Reopens a filed tax return back to draft so its figures can be revised before filing again. Returns that have already been paid cannot be reopened.
// @Tags Tax Returns
// @Accept json
// @Produce json
// @Param id path integer true "Tax return ID"
// @Success 200 {object} UpdateTaxReturnResponseEnvelope "Tax return reopened successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax return not found"
// @Failure 422 {object} httpx.ErrorResponse "Tax return is already paid"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-returns/{id}/open [post]
func (h TaxReturnHandler) Open(c fiber.Ctx) error {
	taxReturnID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax return id provided.", nil)
	}

	taxReturn, err := h.svc.Draft(c, taxReturnID)
	if err != nil {
		return writeTaxReturnError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Tax return reopened successfully.", UpdateTaxReturnResponse{
		TaxReturn: newTaxReturnResponse(taxReturn),
	})
}
