package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Create customer credit note
// @Description Reverses a posted customer invoice and creates a linked credit note with a mirrored journal entry, using the supplied date and due date. Only posted invoices can be reversed, each at most once, and the credit note links back to the original document.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path integer true "Original invoice ID"
// @Param body body CreateCreditNoteRequest true "Credit note details"
// @Success 201 {object} CreateCreditNoteResponseEnvelope "Credit note created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or missing configuration"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/invoices/{id}/credit-note [post]
func (h InvoiceHandler) CreateCreditNote(c fiber.Ctx) error {
	return h.creditNote(c, false)
}

// @Summary Create supplier credit note
// @Description Reverses a posted supplier bill and creates a linked credit note with a mirrored journal entry, using the supplied date and due date. Only posted bills can be reversed, each at most once, and the credit note links back to the original document.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path integer true "Original supplier bill ID"
// @Param body body CreateCreditNoteRequest true "Credit note details"
// @Success 201 {object} CreateCreditNoteResponseEnvelope "Credit note created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or missing configuration"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-bills/{id}/credit-note [post]
func (h InvoiceHandler) CreateVendorCreditNote(c fiber.Ctx) error {
	return h.creditNote(c, true)
}

func (h InvoiceHandler) creditNote(c fiber.Ctx, supplier bool) error {
	originalInvoiceID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid invoice id provided.", nil)
	}

	var request CreateCreditNoteRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	date, dueDate, err := h.parseDates(CreateInvoiceRequest{Date: request.Date, DueDate: request.DueDate})
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Dates must be in YYYY-MM-DD format.", nil)
	}

	accountingRequest := accounting.CreateCreditNoteRequest{
		OrganizationID:    *organizationID,
		OriginalInvoiceID: originalInvoiceID,
		JournalID:         request.JournalID,
		Date:              invoiceDate(date),
		DueDate:           dueDate,
		Reference:         request.Reference,
	}
	var invoice *accounting.Invoice
	if supplier {
		invoice, err = h.svc.CreateVendorCreditNote(c, accountingRequest)
	} else {
		invoice, err = h.svc.CreateCreditNote(c, accountingRequest)
	}
	if err != nil {
		return writeInvoiceError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Credit note created successfully.", CreateCreditNoteResponse{
		Invoice: newInvoiceResponse(invoice),
	})
}

// @Summary Post draft invoice
// @Description Posts a draft invoice or supplier bill, creating the receivable or payable journal entry. Only draft documents can be posted.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path integer true "Invoice ID"
// @Success 200 {object} CreateInvoiceResponseEnvelope "Invoice posted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Invoice not found"
// @Failure 422 {object} httpx.ErrorResponse "Only draft invoices can be posted"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/invoices/{id}/post [post]
func (h InvoiceHandler) Post(c fiber.Ctx) error {
	invoiceID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid invoice id provided.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	invoice, err := h.svc.Post(c, invoiceID, *organizationID)
	if err != nil {
		return writeInvoiceError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Invoice posted successfully.", CreateInvoiceResponse{
		Invoice: newInvoiceResponse(invoice),
	})
}

// @Summary Write off bad debt
// @Description Writes off the open balance of a posted customer invoice to a bad-debt expense account, clearing the residual and excluding the invoice from collections.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path integer true "Invoice ID"
// @Param body body WriteOffBadDebtRequest true "Write-off details"
// @Success 200 {object} CreateInvoiceResponseEnvelope "Bad debt written off successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Invoice not found"
// @Failure 422 {object} httpx.ErrorResponse "Only posted customer invoices with an open balance can be written off"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/invoices/{id}/write-off [post]
func (h InvoiceHandler) WriteOff(c fiber.Ctx) error {
	invoiceID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid invoice id provided.", nil)
	}
	var request WriteOffBadDebtRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	date, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	invoice, err := h.svc.WriteOffBadDebt(c, accounting.WriteOffBadDebtRequest{
		InvoiceID:        invoiceID,
		OrganizationID:   *organizationID,
		ExpenseAccountID: request.ExpenseAccountID,
		Date:             invoiceDate(date),
	})
	if err != nil {
		return writeInvoiceError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Bad debt written off successfully.", CreateInvoiceResponse{
		Invoice: newInvoiceResponse(invoice),
	})
}

// @Summary Get receivables aging
// @Description Groups open customer invoice balances per contact into current and 30/60/90-day overdue buckets as of a given date, feeding reminder prioritization.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param as_of query string false "As-of date (YYYY-MM-DD)"
// @Success 200 {object} AgingReportResponseEnvelope "Aging report retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid as-of date"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/invoices/aging [get]
func (h InvoiceHandler) Aging(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	var asOf *time.Time
	if raw := c.Query("as_of"); raw != "" {
		parsed, err := helper.ParseDate(&raw)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "As-of date must be in YYYY-MM-DD format.", nil)
		}
		asOf = parsed
	}

	rows, err := h.svc.AgingReport(c, *organizationID, asOf)
	if err != nil {
		httpx.RequestLog(c).Error("aging report failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve aging report.", err)
	}
	items := make([]AgingReportRowResponse, len(rows))
	for i, row := range rows {
		items[i] = AgingReportRowResponse{
			ContactID: row.ContactID,
			Current:   row.Current,
			Days1_30:  row.Days1_30,
			Days31_60: row.Days31_60,
			Days61_90: row.Days61_90,
			Over90:    row.Over90,
			Total:     row.Total,
		}
	}
	return httpx.CreateSuccessResponse(c, "Aging report retrieved successfully.", AgingReportResponse{
		Rows: items,
	})
}
