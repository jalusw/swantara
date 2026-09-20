package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type WithholdingTaxHandler struct {
	svc accounting.WithholdingService
}

func NewWithholdingTaxHandler(svc accounting.WithholdingService) WithholdingTaxHandler {
	return WithholdingTaxHandler{svc: svc}
}

type WithholdingTaxResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID *uint64   `json:"organization_id"`
	Name           *string   `json:"name"`
	RatePct        float64   `json:"rate_pct"`
	AccountID      *uint64   `json:"account_id"`
	Scope          string    `json:"scope"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newWithholdingTaxResponse(tax *accounting.WithholdingTax) WithholdingTaxResponse {
	return WithholdingTaxResponse{
		ID:             tax.ID,
		OrganizationID: tax.OrganizationID,
		Name:           tax.Name,
		RatePct:        tax.RatePct,
		AccountID:      tax.AccountID,
		Scope:          tax.Scope,
		Active:         tax.Active,
		CreatedAt:      tax.CreatedAt,
		UpdatedAt:      tax.UpdatedAt,
	}
}

var withholdingTaxQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"rate_pct":        {},
	"account_id":      {},
	"scope":           {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}

func writeWithholdingTaxError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrWithholdingNotFound):
		return httpx.CreateNotFoundResponse(c, "Withholding tax not found.")
	case errors.Is(err, accounting.ErrWithholdingNoAccount), errors.Is(err, accounting.ErrWithholdingScope):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Withholding tax must have an account and matching scope.", nil)
	default:
		httpx.RequestLog(c).Error("withholding tax write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save withholding tax.", err)
	}
}
