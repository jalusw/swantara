package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type IntegrityHandler struct {
	svc accounting.ReportIntegrityService
}

func NewIntegrityHandler(svc accounting.ReportIntegrityService) IntegrityHandler {
	return IntegrityHandler{svc: svc}
}

type IntegrityCheckResponse struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

type IntegrityResponse struct {
	PeriodID uint64                   `json:"period_id"`
	Passed   bool                     `json:"passed"`
	Checks   []IntegrityCheckResponse `json:"checks"`
}

type IntegrityResponseEnvelope struct {
	httpx.EnvelopeBase
	Data IntegrityResponse `json:"data"`
}

// @Summary Check period integrity
// @Description Runs trial-balance, cash-flow, and equity cross-checks for a tax period
// @Tags Accounting Reports
// @Produce json
// @Param period_id query integer true "Tax period ID"
// @Success 200 {object} IntegrityResponseEnvelope "Integrity check completed."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid tax period"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /reports/integrity [get]
func (h IntegrityHandler) Check(c fiber.Ctx) error {
	orgID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}
	periodID, err := strconv.ParseUint(c.Query("period_id", c.Params("id")), 10, 64)
	if err != nil || periodID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid period_id provided.", nil)
	}
	report, err := h.svc.CheckPeriod(c, orgID, periodID)
	if err != nil {
		return writeTrialBalanceError(c, err)
	}
	checks := make([]IntegrityCheckResponse, len(report.Checks))
	for i, check := range report.Checks {
		checks[i] = IntegrityCheckResponse{Name: check.Name, Passed: check.Passed, Detail: check.Detail}
	}
	return httpx.CreateSuccessResponse(c, "Integrity check completed.", IntegrityResponse{PeriodID: report.PeriodID, Passed: report.Passed, Checks: checks})
}
