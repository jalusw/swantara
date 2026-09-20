package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

type SaleOrderHandler struct {
	svc sales.SaleOrderService
}

func NewSaleOrderHandler(svc sales.SaleOrderService) SaleOrderHandler {
	return SaleOrderHandler{svc: svc}
}

type SaleOrderResponse struct {
	ID             uint64                  `json:"id"`
	OrganizationID *uint64                 `json:"organization_id"`
	Name           *string                 `json:"name"`
	ContactID      uint64                  `json:"contact_id"`
	ShipAddressID  *uint64                 `json:"ship_address_id"`
	BillAddressID  *uint64                 `json:"bill_address_id"`
	PriceBookID    *uint64                 `json:"price_book_id"`
	CurrencyCode   *string                 `json:"currency_code"`
	SalespersonID  *uint64                 `json:"salesperson_id"`
	SalesGroupID   *uint64                 `json:"sales_group_id"`
	ProspectID     *uint64                 `json:"prospect_id"`
	WarehouseID    *uint64                 `json:"warehouse_id"`
	State          string                  `json:"state"`
	OrderDate      *time.Time              `json:"order_date"`
	ExpectedDate   *time.Time              `json:"expected_date"`
	ValidityDate   *time.Time              `json:"validity_date"`
	PaymentTermID  *uint64                 `json:"payment_term_id"`
	Incoterm       *string                 `json:"incoterm"`
	CustomerPORef  *string                 `json:"customer_po_ref"`
	AmountUntaxed  float64                 `json:"amount_untaxed"`
	AmountTax      float64                 `json:"amount_tax"`
	AmountTotal    float64                 `json:"amount_total"`
	InvoiceStatus  string                  `json:"invoice_status"`
	DeliveryStatus string                  `json:"delivery_status"`
	Note           *string                 `json:"note"`
	Lines          []SaleOrderLineResponse `json:"lines,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

type SaleOrderLineResponse struct {
	ID            uint64            `json:"id"`
	OrderID       uint64            `json:"order_id"`
	Sequence      int               `json:"sequence"`
	ItemID        *uint64           `json:"item_id"`
	Description   *string           `json:"description"`
	QtyOrdered    float64           `json:"qty_ordered"`
	QtyDelivered  float64           `json:"qty_delivered"`
	QtyInvoiced   float64           `json:"qty_invoiced"`
	UnitID        *uint64           `json:"unit_id"`
	UnitPrice     float64           `json:"unit_price"`
	DiscountPct   float64           `json:"discount_pct"`
	TaxIDs        helper.Int64Array `json:"tax_ids"`
	DimensionID   *uint64           `json:"dimension_id"`
	PriceSubtotal float64           `json:"price_subtotal"`
	PriceTax      float64           `json:"price_tax"`
	PriceTotal    float64           `json:"price_total"`
}

func newSaleOrderResponse(order *sales.SaleOrder) SaleOrderResponse {
	return SaleOrderResponse{
		ID:             order.ID,
		OrganizationID: order.OrganizationID,
		Name:           order.Name,
		ContactID:      order.ContactID,
		ShipAddressID:  order.ShipAddressID,
		BillAddressID:  order.BillAddressID,
		PriceBookID:    order.PriceBookID,
		CurrencyCode:   order.CurrencyCode,
		SalespersonID:  order.SalespersonID,
		SalesGroupID:   order.SalesGroupID,
		ProspectID:     order.ProspectID,
		WarehouseID:    order.WarehouseID,
		State:          order.State,
		OrderDate:      order.OrderDate,
		ExpectedDate:   order.ExpectedDate,
		ValidityDate:   order.ValidityDate,
		PaymentTermID:  order.PaymentTermID,
		Incoterm:       order.Incoterm,
		CustomerPORef:  order.CustomerPORef,
		AmountUntaxed:  order.AmountUntaxed,
		AmountTax:      order.AmountTax,
		AmountTotal:    order.AmountTotal,
		InvoiceStatus:  order.InvoiceStatus,
		DeliveryStatus: order.DeliveryStatus,
		Note:           order.Note,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

func newSaleOrderLineResponse(line *sales.SaleOrderLine) SaleOrderLineResponse {
	return SaleOrderLineResponse{
		ID:            line.ID,
		OrderID:       line.OrderID,
		Sequence:      line.Sequence,
		ItemID:        line.ItemID,
		Description:   line.Description,
		QtyOrdered:    line.QtyOrdered,
		QtyDelivered:  line.QtyDelivered,
		QtyInvoiced:   line.QtyInvoiced,
		UnitID:        line.UnitID,
		UnitPrice:     line.UnitPrice,
		DiscountPct:   line.DiscountPct,
		TaxIDs:        line.TaxIDs,
		DimensionID:   line.DimensionID,
		PriceSubtotal: line.PriceSubtotal,
		PriceTax:      line.PriceTax,
		PriceTotal:    line.PriceTotal,
	}
}

var saleOrderQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"contact_id":      {},
	"state":           {},
	"price_book_id":   {},
	"salesperson_id":  {},
	"sales_group_id":  {},
	"prospect_id":     {},
	"warehouse_id":    {},
	"order_date":      {},
	"expected_date":   {},
	"amount_untaxed":  {},
	"amount_tax":      {},
	"amount_total":    {},
	"invoice_status":  {},
	"delivery_status": {},
	"created_at":      {},
	"updated_at":      {},
}

type ListSaleOrdersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSaleOrdersResponse `json:"data"`
}
type ListSaleOrdersResponse struct {
	Orders []SaleOrderResponse `json:"orders"`
}

// @Summary List sale orders
// @Description Lists sale orders with pagination, sorting, and filtering, scoped to the caller's organization. Supports filtering by state, contact, salesperson, warehouse, and delivery or invoice status, and returns a CSV export when requested.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. state:eq:draft)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListSaleOrdersResponseEnvelope "Sale orders retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders [get]
func (h SaleOrderHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, saleOrderQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("sale order list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve sale orders.", err)
	}

	items := make([]SaleOrderResponse, len(page.Items))
	for i, order := range page.Items {
		items[i] = newSaleOrderResponse(order)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "sale-orders.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Sale orders retrieved successfully.", ListSaleOrdersResponse{
		Orders: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetSaleOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetSaleOrderResponse `json:"data"`
}
type GetSaleOrderResponse struct {
	Order SaleOrderResponse `json:"order"`
}

// @Summary Get sale order
// @Description Gets a single sale order by id together with its full set of lines, each with pricing, tax, delivery, and invoicing quantities. Sale orders not belonging to the caller's organization are treated as not found.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Success 200 {object} GetSaleOrderResponseEnvelope "Sale order retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id} [get]
func (h SaleOrderHandler) Get(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	order, err := h.svc.Find(c, orderID)
	if err != nil {
		if errors.Is(err, sales.ErrOrderNotFound) {
			return httpx.CreateNotFoundResponse(c, "Sale order not found.")
		}
		httpx.RequestLog(c).Error("sale order lookup failed", "order_id", orderID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get sale order.", err)
	}
	if !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Sale order not found.")
	}

	lines, err := h.svc.ListLines(c, orderID)
	if err != nil {
		httpx.RequestLog(c).Error("sale order line list failed", "order_id", orderID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get sale order.", err)
	}

	response := newSaleOrderResponse(order)
	response.Lines = make([]SaleOrderLineResponse, len(lines))
	for i, line := range lines {
		response.Lines[i] = newSaleOrderLineResponse(line)
	}

	return httpx.CreateSuccessResponse(c, "Sale order retrieved successfully.", GetSaleOrderResponse{
		Order: response,
	})
}

type SaleOrderLineRequest struct {
	Sequence    int               `json:"sequence"`
	ItemID      *uint64           `json:"item_id" validate:"required"`
	Description *string           `json:"description"`
	QtyOrdered  float64           `json:"qty_ordered" validate:"required,gt=0"`
	UnitID      *uint64           `json:"unit_id"`
	DiscountPct float64           `json:"discount_pct"`
	TaxIDs      helper.Int64Array `json:"tax_ids"`
	DimensionID *uint64           `json:"dimension_id"`
}

type CreateSaleOrderRequest struct {
	OrganizationID *uint64                `json:"organization_id"`
	ContactID      uint64                 `json:"contact_id" validate:"required,gt=0"`
	ShipAddressID  *uint64                `json:"ship_address_id"`
	BillAddressID  *uint64                `json:"bill_address_id"`
	PriceBookID    *uint64                `json:"price_book_id" validate:"required"`
	SalespersonID  *uint64                `json:"salesperson_id"`
	SalesGroupID   *uint64                `json:"sales_group_id"`
	ProspectID     *uint64                `json:"prospect_id"`
	WarehouseID    *uint64                `json:"warehouse_id" validate:"required"`
	OrderDate      *string                `json:"order_date"`
	ExpectedDate   *string                `json:"expected_date"`
	ValidityDate   *string                `json:"validity_date"`
	PaymentTermID  *uint64                `json:"payment_term_id"`
	Incoterm       *string                `json:"incoterm"`
	CustomerPORef  *string                `json:"customer_po_ref"`
	Note           *string                `json:"note"`
	Lines          []SaleOrderLineRequest `json:"lines"`
}

type CreateSaleOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateSaleOrderResponse `json:"data"`
}
type CreateSaleOrderResponse struct {
	Order SaleOrderResponse `json:"order"`
}

// @Summary Create sale order
// @Description Creates a draft sale order with its lines, contact, price_book, and warehouse. Prices are resolved from the price_book and amounts are computed per line. When a CRM opportunity is provided it must be won; walk-in sales can omit the opportunity.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param body body CreateSaleOrderRequest true "Sale order details"
// @Success 201 {object} CreateSaleOrderResponseEnvelope "Sale order created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown reference"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders [post]
func (h SaleOrderHandler) Create(c fiber.Ctx) error {
	var request CreateSaleOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	lines := make([]*sales.SaleOrderLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &sales.SaleOrderLine{
			Sequence:    line.Sequence,
			ItemID:      line.ItemID,
			Description: line.Description,
			QtyOrdered:  line.QtyOrdered,
			UnitID:      line.UnitID,
			DiscountPct: line.DiscountPct,
			TaxIDs:      line.TaxIDs,
			DimensionID: line.DimensionID,
		}
	}

	orderDate, err := parseSaleDate(request.OrderDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid order date provided.", nil)
	}
	expectedDate, err := parseSaleDate(request.ExpectedDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expected date provided.", nil)
	}
	validityDate, err := parseSaleDate(request.ValidityDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid validity date provided.", nil)
	}

	order, err := h.svc.Create(c, &sales.SaleOrder{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		ContactID:      request.ContactID,
		ShipAddressID:  request.ShipAddressID,
		BillAddressID:  request.BillAddressID,
		PriceBookID:    request.PriceBookID,
		SalespersonID:  request.SalespersonID,
		SalesGroupID:   request.SalesGroupID,
		ProspectID:     request.ProspectID,
		WarehouseID:    request.WarehouseID,
		OrderDate:      orderDate,
		ExpectedDate:   expectedDate,
		ValidityDate:   validityDate,
		PaymentTermID:  request.PaymentTermID,
		Incoterm:       request.Incoterm,
		CustomerPORef:  request.CustomerPORef,
		Note:           request.Note,
	}, lines)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Sale order created successfully.", CreateSaleOrderResponse{
		Order: newSaleOrderResponse(order),
	})
}

type UpdateSaleOrderRequest struct {
	OrganizationID *uint64                `json:"organization_id"`
	ContactID      uint64                 `json:"contact_id" validate:"required,gt=0"`
	ShipAddressID  *uint64                `json:"ship_address_id"`
	BillAddressID  *uint64                `json:"bill_address_id"`
	PriceBookID    *uint64                `json:"price_book_id" validate:"required"`
	SalespersonID  *uint64                `json:"salesperson_id"`
	SalesGroupID   *uint64                `json:"sales_group_id"`
	WarehouseID    *uint64                `json:"warehouse_id" validate:"required"`
	OrderDate      *string                `json:"order_date"`
	ExpectedDate   *string                `json:"expected_date"`
	ValidityDate   *string                `json:"validity_date"`
	PaymentTermID  *uint64                `json:"payment_term_id"`
	Incoterm       *string                `json:"incoterm"`
	CustomerPORef  *string                `json:"customer_po_ref"`
	Note           *string                `json:"note"`
	Lines          []SaleOrderLineRequest `json:"lines"`
}

type UpdateSaleOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateSaleOrderResponse `json:"data"`
}
type UpdateSaleOrderResponse struct {
	Order SaleOrderResponse `json:"order"`
}

// @Summary Update draft sale order
// @Description Updates a draft sale order by replacing its lines and recomputing untaxed, tax, and total amounts from the price_book. Only orders still in draft can be updated, and the order's name, currency, and source opportunity are preserved.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Param body body UpdateSaleOrderRequest true "Sale order details"
// @Success 200 {object} UpdateSaleOrderResponseEnvelope "Sale order updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 409 {object} httpx.ErrorResponse "Sale order is not in draft state"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown reference"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id} [put]
func (h SaleOrderHandler) Update(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	var request UpdateSaleOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	existing, err := h.svc.Find(c, orderID)
	if err != nil {
		if errors.Is(err, sales.ErrOrderNotFound) {
			return httpx.CreateNotFoundResponse(c, "Sale order not found.")
		}
		httpx.RequestLog(c).Error("sale order lookup failed", "order_id", orderID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update sale order.", err)
	}
	if !httpx.OwnsTenant(c, existing.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Sale order not found.")
	}

	lines := make([]*sales.SaleOrderLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &sales.SaleOrderLine{
			Sequence:    line.Sequence,
			ItemID:      line.ItemID,
			Description: line.Description,
			QtyOrdered:  line.QtyOrdered,
			UnitID:      line.UnitID,
			DiscountPct: line.DiscountPct,
			TaxIDs:      line.TaxIDs,
			DimensionID: line.DimensionID,
		}
	}

	orderDate, err := parseSaleDate(request.OrderDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid order date provided.", nil)
	}
	expectedDate, err := parseSaleDate(request.ExpectedDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expected date provided.", nil)
	}
	validityDate, err := parseSaleDate(request.ValidityDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid validity date provided.", nil)
	}

	updated, err := h.svc.UpdateDraft(c, orderID, &sales.SaleOrder{
		OrganizationID: httpx.TenantOrganizationID(c, existing.OrganizationID),
		ContactID:      request.ContactID,
		ShipAddressID:  request.ShipAddressID,
		BillAddressID:  request.BillAddressID,
		PriceBookID:    request.PriceBookID,
		SalespersonID:  request.SalespersonID,
		SalesGroupID:   request.SalesGroupID,
		WarehouseID:    request.WarehouseID,
		OrderDate:      orderDate,
		ExpectedDate:   expectedDate,
		ValidityDate:   validityDate,
		PaymentTermID:  request.PaymentTermID,
		Incoterm:       request.Incoterm,
		CustomerPORef:  request.CustomerPORef,
		Note:           request.Note,
	}, lines)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Sale order updated successfully.", UpdateSaleOrderResponse{
		Order: newSaleOrderResponse(updated),
	})
}

// @Summary Delete sale order
// @Description Permanently deletes a sale order by id. Only sale orders belonging to the caller's organization can be deleted, and unknown orders return not found.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id} [delete]
func (h SaleOrderHandler) Delete(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	order, err := h.svc.Find(c, orderID)
	if err != nil {
		if errors.Is(err, sales.ErrOrderNotFound) {
			return httpx.CreateNotFoundResponse(c, "Sale order not found.")
		}
		httpx.RequestLog(c).Error("sale order lookup failed", "order_id", orderID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete sale order.", err)
	}
	if !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Sale order not found.")
	}

	if err := h.svc.Delete(c, orderID); err != nil {
		httpx.RequestLog(c).Error("sale order deletion failed", "order_id", orderID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete sale order.", err)
	}

	return httpx.CreateNoContentResponse(c)
}

// @Summary Send sale order
// @Description Sends a draft sale order, moving it from draft to sent so the customer can review and confirm it. Only draft orders can be sent; otherwise a conflict is returned.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Success 200 {object} GetSaleOrderResponseEnvelope "Sale order sent successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 409 {object} httpx.ErrorResponse "Sale order cannot be sent in its current state"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id}/send [post]
func (h SaleOrderHandler) Send(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	order, err := h.svc.Send(c, orderID)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Sale order sent successfully.", GetSaleOrderResponse{
		Order: newSaleOrderResponse(order),
	})
}

// @Summary Confirm sale order
// @Description Confirms a draft or sent sale order by checking stock availability, reserving quantities, and generating the outgoing shipment with its stock movements. Confirmation fails with a conflict when stock is insufficient or the order is not in a confirmable state.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Success 200 {object} GetSaleOrderResponseEnvelope "Sale order confirmed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 409 {object} httpx.ErrorResponse "Sale order cannot be confirmed or insufficient stock"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or missing configuration"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id}/confirm [post]
func (h SaleOrderHandler) Confirm(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	order, err := h.svc.Confirm(c, orderID)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Sale order confirmed successfully.", GetSaleOrderResponse{
		Order: newSaleOrderResponse(order),
	})
}

// @Summary Cancel sale order
// @Description Cancels a draft, sent, or confirmed sale order, releasing any stock reservations and cancelling the order's shipments and their movements. Orders already marked done cannot be cancelled.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Success 200 {object} GetSaleOrderResponseEnvelope "Sale order cancelled successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 409 {object} httpx.ErrorResponse "Sale order cannot be cancelled in its current state"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id}/cancel [post]
func (h SaleOrderHandler) Cancel(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	order, err := h.svc.Cancel(c, orderID)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Sale order cancelled successfully.", GetSaleOrderResponse{
		Order: newSaleOrderResponse(order),
	})
}

// @Summary Mark sale order done
// @Description Marks a confirmed sale order as done, completing the order lifecycle. Only confirmed orders can be marked done.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Success 200 {object} GetSaleOrderResponseEnvelope "Sale order marked done successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 409 {object} httpx.ErrorResponse "Sale order cannot be marked done in its current state"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id}/done [post]
func (h SaleOrderHandler) Done(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	order, err := h.svc.Done(c, orderID)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Sale order marked done successfully.", GetSaleOrderResponse{
		Order: newSaleOrderResponse(order),
	})
}

// @Summary Recompute sale order statuses
// @Description Recomputes the delivery and invoice statuses of a sale order from the ordered, delivered, and invoiced quantities on its lines, deriving pending, partial, done, and invoiced statuses accordingly.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Success 200 {object} GetSaleOrderResponseEnvelope "Sale order statuses recomputed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id}/recompute-statuses [post]
func (h SaleOrderHandler) RecomputeStatuses(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	order, err := h.svc.RecomputeStatuses(c, orderID)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Sale order statuses recomputed successfully.", GetSaleOrderResponse{
		Order: newSaleOrderResponse(order),
	})
}

func parseSaleDate(value *string) (*time.Time, error) {
	return helper.ParseDate(value)
}

func writeSaleOrderError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, sales.ErrOrderNotFound):
		return httpx.CreateNotFoundResponse(c, "Sale order not found.")
	case errors.Is(err, sales.ErrOrderState):
		return httpx.CreateConflictResponse(c, "Sale order cannot be processed in its current state.", err)
	case errors.Is(err, sales.ErrOrderNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order must have at least one line.", nil)
	case errors.Is(err, sales.ErrOrderQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order line quantity must be positive.", nil)
	case errors.Is(err, sales.ErrOrderDiscount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order line discount must be between 0 and 100.", nil)
	case errors.Is(err, sales.ErrOrderContact):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order contact does not exist.", nil)
	case errors.Is(err, sales.ErrOrderPriceBook):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order price_book does not exist.", nil)
	case errors.Is(err, sales.ErrOrderWarehouse):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order warehouse does not exist.", nil)
	case errors.Is(err, sales.ErrOrderContactOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order contact does not belong to this organization.", nil)
	case errors.Is(err, sales.ErrOrderPriceBookOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order price_book does not belong to this organization.", nil)
	case errors.Is(err, sales.ErrOrderWarehouseOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order warehouse does not belong to this organization.", nil)
	case errors.Is(err, sales.ErrOrderVariant):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order item variant does not exist.", nil)
	case errors.Is(err, sales.ErrOrderVariantOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order line item does not belong to this organization.", nil)
	case errors.Is(err, sales.ErrOrderLead):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order source opportunity does not exist.", nil)
	case errors.Is(err, sales.ErrOrderLeadOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order source opportunity does not belong to this organization.", nil)
	case errors.Is(err, sales.ErrOrderLeadNotWon):
		return httpx.CreateConflictResponse(c, "Sale order source opportunity must be won.", err)
	case errors.Is(err, sales.ErrOrderTax):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order tax does not exist.", nil)
	case errors.Is(err, sales.ErrOrderTaxInvalid):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order tax is not a sale tax.", nil)
	case errors.Is(err, sales.ErrOrderTaxOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order tax does not belong to this organization.", nil)
	case errors.Is(err, sales.ErrOrderCurrency):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order currency does not exist.", nil)
	case errors.Is(err, sales.ErrOrderLocation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Sale order warehouse stock location not found.", nil)
	case errors.Is(err, sales.ErrOrderCustomerLocation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No customer stock location configured.", nil)
	case errors.Is(err, sales.ErrOrderStockUnavailable):
		return httpx.CreateConflictResponse(c, "Sale order line quantity is not available to reserve.", err)
	case errors.Is(err, sales.ErrBalanceNotFound):
		return httpx.CreateConflictResponse(c, "No stock quant found to reserve.", err)
	case errors.Is(err, sales.ErrHoldOverflow):
		return httpx.CreateConflictResponse(c, "Insufficient stock to reserve this sale order.", err)
	case errors.Is(err, sales.ErrOrderShipmentNotFound):
		return httpx.CreateConflictResponse(c, "No outgoing shipment found for this sale order.", err)
	case errors.Is(err, sales.ErrOrderNothingToDeliver):
		return httpx.CreateConflictResponse(c, "Sale order has nothing ready to deliver.", err)
	case errors.Is(err, sales.ErrOrderNoInvoices):
		return httpx.CreateConflictResponse(c, "No open invoices found for this customer.", err)
	case errors.Is(err, accounting.ErrNoReceivableAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No receivable account configured for the organization.", nil)
	case errors.Is(err, accounting.ErrNoRevenueAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No revenue account configured for the item.", nil)
	case errors.Is(err, accounting.ErrInvoiceNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invoice must have at least one line.", nil)
	case errors.Is(err, accounting.ErrInvoiceTaxInvalid):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invoice line references an invalid sale tax.", nil)
	case errors.Is(err, accounting.ErrInvoiceSequence):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invoice document sequence is not configured.", nil)
	case errors.Is(err, sales.ErrNoIncomeAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category has no income account configured.", nil)
	case errors.Is(err, accounting.ErrPaymentAmount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payment amount must be positive.", nil)
	case errors.Is(err, accounting.ErrPaymentNoInvoices):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payment must allocate at least one invoice.", nil)
	case errors.Is(err, accounting.ErrOverAllocation):
		return httpx.CreateConflictResponse(c, "Payment exceeds the total open invoice balance.", err)
	case errors.Is(err, accounting.ErrNoBankAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No bank or cash account configured for the journal.", nil)
	default:
		httpx.RequestLog(c).Error("sale order write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process sale order.", err)
	}
}
