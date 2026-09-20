package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ProspectHandler struct {
	stages crm.PipelineStageService
	svc    crm.ProspectService
}

func NewProspectHandler(stages crm.PipelineStageService, svc crm.ProspectService) ProspectHandler {
	return ProspectHandler{stages: stages, svc: svc}
}

func writeProspectError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, crm.ErrLeadNotFound):
		return httpx.CreateNotFoundResponse(c, "Prospect not found.")
	case errors.Is(err, crm.ErrStageNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "CRM stage does not exist.", nil)
	case errors.Is(err, crm.ErrStageOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "CRM stage does not belong to this organization.", nil)
	case errors.Is(err, crm.ErrStageRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A pipeline stage is required for an opportunity.", nil)
	case errors.Is(err, crm.ErrContactNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "CRM contact does not exist.", nil)
	case errors.Is(err, crm.ErrSalespersonNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Salesperson does not exist.", nil)
	case errors.Is(err, crm.ErrSalesGroupNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sales team does not exist.", nil)
	case errors.Is(err, crm.ErrInvalidLeadType):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Prospect type must be prospect or opportunity.", nil)
	case errors.Is(err, crm.ErrInvalidProbability):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Probability must be between 0 and 100.", nil)
	case errors.Is(err, crm.ErrInvalidRevenue):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Expected revenue cannot be negative.", nil)
	case errors.Is(err, crm.ErrInvalidPriority):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Priority cannot be negative.", nil)
	case errors.Is(err, crm.ErrLeadNameRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Prospect name is required.", nil)
	case errors.Is(err, crm.ErrNotLead):
		return httpx.CreateConflictResponse(c, "Only a prospect can be promoted to an opportunity.", err)
	case errors.Is(err, crm.ErrNotOpportunity):
		return httpx.CreateConflictResponse(c, "Record is not an opportunity.", err)
	case errors.Is(err, crm.ErrLeadClosed):
		return httpx.CreateConflictResponse(c, "Prospect is already closed.", err)
	case errors.Is(err, crm.ErrNoWonStage):
		return httpx.CreateConflictResponse(c, "No won pipeline stage is configured.", err)
	case errors.Is(err, crm.ErrLostReasonRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A lost reason is required.", nil)
	default:
		httpx.RequestLog(c).Error("crm prospect write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save prospect.", err)
	}
}
