package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

var crmLeadQueryAllowlist = map[string]struct{}{
	"organization_id":  {},
	"name":             {},
	"contact_id":       {},
	"contact_name":     {},
	"email":            {},
	"phone":            {},
	"job_position":     {},
	"stage_id":         {},
	"expected_revenue": {},
	"probability":      {},
	"priority":         {},
	"salesperson_id":   {},
	"sales_group_id":   {},
	"source":           {},
	"medium":           {},
	"campaign":         {},
	"expected_close":   {},
	"closed_at":        {},
	"created_at":       {},
	"updated_at":       {},
}

type ListProspectsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListProspectsResponse `json:"data"`
}
type ListProspectsResponse struct {
	Leads []ProspectResponse `json:"leads"`
}

// @Summary List leads
// @Description Lists CRM leads (type prospect only) scoped to the caller's organization with pagination, sorting, and filtering. The tenant scope is always enforced and filters are limited to an allowlist of prospect fields, and each returned prospect is enriched with its pipeline stage. Results can be exported as CSV.
// @Tags CRM Leads
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. stage_id:eq:1)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListProspectsResponseEnvelope "Leads retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/leads [get]
func (h ProspectHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, crmLeadQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	parsedQuery.Filters = append(parsedQuery.Filters, query.Filter{
		Field: "type", Operator: query.Equal, Value: crm.ProspectKindLead,
	})

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("crm prospect list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve leads.", err)
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve leads.", err)
	}

	items := make([]ProspectResponse, len(page.Items))
	for i, prospect := range page.Items {
		items[i] = newProspectResponse(prospect, stages)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "crm-leads.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Leads retrieved successfully.", ListProspectsResponse{
		Leads: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetProspectResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetProspectResponse `json:"data"`
}
type GetProspectResponse struct {
	Prospect ProspectResponse `json:"prospect"`
}

// @Summary Get prospect
// @Description Gets a single CRM prospect by id after enforcing tenant ownership, returning 404 when the prospect does not exist or belongs to another organization. The returned record is enriched with its pipeline stage.
// @Tags CRM Leads
// @Accept json
// @Produce json
// @Param id path integer true "CRM prospect ID"
// @Success 200 {object} GetProspectResponseEnvelope "Prospect retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Prospect not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/leads/{id} [get]
func (h ProspectHandler) Get(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid prospect id provided.", nil)
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm prospect lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get prospect.", err)
	}
	if prospect == nil || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Prospect not found.")
	}

	stages, err := loadStageMap(c, h.stages)
	if err != nil {
		httpx.RequestLog(c).Error("crm stage lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get prospect.", err)
	}

	return httpx.CreateSuccessResponse(c, "Prospect retrieved successfully.", GetProspectResponse{
		Prospect: newProspectResponse(prospect, stages),
	})
}

type CreateProspectRequest struct {
	OrganizationID  *uint64 `json:"organization_id"`
	Name            string  `json:"name" validate:"required"`
	Type            string  `json:"type"`
	ContactID       *uint64 `json:"contact_id"`
	ContactName     *string `json:"contact_name"`
	Email           *string `json:"email"`
	Phone           *string `json:"phone"`
	JobPosition     *string `json:"job_position"`
	StageID         *uint64 `json:"stage_id"`
	ExpectedRevenue float64 `json:"expected_revenue"`
	Probability     float64 `json:"probability"`
	Priority        int16   `json:"priority"`
	SalespersonID   *uint64 `json:"salesperson_id"`
	SalesGroupID    *uint64 `json:"sales_group_id"`
	Source          *string `json:"source"`
	Medium          *string `json:"medium"`
	Campaign        *string `json:"campaign"`
	ExpectedClose   *string `json:"expected_close"`
}

type CreateProspectResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateProspectResponse `json:"data"`
}
type CreateProspectResponse struct {
	Prospect ProspectResponse `json:"prospect"`
}

