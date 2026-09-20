package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type PurchaseOrderHandler struct {
	svc procurement.PurchaseOrderService
}

func NewPurchaseOrderHandler(
	svc procurement.PurchaseOrderService,
) PurchaseOrderHandler {
	return PurchaseOrderHandler{svc: svc}
}

type PurchaseOrderResponse struct {
	ID             uint64                      `json:"id"`
	OrganizationID *uint64                     `json:"organization_id"`
	Name           *string                     `json:"name"`
	SupplierID     uint64                      `json:"supplier_id"`
	SupplierRef    *string                     `json:"supplier_ref"`
	CurrencyCode   *string                     `json:"currency_code"`
	WarehouseID    *uint64                     `json:"warehouse_id"`
	DestLocationID *uint64                     `json:"dest_location_id"`
	State          string                      `json:"state"`
	OrderDate      *time.Time                  `json:"order_date"`
	ExpectedDate   *time.Time                  `json:"expected_date"`
	PaymentTermID  *uint64                     `json:"payment_term_id"`
	Incoterm       *string                     `json:"incoterm"`
	AmountUntaxed  amount.Amount               `json:"amount_untaxed"`
	AmountTax      amount.Amount               `json:"amount_tax"`
	AmountTotal    amount.Amount               `json:"amount_total"`
	InvoiceStatus  string                      `json:"invoice_status"`
	ReceiptStatus  string                      `json:"receipt_status"`
	Lines          []PurchaseOrderLineResponse `json:"lines,omitempty"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}

type PurchaseOrderLineResponse struct {
	ID            uint64            `json:"id"`
	OrderID       uint64            `json:"order_id"`
	Sequence      int               `json:"sequence"`
	ItemID        *uint64           `json:"item_id"`
	Description   *string           `json:"description"`
	QtyOrdered    float64           `json:"qty_ordered"`
	QtyReceived   float64           `json:"qty_received"`
	QtyBilled     float64           `json:"qty_billed"`
	UnitID        *uint64           `json:"unit_id"`
	UnitPrice     float64           `json:"unit_price"`
	DiscountPct   float64           `json:"discount_pct"`
	TaxIDs        helper.Int64Array `json:"tax_ids"`
	DimensionID   *uint64           `json:"dimension_id"`
	PriceSubtotal float64           `json:"price_subtotal"`
}

func newPurchaseOrderResponse(order *procurement.PurchaseOrder) PurchaseOrderResponse {
	return PurchaseOrderResponse{
		ID:             order.ID,
		OrganizationID: order.OrganizationID,
		Name:           order.Name,
		SupplierID:     order.SupplierID,
		SupplierRef:    order.SupplierRef,
		CurrencyCode:   order.CurrencyCode,
		WarehouseID:    order.WarehouseID,
		DestLocationID: order.DestLocationID,
		State:          order.State,
		OrderDate:      order.OrderDate,
		ExpectedDate:   order.ExpectedDate,
		PaymentTermID:  order.PaymentTermID,
		Incoterm:       order.Incoterm,
		AmountUntaxed:  amount.FromFloat64(order.AmountUntaxed),
		AmountTax:      amount.FromFloat64(order.AmountTax),
		AmountTotal:    amount.FromFloat64(order.AmountTotal),
		InvoiceStatus:  order.InvoiceStatus,
		ReceiptStatus:  order.ReceiptStatus,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

func newPurchaseOrderLineResponse(line *procurement.PurchaseOrderLine) PurchaseOrderLineResponse {
	return PurchaseOrderLineResponse{
		ID:            line.ID,
		OrderID:       line.OrderID,
		Sequence:      line.Sequence,
		ItemID:        line.ItemID,
		Description:   line.Description,
		QtyOrdered:    line.QtyOrdered,
		QtyReceived:   line.QtyReceived,
		QtyBilled:     line.QtyBilled,
		UnitID:        line.UnitID,
		UnitPrice:     line.UnitPrice,
		DiscountPct:   line.DiscountPct,
		TaxIDs:        line.TaxIDs,
		DimensionID:   line.DimensionID,
		PriceSubtotal: line.PriceSubtotal,
	}
}

var purchaseOrderQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"supplier_id":     {},
	"state":           {},
	"warehouse_id":    {},
	"order_date":      {},
	"expected_date":   {},
	"amount_untaxed":  {},
	"amount_tax":      {},
	"amount_total":    {},
	"invoice_status":  {},
	"receipt_status":  {},
	"created_at":      {},
	"updated_at":      {},
}

type ListPurchaseOrdersResponse struct {
	Orders []PurchaseOrderResponse `json:"orders"`
}

type ListPurchaseOrdersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPurchaseOrdersResponse `json:"data"`
}

// @Summary List purchase orders
// @Description Lists purchase orders across the caller's organization with pagination, sorting, and filtering on attributes such as supplier, state, warehouse, order date, and invoice or receipt status. The result set is always scoped to the caller's organization even when no organization_id is supplied. Returns the matching orders together with pagination metadata.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListPurchaseOrdersResponseEnvelope "Purchase orders retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders [get]
func (h PurchaseOrderHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, purchaseOrderQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("purchase order list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve purchase orders.", err)
	}
	items := make([]PurchaseOrderResponse, len(page.Items))
	for i, order := range page.Items {
		items[i] = newPurchaseOrderResponse(order)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Purchase orders retrieved successfully.", ListPurchaseOrdersResponse{
		Orders: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetPurchaseOrderResponse struct {
	Order PurchaseOrderResponse `json:"order"`
}

type GetPurchaseOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPurchaseOrderResponse `json:"data"`
}

// @Summary Get purchase order
// @Description Gets a single purchase order by id together with all of its order lines, including ordered, received, and billed quantities and pricing. The lookup is scoped to the caller's organization, and a 404 is returned when the order does not exist or belongs to another tenant.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path integer true "Purchase order ID"
// @Success 200 {object} GetPurchaseOrderResponseEnvelope "Purchase order retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase order not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders/{id} [get]
func (h PurchaseOrderHandler) Get(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase order id provided.", nil)
	}
	order, err := h.svc.Find(c, orderID)
	if err != nil {
		httpx.RequestLog(c).Error("purchase order lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get purchase order.", err)
	}
	if order == nil || !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Purchase order not found.")
	}
	lines, err := h.svc.ListLines(c, orderID)
	if err != nil {
		httpx.RequestLog(c).Error("purchase order line list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get purchase order.", err)
	}
	response := newPurchaseOrderResponse(order)
	response.Lines = make([]PurchaseOrderLineResponse, len(lines))
	for i, line := range lines {
		response.Lines[i] = newPurchaseOrderLineResponse(line)
	}
	return httpx.CreateSuccessResponse(c, "Purchase order retrieved successfully.", GetPurchaseOrderResponse{
		Order: response,
	})
}

type PurchaseOrderLineRequest struct {
	Sequence    int               `json:"sequence"`
	ItemID      *uint64           `json:"item_id"`
	Description *string           `json:"description"`
	QtyOrdered  float64           `json:"qty_ordered" validate:"required,gt=0"`
	UnitID      *uint64           `json:"unit_id"`
	UnitPrice   float64           `json:"unit_price"`
	DiscountPct float64           `json:"discount_pct"`
	TaxIDs      helper.Int64Array `json:"tax_ids"`
	DimensionID *uint64           `json:"dimension_id"`
}

type CreatePurchaseOrderRequest struct {
	OrganizationID *uint64                    `json:"organization_id"`
	RequestID      *uint64                    `json:"request_id"`
	SupplierID     uint64                     `json:"supplier_id" validate:"required,gt=0"`
	SupplierRef    *string                    `json:"supplier_ref"`
	CurrencyCode   *string                    `json:"currency_code"`
	WarehouseID    *uint64                    `json:"warehouse_id" validate:"required"`
	DestLocationID *uint64                    `json:"dest_location_id"`
	OrderDate      *string                    `json:"order_date"`
	ExpectedDate   *string                    `json:"expected_date"`
	PaymentTermID  *uint64                    `json:"payment_term_id"`
	Incoterm       *string                    `json:"incoterm"`
	Lines          []PurchaseOrderLineRequest `json:"lines"`
}

type CreatePurchaseOrderResponse struct {
	Order PurchaseOrderResponse `json:"order"`
}

type CreatePurchaseOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreatePurchaseOrderResponse `json:"data"`
}

// @Summary Create purchase order
// @Description Creates a new draft purchase order, either directly from the supplied order lines or by converting an approved purchase requisition. It validates that the supplier is an active supplier and that the warehouse exists, resolves any missing unit prices from the supplier's catalog offers, and defaults the currency to IDR. Totals are recomputed across taxes and discounts, a sequence number is generated for the order, and the order is saved in draft state.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param body body CreatePurchaseOrderRequest true "Purchase order details"
// @Success 201 {object} CreatePurchaseOrderResponseEnvelope "Purchase order created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders [post]
func (h PurchaseOrderHandler) Create(c fiber.Ctx) error {
	var request CreatePurchaseOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	orderDate, err := helper.ParseDate(request.OrderDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid order date provided.", nil)
	}
	expectedDate, err := helper.ParseDate(request.ExpectedDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expected date provided.", nil)
	}
	lines := make([]*procurement.PurchaseOrderLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &procurement.PurchaseOrderLine{
			Sequence:    line.Sequence,
			ItemID:      line.ItemID,
			Description: line.Description,
			QtyOrdered:  line.QtyOrdered,
			UnitID:      line.UnitID,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
			TaxIDs:      line.TaxIDs,
			DimensionID: line.DimensionID,
		}
	}
	order := &procurement.PurchaseOrder{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		SupplierID:     request.SupplierID,
		SupplierRef:    request.SupplierRef,
		CurrencyCode:   request.CurrencyCode,
		WarehouseID:    request.WarehouseID,
		DestLocationID: request.DestLocationID,
		OrderDate:      orderDate,
		ExpectedDate:   expectedDate,
		PaymentTermID:  request.PaymentTermID,
		Incoterm:       request.Incoterm,
	}

	var created *procurement.PurchaseOrder
	if request.RequestID != nil {
		created, err = h.svc.CreateFromRequest(c, *request.RequestID, order)
	} else {
		created, err = h.svc.Create(c, order, lines)
	}
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Purchase order created successfully.", CreatePurchaseOrderResponse{
		Order: newPurchaseOrderResponse(created),
	})
}

type ConfirmPurchaseOrderRequest struct {
	ByUserID *uint64 `json:"by_user_id"`
}

// @Summary Update purchase order draft
// @Description Replaces the header fields and lines of an existing purchase order, which must still be in draft state. The supplier and warehouse are revalidated, lines without an explicit unit price are re-priced from the supplier catalog, and the order lines are replaced before totals are recomputed. Any state other than draft is rejected.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path integer true "Purchase order ID"
// @Param body body CreatePurchaseOrderRequest true "Purchase order details"
// @Success 200 {object} CreatePurchaseOrderResponseEnvelope "Purchase order draft updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase order not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders/{id} [put]
func (h PurchaseOrderHandler) UpdateDraft(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase order id provided.", nil)
	}
	var request CreatePurchaseOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	orderDate, err := helper.ParseDate(request.OrderDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid order date provided.", nil)
	}
	expectedDate, err := helper.ParseDate(request.ExpectedDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expected date provided.", nil)
	}
	lines := make([]*procurement.PurchaseOrderLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &procurement.PurchaseOrderLine{
			Sequence:    line.Sequence,
			ItemID:      line.ItemID,
			Description: line.Description,
			QtyOrdered:  line.QtyOrdered,
			UnitID:      line.UnitID,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
			TaxIDs:      line.TaxIDs,
			DimensionID: line.DimensionID,
		}
	}
	order := &procurement.PurchaseOrder{
		Base:           model.Base{ID: orderID},
		SupplierID:     request.SupplierID,
		SupplierRef:    request.SupplierRef,
		CurrencyCode:   request.CurrencyCode,
		WarehouseID:    request.WarehouseID,
		DestLocationID: request.DestLocationID,
		OrderDate:      orderDate,
		ExpectedDate:   expectedDate,
		PaymentTermID:  request.PaymentTermID,
		Incoterm:       request.Incoterm,
	}
	updated, err := h.svc.UpdateDraft(c, order, lines)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase order draft updated successfully.", CreatePurchaseOrderResponse{
		Order: newPurchaseOrderResponse(updated),
	})
}

// @Summary Confirm purchase order
// @Description Confirms a purchase order, transitioning it from draft to sent. When the order total reaches or exceeds the configured approval threshold and the order has not yet been approved, it is routed into the approval workflow and the confirmation is rejected until the configured approvers act on it. Otherwise the order movements directly to the sent state.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path integer true "Purchase order ID"
// @Param body body ConfirmPurchaseOrderRequest true "Confirmation details"
// @Success 200 {object} GetPurchaseOrderResponseEnvelope "Purchase order confirmed successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase order not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders/{id}/confirm [post]
func (h PurchaseOrderHandler) Confirm(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase order id provided.", nil)
	}
	var request ConfirmPurchaseOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	byUserID, ok := httpx.CallerID(c)
	if !ok || request.ByUserID != nil {
		if request.ByUserID != nil {
			byUserID = *request.ByUserID
		}
	}
	order, err := h.svc.Confirm(c, orderID, byUserID)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase order confirmed successfully.", GetPurchaseOrderResponse{
		Order: newPurchaseOrderResponse(order),
	})
}

// @Summary Cancel purchase order
// @Description Cancels a purchase order that is in draft, sent, or confirmed state, moving it to cancelled. The state machine rejects cancellation from any other state, so orders that have already been received and moved to done cannot be cancelled.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path integer true "Purchase order ID"
// @Success 200 {object} GetPurchaseOrderResponseEnvelope "Purchase order cancelled successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase order not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders/{id}/cancel [post]
func (h PurchaseOrderHandler) Cancel(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase order id provided.", nil)
	}
	order, err := h.svc.Cancel(c, orderID)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase order cancelled successfully.", GetPurchaseOrderResponse{
		Order: newPurchaseOrderResponse(order),
	})
}

type ReceivePurchaseOrderRequest struct {
	JournalID uint64  `json:"journal_id" validate:"required,gt=0"`
	Date      *string `json:"date"`
}

// @Summary Receive purchase order
// @Description Records receipt of the remaining outstanding quantities on a sent or confirmed purchase order, confirming the order first when it is still in the sent state. An inbound shipment with stock movements is created and each line's remaining quantity is received into stock, updating qty_received and triggering quality checks. The operation requires any pending approval to be satisfied when the total is above the configured threshold and fails with a conflict when nothing is left to receive.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path integer true "Purchase order ID"
// @Param body body ReceivePurchaseOrderRequest true "Receipt details"
// @Success 200 {object} GetPurchaseOrderResponseEnvelope "Purchase order received successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase order not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders/{id}/receive [post]
func (h PurchaseOrderHandler) Receive(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase order id provided.", nil)
	}
	var request ReceivePurchaseOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	receiveDate, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid receipt date provided.", nil)
	}
	order, err := h.svc.Receive(c, orderID, request.JournalID, helper.Deref(receiveDate, time.Now().UTC()))
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase order received successfully.", GetPurchaseOrderResponse{
		Order: newPurchaseOrderResponse(order),
	})
}

func writePurchaseOrderError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, procurement.ErrPurchaseOrderNotFound):
		return httpx.CreateNotFoundResponse(c, "Purchase order not found.")
	case errors.Is(err, procurement.ErrPurchaseOrderState):
		return httpx.CreateConflictResponse(c, "Purchase order cannot be processed in its current state.", err)
	case errors.Is(err, procurement.ErrPurchaseOrderNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase order must have at least one line.", nil)
	case errors.Is(err, procurement.ErrPurchaseOrderLineQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase order line quantity must be positive.", nil)
	case errors.Is(err, procurement.ErrPurchaseOrderLineDiscount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase order line discount must be between 0 and 100.", nil)
	case errors.Is(err, procurement.ErrPurchaseOrderVendor):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase order supplier does not exist.", nil)
	case errors.Is(err, procurement.ErrPurchaseOrderVendorNotSupplier):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase order supplier is not an active supplier.", nil)
	case errors.Is(err, procurement.ErrPurchaseOrderWarehouse):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase order warehouse does not exist.", nil)
	case errors.Is(err, procurement.ErrPurchaseOrderNoOffer):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No supplier catalog price available for a purchase order line.", nil)
	case errors.Is(err, procurement.ErrPurchaseOrderApprovalPending):
		return httpx.CreateConflictResponse(c, "Purchase order approval is still pending.", err)
	case errors.Is(err, procurement.ErrPurchaseOrderApprovalRequired):
		return httpx.CreateConflictResponse(c, "Purchase order requires approval but no approvers are configured.", err)
	case errors.Is(err, procurement.ErrPurchaseOrderNothingToReceive):
		return httpx.CreateConflictResponse(c, "Purchase order has nothing ready to receive.", err)
	case errors.Is(err, procurement.ErrPurchaseOrderNothingToBill):
		return httpx.CreateConflictResponse(c, "Purchase order has nothing ready to bill.", err)
	case errors.Is(err, procurement.ErrPurchaseOrderBillQualityBlocked):
		return httpx.CreateConflictResponse(c, "Purchase order receipt has unresolved quality failures.", err)
	case errors.Is(err, procurement.ErrPurchaseOrderNoOpenBills):
		return httpx.CreateConflictResponse(c, "No open supplier bills found for this supplier.", err)
	case errors.Is(err, procurement.ErrRequisitionNotFound):
		return httpx.CreateNotFoundResponse(c, "Purchase requisition not found.")
	case errors.Is(err, procurement.ErrRequisitionNotConvertible):
		return httpx.CreateConflictResponse(c, "Purchase requisition must be approved before conversion.", err)
	case errors.Is(err, procurement.ErrRFQNotFound):
		return httpx.CreateNotFoundResponse(c, "Purchase QuoteRequest not found.")
	case errors.Is(err, procurement.ErrQuoteRequestState):
		return httpx.CreateConflictResponse(c, "Purchase QuoteRequest cannot be processed in its current state.", err)
	case errors.Is(err, procurement.ErrRFQNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase QuoteRequest must have at least one line.", nil)
	case errors.Is(err, procurement.ErrRFQLineQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase QuoteRequest line quantity must be positive.", nil)
	case errors.Is(err, procurement.ErrRFQRequester):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase QuoteRequest requester does not exist.", nil)
	case errors.Is(err, procurement.ErrRFQVendor):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase QuoteRequest supplier does not exist.", nil)
	case errors.Is(err, procurement.ErrRFQNotConvertible):
		return httpx.CreateConflictResponse(c, "Purchase QuoteRequest must be sent before conversion.", err)
	case errors.Is(err, procurement.ErrRFQQuoteNotFound):
		return httpx.CreateNotFoundResponse(c, "Purchase quotation not found.")
	case errors.Is(err, procurement.ErrSupplierQuoteState):
		return httpx.CreateConflictResponse(c, "Purchase quotation cannot be processed in its current state.", err)
	case errors.Is(err, procurement.ErrRFQQuoteNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase quotation must have at least one line.", nil)
	case errors.Is(err, procurement.ErrRFQQuoteLineQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase quotation line quantity must be positive.", nil)
	case errors.Is(err, procurement.ErrRFQQuoteNoAccepted):
		return httpx.CreateConflictResponse(c, "Purchase QuoteRequest has no accepted quotation.", err)
	case errors.Is(err, procurement.ErrRFQQuoteVendor):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase quotation supplier does not exist.", nil)
	case errors.Is(err, procurement.ErrRFQQuoteNotSubmitted):
		return httpx.CreateConflictResponse(c, "Purchase quotation must be submitted before it can be accepted.", err)
	case errors.Is(err, procurement.ErrRFQQuotePrice):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase quotation line price cannot be negative.", nil)
	case errors.Is(err, procurement.ErrRFQQuoteDuplicateLine):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase quotation references an unknown request line.", nil)
	case errors.Is(err, accounting.ErrNoPayableAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No payable account configured for the organization.", nil)
	case errors.Is(err, accounting.ErrInvoiceNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Supplier bill must have at least one line.", nil)
	case errors.Is(err, accounting.ErrInvoiceTaxInvalid):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Supplier bill line references an invalid purchase tax.", nil)
	default:
		httpx.RequestLog(c).Error("purchase order write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process purchase order.", err)
	}
}
