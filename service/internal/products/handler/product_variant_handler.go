package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

type ListItemVariantsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListItemVariantsResponse `json:"data"`
}
type ListItemVariantsResponse struct {
	Variants []ItemVariantResponse `json:"variants"`
}

// @Summary List item variants
// @Description Lists the variants of a item template, each with its SKU, barcode, attribute JSON, and extra cost. The template must belong to the caller's organization, otherwise a 404 is returned; results can be exported as JSON, XML, or CSV.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path integer true "Item template ID"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListItemVariantsResponseEnvelope "Variants retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Item not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/products/{id}/variants [get]
func (h ProductHandler) ListVariants(c fiber.Ctx) error {
	templateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid item id provided.", nil)
	}

	template, err := h.svc.FindTemplate(c, templateID)
	if err != nil {
		httpx.RequestLog(c).Error("item template lookup failed", "item_id", templateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve variants.", err)
	}
	if template == nil || !httpx.OwnsTenant(c, template.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Item not found.")
	}

	variants, err := h.svc.ListVariantsByTemplate(c, templateID)
	if err != nil {
		httpx.RequestLog(c).Error("item variant list failed", "item_id", templateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve variants.", err)
	}

	items := make([]ItemVariantResponse, len(variants))
	for i, variant := range variants {
		items[i] = newItemVariantResponse(variant)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "item-variants.csv", items)
	}

	return httpx.CreateSuccessResponse(c, "Variants retrieved successfully.", ListItemVariantsResponse{
		Variants: items,
	})
}

// @Summary Add item variant
// @Description Adds a variant with an optional SKU, barcode, attribute JSON, and extra cost to a item template. The template must belong to the caller's organization, the SKU, if provided, must be unique, and new variants default to active.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path integer true "Item template ID"
// @Param body body CreateVariantRequest true "Variant details"
// @Success 201 {object} ItemVariantResponseEnvelope "Variant created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Item not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or duplicate SKU"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/products/{id}/variants [post]
func (h ProductHandler) CreateVariant(c fiber.Ctx) error {
	templateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid item id provided.", nil)
	}

	var request CreateVariantRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	template, err := h.svc.FindTemplate(c, templateID)
	if err != nil {
		httpx.RequestLog(c).Error("item template lookup failed", "item_id", templateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create variant.", err)
	}
	if template == nil || !httpx.OwnsTenant(c, template.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Item not found.")
	}

	active := true
	if request.Active != nil {
		active = *request.Active
	}
	variant, err := h.svc.CreateVariant(c, &products.ItemVariant{
		ItemID:        templateID,
		Sku:           request.Sku,
		Barcode:       request.Barcode,
		AttributeJSON: request.AttributeJSON,
		ExtraCost:     request.ExtraCost,
		Active:        active,
	})
	if err != nil {
		return writeProductError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Variant created successfully.", newItemVariantResponse(variant))
}
