package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type InvoiceResponse struct {
	ID             uint64        `json:"id"`
	OrganizationID *uint64       `json:"organization_id"`
	EntryID        *uint64       `json:"entry_id"`
	Type           string        `json:"type"`
	ContactID      uint64        `json:"contact_id"`
	Name           *string       `json:"name"`
	Reference      *string       `json:"reference"`
	InvoiceDate    *time.Time    `json:"invoice_date"`
	DueDate        *time.Time    `json:"due_date"`
	CurrencyCode   *string       `json:"currency_code"`
	JournalID      *uint64       `json:"journal_id"`
	PaymentTermID  *uint64       `json:"payment_term_id"`
	TaxRuleID      *uint64       `json:"tax_rule_id"`
	State          string        `json:"state"`
	PaymentState   string        `json:"payment_state"`
	AmountUntaxed  amount.Amount `json:"amount_untaxed"`
	AmountTax      amount.Amount `json:"amount_tax"`
	AmountTotal    amount.Amount `json:"amount_total"`
	AmountResidual amount.Amount `json:"amount_residual"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

func newInvoiceResponse(invoice *accounting.Invoice) InvoiceResponse {
	return InvoiceResponse{
		ID:             invoice.ID,
		OrganizationID: invoice.OrganizationID,
		EntryID:        invoice.EntryID,
		Type:           invoice.Type,
		ContactID:      invoice.ContactID,
		Name:           invoice.Name,
		Reference:      invoice.Reference,
		InvoiceDate:    invoice.InvoiceDate,
		DueDate:        invoice.DueDate,
		CurrencyCode:   invoice.CurrencyCode,
		JournalID:      invoice.JournalID,
		PaymentTermID:  invoice.PaymentTermID,
		TaxRuleID:      invoice.TaxRuleID,
		State:          invoice.State,
		PaymentState:   invoice.PaymentState,
		AmountUntaxed:  invoice.AmountUntaxed,
		AmountTax:      invoice.AmountTax,
		AmountTotal:    invoice.AmountTotal,
		AmountResidual: invoice.AmountResidual,
		CreatedAt:      invoice.CreatedAt,
		UpdatedAt:      invoice.UpdatedAt,
	}
}

type InvoiceLineResponse struct {
	ID            uint64            `json:"id"`
	InvoiceID     uint64            `json:"invoice_id"`
	Sequence      int               `json:"sequence"`
	ItemID        *uint64           `json:"item_id"`
	Description   *string           `json:"description"`
	Qty           float64           `json:"qty"`
	UnitID        *uint64           `json:"unit_id"`
	UnitPrice     amount.Amount     `json:"unit_price"`
	DiscountPct   float64           `json:"discount_pct"`
	TaxIDs        helper.Int64Array `json:"tax_ids"`
	AccountID     *uint64           `json:"account_id"`
	DimensionID   *uint64           `json:"dimension_id"`
	PriceSubtotal amount.Amount     `json:"price_subtotal"`
}

func newInvoiceLineResponse(line *accounting.InvoiceLine) InvoiceLineResponse {
	return InvoiceLineResponse{
		ID:            line.ID,
		InvoiceID:     line.InvoiceID,
		Sequence:      line.Sequence,
		ItemID:        line.ItemID,
		Description:   line.Description,
		Qty:           line.Qty,
		UnitID:        line.UnitID,
		UnitPrice:     line.UnitPrice,
		DiscountPct:   line.DiscountPct,
		TaxIDs:        line.TaxIDs,
		AccountID:     line.AccountID,
		DimensionID:   line.DimensionID,
		PriceSubtotal: line.PriceSubtotal,
	}
}

type InvoiceTaxResponse struct {
	ID        uint64        `json:"id"`
	InvoiceID uint64        `json:"invoice_id"`
	TaxID     *uint64       `json:"tax_id"`
	Base      amount.Amount `json:"base"`
	Amount    amount.Amount `json:"amount"`
	AccountID *uint64       `json:"account_id"`
}

func newInvoiceTaxResponse(tax *accounting.InvoiceTax) InvoiceTaxResponse {
	return InvoiceTaxResponse{
		ID:        tax.ID,
		InvoiceID: tax.InvoiceID,
		TaxID:     tax.TaxID,
		Base:      tax.BaseAmount,
		Amount:    tax.Amount,
		AccountID: tax.AccountID,
	}
}

type ListInvoicesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListInvoicesResponse `json:"data"`
}
type ListInvoicesResponse struct {
	Invoices []InvoiceResponse `json:"invoices"`
}

type GetInvoiceResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetInvoiceResponse `json:"data"`
}
type GetInvoiceResponse struct {
	Invoice      InvoiceResponse              `json:"invoice"`
	Lines        []InvoiceLineResponse        `json:"lines"`
	Taxes        []InvoiceTaxResponse         `json:"taxes"`
	Installments []InvoiceInstallmentResponse `json:"installments"`
}

type InvoiceInstallmentResponse struct {
	ID        uint64        `json:"id"`
	InvoiceID uint64        `json:"invoice_id"`
	Sequence  int           `json:"sequence"`
	DueDate   *time.Time    `json:"due_date"`
	Amount    amount.Amount `json:"amount"`
	State     string        `json:"state"`
}

func newInvoiceInstallmentResponse(installment *accounting.InvoiceInstallment) InvoiceInstallmentResponse {
	return InvoiceInstallmentResponse{
		ID:        installment.ID,
		InvoiceID: installment.InvoiceID,
		Sequence:  installment.Sequence,
		DueDate:   installment.DueDate,
		Amount:    installment.Amount,
		State:     installment.State,
	}
}

type CreateInvoiceResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateInvoiceResponse `json:"data"`
}
type CreateInvoiceResponse struct {
	Invoice InvoiceResponse `json:"invoice"`
}

type CreateCreditNoteResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateCreditNoteResponse `json:"data"`
}
type CreateCreditNoteResponse struct {
	Invoice InvoiceResponse `json:"invoice"`
}

type AgingReportRowResponse struct {
	ContactID uint64  `json:"contact_id"`
	Current   float64 `json:"current"`
	Days1_30  float64 `json:"days_1_30"`
	Days31_60 float64 `json:"days_31_60"`
	Days61_90 float64 `json:"days_61_90"`
	Over90    float64 `json:"over_90"`
	Total     float64 `json:"total"`
}

type AgingReportResponseEnvelope struct {
	httpx.EnvelopeBase
	Data AgingReportResponse `json:"data"`
}
type AgingReportResponse struct {
	Rows []AgingReportRowResponse `json:"rows"`
}
