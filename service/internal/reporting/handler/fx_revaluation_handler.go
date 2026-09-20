package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

type FxRevaluationHandler struct {
	svc reporting.FxRevaluationService
}

func NewFxRevaluationHandler(svc reporting.FxRevaluationService) FxRevaluationHandler {
	return FxRevaluationHandler{svc: svc}
}

type FxRevaluationResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID uint64    `json:"organization_id"`
	PeriodID       uint64    `json:"period_id"`
	Name           *string   `json:"name"`
	Date           time.Time `json:"date"`
	State          string    `json:"state"`
	TotalGainLoss  float64   `json:"total_gain_loss"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newFxRevaluationResponse(run *reporting.FxRevaluation) FxRevaluationResponse {
	return FxRevaluationResponse{
		ID:             run.ID,
		OrganizationID: run.OrganizationID,
		PeriodID:       run.PeriodID,
		Name:           run.Name,
		Date:           run.Date,
		State:          run.State,
		TotalGainLoss:  run.TotalGainLoss,
		CreatedAt:      run.CreatedAt,
		UpdatedAt:      run.UpdatedAt,
	}
}

type ListFxRevaluationsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []FxRevaluationResponse `json:"data"`
}

type RevalueRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	PeriodID       uint64  `json:"period_id" validate:"required,gt=0"`
	Date           string  `json:"date"`
}

// @Summary Run FX revaluation
// @Description Revalues open foreign-currency AR/AP positions at the closing rate for a period, posting unrealized FX gain/loss entries and reversing the prior period's unrealized entries.
// @Tags FX Revaluation
// @Accept json
// @Produce json
// @Param body body RevalueRequest true "FX revaluation details"
// @Success 200 {object} reporting.RevaluationResult "FX revaluation completed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/fx-revaluations [post]
func (h FxRevaluationHandler) Revalue(c fiber.Ctx) error {
	var request RevalueRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	date := time.Now().UTC()
	if request.Date != "" {
		parsed, err := helper.ParseDate(&request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
		}
		date = *parsed
	}

	result, err := h.svc.Revalue(c, *organizationID, request.PeriodID, date)
	if err != nil {
		return writeRevaluationError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "FX revaluation completed successfully.", result)
}

// @Summary List FX revaluations
// @Description Lists FX revaluation runs for the tenant.
// @Tags FX Revaluation
// @Accept json
// @Produce json
// @Success 200 {object} ListFxRevaluationsResponseEnvelope "FX revaluations retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/fx-revaluations [get]
func (h FxRevaluationHandler) List(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	runs, err := h.svc.ListRunsByOrganization(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("fx revaluation list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to list FX revaluations.", err)
	}
	items := make([]FxRevaluationResponse, len(runs))
	for i, run := range runs {
		items[i] = newFxRevaluationResponse(run)
	}
	return httpx.CreateSuccessResponse(c, "FX revaluations retrieved successfully.", items)
}

func writeRevaluationError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reporting.ErrConfigMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Reporting configuration is missing (journal or FX gain/loss account).", nil)
	case errors.Is(err, reporting.ErrClosingRateMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No closing rate available for a foreign currency position.", nil)
	case errors.Is(err, reporting.ErrBaseCurrencyMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization base currency is not set.", nil)
	default:
		httpx.RequestLog(c).Error("fx revaluation failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to run FX revaluation.", err)
	}
}
