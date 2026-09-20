package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type DeliverSaleOrderRequest struct {
	JournalID uint64  `json:"journal_id" validate:"required,gt=0"`
	Date      *string `json:"date"`
}

type InvoiceSaleOrderRequest struct {
	JournalID uint64  `json:"journal_id" validate:"required,gt=0"`
	Date      *string `json:"date"`
}

type PaySaleOrderRequest struct {
	JournalID uint64  `json:"journal_id" validate:"required,gt=0"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Date      *string `json:"date"`
}

// @Summary Deliver sale order
// @Description Delivers a confirmed sale order by shipping the movements of its assigned shipments, posting cost of goods sold to the given journal and date, releasing reservations, and incrementing delivered quantities on the order's lines.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Param body body DeliverSaleOrderRequest true "Delivery details"
// @Success 200 {object} GetSaleOrderResponseEnvelope "Sale order delivered successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 409 {object} httpx.ErrorResponse "Sale order cannot be delivered in its current state"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or missing configuration"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id}/deliver [post]
func (h SaleOrderHandler) Deliver(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	var request DeliverSaleOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := parseSaleDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid delivery date provided.", nil)
	}
	deliveryDate := time.Now().UTC()
	if date != nil {
		deliveryDate = *date
	}

	order, err := h.svc.Deliver(c, orderID, request.JournalID, deliveryDate)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Sale order delivered successfully.", GetSaleOrderResponse{
		Order: newSaleOrderResponse(order),
	})
}

type InvoiceResponse struct {
	ID             uint64        `json:"id"`
	Type           string        `json:"type"`
	ContactID      uint64        `json:"contact_id"`
	Name           *string       `json:"name"`
	State          string        `json:"state"`
	PaymentState   string        `json:"payment_state"`
	AmountUntaxed  amount.Amount `json:"amount_untaxed"`
	AmountTax      amount.Amount `json:"amount_tax"`
	AmountTotal    amount.Amount `json:"amount_total"`
	AmountResidual amount.Amount `json:"amount_residual"`
	InvoiceDate    *time.Time    `json:"invoice_date"`
}

func newInvoiceResponse(invoice *accounting.Invoice) InvoiceResponse {
	return InvoiceResponse{
		ID:             invoice.ID,
		Type:           invoice.Type,
		ContactID:      invoice.ContactID,
		Name:           invoice.Name,
		State:          invoice.State,
		PaymentState:   invoice.PaymentState,
		AmountUntaxed:  invoice.AmountUntaxed,
		AmountTax:      invoice.AmountTax,
		AmountTotal:    invoice.AmountTotal,
		AmountResidual: invoice.AmountResidual,
		InvoiceDate:    invoice.InvoiceDate,
	}
}

// @Summary Invoice sale order
// @Description Invoices a sale order by generating a customer invoice for the delivered-but-uninvoiced quantities on its lines, posting to the given journal and date, and incrementing each line's invoiced quantity. Requires a configured income account for every item line.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Param body body InvoiceSaleOrderRequest true "Invoice details"
// @Success 201 {object} CreateInvoiceResponseEnvelope "Invoice generated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or missing configuration"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id}/invoice [post]
func (h SaleOrderHandler) Invoice(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	var request InvoiceSaleOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := parseSaleDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid invoice date provided.", nil)
	}
	invoiceDate := time.Now().UTC()
	if date != nil {
		invoiceDate = *date
	}

	invoice, err := h.svc.CreateInvoice(c, orderID, request.JournalID, invoiceDate)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Invoice generated successfully.", CreateInvoiceResponse{
		Invoice: newInvoiceResponse(invoice),
	})
}

type CreateInvoiceResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateInvoiceResponse `json:"data"`
}
type CreateInvoiceResponse struct {
	Invoice InvoiceResponse `json:"invoice"`
}

type PaymentResponse struct {
	ID        uint64    `json:"id"`
	Name      *string   `json:"name"`
	ContactID uint64    `json:"contact_id"`
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	Date      time.Time `json:"date"`
	State     string    `json:"state"`
}

func newPaymentResponse(payment *accounting.Payment) PaymentResponse {
	return PaymentResponse{
		ID:        payment.ID,
		Name:      payment.Name,
		ContactID: payment.ContactID,
		Type:      payment.Type,
		Amount:    payment.Amount,
		Date:      payment.Date,
		State:     payment.State,
	}
}

// @Summary Pay sale order
// @Description Records an inbound customer payment against the sale order's contact, allocating the amount across the contact's open invoices and reconciling them. The payment cannot exceed the total open invoice balance and requires a bank or cash account on the journal.
// @Tags Sale Orders
// @Accept json
// @Produce json
// @Param id path integer true "Sale order ID"
// @Param body body PaySaleOrderRequest true "Payment details"
// @Success 201 {object} CreatePaymentResponseEnvelope "Payment recorded successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Sale order not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or missing configuration"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/sale-orders/{id}/pay [post]
func (h SaleOrderHandler) Pay(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid sale order id provided.", nil)
	}

	var request PaySaleOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := parseSaleDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payment date provided.", nil)
	}
	paymentDate := time.Now().UTC()
	if date != nil {
		paymentDate = *date
	}

	payment, err := h.svc.CollectPayment(c, orderID, request.JournalID, request.Amount, paymentDate)
	if err != nil {
		return writeSaleOrderError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Payment recorded successfully.", CreatePaymentResponse{
		Payment: newPaymentResponse(payment),
	})
}

type CreatePaymentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreatePaymentResponse `json:"data"`
}
type CreatePaymentResponse struct {
	Payment PaymentResponse `json:"payment"`
}
