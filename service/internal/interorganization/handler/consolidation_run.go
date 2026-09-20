package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
)

type CreateConsolidationRunRequest struct {
	PeriodID          uint64 `json:"period_id" validate:"required,gt=0"`
	ReportingCurrency string `json:"reporting_currency" validate:"required,len=3"`
}

type CreateConsolidationRunResponse struct {
	ConsolidationRun ConsolidationRunResponse `json:"consolidation_run"`
}

type CreateConsolidationRunResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateConsolidationRunResponse `json:"data"`
}

// @Summary Create consolidation run
// @Description Creates a draft consolidation run for the caller's organization as the group parent.
// @Tags Consolidation
// @Accept json
// @Produce json
// @Param body body CreateConsolidationRunRequest true "Run details"
// @Success 201 {object} CreateConsolidationRunResponseEnvelope "Consolidation run created successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/consolidation-runs [post]
func (h ConsolidationHandler) CreateRun(c fiber.Ctx) error {
	var request CreateConsolidationRunRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	run, err := h.svc.CreateRun(c, interorganization.CreateConsolidationRunRequest{
		GroupOrganizationID: *organizationID,
		PeriodID:            request.PeriodID,
		ReportingCurrency:   request.ReportingCurrency,
	})
	if err != nil {
		return writeInterorganizationError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Consolidation run created successfully.", CreateConsolidationRunResponse{
		ConsolidationRun: newConsolidationRunResponse(run),
	})
}

type ListConsolidationRunsResponse struct {
	ConsolidationRuns []ConsolidationRunResponse `json:"consolidation_runs"`
}

type ListConsolidationRunsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListConsolidationRunsResponse `json:"data"`
}

// @Summary List consolidation runs
// @Description Lists the consolidation runs of the caller's organization.
// @Tags Consolidation
// @Accept json
// @Produce json
// @Success 200 {object} ListConsolidationRunsResponseEnvelope "Consolidation runs retrieved successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/consolidation-runs [get]
func (h ConsolidationHandler) ListRuns(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	runs, err := h.svc.ListRuns(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("consolidation runs list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve consolidation runs.", err)
	}
	items := make([]ConsolidationRunResponse, len(runs))
	for i, run := range runs {
		items[i] = newConsolidationRunResponse(run)
	}
	return httpx.CreateSuccessResponse(c, "Consolidation runs retrieved successfully.", ListConsolidationRunsResponse{
		ConsolidationRuns: items,
	})
}

type GetConsolidationRunResponse struct {
	ConsolidationRun ConsolidationRunResponse           `json:"consolidation_run"`
	Eliminations     []ConsolidationEliminationResponse `json:"eliminations"`
}

type GetConsolidationRunResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetConsolidationRunResponse `json:"data"`
}

// @Summary Get consolidation run
// @Description Retrieves a consolidation run and its recorded eliminations.
// @Tags Consolidation
// @Accept json
// @Produce json
// @Param id path integer true "Consolidation run ID"
// @Success 200 {object} GetConsolidationRunResponseEnvelope "Consolidation run retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Consolidation run not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/consolidation-runs/{id} [get]
func (h ConsolidationHandler) GetRun(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	run, eliminations, err := h.svc.GetRun(c, id, *organizationID)
	if err != nil {
		return writeInterorganizationError(c, err)
	}
	items := make([]ConsolidationEliminationResponse, len(eliminations))
	for i, elimination := range eliminations {
		items[i] = newConsolidationEliminationResponse(elimination)
	}
	return httpx.CreateSuccessResponse(c, "Consolidation run retrieved successfully.", GetConsolidationRunResponse{
		ConsolidationRun: newConsolidationRunResponse(run),
		Eliminations:     items,
	})
}

type RunConsolidationRunResponse struct {
	ConsolidationRun ConsolidationRunResponse           `json:"consolidation_run"`
	MemberBalances   []ConsolidatedBalanceResponse      `json:"member_balances"`
	Eliminations     []ConsolidationEliminationResponse `json:"eliminations"`
}

type RunConsolidationRunResponseEnvelope struct {
	httpx.EnvelopeBase
	Data RunConsolidationRunResponse `json:"data"`
}

// @Summary Run consolidation
// @Description Translates each group member's posted balances into the reporting currency and computes intra-group and unrealized-profit eliminations, marking the run done.
// @Tags Consolidation
// @Accept json
// @Produce json
// @Param id path integer true "Consolidation run ID"
// @Success 200 {object} RunConsolidationRunResponseEnvelope "Consolidation run completed successfully."
// @Failure 404 {object} httpx.ErrorResponse "Consolidation run not found"
// @Failure 422 {object} httpx.ErrorResponse "Run must be draft"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/consolidation-runs/{id}/run [post]
func (h ConsolidationHandler) Run(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	result, err := h.svc.Run(c, id, *organizationID)
	if err != nil {
		return writeInterorganizationError(c, err)
	}
	balances := make([]ConsolidatedBalanceResponse, len(result.MemberBalances))
	for i, balance := range result.MemberBalances {
		balances[i] = ConsolidatedBalanceResponse{
			OrganizationID: balance.OrganizationID, AccountID: balance.AccountID, Amount: balance.Amount,
		}
	}
	eliminations := make([]ConsolidationEliminationResponse, len(result.Eliminations))
	for i, elimination := range result.Eliminations {
		eliminations[i] = newConsolidationEliminationResponse(elimination)
	}
	return httpx.CreateSuccessResponse(c, "Consolidation run completed successfully.", RunConsolidationRunResponse{
		ConsolidationRun: newConsolidationRunResponse(result.Run),
		MemberBalances:   balances,
		Eliminations:     eliminations,
	})
}
