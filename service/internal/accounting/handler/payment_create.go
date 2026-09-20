package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h PaymentHandler) register(c fiber.Ctx, request CreatePaymentRequest) (time.Time, error) {
	date, err := helper.ParseDate(request.Date)
	if err != nil {
		return time.Time{}, err
	}
	if date == nil || date.IsZero() {
		return time.Now().UTC(), nil
	}
	return *date, nil
}

// @Summary Create inbound payment
// @Description Records a posted customer payment against one or more open posted invoices, allocating the amount across them and reducing their residual balances. The payment must reference at least one open posted invoice, have a positive amount, and not exceed the open balance, and it posts the journal entry debiting the bank or cash account and crediting receivables.
// @Tags Payments
// @Accept json
// @Produce json
// @Param body body CreatePaymentRequest true "Payment details"
// @Success 201 {object} CreatePaymentResponseEnvelope "Payment created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or allocation failure"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payments [post]
func (h PaymentHandler) Create(c fiber.Ctx) error {
	var request CreatePaymentRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	date, err := h.register(c, request)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	payment, err := h.svc.Create(c, accounting.CreatePaymentRequest{
		OrganizationID:    *organizationID,
		ContactID:         request.ContactID,
		JournalID:         request.JournalID,
		Amount:            request.Amount,
		CurrencyCode:      request.CurrencyCode,
		Date:              date,
		Reference:         request.Reference,
		InvoiceIDs:        request.InvoiceIDs,
		Tolerance:         request.Tolerance,
		AllowAdvance:      request.AllowAdvance,
		DiscountAmount:    request.DiscountAmount,
		WithholdingAmount: request.WithholdingAmount,
		WriteOffAccountID: request.WriteOffAccountID,
	})
	if err != nil {
		return writePaymentError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Payment created successfully.", CreatePaymentResponse{
		Payment: newPaymentResponse(payment),
	})
}

// @Summary Create outbound payment
// @Description Records a posted supplier payment against one or more open posted supplier bills, allocating the amount across them and reducing their residual balances. The payment must reference at least one open posted bill, have a positive amount, and not exceed the open balance, and it posts the journal entry debiting payables and crediting the bank or cash account.
// @Tags Payments
// @Accept json
// @Produce json
// @Param body body CreatePaymentRequest true "Payment details"
// @Success 201 {object} CreatePaymentResponseEnvelope "Payment created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or allocation failure"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payments/outbound [post]
func (h PaymentHandler) CreateOutbound(c fiber.Ctx) error {
	var request CreatePaymentRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	date, err := h.register(c, request)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	payment, err := h.svc.CreateOutbound(c, accounting.CreatePaymentRequest{
		OrganizationID:    *organizationID,
		ContactID:         request.ContactID,
		JournalID:         request.JournalID,
		Amount:            request.Amount,
		CurrencyCode:      request.CurrencyCode,
		Date:              date,
		Reference:         request.Reference,
		InvoiceIDs:        request.InvoiceIDs,
		Tolerance:         request.Tolerance,
		AllowAdvance:      request.AllowAdvance,
		DiscountAmount:    request.DiscountAmount,
		WithholdingAmount: request.WithholdingAmount,
		WriteOffAccountID: request.WriteOffAccountID,
	})
	if err != nil {
		return writePaymentError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Payment created successfully.", CreatePaymentResponse{
		Payment: newPaymentResponse(payment),
	})
}
