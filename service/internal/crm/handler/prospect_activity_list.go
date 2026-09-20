package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListCRMActivitiesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCRMActivitiesResponse `json:"data"`
}

type ListCRMActivitiesResponse struct {
	Activities []ProspectActivityResponse `json:"activities"`
}

// @Summary List activities
// @Description Lists CRM activities with pagination, sorting, and filtering by prospect, contact, type, summary, user, due date, or done status. Each activity carries its type, completion state, and timestamps, and results can be exported as CSV.
// @Tags CRM Activities
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. due_date:asc)"
// @Param filter query string false "Filters (repeatable, e.g. lead_id:eq:1)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListCRMActivitiesResponseEnvelope "Activities retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /crm/activities [get]
func (h ProspectActivityHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, prospectActivityQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	page, err := h.svc.ListInOrg(c, parsedQuery, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("crm activity list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve activities.", err)
	}

	items := make([]ProspectActivityResponse, len(page.Items))
	for i, activity := range page.Items {
		items[i] = newProspectActivityResponse(activity)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "crm-activities.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Activities retrieved successfully.", ListCRMActivitiesResponse{
		Activities: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
