package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

type ReportHandler struct {
	svc reporting.ReportService
}

func NewReportHandler(svc reporting.ReportService) ReportHandler {
	return ReportHandler{svc: svc}
}

type TrialBalanceResponseEnvelope struct {
	httpx.EnvelopeBase
	Data reporting.TrialBalance `json:"data"`
}

// @Summary Trial balance
// @Description Returns the trial balance for a tax period: opening, period activity, and closing debit/credit per account, reconciling to the posted general ledger.
// @Tags Reporting
// @Accept json
// @Produce json
// @Param period_id query integer true "Tax period ID"
// @Success 200 {object} TrialBalanceResponseEnvelope "Trial balance retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reports/trial-balance [get]
func (h ReportHandler) TrialBalance(c fiber.Ctx) error {
	periodID, err := strconv.ParseUint(c.Query("period_id"), 10, 64)
	if err != nil || periodID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid period_id provided.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	trial, err := h.svc.TrialBalance(c, *organizationID, periodID)
	if err != nil {
		return writeReportError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Trial balance retrieved successfully.", trial)
}

type AgingReportResponse struct {
	AsOf time.Time            `json:"as_of"`
	Rows []reporting.AgingRow `json:"rows"`
}

// @Summary AR/AP aging
// @Description Returns open customer and supplier balances aged into buckets (current, 1-30, 31-60, 61-90, 90+) as of a date, from open posted invoices.
// @Tags Reporting
// @Accept json
// @Produce json
// @Param as_of query string false "As-of date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} AgingReportResponse "Aging report retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reports/aging [get]
func (h ReportHandler) Aging(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	asOf := time.Now().UTC()
	if raw := c.Query("as_of"); raw != "" {
		parsed, err := helper.ParseDate(&raw)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "as_of must be in YYYY-MM-DD format.", nil)
		}
		asOf = *parsed
	}

	rows, err := h.svc.AgingReport(c, *organizationID, asOf)
	if err != nil {
		return writeReportError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Aging report retrieved successfully.", AgingReportResponse{
		AsOf: asOf,
		Rows: rows,
	})
}

type InventoryValuationResponse struct {
	Rows []reporting.InventoryValueRow `json:"rows"`
}

// @Summary Inventory valuation
// @Description Returns current on-hand inventory value per item from open valuation layers, reconciling to the inventory control account.
// @Tags Reporting
// @Accept json
// @Produce json
// @Success 200 {object} InventoryValuationResponse "Inventory valuation retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reports/inventory-valuation [get]
func (h ReportHandler) InventoryValuation(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	rows, err := h.svc.InventoryValuation(c, *organizationID)
	if err != nil {
		return writeReportError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Inventory valuation retrieved successfully.", InventoryValuationResponse{
		Rows: rows,
	})
}

func writeReportError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reporting.ErrPeriodNotFound):
		return httpx.CreateNotFoundResponse(c, "Tax period not found.")
	default:
		httpx.RequestLog(c).Error("report query failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to generate report.", err)
	}
}
