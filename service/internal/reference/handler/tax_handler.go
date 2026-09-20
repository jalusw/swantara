package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type TaxHandler struct {
	svc reference.TaxService
}

func NewTaxHandler(svc reference.TaxService) TaxHandler {
	return TaxHandler{svc: svc}
}

var taxQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"type":            {},
	"scope":           {},
	"price_include":   {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}

func writeTaxError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reference.ErrInvalidTaxType):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax type is not valid.", nil)
	case errors.Is(err, reference.ErrInvalidTaxScope):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax scope is not valid.", nil)
	case errors.Is(err, reference.ErrTaxAmountMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Percent and fixed taxes require a positive amount.", nil)
	case errors.Is(err, reference.ErrInvalidTaxAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax account does not exist.", nil)
	default:
		httpx.RequestLog(c).Error("tax write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save tax.", err)
	}
}
