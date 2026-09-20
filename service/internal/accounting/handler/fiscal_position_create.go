package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Create fiscal position
// @Description Creates an active fiscal position under the caller's organization, capturing its name, country code, and auto-apply flag together with its tax and account remap mappings. The mappings define how source taxes and accounts are translated when the position is applied.
// @Tags Fiscal Positions
// @Accept json
// @Produce json
// @Param body body CreateTaxRuleRequest true "Fiscal position details"
// @Success 201 {object} CreateTaxRuleResponseEnvelope "Fiscal position created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-rules [post]
func (h TaxRuleHandler) Create(c fiber.Ctx) error {
	var request CreateTaxRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	taxMaps := make([]*accounting.TaxRuleTaxMap, len(request.TaxMaps))
	for i, taxMap := range request.TaxMaps {
		taxMaps[i] = &accounting.TaxRuleTaxMap{SrcTaxID: taxMap.SrcTaxID, DestTaxID: taxMap.DestTaxID}
	}
	accountMaps := make([]*accounting.TaxRuleAccountMap, len(request.AccountMaps))
	for i, accountMap := range request.AccountMaps {
		accountMaps[i] = &accounting.TaxRuleAccountMap{SrcAccountID: accountMap.SrcAccountID, DestAccountID: accountMap.DestAccountID}
	}

	position, err := h.resolver.Create(c, accounting.CreateTaxRuleRequest{
		OrganizationID: *organizationID,
		Name:           request.Name,
		CountryCode:    request.CountryCode,
		AutoApply:      request.AutoApply,
		TaxMaps:        taxMaps,
		AccountMaps:    accountMaps,
	})
	if err != nil {
		return writeTaxRuleError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Fiscal position created successfully.", CreateTaxRuleResponse{
		TaxRule: newTaxRuleResponse(position),
	})
}
