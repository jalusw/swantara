package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

var itemQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"category_id":     {},
	"type":            {},
	"tracking":        {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}

type ListProductsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListProductsResponse `json:"data"`
}
type ListProductsResponse struct {
	Products []ItemResponse `json:"products"`
}

// @Summary List item templates
// @Description Lists tenant-scoped item templates with pagination, sorting, and filtering, automatically restricting results to the caller's organization; the response can be exported as JSON, XML, or CSV.
// @Tags Products
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. category_id:eq:1)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListProductsResponseEnvelope "Products retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/products [get]
func (h ProductHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, itemQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.ListTemplates(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("item template list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve products.", err)
	}

	items := make([]ItemResponse, len(page.Items))
	for i, template := range page.Items {
		items[i] = newItemResponse(template)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "products.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Products retrieved successfully.", ListProductsResponse{
		Products: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetProductResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetProductResponse `json:"data"`
}
type GetProductResponse struct {
	Item ItemResponse `json:"item"`
}

// @Summary Get item template
// @Description Returns a single item template by id with its pricing, UoMs, tracking, and sales or purchase descriptions. The template must belong to the caller's organization, otherwise a 404 is returned.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path integer true "Item template ID"
// @Success 200 {object} GetProductResponseEnvelope "Item retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Item not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/products/{id} [get]
func (h ProductHandler) Get(c fiber.Ctx) error {
	templateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid item id provided.", nil)
	}

	template, err := h.svc.FindTemplate(c, templateID)
	if err != nil {
		httpx.RequestLog(c).Error("item template lookup failed", "item_id", templateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get item.", err)
	}
	if template == nil || !httpx.OwnsTenant(c, template.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Item not found.")
	}

	return httpx.CreateSuccessResponse(c, "Item retrieved successfully.", GetProductResponse{
		Item: newItemResponse(template),
	})
}

type CreateProductRequest struct {
	OrganizationID  *uint64                    `json:"organization_id"`
	Name            string                     `json:"name" validate:"required"`
	CategoryID      *uint64                    `json:"category_id"`
	Type            string                     `json:"type"`
	UnitID          *uint64                    `json:"unit_id"`
	PurchaseUnitID  *uint64                    `json:"purchase_unit_id"`
	ListPrice       float64                    `json:"list_price"`
	StandardCost    float64                    `json:"standard_cost"`
	IsPurchasable   *bool                      `json:"is_purchasable"`
	IsSellable      *bool                      `json:"is_sellable"`
	IsManufactured  bool                       `json:"is_manufactured"`
	Tracking        string                     `json:"tracking"`
	Weight          float64                    `json:"weight"`
	Volume          float64                    `json:"volume"`
	HsCode          *string                    `json:"hs_code"`
	DescriptionSale *string                    `json:"description_sale"`
	DescriptionPur  *string                    `json:"description_purchase"`
	Active          *bool                      `json:"active"`
	Variants        []CreateVariantRequest     `json:"variants"`
	AttributeMatrix []products.AttributeOption `json:"attribute_matrix"`
}

type CreateVariantRequest struct {
	Sku           *string         `json:"sku"`
	Barcode       *string         `json:"barcode"`
	AttributeJSON json.RawMessage `json:"attribute_json" swaggertype:"object"`
	ExtraCost     float64         `json:"extra_cost"`
	Active        *bool           `json:"active"`
}

type CreateProductResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateProductResponse `json:"data"`
}
type CreateProductResponse struct {
	Item     ItemResponse          `json:"item"`
	Variants []ItemVariantResponse `json:"variants,omitempty"`
}

// @Summary Create item template with variants
// @Description Creates a item template and generates its variants either from an attribute matrix (cartesian item) or a provided variant list. The category, if given, must exist, the attribute matrix must contain at least one attribute, and SKUs must be unique; new products default to purchasable, sellable, and active.
// @Tags Products
// @Accept json
// @Produce json
// @Param body body CreateProductRequest true "Item details"
// @Success 201 {object} CreateProductResponseEnvelope "Item created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error, unknown category, or duplicate SKU"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/products [post]
func (h ProductHandler) Create(c fiber.Ctx) error {
	var request CreateProductRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	template := &products.Item{
		OrganizationID:  httpx.TenantOrganizationID(c, request.OrganizationID),
		Name:            request.Name,
		CategoryID:      request.CategoryID,
		Type:            request.Type,
		UnitID:          request.UnitID,
		PurchaseUnitID:  request.PurchaseUnitID,
		ListPrice:       request.ListPrice,
		StandardCost:    request.StandardCost,
		IsPurchasable:   true,
		IsSellable:      true,
		IsManufactured:  request.IsManufactured,
		Tracking:        request.Tracking,
		Weight:          request.Weight,
		Volume:          request.Volume,
		HsCode:          request.HsCode,
		DescriptionSale: request.DescriptionSale,
		DescriptionPur:  request.DescriptionPur,
		Active:          true,
	}
	if request.IsPurchasable != nil {
		template.IsPurchasable = *request.IsPurchasable
	}
	if request.IsSellable != nil {
		template.IsSellable = *request.IsSellable
	}
	if request.Active != nil {
		template.Active = *request.Active
	}

	created, err := h.svc.CreateTemplate(c, template)
	if err != nil {
		return writeProductError(c, err)
	}

	variants, err := h.createVariants(c, created.ID, request)
	if err != nil {
		return writeProductError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Item created successfully.", CreateProductResponse{
		Item:     newItemResponse(created),
		Variants: variants,
	})
}

