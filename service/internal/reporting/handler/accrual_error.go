package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

func writeAccrualError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reporting.ErrUnbalancedAccrual):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Accrual lines must be balanced with positive debit or credit amounts.", nil)
	case errors.Is(err, reporting.ErrConfigMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Reporting configuration is missing (journal).", nil)
	default:
		httpx.RequestLog(c).Error("accrual write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save accrual.", err)
	}
}
