package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type ApplyCreditRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	CreditNoteID   uint64  `json:"credit_note_id" validate:"required,gt=0"`
	Amount         float64 `json:"amount" validate:"gte=0"`
}

type ContraSettleRequest struct {
	OrganizationID    *uint64 `json:"organization_id"`
	CustomerInvoiceID uint64  `json:"customer_invoice_id" validate:"required,gt=0"`
	SupplierBillID    uint64  `json:"supplier_bill_id" validate:"required,gt=0"`
	JournalID         uint64  `json:"journal_id" validate:"required,gt=0"`
	Date              *string `json:"date"`
	Amount            float64 `json:"amount" validate:"gte=0"`
}

type CreditApplicationResponse struct {
	ID           uint64        `json:"id"`
	InvoiceID    uint64        `json:"invoice_id"`
	CreditNoteID uint64        `json:"credit_note_id"`
	Amount       amount.Amount `json:"amount"`
}

type DeductDownPaymentRequest struct {
	OrganizationID   *uint64 `json:"organization_id"`
	AdvanceInvoiceID uint64  `json:"advance_invoice_id" validate:"required,gt=0"`
	Amount           float64 `json:"amount" validate:"gte=0"`
}

type DownPaymentLinkResponse struct {
	ID               uint64        `json:"id"`
	AdvanceInvoiceID uint64        `json:"advance_invoice_id"`
	FinalInvoiceID   uint64        `json:"final_invoice_id"`
	Amount           amount.Amount `json:"amount"`
}

type ContraSettlementResponse struct {
	ID                uint64        `json:"id"`
	CustomerInvoiceID uint64        `json:"customer_invoice_id"`
	SupplierBillID    uint64        `json:"supplier_bill_id"`
	Amount            amount.Amount `json:"amount"`
	EntryID           *uint64       `json:"entry_id"`
	Date              *time.Time    `json:"date"`
}

// @Summary Apply credit note
// @Description Settles an open invoice against a posted credit note for the same contact, reducing both residuals. Amount defaults to the smaller open balance.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path integer true "Invoice ID"
// @Param body body ApplyCreditRequest true "Credit application details"
// @Success 201 {object} ApplyCreditResponseEnvelope "Credit applied successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Invoice not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or open balance exceeded"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/invoices/{id}/apply-credit [post]
func (h InvoiceHandler) ApplyCredit(c fiber.Ctx) error {
	invoiceID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid invoice id provided.", nil)
	}
	var request ApplyCreditRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	applied, err := h.svc.ApplyCredit(c, accounting.ApplyCreditRequest{
		OrganizationID: *organizationID,
		InvoiceID:      invoiceID,
		CreditNoteID:   request.CreditNoteID,
		Amount:         request.Amount,
	})
	if err != nil {
		return writeInvoiceError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Credit applied successfully.", CreditApplicationResponse{
		ID:           applied.ID,
		InvoiceID:    applied.InvoiceID,
		CreditNoteID: applied.CreditNoteID,
		Amount:       applied.Amount,
	})
}

// @Summary Settle AR/AP contra
// @Description Nets a customer invoice against a supplier bill for the same contact, posting a contra journal entry for the settled amount.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param body body ContraSettleRequest true "Contra settlement details"
// @Success 201 {object} ContraSettleResponseEnvelope "Contra settled successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Invoice not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or open balance exceeded"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/contra-settlements [post]
func (h InvoiceHandler) ContraSettle(c fiber.Ctx) error {
	var request ContraSettleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	var date time.Time
	if request.Date != nil {
		parsed, err := helper.ParseDate(request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
		}
		date = *parsed
	}

	settlement, err := h.svc.ContraSettle(c, accounting.ContraSettleRequest{
		OrganizationID:    *organizationID,
		CustomerInvoiceID: request.CustomerInvoiceID,
		SupplierBillID:    request.SupplierBillID,
		JournalID:         request.JournalID,
		Date:              date,
		Amount:            request.Amount,
	})
	if err != nil {
		return writeInvoiceError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Contra settled successfully.", ContraSettlementResponse{
		ID:                settlement.ID,
		CustomerInvoiceID: settlement.CustomerInvoiceID,
		SupplierBillID:    settlement.SupplierBillID,
		Amount:            settlement.Amount,
		EntryID:           settlement.EntryID,
		Date:              settlement.Date,
	})
}

// @Summary Deduct down payment
// @Description Deducts a posted advance invoice from a posted final invoice for the same contact, reducing both residuals. Amount defaults to the smaller open balance.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path integer true "Final invoice ID"
// @Param body body DeductDownPaymentRequest true "Down payment deduction details"
// @Success 201 {object} DeductDownPaymentResponseEnvelope "Down payment deducted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Invoice not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or open balance exceeded"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/invoices/{id}/deduct-down-payment [post]
func (h InvoiceHandler) DeductDownPayment(c fiber.Ctx) error {
	invoiceID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid invoice id provided.", nil)
	}
	var request DeductDownPaymentRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	link, err := h.svc.DeductDownPayment(c, accounting.DeductDownPaymentRequest{
		OrganizationID:   *organizationID,
		FinalInvoiceID:   invoiceID,
		AdvanceInvoiceID: request.AdvanceInvoiceID,
		Amount:           request.Amount,
	})
	if err != nil {
		return writeInvoiceError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Down payment deducted successfully.", DownPaymentLinkResponse{
		ID:               link.ID,
		AdvanceInvoiceID: link.AdvanceInvoiceID,
		FinalInvoiceID:   link.FinalInvoiceID,
		Amount:           link.Amount,
	})
}

type ApplyCreditResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreditApplicationResponse `json:"data"`
}

type DeductDownPaymentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data DownPaymentLinkResponse `json:"data"`
}

type ContraSettleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ContraSettlementResponse `json:"data"`
}
