package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

type PeriodCloseHandler struct {
	svc reporting.PeriodCloseService
}

func NewPeriodCloseHandler(svc reporting.PeriodCloseService) PeriodCloseHandler {
	return PeriodCloseHandler{svc: svc}
}

type ClosePeriodRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	PeriodID       uint64  `json:"period_id" validate:"required,gt=0"`
}

// @Summary Close period
// @Description Runs the period-close gate: depreciation, FX revaluation, accrual reversals, and deferral recognition, then closes and locks the tax period.
// @Tags Period Close
// @Accept json
// @Produce json
// @Param body body ClosePeriodRequest true "Period close details"
// @Success 200 {object} reporting.CloseResult "Period closed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/period-close [post]
func (h PeriodCloseHandler) Close(c fiber.Ctx) error {
	var request ClosePeriodRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	result, err := h.svc.Close(c, *organizationID, request.PeriodID)
	if err != nil {
		return writeCloseError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Period closed successfully.", result)
}

func writeCloseError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reporting.ErrPeriodNotFound):
		return httpx.CreateNotFoundResponse(c, "Tax period not found.")
	case errors.Is(err, reporting.ErrPeriodLocked):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax period is locked.", nil)
	case errors.Is(err, reporting.ErrConfigMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Reporting configuration is missing (journal or FX gain/loss account).", nil)
	default:
		httpx.RequestLog(c).Error("period close failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to close the period.", err)
	}
}
