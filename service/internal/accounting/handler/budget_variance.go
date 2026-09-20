package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Get budget variance
// @Description Computes the planned versus practical variance for a budget by summing posted amounts within the budget period. Each budget line returns its planned amount, actual posted amount, and the resulting variance, or 404 when the budget is not found.
// @Tags Budgets
// @Accept json
// @Produce json
// @Param id path integer true "Budget ID"
// @Success 200 {object} BudgetVarianceResponseEnvelope "Budget variance retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Budget not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/budgets/{id}/variance [get]
func (h BudgetHandler) Variance(c fiber.Ctx) error {
	budgetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid budget id provided.", nil)
	}

	variance, err := h.svc.Variance(c, budgetID)
	if err != nil {
		return writeBudgetError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Budget variance retrieved successfully.", BudgetVarianceResponse{
		Variance: variance,
	})
}
