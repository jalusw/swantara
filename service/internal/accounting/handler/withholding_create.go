package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type CreateWithholdingTaxRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name" validate:"required"`
	RatePct        float64 `json:"rate_pct" validate:"required,gte=0"`
	AccountID      *uint64 `json:"account_id" validate:"required,gt=0"`
	Scope          string  `json:"scope" validate:"required,oneof=sale purchase"`
}

type CreateWithholdingTaxResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateWithholdingTaxResponse `json:"data"`
}
type CreateWithholdingTaxResponse struct {
	WithholdingTax WithholdingTaxResponse `json:"withholding_tax"`
}

// @Summary Create withholding tax
// @Description Creates an active withholding tax definition under the caller's organization, capturing its rate percentage, posting account, and a sale or purchase scope. The scope must be either sale or purchase, and the rate must be non-negative.
// @Tags Withholding Taxes
// @Accept json
// @Produce json
// @Param body body CreateWithholdingTaxRequest true "Withholding tax details"
// @Success 201 {object} CreateWithholdingTaxResponseEnvelope "Withholding tax created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/withholding-taxes [post]
func (h WithholdingTaxHandler) Create(c fiber.Ctx) error {
	var request CreateWithholdingTaxRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	tax, err := h.svc.Create(c, &accounting.WithholdingTax{
		OrganizationID: organizationID,
		Name:           helper.Ptr(request.Name),
		RatePct:        request.RatePct,
		AccountID:      request.AccountID,
		Scope:          request.Scope,
	})
	if err != nil {
		return writeWithholdingTaxError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Withholding tax created successfully.", CreateWithholdingTaxResponse{
		WithholdingTax: newWithholdingTaxResponse(tax),
	})
}
