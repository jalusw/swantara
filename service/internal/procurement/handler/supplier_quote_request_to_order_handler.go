package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type CreatePurchaseOrderFromRFQRequest struct {
	SupplierID    *uint64    `json:"supplier_id"`
	CurrencyCode  *string    `json:"currency_code"`
	WarehouseID   *uint64    `json:"warehouse_id"`
	SupplierRef   *string    `json:"supplier_ref"`
	OrderDate     *time.Time `json:"order_date"`
	ExpectedDate  *time.Time `json:"expected_date"`
	PaymentTermID *uint64    `json:"payment_term_id"`
	Incoterm      *string    `json:"incoterm"`
}

// @Summary Create purchase order from QuoteRequest
// @Description Converts a done purchase QuoteRequest into a purchase order using the accepted quotation's supplier and line pricing. The order lines are built from the accepted quotation's lines, any fields not supplied on the request are inherited from the QuoteRequest, and the resulting draft purchase order is returned.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param id path integer true "Purchase QuoteRequest ID"
// @Param body body CreatePurchaseOrderFromRFQRequest true "Purchase order details"
// @Success 201 {object} CreatePurchaseOrderResponseEnvelope "Purchase order created from QuoteRequest successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase QuoteRequest not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests/{id}/purchase-order [post]
func (h SupplierQuoteRequestHandler) CreatePurchaseOrder(c fiber.Ctx) error {
	rfqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase QuoteRequest id provided.", nil)
	}
	var request CreatePurchaseOrderFromRFQRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	order, err := h.svc.CreatePurchaseOrder(c, rfqID, &procurement.PurchaseOrder{
		SupplierID:    derefUint64(request.SupplierID),
		CurrencyCode:  request.CurrencyCode,
		WarehouseID:   request.WarehouseID,
		SupplierRef:   request.SupplierRef,
		OrderDate:     request.OrderDate,
		ExpectedDate:  request.ExpectedDate,
		PaymentTermID: request.PaymentTermID,
		Incoterm:      request.Incoterm,
	})
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Purchase order created from QuoteRequest successfully.", CreatePurchaseOrderResponse{
		Order: newPurchaseOrderResponse(order),
	})
}

func derefUint64(value *uint64) uint64 {
	if value == nil {
		return 0
	}
	return *value
}
