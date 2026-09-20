package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

type ProductionOrderHandler struct {
	svc manufacturing.ProductionOrderService
}

func NewProductionOrderHandler(svc manufacturing.ProductionOrderService) ProductionOrderHandler {
	return ProductionOrderHandler{svc: svc}
}

// @Summary List production orders
// @Description Lists tenant-scoped production orders with pagination, sorting, and filtering on attributes such as item, recipe, state, and planned dates. Results can be exported in CSV format when requested. Returns the matching orders together with pagination metadata.
// @Tags Manufacturing Orders
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. state:eq:draft)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListProductionOrdersResponseEnvelope "Production orders retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders [get]
func (h ProductionOrderHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, manufacturingOrderQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("production order list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve production orders.", err)
	}

	items := make([]ProductionOrderResponse, len(page.Items))
	for i, productionOrder := range page.Items {
		items[i] = newProductionOrderResponse(productionOrder)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "manufacturing-orders.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Production orders retrieved successfully.", ListProductionOrdersResponse{
		ProductionOrders: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get production order
// @Description Gets a single tenant-scoped production order by id, including its planned dates, produced quantities, and state. A 404 is returned when the order does not exist or belongs to another organization.
// @Tags Manufacturing Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Success 200 {object} GetProductionOrderResponseEnvelope "Production order retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id} [get]
func (h ProductionOrderHandler) Get(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	productionOrder, err := h.svc.Find(c, productionOrderID)
	if err != nil {
		httpx.RequestLog(c).Error("production order lookup failed", "production_order_id", productionOrderID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get production order.", err)
	}
	if productionOrder == nil || !httpx.OwnsTenant(c, productionOrder.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Production order not found.")
	}

	return httpx.CreateSuccessResponse(c, "Production order retrieved successfully.", GetProductionOrderResponse{
		ProductionOrder: newProductionOrderResponse(productionOrder),
	})
}

// @Summary Create production order
// @Description Creates a draft production order and explodes its recipe into planned component lines, validating that the item variant and recipe exist and match, and that the source and destination locations are valid. A sequence number is generated for the order, and the order is saved together with its exploded components in draft state.
// @Tags Manufacturing Orders
// @Accept json
// @Produce json
// @Param body body CreateProductionOrderRequest true "Production order details"
// @Success 201 {object} CreateProductionOrderResponseEnvelope "Production order created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown references"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders [post]
func (h ProductionOrderHandler) Create(c fiber.Ctx) error {
	var request CreateProductionOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	productionOrder, err := h.svc.Create(c, &manufacturing.ProductionOrder{
		OrganizationID:    httpx.TenantOrganizationID(c, request.OrganizationID),
		ItemID:            request.ItemID,
		RecipeID:          request.RecipeID,
		QtyToProduce:      request.QtyToProduce,
		UnitID:            request.UnitID,
		SrcLocationID:     request.SrcLocationID,
		DstLocationID:     request.DstLocationID,
		DatePlannedStart:  request.DatePlannedStart,
		DatePlannedFinish: request.DatePlannedFinish,
		Origin:            request.Origin,
		Priority:          request.Priority,
	})
	if err != nil {
		return writeProductionOrderError(c, err)
	}

	components, err := h.svc.ListComponents(c, productionOrder.ID)
	if err != nil {
		httpx.RequestLog(c).Error("production order components list failed", "production_order_id", productionOrder.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve production order components.", err)
	}

	componentResponses := make([]ConsumedMaterialResponse, len(components))
	for i, component := range components {
		componentResponses[i] = newConsumedMaterialResponse(component)
	}

	return httpx.CreateCreatedResponse(c, "Production order created successfully.", CreateProductionOrderResponse{
		ProductionOrder: newProductionOrderResponse(productionOrder),
		Components:      componentResponses,
	})
}
