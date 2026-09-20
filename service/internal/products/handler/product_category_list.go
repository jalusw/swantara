package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListProductCategoriesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListProductCategoriesResponse `json:"data"`
}
type ListProductCategoriesResponse struct {
	Categories []ItemCategoryResponse `json:"categories"`
}

// @Summary List item categories
// @Description Lists item categories with pagination, sorting, and filtering, including their GL account mappings, cost method, and valuation; the response can be exported as JSON, XML, or CSV. The tenant filter is always enforced so callers only see their own organization's categories.
// @Tags Item Categories
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. cost_method:eq:average)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListProductCategoriesResponseEnvelope "Categories retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/item-categories [get]
func (h ItemCategoryHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, itemCategoryQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.categories.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("item category list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve item categories.", err)
	}

	items := make([]ItemCategoryResponse, len(page.Items))
	for i, category := range page.Items {
		items[i] = newItemCategoryResponse(category)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "item-categories.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Item categories retrieved successfully.", ListProductCategoriesResponse{
		Categories: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
