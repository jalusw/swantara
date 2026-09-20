package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type TaxReturnHandler struct {
	svc accounting.TaxReturnService
}

func NewTaxReturnHandler(svc accounting.TaxReturnService) TaxReturnHandler {
	return TaxReturnHandler{svc: svc}
}

type TaxReturnResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID *uint64   `json:"organization_id"`
	PeriodID       uint64    `json:"period_id"`
	Type           string    `json:"type"`
	OutputTax      float64   `json:"output_tax"`
	InputTax       float64   `json:"input_tax"`
	NetPayable     float64   `json:"net_payable"`
	State          string    `json:"state"`
	FiledAt        *string   `json:"filed_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newTaxReturnResponse(taxReturn *accounting.TaxReturn) TaxReturnResponse {
	return TaxReturnResponse{
		ID:             taxReturn.ID,
		OrganizationID: taxReturn.OrganizationID,
		PeriodID:       taxReturn.PeriodID,
		Type:           taxReturn.Type,
		OutputTax:      taxReturn.OutputTax,
		InputTax:       taxReturn.InputTax,
		NetPayable:     taxReturn.NetPayable,
		State:          taxReturn.State,
		FiledAt:        helper.FormatDatePtr(taxReturn.FiledAt),
		CreatedAt:      taxReturn.CreatedAt,
		UpdatedAt:      taxReturn.UpdatedAt,
	}
}

func writeTaxReturnError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrTaxReturnNotFound), errors.Is(err, accounting.ErrPeriodNotFound):
		return httpx.CreateNotFoundResponse(c, "Tax return or period not found.")
	case errors.Is(err, accounting.ErrTaxReturnExists):
		return httpx.CreateConflictResponse(c, "A tax return already exists for the period.", err)
	case errors.Is(err, accounting.ErrTaxReturnNotDraft), errors.Is(err, accounting.ErrTaxReturnNotFiled), errors.Is(err, accounting.ErrTaxReturnPaid):
		return httpx.CreateConflictResponse(c, "Tax return state does not allow the requested transition.", err)
	default:
		httpx.RequestLog(c).Error("tax return write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save tax return.", err)
	}
}
