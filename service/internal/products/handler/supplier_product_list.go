package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListSupplierProductsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSupplierProductsResponse `json:"data"`
}
type ListSupplierProductsResponse struct {
	SupplierProducts []SupplierProductResponse `json:"supplier_products"`
}

// @Summary List supplier products
// @Description Lists supplier catalog entries with pagination, sorting, and filtering, including supplier pricing, lead time, and validity window; the response can be exported as JSON, XML, or CSV.
// @Tags Supplier Products
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. priority:asc)"
// @Param filter query string false "Filters (repeatable, e.g. item_id:eq:1)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListSupplierProductsResponseEnvelope "Supplier products retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-products [get]
func (h SupplierProductHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, supplierProductQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	page, err := h.svc.ListInOrg(c, parsedQuery, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("supplier item list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve supplier products.", err)
	}

	items := make([]SupplierProductResponse, len(page.Items))
	for i, offer := range page.Items {
		items[i] = newSupplierProductResponse(offer)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "supplier-products.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Supplier products retrieved successfully.", ListSupplierProductsResponse{
		SupplierProducts: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
