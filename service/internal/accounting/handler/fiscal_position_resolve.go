package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Resolve fiscal position
// @Description Applies a fiscal position's remap rules to a document's source tax and account, returning the mapped destination tax and account. Source values without a matching remap are passed through unchanged.
// @Tags Fiscal Positions
// @Accept json
// @Produce json
// @Param id path integer true "Fiscal position ID"
// @Param body body ResolveTaxRuleRequest true "Document reference"
// @Success 200 {object} ResolveTaxRuleResponseEnvelope "Fiscal position resolved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Fiscal position not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-rules/{id}/resolve [post]
func (h TaxRuleHandler) Resolve(c fiber.Ctx) error {
	positionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid fiscal position id provided.", nil)
	}

	var request ResolveTaxRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	result, err := h.resolver.Resolve(c, positionID, request.TaxID, request.AccountID)
	if err != nil {
		return writeTaxRuleError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Fiscal position resolved successfully.", ResolveTaxRuleResponse{
		TaxID:     result.TaxAccount,
		AccountID: result.Account,
	})
}
