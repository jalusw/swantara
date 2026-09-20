package handler

import (
	"strconv"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type PurchaseInvoiceResponse struct {
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

func newPurchaseInvoiceResponse(invoice *accounting.Invoice) PurchaseInvoiceResponse {
	return PurchaseInvoiceResponse{
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

type BillPurchaseOrderRequest struct {
	JournalID uint64  `json:"journal_id" validate:"required,gt=0"`
	Date      *string `json:"date"`
	Override  bool    `json:"override"`
}

type BillPurchaseOrderResponse struct {
	Invoice PurchaseInvoiceResponse `json:"invoice"`
}

type BillPurchaseOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data BillPurchaseOrderResponse `json:"data"`
}

// @Summary Create supplier bill
// @Description Creates a supplier bill for the received but not yet billed quantities of a confirmed or done purchase order, resolving the stock input or expense account for each line and posting the bill against the supplier. Billing is blocked while the receipt has unresolved quality failures unless the override flag is set. Each line's qty_billed is updated and the order's invoice and receipt statuses are recomputed.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path integer true "Purchase order ID"
// @Param body body BillPurchaseOrderRequest true "Bill details"
// @Success 201 {object} BillPurchaseOrderResponseEnvelope "Supplier bill created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase order not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders/{id}/supplier-bill [post]
func (h PurchaseOrderHandler) CreateSupplierBill(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase order id provided.", nil)
	}
	var request BillPurchaseOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	billDate, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid bill date provided.", nil)
	}
	invoice, err := h.svc.CreateSupplierBill(c, orderID, request.JournalID, helper.Deref(billDate, time.Now().UTC()), request.Override)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Supplier bill created successfully.", BillPurchaseOrderResponse{
		Invoice: newPurchaseInvoiceResponse(invoice),
	})
}

type PayPurchaseOrderRequest struct {
	JournalID uint64  `json:"journal_id" validate:"required,gt=0"`
	Date      *string `json:"date"`
}

type PurchasePaymentResponse struct {
	ID        uint64    `json:"id"`
	Name      *string   `json:"name"`
	ContactID uint64    `json:"contact_id"`
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	Date      time.Time `json:"date"`
	State     string    `json:"state"`
}

func newPurchasePaymentResponse(payment *accounting.Payment) PurchasePaymentResponse {
	return PurchasePaymentResponse{
		ID:        payment.ID,
		Name:      payment.Name,
		ContactID: payment.ContactID,
		Type:      payment.Type,
		Amount:    payment.Amount,
		Date:      payment.Date,
		State:     payment.State,
	}
}

type PayPurchaseOrderResponse struct {
	Payment PurchasePaymentResponse `json:"payment"`
}

type PayPurchaseOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PayPurchaseOrderResponse `json:"data"`
}

// @Summary Pay supplier bill
// @Description Pays the supplier bill for a purchase order by recording an outbound payment for the order total and allocating it across the supplier's open bills, reconciling them in the process. The operation fails when the supplier has no open bills to allocate the payment against.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path integer true "Purchase order ID"
// @Param body body PayPurchaseOrderRequest true "Payment details"
// @Success 201 {object} PayPurchaseOrderResponseEnvelope "Supplier payment recorded successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase order not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders/{id}/pay [post]
func (h PurchaseOrderHandler) Pay(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase order id provided.", nil)
	}
	var request PayPurchaseOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	payDate, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payment date provided.", nil)
	}
	payment, err := h.svc.PaySupplierBill(c, orderID, request.JournalID, helper.Deref(payDate, time.Now().UTC()))
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Supplier payment recorded successfully.", PayPurchaseOrderResponse{
		Payment: newPurchasePaymentResponse(payment),
	})
}