// @Summary Create prospect
// @Description Creates a CRM prospect or opportunity for the caller's organization, optionally assigning a pipeline stage and sales context such as salesperson, sales team, source, and expected close date. When a stage is referenced its probability is synced onto the record, and opportunities are required to have a stage assigned. Validation enforces a required name, a valid prospect or opportunity type, non-negative revenue and priority, and a probability between 0 and 100.
// @Tags CRM Leads
// @Accept json
// @Produce json
// @Param body body CreateProspectRequest true "Prospect details"
// @Success 201 {object} CreateProspectResponseEnvelope "Prospect created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown reference"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/leads [post]
func (h ProspectHandler) Create(c fiber.Ctx) error {
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
		Type:            request.Type,
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
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create prospect.", err)
	}

	return httpx.CreateCreatedResponse(c, "Prospect created successfully.", CreateProspectResponse{
		Prospect: newProspectResponse(prospect, stages),
	})
}

type UpdateProspectRequest struct {
	OrganizationID  *uint64 `json:"organization_id"`
	Name            string  `json:"name" validate:"required"`
	Type            string  `json:"type"`
	ContactID       *uint64 `json:"contact_id"`
	ContactName     *string `json:"contact_name"`
	Email           *string `json:"email"`
	Phone           *string `json:"phone"`
	JobPosition     *string `json:"job_position"`
	StageID         *uint64 `json:"stage_id"`
	ExpectedRevenue float64 `json:"expected_revenue"`
	Probability     float64 `json:"probability"`
	Priority        int16   `json:"priority"`
	SalespersonID   *uint64 `json:"salesperson_id"`
	SalesGroupID    *uint64 `json:"sales_group_id"`
	Source          *string `json:"source"`
	Medium          *string `json:"medium"`
	Campaign        *string `json:"campaign"`
	ExpectedClose   *string `json:"expected_close"`
}

type UpdateProspectResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateProspectResponse `json:"data"`
}
type UpdateProspectResponse struct {
	Prospect ProspectResponse `json:"prospect"`
}

// @Summary Update prospect
// @Description Updates a CRM prospect or opportunity's attributes after enforcing tenant ownership, returning 404 when the record is missing or not owned by the caller's organization. Reassigning a stage syncs its probability onto the record, and opportunities must always keep a stage assigned.
// @Tags CRM Leads
// @Accept json
// @Produce json
// @Param id path integer true "CRM prospect ID"
// @Param body body UpdateProspectRequest true "Prospect details"
// @Success 200 {object} UpdateProspectResponseEnvelope "Prospect updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Prospect not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown reference"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/leads/{id} [put]
func (h ProspectHandler) Update(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid prospect id provided.", nil)
	}

	var request UpdateProspectRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm prospect lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update prospect.", err)
	}
	if prospect == nil || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Prospect not found.")
	}

	expectedClose, err := helper.ParseDate(request.ExpectedClose)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expected close date provided.", nil)
	}

	prospect.OrganizationID = httpx.TenantOrganizationID(c, prospect.OrganizationID)
	prospect.Name = request.Name
	prospect.Type = request.Type
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
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update prospect.", err)
	}

	return httpx.CreateSuccessResponse(c, "Prospect updated successfully.", UpdateProspectResponse{
		Prospect: newProspectResponse(updated, stages),
	})
}

// @Summary Delete prospect
// @Description Deletes a CRM prospect by id after enforcing tenant ownership, returning 404 when the prospect does not exist or belongs to another organization. The record is removed permanently and the endpoint responds with 204 No Content on success.
// @Tags CRM Leads
// @Accept json
// @Produce json
// @Param id path integer true "CRM prospect ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Prospect not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/crm/leads/{id} [delete]
func (h ProspectHandler) Delete(c fiber.Ctx) error {
	prospectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid prospect id provided.", nil)
	}

	prospect, err := h.svc.Find(c, prospectID)
	if err != nil {
		httpx.RequestLog(c).Error("crm prospect lookup failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete prospect.", err)
	}
	if prospect == nil || !httpx.OwnsTenant(c, prospect.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Prospect not found.")
	}

	if err := h.svc.Delete(c, prospectID); err != nil {
		httpx.RequestLog(c).Error("crm prospect deletion failed", "lead_id", prospectID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete prospect.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
