package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

type CreateAccrualRequest struct {
	OrganizationID *uint64                    `json:"organization_id"`
	PeriodID       uint64                     `json:"period_id" validate:"required,gt=0"`
	Name           string                     `json:"name" validate:"required"`
	Description    string                     `json:"description"`
	Date           string                     `json:"date"`
	ReversalDate   string                     `json:"reversal_date"`
	Lines          []CreateAccrualLineRequest `json:"lines" validate:"required,min=2"`
}

type CreateAccrualLineRequest struct {
	AccountID uint64  `json:"account_id" validate:"required,gt=0"`
	Name      string  `json:"name"`
	Debit     float64 `json:"debit"`
	Credit    float64 `json:"credit"`
}

// @Summary Create accrual
// @Description Posts a balanced accrual entry for a period with an optional automatic reversal date, returning the posted accrual with its movement.
// @Tags Accruals
// @Accept json
// @Produce json
// @Param body body CreateAccrualRequest true "Accrual details"
// @Success 201 {object} reporting.AccrualResult "Accrual created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/accruals [post]
func (h AccrualHandler) Create(c fiber.Ctx) error {
	var request CreateAccrualRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	lines := make([]reporting.AccrualLineRequest, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = reporting.AccrualLineRequest{
			AccountID: line.AccountID,
			Name:      line.Name,
			Debit:     line.Debit,
			Credit:    line.Credit,
		}
	}
	accrualRequest := reporting.AccrualRequest{
		OrganizationID: *organizationID,
		PeriodID:       request.PeriodID,
		Name:           request.Name,
		Description:    request.Description,
		Lines:          lines,
	}
	if request.Date != "" {
		parsed, err := helper.ParseDate(&request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
		}
		accrualRequest.Date = *parsed
	}
	if request.ReversalDate != "" {
		parsed, err := helper.ParseDate(&request.ReversalDate)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Reversal date must be in YYYY-MM-DD format.", nil)
		}
		accrualRequest.ReversalDate = parsed
	}

	result, err := h.svc.Create(c, accrualRequest)
	if err != nil {
		return writeAccrualError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Accrual created successfully.", result)
}
