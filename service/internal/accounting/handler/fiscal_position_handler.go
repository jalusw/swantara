package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type TaxRuleHandler struct {
	resolver accounting.TaxRuleResolver
}

func NewTaxRuleHandler(resolver accounting.TaxRuleResolver) TaxRuleHandler {
	return TaxRuleHandler{resolver: resolver}
}

var taxRuleQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"country_code":    {},
	"auto_apply":      {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}

func writeTaxRuleError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrTaxRuleNotFound):
		return httpx.CreateNotFoundResponse(c, "Fiscal position not found.")
	default:
		httpx.RequestLog(c).Error("fiscal position write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save fiscal position.", err)
	}
}
