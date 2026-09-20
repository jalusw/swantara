package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

type PlanningHandler struct {
	svc manufacturing.PlanningService
}

func NewPlanningHandler(svc manufacturing.PlanningService) PlanningHandler {
	return PlanningHandler{svc: svc}
}

// @Summary Run Planning
// @Description Runs a material requirements planning pass for the caller's organization over the given horizon, collecting demand from confirmed sale orders, demand forecasts, and reorder point candidates within the horizon. Demand is netted against on-hand and confirmed incoming stock, and any shortfalls generate planned manufacture or purchase orders, recursively exploding recipes to cover component demand. The completed run is returned together with its demands and planned orders.
// @Tags Planning
// @Accept json
// @Produce json
// @Param body body RunMrpRequest true "Planning run parameters"
// @Success 200 {object} PlanningRunResponseEnvelope "Planning run completed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/planning/runs [post]
func (h PlanningHandler) Run(c fiber.Ctx) error {
	var request RunMrpRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}
	runDate, err := helper.ParseDateOrToday(request.RunDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid run date provided.", nil)
	}
	run, err := h.svc.Run(c, *organizationID, request.HorizonDays, runDate)
	if err != nil {
		return writeMrpError(c, err)
	}
	demands, err := h.svc.ListDemands(c, run.ID)
	if err != nil {
		httpx.RequestLog(c).Error("planning run demands list failed", "planning_run_id", run.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to load planning demands.", err)
	}
	planned, err := h.svc.ListPlanned(c, run.ID)
	if err != nil {
		httpx.RequestLog(c).Error("planning run planned orders list failed", "planning_run_id", run.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to load planned orders.", err)
	}
	return httpx.CreateSuccessResponse(c, "Planning run completed successfully.", newPlanningRunResponse(run, demands, planned))
}

// @Summary Get Planning run
// @Description Gets a single tenant-scoped Planning run by id together with the demands it collected and the planned orders it generated. A 404 is returned when the run does not exist.
// @Tags Planning
// @Produce json
// @Param id path integer true "Planning run ID"
// @Success 200 {object} PlanningRunResponseEnvelope "Planning run retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Planning run not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/planning/runs/{id} [get]
func (h PlanningHandler) Get(c fiber.Ctx) error {
	runID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid planning run id.", nil)
	}
	run, err := h.svc.FindRun(c, runID)
	if err != nil {
		httpx.RequestLog(c).Error("planning run find failed", "planning_run_id", runID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to load planning run.", err)
	}
	if run == nil {
		return httpx.CreateNotFoundResponse(c, "Planning run not found.")
	}
	demands, err := h.svc.ListDemands(c, runID)
	if err != nil {
		httpx.RequestLog(c).Error("planning run demands list failed", "planning_run_id", runID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to load planning demands.", err)
	}
	planned, err := h.svc.ListPlanned(c, runID)
	if err != nil {
		httpx.RequestLog(c).Error("planning run planned orders list failed", "planning_run_id", runID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to load planned orders.", err)
	}
	return httpx.CreateSuccessResponse(c, "Planning run retrieved successfully.", newPlanningRunResponse(run, demands, planned))
}

// @Summary List Planning runs
// @Description Lists tenant-scoped Planning runs with pagination, returning each run's header without its demands or planned orders. The result set is always scoped to the caller's organization even when no organization_id is supplied.
// @Tags Planning
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Success 200 {object} ListPlanningRunsResponseEnvelope "Planning runs retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/planning/runs [get]
func (h PlanningHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, map[string]struct{}{
		"organization_id": {},
	})
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.ListRuns(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("planning run list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve Planning runs.", err)
	}

	items := make([]PlanningRunResponse, len(page.Items))
	for i, run := range page.Items {
		items[i] = newPlanningRunResponse(run, nil, nil)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Planning runs retrieved successfully.", ListPlanningRunsResponse{
		Runs: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