type CreateCreditMemoRequest struct {
	JournalID uint64  `json:"journal_id" validate:"required"`
	Date      *string `json:"date"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Reason    string  `json:"reason" validate:"required"`
}

type CreditMemoResponse struct {
	ID            uint64     `json:"id"`
	OrderID       uint64     `json:"order_id"`
	JournalID     uint64     `json:"journal_id"`
	ContactID     uint64     `json:"contact_id"`
	Date          *time.Time `json:"date"`
	State         string     `json:"state"`
	AmountUntaxed float64    `json:"amount_untaxed"`
	AmountTotal   float64    `json:"amount_total"`
	Reason        *string    `json:"reason"`
}

func newCreditMemoResponse(memo *procurement.PurchaseCreditMemo) CreditMemoResponse {
	return CreditMemoResponse{
		ID:            memo.ID,
		OrderID:       memo.OrderID,
		JournalID:     memo.JournalID,
		ContactID:     memo.ContactID,
		Date:          memo.Date,
		State:         memo.State,
		AmountUntaxed: memo.AmountUntaxed,
		AmountTotal:   memo.AmountTotal,
		Reason:        memo.Reason,
	}
}

type CreateCreditMemoResponse struct {
	Memo CreditMemoResponse `json:"memo"`
}

type CreateCreditMemoResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateCreditMemoResponse `json:"data"`
}

// @Summary Create supplier credit memo
// @Description Creates a credit memo (supplier credit note) for a purchase order. Reduces the order total.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path integer true "Purchase order ID"
// @Param body body CreateCreditMemoRequest true "Credit memo details"
// @Success 201 {object} CreateCreditMemoResponseEnvelope "Credit memo created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase order not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders/{id}/credit-memo [post]
func (h PurchaseOrderHandler) CreateCreditMemo(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase order id provided.", nil)
	}
	var request CreateCreditMemoRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	memo, err := h.svc.CreateVendorCreditMemo(c, orderID, request.JournalID, helper.Deref(date, time.Now().UTC()), request.Amount, request.Reason)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Credit memo created successfully.", CreateCreditMemoResponse{
		Memo: newCreditMemoResponse(memo),
	})
}

type CreateDebitMemoRequest struct {
	JournalID uint64  `json:"journal_id" validate:"required"`
	Date      *string `json:"date"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
	Reason    string  `json:"reason" validate:"required"`
}

type DebitMemoResponse struct {
	ID            uint64     `json:"id"`
	OrderID       uint64     `json:"order_id"`
	JournalID     uint64     `json:"journal_id"`
	ContactID     uint64     `json:"contact_id"`
	Date          *time.Time `json:"date"`
	State         string     `json:"state"`
	AmountUntaxed float64    `json:"amount_untaxed"`
	AmountTotal   float64    `json:"amount_total"`
	Reason        *string    `json:"reason"`
}

func newDebitMemoResponse(memo *procurement.PurchaseDebitMemo) DebitMemoResponse {
	return DebitMemoResponse{
		ID:            memo.ID,
		OrderID:       memo.OrderID,
		JournalID:     memo.JournalID,
		ContactID:     memo.ContactID,
		Date:          memo.Date,
		State:         memo.State,
		AmountUntaxed: memo.AmountUntaxed,
		AmountTotal:   memo.AmountTotal,
		Reason:        memo.Reason,
	}
}

type CreateDebitMemoResponse struct {
	Memo DebitMemoResponse `json:"memo"`
}

type CreateDebitMemoResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateDebitMemoResponse `json:"data"`
}

// @Summary Create supplier debit memo
// @Description Creates a debit memo (supplier debit note) for a purchase order. Increases the order total.
// @Tags Purchase Orders
// @Accept json
// @Produce json
// @Param id path integer true "Purchase order ID"
// @Param body body CreateDebitMemoRequest true "Debit memo details"
// @Success 201 {object} CreateDebitMemoResponseEnvelope "Debit memo created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase order not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-orders/{id}/debit-memo [post]
func (h PurchaseOrderHandler) CreateDebitMemo(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase order id provided.", nil)
	}
	var request CreateDebitMemoRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	memo, err := h.svc.CreateVendorDebitMemo(c, orderID, request.JournalID, helper.Deref(date, time.Now().UTC()), request.Amount, request.Reason)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Debit memo created successfully.", CreateDebitMemoResponse{
		Memo: newDebitMemoResponse(memo),
	})
}
