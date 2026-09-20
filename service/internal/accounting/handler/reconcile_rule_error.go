package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func writeReconcileRuleError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrLineNotFound), errors.Is(err, accounting.ErrReconcileRuleNotFound):
		return httpx.CreateNotFoundResponse(c, "Reconciliation record not found.")
	case errors.Is(err, accounting.ErrLineNotInOrganization):
		return httpx.CreateNotFoundResponse(c, "Account entry line not found.")
	case errors.Is(err, accounting.ErrReconcileAmount), errors.Is(err, accounting.ErrReconcileExceedsBalance), errors.Is(err, accounting.ErrInvalidLine), errors.Is(err, accounting.ErrLinesDifferentAccount), errors.Is(err, accounting.ErrReconcileRuleNoAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Reconciliation amount is invalid.", nil)
	default:
		httpx.RequestLog(c).Error("reconcile rule apply failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to apply reconcile rules.", err)
	}
}
