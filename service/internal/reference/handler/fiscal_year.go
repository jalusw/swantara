package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type TaxYearHandler struct {
	svc reference.TaxYearService
}

func NewTaxYearHandler(svc reference.TaxYearService) TaxYearHandler {
	return TaxYearHandler{svc: svc}
}

var taxYearQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"state":           {},
	"created_at":      {},
	"updated_at":      {},
}

func writeTaxYearError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reference.ErrInvalidTaxYear):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax year date range is invalid.", nil)
	default:
		httpx.RequestLog(c).Error("tax year write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save tax year.", err)
	}
}
