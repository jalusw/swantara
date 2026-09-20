package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

// @Summary List opportunities
// @Description Lists CRM opportunities (type opportunity only) scoped to the caller's organization with pagination, sorting, and filtering. The tenant scope is always enforced and only records typed as opportunities are returned, each enriched with its pipeline stage. Results can be exported as CSV.
// @Tags CRM Opportunities
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. expected_revenue:desc)"
// @Param filter query string false "Filters (repeatable, e.g. stage_id:eq:1)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListCRMOpportunitiesResponseEnvelope "Opportunities retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/opportunities [get]
func (h OpportunityHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, crmLeadQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	parsedQuery.Filters = append(parsedQuery.Filters, query.Filter{
		Field: "type", Operator: query.Equal, Value: crm.ProspectKindOpportunity,
	})

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("crm opportunity list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve opportunities.", err)
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve opportunities.", err)
	}

	items := make([]ProspectResponse, len(page.Items))
	for i, prospect := range page.Items {
		items[i] = newProspectResponse(prospect, stages)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "crm-opportunities.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Opportunities retrieved successfully.", ListCRMOpportunitiesResponse{
		Opportunities: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get opportunity
// @Description Gets a single CRM opportunity by id, returning 404 when the record does not exist, is not typed as an opportunity, or belongs to another organization. The returned record is enriched with its pipeline stage.
// @Tags CRM Opportunities
// @Accept json
// @Produce json
// @Param id path integer true "CRM opportunity ID"
// @Success 200 {object} GetCRMOpportunityResponseEnvelope "Opportunity retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Opportunity not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/opportunities/{id} [get]
func (h OpportunityHandler) Get(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid opportunity id provided.", nil)
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm opportunity lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get opportunity.", err)
	}
	if prospect == nil || prospect.Type != crm.ProspectKindOpportunity || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Opportunity not found.")
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get opportunity.", err)
	}

	return httpx.CreateSuccessResponse(c, "Opportunity retrieved successfully.", GetCRMOpportunityResponse{
		Opportunity: newProspectResponse(prospect, stages),
	})
}

// @Summary Create opportunity
// @Description Creates a CRM opportunity directly for the caller's organization, forcing the record type to opportunity and requiring a valid pipeline stage reference. The stage's probability is synced onto the record, and validation enforces a required name, valid references, non-negative revenue and priority, and a probability between 0 and 100.
// @Tags CRM Opportunities
// @Accept json
// @Produce json
// @Param body body CreateProspectRequest true "Opportunity details"
// @Success 201 {object} CreateCRMOpportunityResponseEnvelope "Opportunity created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown reference"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/opportunities [post]
func (h OpportunityHandler) Create(c fiber.Ctx) error {
	var request CreateProspectRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	expectedClose, err := helper.ParseDate(request.ExpectedClose)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expected close date provided.", nil)
	}

	prospect, err := h.svc.CreateLead(c, &crm.Prospect{
		OrganizationID:  httpx.TenantOrganizationID(c, request.OrganizationID),
		Name:            request.Name,
		Type:            crm.ProspectKindOpportunity,
		ContactID:       request.ContactID,
		ContactName:     request.ContactName,
		Email:           request.Email,
		Phone:           request.Phone,
		JobPosition:     request.JobPosition,
		StageID:         request.StageID,
		ExpectedRevenue: request.ExpectedRevenue,
		Probability:     request.Probability,
		Priority:        request.Priority,
		SalespersonID:   request.SalespersonID,
		SalesGroupID:    request.SalesGroupID,
		Source:          request.Source,
		Medium:          request.Medium,
		Campaign:        request.Campaign,
		ExpectedClose:   expectedClose,
	})
	if err != nil {
		return writeProspectError(c, err)
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create opportunity.", err)
	}

	return httpx.CreateCreatedResponse(c, "Opportunity created successfully.", CreateCRMOpportunityResponse{
		Opportunity: newProspectResponse(prospect, stages),
	})
}

// @Summary Update opportunity
// @Description Updates a CRM opportunity's attributes after verifying the record is typed as an opportunity and owned by the caller's organization, returning 404 otherwise. Reassigning a stage syncs its probability onto the record, and opportunities must always keep a stage assigned.
// @Tags CRM Opportunities
// @Accept json
// @Produce json
// @Param id path integer true "CRM opportunity ID"
// @Param body body UpdateProspectRequest true "Opportunity details"
// @Success 200 {object} UpdateCRMOpportunityResponseEnvelope "Opportunity updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Opportunity not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown reference"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/opportunities/{id} [put]
func (h OpportunityHandler) Update(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid opportunity id provided.", nil)
	}

	var request UpdateProspectRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm opportunity lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update opportunity.", err)
	}
	if prospect == nil || prospect.Type != crm.ProspectKindOpportunity || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Opportunity not found.")
	}

	expectedClose, err := helper.ParseDate(request.ExpectedClose)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expected close date provided.", nil)
	}

	prospect.Name = request.Name
	prospect.Type = crm.ProspectKindOpportunity
	prospect.ContactID = request.ContactID
	prospect.ContactName = request.ContactName
	prospect.Email = request.Email
	prospect.Phone = request.Phone
	prospect.JobPosition = request.JobPosition
	prospect.StageID = request.StageID
	prospect.ExpectedRevenue = request.ExpectedRevenue
	prospect.Probability = request.Probability
	prospect.Priority = request.Priority
	prospect.SalespersonID = request.SalespersonID
	prospect.SalesGroupID = request.SalesGroupID
	prospect.Source = request.Source
	prospect.Medium = request.Medium
	prospect.Campaign = request.Campaign
	prospect.ExpectedClose = expectedClose

	updated, err := h.svc.UpdateLead(c, prospect)
	if err != nil {
		return writeProspectError(c, err)
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update opportunity.", err)
	}

	return httpx.CreateSuccessResponse(c, "Opportunity updated successfully.", UpdateCRMOpportunityResponse{
		Opportunity: newProspectResponse(updated, stages),
	})
}

// @Summary Delete opportunity
// @Description Deletes a CRM opportunity by id after verifying the record is typed as an opportunity and owned by the caller's organization, returning 404 otherwise. The record is removed permanently and the endpoint responds with 204 No Content on success.
// @Tags CRM Opportunities
// @Accept json
// @Produce json
// @Param id path integer true "CRM opportunity ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Opportunity not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/opportunities/{id} [delete]
func (h OpportunityHandler) Delete(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid opportunity id provided.", nil)
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm opportunity lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete opportunity.", err)
	}
	if prospect == nil || prospect.Type != crm.ProspectKindOpportunity || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Opportunity not found.")
	}

	if err := h.svc.Delete(c, prospectID); err != nil {
		httpx.RequestLog(c).Error("crm opportunity deletion failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete opportunity.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
