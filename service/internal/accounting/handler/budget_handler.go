package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type BudgetHandler struct {
	svc accounting.BudgetService
}

func NewBudgetHandler(svc accounting.BudgetService) BudgetHandler {
	return BudgetHandler{svc: svc}
}

func writeBudgetError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrBudgetNotFound):
		return httpx.CreateNotFoundResponse(c, "Budget not found.")
	case errors.Is(err, accounting.ErrBudgetNoLines), errors.Is(err, accounting.ErrBudgetInvalidDates):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Budget must have at least one line and a valid date range.", nil)
	default:
		httpx.RequestLog(c).Error("budget write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save budget.", err)
	}
}
