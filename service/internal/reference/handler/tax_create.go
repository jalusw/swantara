package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CreateTaxRequest struct {
	OrganizationID     *uint64  `json:"organization_id"`
	Name               string   `json:"name" validate:"required"`
	Amount             *float64 `json:"amount"`
	Type               string   `json:"type" validate:"required"`
	Scope              string   `json:"scope" validate:"required"`
	PriceInclude       bool     `json:"price_include"`
	TaxAccountID       *uint64  `json:"tax_account_id"`
	RefundTaxAccountID *uint64  `json:"refund_tax_account_id"`
}

// @Summary Create tax
// @Description Creates a tax definition with a required name, type, and scope, plus an optional amount and tax accounts. Percent and fixed taxes require a positive amount, the type and scope must be valid, and referenced accounts must exist; new taxes default to active.
// @Tags Taxes
// @Accept json
// @Produce json
// @Param body body CreateTaxRequest true "Tax details"
// @Success 201 {object} CreateTaxResponseEnvelope "Tax created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/taxes [post]
func (h TaxHandler) Create(c fiber.Ctx) error {
	var request CreateTaxRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	tax, err := h.svc.Create(c, &reference.Tax{
		OrganizationID:     organizationID,
		Name:               request.Name,
		Amount:             request.Amount,
		Type:               request.Type,
		Scope:              request.Scope,
		PriceInclude:       request.PriceInclude,
		TaxAccountID:       request.TaxAccountID,
		RefundTaxAccountID: request.RefundTaxAccountID,
		Active:             true,
	})
	if err != nil {
		return writeTaxError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Tax created successfully.", CreateTaxResponse{
		Tax: newTaxResponse(tax),
	})
}
