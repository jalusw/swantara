package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type PipelineStageHandler struct {
	stages crm.PipelineStageService
}

func NewPipelineStageHandler(stages crm.PipelineStageService) PipelineStageHandler {
	return PipelineStageHandler{stages: stages}
}

type PipelineStageResponse struct {
	ID             uint64  `json:"id"`
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	Sequence       int     `json:"sequence"`
	IsWon          bool    `json:"is_won"`
	Probability    float64 `json:"probability"`
}

type ListPipelineStagesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPipelineStagesResponse `json:"data"`
}
type ListPipelineStagesResponse struct {
	Stages []PipelineStageResponse `json:"stages"`
}

// @Summary List CRM stages
// @Description Lists the caller's organization's CRM pipeline stages ordered by their sequence, including each stage's name, win flag, and probability. These stages define the pipeline used when creating, promoting, advancing, winning, or losing opportunities.
// @Tags CRM Stages
// @Accept json
// @Produce json
// @Success 200 {object} ListPipelineStagesResponseEnvelope "Stages retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/stages [get]
func (h PipelineStageHandler) List(c fiber.Ctx) error {
	parsedQuery := &query.Query{
		Sorts: []query.Sort{{Field: "sequence", Direction: query.Ascending}},
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.stages.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve stages.", err)
	}

	items := make([]PipelineStageResponse, len(page.Items))
	for i, stage := range page.Items {
		items[i] = PipelineStageResponse{
			ID:             stage.ID,
			OrganizationID: stage.OrganizationID,
			Name:           stage.Name,
			Sequence:       stage.Sequence,
			IsWon:          stage.IsWon,
			Probability:    stage.Probability,
		}
	}

	return httpx.CreateSuccessResponse(c, "Stages retrieved successfully.", ListPipelineStagesResponse{
		Stages: items,
	})
}

type SalesGroupHandler struct {
	teams crm.SalesGroupService
}

func NewSalesGroupHandler(teams crm.SalesGroupService) SalesGroupHandler {
	return SalesGroupHandler{teams: teams}
}

type SalesGroupResponse struct {
	ID             uint64  `json:"id"`
	Name           string  `json:"name"`
	LeaderID       *uint64 `json:"leader_id"`
	OrganizationID *uint64 `json:"organization_id"`
}

type ListSalesGroupsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSalesGroupsResponse `json:"data"`
}
type ListSalesGroupsResponse struct {
	Teams []SalesGroupResponse `json:"teams"`
}

// @Summary List sales teams
// @Description Lists CRM sales teams scoped to the caller's organization, including each team's leader and organization. The tenant filter is always enforced so callers only see their own organization's teams.
// @Tags CRM Teams
// @Accept json
// @Produce json
// @Success 200 {object} ListSalesGroupsResponseEnvelope "Sales teams retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/teams [get]
func (h SalesGroupHandler) List(c fiber.Ctx) error {
	parsedQuery := &query.Query{}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.teams.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("crm sales team list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve sales teams.", err)
	}

	items := make([]SalesGroupResponse, len(page.Items))
	for i, team := range page.Items {
		items[i] = SalesGroupResponse{
			ID:             team.ID,
			Name:           team.Name,
			LeaderID:       team.LeaderID,
			OrganizationID: team.OrganizationID,
		}
	}

	return httpx.CreateSuccessResponse(c, "Sales teams retrieved successfully.", ListSalesGroupsResponse{
		Teams: items,
	})
}
