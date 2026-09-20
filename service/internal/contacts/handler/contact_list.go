package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListContactsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListContactsResponse `json:"data"`
}
type ListContactsResponse struct {
	Contacts []ContactResponse `json:"contacts"`
}

// @Summary List contacts
// @Description Lists contacts with pagination, sorting, and filtering, always scoped to the caller's organization via an enforced tenant filter. Filters are limited to an allowlist of contact fields, and results can be exported as CSV.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. organization_id:eq:1)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListContactsResponseEnvelope "Contacts retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts [get]
func (h ContactHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, contactQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("contact list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve contacts.", err)
	}

	items := make([]ContactResponse, len(page.Items))
	for i, contact := range page.Items {
		items[i] = newContactResponse(contact)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "contacts.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Contacts retrieved successfully.", ListContactsResponse{
		Contacts: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
