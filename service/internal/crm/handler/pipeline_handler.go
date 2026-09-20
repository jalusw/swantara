package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type PipelineHandler struct {
	svc crm.PipelineService
}

func NewPipelineHandler(svc crm.PipelineService) PipelineHandler {
	return PipelineHandler{svc: svc}
}

type GetPipelineResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPipelineResponse `json:"data"`
}
type GetPipelineResponse struct {
	Forecast crm.PipelineForecast `json:"forecast"`
}

// @Summary Get CRM pipeline forecast
// @Description Returns the real-time weighted pipeline forecast for the caller's organization, summing the expected revenue of open opportunities weighted by each stage's probability and computing the overall win rate from won and lost records. The response also breaks down expected and weighted revenue per stage, including open opportunity counts.
// @Tags CRM Pipeline
// @Accept json
// @Produce json
// @Success 200 {object} GetPipelineResponseEnvelope "Pipeline forecast retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/pipeline [get]
func (h PipelineHandler) Forecast(c fiber.Ctx) error {
	var organizationID *uint64
	if id, ok := httpx.CallerOrganizationID(c); ok {
		organizationID = helper.Ptr(id)
	}
	forecast, err := h.svc.Forecast(c, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("crm pipeline forecast failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve pipeline forecast.", err)
	}

	return httpx.CreateSuccessResponse(c, "Pipeline forecast retrieved successfully.", GetPipelineResponse{
		Forecast: forecast,
	})
}
