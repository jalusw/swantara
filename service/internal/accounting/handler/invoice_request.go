package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type InvoiceLineRequest struct {
	ItemID      *uint64           `json:"item_id"`
	Description string            `json:"description"`
	Qty         float64           `json:"qty" validate:"required,gt=0"`
	UnitID      *uint64           `json:"unit_id"`
	UnitPrice   float64           `json:"unit_price" validate:"gte=0"`
	DiscountPct float64           `json:"discount_pct" validate:"gte=0,lte=100"`
	TaxIDs      helper.Int64Array `json:"tax_ids"`
	AccountID   uint64            `json:"account_id" validate:"required,gt=0"`
	DimensionID *uint64           `json:"dimension_id"`
}

type CreateInvoiceRequest struct {
	OrganizationID *uint64              `json:"organization_id"`
	JournalID      uint64               `json:"journal_id" validate:"required,gt=0"`
	ContactID      uint64               `json:"contact_id" validate:"required,gt=0"`
	Date           *string              `json:"date"`
	DueDate        *string              `json:"due_date"`
	Reference      string               `json:"reference"`
	CurrencyCode   *string              `json:"currency_code" validate:"omitempty,len=3"`
	Draft          bool                 `json:"draft"`
	PaymentTermID  *uint64              `json:"payment_term_id"`
	TaxRuleID      *uint64              `json:"tax_rule_id"`
	Lines          []InvoiceLineRequest `json:"lines" validate:"required,min=1,dive"`
}

type WriteOffBadDebtRequest struct {
	OrganizationID   *uint64 `json:"organization_id"`
	ExpenseAccountID uint64  `json:"expense_account_id" validate:"required,gt=0"`
	Date             *string `json:"date"`
}

type CreateCreditNoteRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	JournalID      uint64  `json:"journal_id" validate:"required,gt=0"`
	Date           *string `json:"date"`
	DueDate        *string `json:"due_date"`
	Reference      string  `json:"reference"`
}

var invoiceQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"contact_id":      {},
	"type":            {},
	"state":           {},
	"payment_state":   {},
	"journal_id":      {},
	"name":            {},
	"reference":       {},
	"invoice_date":    {},
	"due_date":        {},
	"created_at":      {},
	"updated_at":      {},
}

func invoiceDate(date *time.Time) time.Time {
	if date == nil || date.IsZero() {
		return time.Now().UTC()
	}
	return *date
}

func writeInvoiceError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrInvoiceNotFound):
		return httpx.CreateNotFoundResponse(c, "Invoice not found.")
	case errors.Is(err, accounting.ErrInvoiceNotPosted):
		return httpx.CreateConflictResponse(c, "Only posted invoices can be reversed.", err)
	case errors.Is(err, accounting.ErrInvoiceNotDraft):
		return httpx.CreateConflictResponse(c, "Only draft invoices can be posted.", err)
	case errors.Is(err, accounting.ErrBadDebtNotAllowed):
		return httpx.CreateConflictResponse(c, "Only posted customer invoices with an open balance can be written off.", err)
	case errors.Is(err, accounting.ErrInvoiceReversed), errors.Is(err, accounting.ErrInvoicePaid):
		return httpx.CreateConflictResponse(c, "Invoice is already settled.", err)
	case errors.Is(err, accounting.ErrInvoiceNoLines), errors.Is(err, accounting.ErrInvoiceLineQty), errors.Is(err, accounting.ErrInvoiceTaxInvalid), errors.Is(err, accounting.ErrInvoiceSequence):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invoice lines are invalid.", nil)
	case errors.Is(err, accounting.ErrTaxRuleNotFound), errors.Is(err, reference.ErrTermNotFound), errors.Is(err, reference.ErrInvalidTermLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payment term or fiscal position is invalid.", nil)
	case errors.Is(err, accounting.ErrNoReceivableAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No receivable account configured for the organization.", nil)
	case errors.Is(err, accounting.ErrNoPayableAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No payable account configured for the organization.", nil)
	case errors.Is(err, accounting.ErrNoRevenueAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No revenue account configured for the item.", nil)
	case errors.Is(err, accounting.ErrCreditMismatch):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Credit note does not match the invoice.", nil)
	case errors.Is(err, accounting.ErrOverAllocation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Settlement amount exceeds the open balance.", nil)
	case errors.Is(err, accounting.ErrPaymentAmount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Settlement amount must be positive.", nil)
	case errors.Is(err, accounting.ErrCurrencyMismatch):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Documents must share the same currency to net.", nil)
	case errors.Is(err, accounting.ErrCreditLimitExceeded):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Contact credit limit exceeded.", nil)
	default:
		httpx.RequestLog(c).Error("invoice write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save invoice.", err)
	}
}
