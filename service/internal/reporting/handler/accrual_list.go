package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListAccrualsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []AccrualResponse `json:"data"`
}

// @Summary List accruals
// @Description Lists accruals for the tenant.
// @Tags Accruals
// @Accept json
// @Produce json
// @Success 200 {object} ListAccrualsResponseEnvelope "Accruals retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/accruals [get]
func (h AccrualHandler) List(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	accruals, err := h.svc.ListByOrganization(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("accrual list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to list accruals.", err)
	}
	items := make([]AccrualResponse, len(accruals))
	for i, accrual := range accruals {
		items[i] = newAccrualResponse(accrual)
	}
	return httpx.CreateSuccessResponse(c, "Accruals retrieved successfully.", items)
}
