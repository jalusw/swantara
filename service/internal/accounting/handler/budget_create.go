package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Create budget
// @Description Creates a draft budget with its planned lines, validating that at least one line is provided and that the date range is valid. The planned amounts are stored per account and optional dimension account, and the budget starts in the draft state.
// @Tags Budgets
// @Accept json
// @Produce json
// @Param body body CreateBudgetRequest true "Budget details"
// @Success 201 {object} CreateBudgetResponseEnvelope "Budget created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/budgets [post]
func (h BudgetHandler) Create(c fiber.Ctx) error {
	var request CreateBudgetRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	dateStart, err := helper.ParseDate(&request.DateStart)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}
	dateEnd, err := helper.ParseDate(&request.DateEnd)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	lines := make([]accounting.BudgetLineRequest, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = accounting.BudgetLineRequest{
			AccountID:     line.AccountID,
			DimensionID:   line.DimensionID,
			PlannedAmount: line.PlannedAmount,
		}
	}

	budget, err := h.svc.Create(c, accounting.CreateBudgetRequest{
		OrganizationID: *organizationID,
		Name:           request.Name,
		DateStart:      *dateStart,
		DateEnd:        *dateEnd,
		Lines:          lines,
	})
	if err != nil {
		return writeBudgetError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Budget created successfully.", CreateBudgetResponse{
		Budget: newBudgetResponse(budget),
	})
}
