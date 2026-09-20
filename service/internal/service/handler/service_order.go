package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/service"
)

type ServiceOrderResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	Name           string     `json:"name"`
	ContactID      *uint64    `json:"contact_id"`
	EquipmentID    *uint64    `json:"equipment_id"`
	ContractID     *uint64    `json:"contract_id"`
	Type           string     `json:"type"`
	Priority       int        `json:"priority"`
	State          string     `json:"state"`
	ScheduledDate  *time.Time `json:"scheduled_date"`
	TechnicianID   *uint64    `json:"technician_id"`
	InvoiceID      *uint64    `json:"invoice_id"`
	DimensionID    *uint64    `json:"dimension_id"`
	ReportedIssue  string     `json:"reported_issue"`
	Resolution     string     `json:"resolution"`
}

func newServiceOrderResponse(order *service.ServiceOrder) ServiceOrderResponse {
	return ServiceOrderResponse{
		ID: order.ID, OrganizationID: order.OrganizationID, Name: order.Name, ContactID: order.ContactID,
		EquipmentID: order.EquipmentID, ContractID: order.ContractID, Type: order.Type, Priority: order.Priority,
		State: order.State, ScheduledDate: order.ScheduledDate, TechnicianID: order.TechnicianID,
		InvoiceID: order.InvoiceID, DimensionID: order.DimensionID,
		ReportedIssue: order.ReportedIssue, Resolution: order.Resolution,
	}
}

var serviceOrderQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"contact_id":      {},
	"equipment_id":    {},
	"contract_id":     {},
	"type":            {},
	"state":           {},
	"technician_id":   {},
	"created_at":      {},
	"updated_at":      {},
}

type ListServiceOrdersResponse struct {
	ServiceOrders []ServiceOrderResponse `json:"service_orders"`
}

type ListServiceOrdersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListServiceOrdersResponse `json:"data"`
}

// @Summary List service orders
// @Description Lists service orders with pagination, sorting, and filtering. Results are scoped to the caller's organization.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Success 200 {object} ListServiceOrdersResponseEnvelope "Service orders retrieved successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders [get]
func (h ServiceHandler) ListOrders(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, serviceOrderQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListOrders(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("service order list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service orders.", err)
	}
	items := make([]ServiceOrderResponse, len(page.Items))
	for i, order := range page.Items {
		items[i] = newServiceOrderResponse(order)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Service orders retrieved successfully.", ListServiceOrdersResponse{
		ServiceOrders: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetServiceOrderResponse struct {
	ServiceOrder ServiceOrderResponse `json:"service_order"`
}

type GetServiceOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetServiceOrderResponse `json:"data"`
}

// @Summary Get service order
// @Description Gets a single service order by id. The order must belong to the caller's organization; a request for an order owned by another tenant returns 404.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service order ID"
// @Success 200 {object} GetServiceOrderResponseEnvelope "Service order retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Service order not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders/{id} [get]
func (h ServiceHandler) GetOrder(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	order, err := h.svc.FindOrder(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service order get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service order.", err)
	}
	if order == nil || !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service order not found.")
	}
	return httpx.CreateSuccessResponse(c, "Service order retrieved successfully.", GetServiceOrderResponse{
		ServiceOrder: newServiceOrderResponse(order),
	})
}

type ServiceOrderLineRequest struct {
	Type              string  `json:"type" validate:"required,oneof=part labor expense"`
	ItemID            *uint64 `json:"item_id"`
	Description       string  `json:"description"`
	Qty               float64 `json:"qty" validate:"required,gt=0"`
	UnitID            *uint64 `json:"unit_id"`
	UnitCost          float64 `json:"unit_cost"`
	UnitPrice         float64 `json:"unit_price"`
	Billable          bool    `json:"billable"`
	CoveredByWarranty bool    `json:"covered_by_warranty"`
}

type CreateServiceOrderRequest struct {
	OrganizationID *uint64                   `json:"organization_id"`
	Name           string                    `json:"name" validate:"required"`
	ContactID      *uint64                   `json:"contact_id"`
	EquipmentID    *uint64                   `json:"equipment_id"`
	ContractID     *uint64                   `json:"contract_id"`
	Type           string                    `json:"type" validate:"required,oneof=repair maintenance installation inspection"`
	Priority       int                       `json:"priority"`
	ScheduledDate  *time.Time                `json:"scheduled_date"`
	TechnicianID   *uint64                   `json:"technician_id"`
	DimensionID    *uint64                   `json:"dimension_id"`
	ReportedIssue  string                    `json:"reported_issue"`
	Lines          []ServiceOrderLineRequest `json:"lines" validate:"required,min=1,dive"`
}

