package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type CreateTaxReturnRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	PeriodID       uint64  `json:"period_id" validate:"required,gt=0"`
	Type           string  `json:"type"`
}

type CreateTaxReturnResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateTaxReturnResponse `json:"data"`
}
type CreateTaxReturnResponse struct {
	TaxReturn TaxReturnResponse `json:"tax_return"`
}

// @Summary Create tax return
// @Description Computes a draft tax return for a tax period by summing output tax from posted customer invoices and input tax from posted supplier bills. Net payable is recorded as output tax minus input tax, and only one tax return can exist per period.
// @Tags Tax Returns
// @Accept json
// @Produce json
// @Param body body CreateTaxReturnRequest true "Tax return details"
// @Success 201 {object} CreateTaxReturnResponseEnvelope "Tax return created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-returns [post]
func (h TaxReturnHandler) Create(c fiber.Ctx) error {
	var request CreateTaxReturnRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	taxReturn, err := h.svc.Create(c, accounting.CreateTaxReturnRequest{
		OrganizationID: *organizationID,
		PeriodID:       request.PeriodID,
		Type:           request.Type,
	})
	if err != nil {
		return writeTaxReturnError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Tax return created successfully.", CreateTaxReturnResponse{
		TaxReturn: newTaxReturnResponse(taxReturn),
	})
}