func (h ProductHandler) createVariants(c fiber.Ctx, templateID uint64, request CreateProductRequest) ([]ItemVariantResponse, error) {
	if len(request.AttributeMatrix) > 0 {
		generated, err := h.svc.GenerateVariants(c, templateID, request.AttributeMatrix)
		if err != nil {
			return nil, err
		}
		variants := make([]ItemVariantResponse, len(generated))
		for i, variant := range generated {
			variants[i] = newItemVariantResponse(variant)
		}
		return variants, nil
	}

	variants := make([]ItemVariantResponse, 0, len(request.Variants))
	for _, item := range request.Variants {
		active := true
		if item.Active != nil {
			active = *item.Active
		}
		created, err := h.svc.CreateVariant(c, &products.ItemVariant{
			ItemID:        templateID,
			Sku:           item.Sku,
			Barcode:       item.Barcode,
			AttributeJSON: item.AttributeJSON,
			ExtraCost:     item.ExtraCost,
			Active:        active,
		})
		if err != nil {
			return nil, err
		}
		variants = append(variants, newItemVariantResponse(created))
	}
	return variants, nil
}

type UpdateProductRequest struct {
	Name            string  `json:"name" validate:"required"`
	CategoryID      *uint64 `json:"category_id"`
	Type            string  `json:"type"`
	UnitID          *uint64 `json:"unit_id"`
	PurchaseUnitID  *uint64 `json:"purchase_unit_id"`
	ListPrice       float64 `json:"list_price"`
	StandardCost    float64 `json:"standard_cost"`
	IsPurchasable   *bool   `json:"is_purchasable"`
	IsSellable      *bool   `json:"is_sellable"`
	IsManufactured  bool    `json:"is_manufactured"`
	Tracking        string  `json:"tracking"`
	Weight          float64 `json:"weight"`
	Volume          float64 `json:"volume"`
	HsCode          *string `json:"hs_code"`
	DescriptionSale *string `json:"description_sale"`
	DescriptionPur  *string `json:"description_purchase"`
	Active          *bool   `json:"active"`
}

type UpdateProductResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateProductResponse `json:"data"`
}
type UpdateProductResponse struct {
	Item ItemResponse `json:"item"`
}

// @Summary Update item template
// @Description Updates a item template's attributes, pricing, UoMs, tracking, and flags. The category, if given, must exist; a 404 is returned if the template does not belong to the caller's organization.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path integer true "Item template ID"
// @Param body body UpdateProductRequest true "Item details"
// @Success 200 {object} UpdateProductResponseEnvelope "Item updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Item not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown category"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/products/{id} [put]
func (h ProductHandler) Update(c fiber.Ctx) error {
	templateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid item id provided.", nil)
	}

	var request UpdateProductRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	template, err := h.svc.FindTemplate(c, templateID)
	if err != nil {
		httpx.RequestLog(c).Error("item template lookup failed", "item_id", templateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update item.", err)
	}
	if template == nil || !httpx.OwnsTenant(c, template.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Item not found.")
	}
	template.OrganizationID = httpx.TenantOrganizationID(c, template.OrganizationID)

	template.Name = request.Name
	template.CategoryID = request.CategoryID
	template.Type = request.Type
	template.UnitID = request.UnitID
	template.PurchaseUnitID = request.PurchaseUnitID
	template.ListPrice = request.ListPrice
	template.StandardCost = request.StandardCost
	template.IsManufactured = request.IsManufactured
	template.Tracking = request.Tracking
	template.Weight = request.Weight
	template.Volume = request.Volume
	template.HsCode = request.HsCode
	template.DescriptionSale = request.DescriptionSale
	template.DescriptionPur = request.DescriptionPur
	if request.IsPurchasable != nil {
		template.IsPurchasable = *request.IsPurchasable
	}
	if request.IsSellable != nil {
		template.IsSellable = *request.IsSellable
	}
	if request.Active != nil {
		template.Active = *request.Active
	}

	updated, err := h.svc.UpdateTemplate(c, template)
	if err != nil {
		return writeProductError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Item updated successfully.", UpdateProductResponse{
		Item: newItemResponse(updated),
	})
}

// @Summary Delete item template
// @Description Deletes a item template by id. The template must belong to the caller's organization; a 404 is returned otherwise.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path integer true "Item template ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Item not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/products/{id} [delete]
func (h ProductHandler) Delete(c fiber.Ctx) error {
	templateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid item id provided.", nil)
	}

	template, err := h.svc.FindTemplate(c, templateID)
	if err != nil {
		httpx.RequestLog(c).Error("item template lookup failed", "item_id", templateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete item.", err)
	}
	if template == nil || !httpx.OwnsTenant(c, template.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Item not found.")
	}

	if err := h.svc.DeleteTemplate(c, templateID); err != nil {
		httpx.RequestLog(c).Error("item template deletion failed", "item_id", templateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete item.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