type CreateServiceOrderResponse struct {
	ServiceOrder ServiceOrderResponse `json:"service_order"`
}

type CreateServiceOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateServiceOrderResponse `json:"data"`
}

// @Summary Create service order
// @Description Creates a service order in the new state with its part/labor/expense lines. Billable lines flow to an invoice on billing; covered-by-warranty parts are cost-only.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param body body CreateServiceOrderRequest true "Service order details"
// @Success 201 {object} CreateServiceOrderResponseEnvelope "Service order created successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders [post]
func (h ServiceHandler) CreateOrder(c fiber.Ctx) error {
	var request CreateServiceOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	lines := make([]service.LineRequest, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = service.LineRequest{
			Type: line.Type, ItemID: line.ItemID, Description: line.Description, Qty: line.Qty,
			UnitID: line.UnitID, UnitCost: line.UnitCost, UnitPrice: line.UnitPrice,
			Billable: line.Billable, CoveredByWarranty: line.CoveredByWarranty,
		}
	}
	created, err := h.svc.CreateOrder(c, service.CreateOrderRequest{
		OrganizationID: *organizationID,
		Name:           request.Name,
		ContactID:      request.ContactID,
		EquipmentID:    request.EquipmentID,
		ContractID:     request.ContractID,
		Type:           request.Type,
		Priority:       request.Priority,
		ScheduledDate:  request.ScheduledDate,
		TechnicianID:   request.TechnicianID,
		DimensionID:    request.DimensionID,
		ReportedIssue:  request.ReportedIssue,
		Lines:          lines,
	})
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Service order created successfully.", CreateServiceOrderResponse{
		ServiceOrder: newServiceOrderResponse(created),
	})
}

type ServiceOrderLineResponse struct {
	ID                uint64  `json:"id"`
	ServiceOrderID    uint64  `json:"service_order_id"`
	Type              string  `json:"type"`
	ItemID            *uint64 `json:"item_id"`
	Description       string  `json:"description"`
	Qty               float64 `json:"qty"`
	UnitID            *uint64 `json:"unit_id"`
	UnitCost          float64 `json:"unit_cost"`
	UnitPrice         float64 `json:"unit_price"`
	StockMovementID   *uint64 `json:"stock_movement_id"`
	Billable          bool    `json:"billable"`
	CoveredByWarranty bool    `json:"covered_by_warranty"`
}

func newServiceOrderLineResponse(line *service.ServiceOrderLine) ServiceOrderLineResponse {
	return ServiceOrderLineResponse{
		ID: line.ID, ServiceOrderID: line.ServiceOrderID, Type: line.Type, ItemID: line.ItemID,
		Description: line.Description, Qty: line.Qty, UnitID: line.UnitID, UnitCost: line.UnitCost,
		UnitPrice: line.UnitPrice, StockMovementID: line.StockMovementID, Billable: line.Billable,
		CoveredByWarranty: line.CoveredByWarranty,
	}
}

type ListServiceOrderLinesResponse struct {
	ServiceOrderLines []ServiceOrderLineResponse `json:"service_order_lines"`
}

type ListServiceOrderLinesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListServiceOrderLinesResponse `json:"data"`
}

// @Summary List service order lines
// @Description Lists the part/labor/expense lines of a service order.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service order ID"
// @Success 200 {object} ListServiceOrderLinesResponseEnvelope "Service order lines retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Service order not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders/{id}/lines [get]
func (h ServiceHandler) ListOrderLines(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	order, err := h.svc.FindOrder(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service order get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service order.", err)
	}
	if order == nil || !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service order not found.")
	}
	lines, err := h.svc.ListOrderLines(c, order.ID)
	if err != nil {
		httpx.RequestLog(c).Error("service order lines list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service order lines.", err)
	}
	items := make([]ServiceOrderLineResponse, len(lines))
	for i, line := range lines {
		items[i] = newServiceOrderLineResponse(line)
	}
	return httpx.CreateSuccessResponse(c, "Service order lines retrieved successfully.", ListServiceOrderLinesResponse{
		ServiceOrderLines: items,
	})
}
