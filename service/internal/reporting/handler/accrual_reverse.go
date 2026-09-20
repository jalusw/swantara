package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ReverseAccrualsRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	AsOf           string  `json:"as_of" validate:"required"`
}

type ReverseAccrualsResponse struct {
	Reversed int `json:"reversed"`
}

// @Summary Reverse due accruals
// @Description Automatically reverses posted accruals whose reversal date is on or before the given date, creating mirrored GL entries.
// @Tags Accruals
// @Accept json
// @Produce json
// @Param body body ReverseAccrualsRequest true "Reversal details"
// @Success 200 {object} ReverseAccrualsResponse "Due accruals reversed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/accruals/reverse-due [post]
func (h AccrualHandler) ReverseDue(c fiber.Ctx) error {
	var request ReverseAccrualsRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	asOf, err := helper.ParseDate(&request.AsOf)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "as_of must be in YYYY-MM-DD format.", nil)
	}

	reversed, err := h.svc.ReverseDue(c, *organizationID, *asOf)
	if err != nil {
		return writeAccrualError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Due accruals reversed successfully.", ReverseAccrualsResponse{
		Reversed: reversed,
	})
}
