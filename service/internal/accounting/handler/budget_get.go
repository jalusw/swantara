package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Get budget
// @Description Gets a single budget by its id together with its planned lines, each showing the planned and practical amounts per account, returning 404 when the budget does not exist or belongs to another tenant.
// @Tags Budgets
// @Accept json
// @Produce json
// @Param id path integer true "Budget ID"
// @Success 200 {object} GetBudgetResponseEnvelope "Budget retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Budget not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/budgets/{id} [get]
func (h BudgetHandler) Get(c fiber.Ctx) error {
	budgetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid budget id provided.", nil)
	}

	budget, err := h.svc.Find(c, budgetID)
	if err != nil {
		httpx.RequestLog(c).Error("budget lookup failed", "budget_id", budgetID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get budget.", err)
	}
	if budget == nil || !httpx.OwnsTenant(c, budget.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Budget not found.")
	}

	lines, err := h.svc.ListLines(c, budget.ID)
	if err != nil {
		httpx.RequestLog(c).Error("budget lines lookup failed", "budget_id", budget.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get budget lines.", err)
	}
	lineItems := make([]BudgetLineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newBudgetLineResponse(line)
	}

	return httpx.CreateSuccessResponse(c, "Budget retrieved successfully.", GetBudgetResponse{
		Budget: newBudgetResponse(budget),
		Lines:  lineItems,
	})
}
