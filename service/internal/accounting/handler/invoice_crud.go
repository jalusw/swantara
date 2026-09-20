package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary List invoices
// @Description Lists customer invoices and supplier bills for the caller's tenant, applying pagination, sorting, and filtering over fields such as contact, type, state, payment state, and dates. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. invoice_date:asc)"
// @Param filter query string false "Filters (repeatable, e.g. type:eq:customer_invoice)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListInvoicesResponseEnvelope "Invoices retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/invoices [get]
func (h InvoiceHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, invoiceQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("invoice list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve invoices.", err)
	}

	items := make([]InvoiceResponse, len(page.Items))
	for i, invoice := range page.Items {
		items[i] = newInvoiceResponse(invoice)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "invoices.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Invoices retrieved successfully.", ListInvoicesResponse{
		Invoices: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get invoice
// @Description Gets a single invoice or supplier bill by its id together with its lines and tax breakdown, returning 404 when the document does not exist or belongs to another tenant. The response includes each line's subtotal and the computed tax amounts for the whole document.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param id path integer true "Invoice ID"
// @Success 200 {object} GetInvoiceResponseEnvelope "Invoice retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Invoice not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/invoices/{id} [get]
func (h InvoiceHandler) Get(c fiber.Ctx) error {
	invoiceID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid invoice id provided.", nil)
	}

	invoice, err := h.svc.Get(c, invoiceID)
	if err != nil {
		httpx.RequestLog(c).Error("invoice lookup failed", "invoice_id", invoiceID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get invoice.", err)
	}
	if invoice == nil || !httpx.OwnsTenant(c, invoice.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Invoice not found.")
	}

	lines, err := h.svc.ListLines(c, invoice.ID)
	if err != nil {
		httpx.RequestLog(c).Error("invoice lines lookup failed", "invoice_id", invoice.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get invoice lines.", err)
	}
	lineItems := make([]InvoiceLineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newInvoiceLineResponse(line)
	}

	taxes, err := h.svc.ListTaxes(c, invoice.ID)
	if err != nil {
		httpx.RequestLog(c).Error("invoice taxes lookup failed", "invoice_id", invoice.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get invoice taxes.", err)
	}
	taxItems := make([]InvoiceTaxResponse, len(taxes))
	for i, tax := range taxes {
		taxItems[i] = newInvoiceTaxResponse(tax)
	}

	installments, err := h.svc.ListInstallments(c, invoice.ID)
	if err != nil {
		httpx.RequestLog(c).Error("invoice installments lookup failed", "invoice_id", invoice.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get invoice installments.", err)
	}
	installmentItems := make([]InvoiceInstallmentResponse, len(installments))
	for i, installment := range installments {
		installmentItems[i] = newInvoiceInstallmentResponse(installment)
	}

	return httpx.CreateSuccessResponse(c, "Invoice retrieved successfully.", GetInvoiceResponse{
		Invoice:      newInvoiceResponse(invoice),
		Lines:        lineItems,
		Taxes:        taxItems,
		Installments: installmentItems,
	})
}

// @Summary Create customer invoice
// @Description Creates and posts a customer invoice by validating the requested lines and their taxes, then computing the untaxed, tax, and total amounts. The posting engine requires a configured receivable account and records the journal entry that debits receivables and credits revenue and output tax.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param body body CreateInvoiceRequest true "Invoice details"
// @Success 201 {object} CreateInvoiceResponseEnvelope "Invoice created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or missing configuration"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/invoices [post]
func (h InvoiceHandler) Create(c fiber.Ctx) error {
	return h.create(c, false)
}

// @Summary Create supplier bill
// @Description Creates and posts a supplier bill by validating the requested lines and their taxes, then computing the untaxed, tax, and total amounts. The posting engine requires a configured payable account and records the journal entry that credits payables and debits expenses and input tax.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param body body CreateInvoiceRequest true "Supplier bill details"
// @Success 201 {object} CreateInvoiceResponseEnvelope "Supplier bill created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or missing configuration"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-bills [post]
func (h InvoiceHandler) CreateSupplierBill(c fiber.Ctx) error {
	return h.create(c, true)
}

func (h InvoiceHandler) create(c fiber.Ctx, supplier bool) error {
	var request CreateInvoiceRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	date, dueDate, err := h.parseDates(request)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Dates must be in YYYY-MM-DD format.", nil)
	}

	var invoice *accounting.Invoice
	if supplier {
		invoice, err = h.svc.CreateSupplierBill(c, accounting.CreateSupplierBillRequest{
			OrganizationID: *organizationID,
			JournalID:      request.JournalID,
			ContactID:      request.ContactID,
			Date:           invoiceDate(date),
			DueDate:        dueDate,
			Reference:      request.Reference,
			CurrencyCode:   request.CurrencyCode,
			Draft:          request.Draft,
			PaymentTermID:  request.PaymentTermID,
			TaxRuleID:      request.TaxRuleID,
			Lines:          h.buildLines(request.Lines),
		})
	} else {
		invoice, err = h.svc.Create(c, accounting.CreateInvoiceRequest{
			OrganizationID: *organizationID,
			JournalID:      request.JournalID,
			ContactID:      request.ContactID,
			Date:           invoiceDate(date),
			DueDate:        dueDate,
			Reference:      request.Reference,
			CurrencyCode:   request.CurrencyCode,
			Draft:          request.Draft,
			PaymentTermID:  request.PaymentTermID,
			TaxRuleID:      request.TaxRuleID,
			Lines:          h.buildLines(request.Lines),
		})
	}
	if err != nil {
		return writeInvoiceError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Invoice created successfully.", CreateInvoiceResponse{
		Invoice: newInvoiceResponse(invoice),
	})
}
