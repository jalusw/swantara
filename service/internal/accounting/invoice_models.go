package accounting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	InvoiceTypeCustomerInvoice = "customer_invoice"
	InvoiceTypeCustomerCredit  = "customer_credit_note"
	InvoiceTypeSupplierBill    = "supplier_bill"
	InvoiceTypeSupplierCredit  = "supplier_credit_note"

	InvoiceStateDraft     = "draft"
	InvoiceStatePosted    = "posted"
	InvoiceStateCancelled = "cancelled"

	PaymentStateNotPaid   = "not_paid"
	PaymentStateInPayment = "in_payment"
	PaymentStatePartial   = "partial"
	PaymentStatePaid      = "paid"
	PaymentStateReversed  = "reversed"
	PaymentStateBadDebt   = "bad_debt"

	SequenceInvoiceCode = "invoice"
	SequencePaymentCode = "payment"
)

type Invoice struct {
	model.Base
	OrganizationID  *uint64       `json:"organization_id"`
	EntryID         *uint64       `json:"entry_id"`
	Type            string        `gorm:"type:text" json:"type"`
	ContactID       uint64        `gorm:"not null" json:"contact_id"`
	Name            *string       `json:"name"`
	Reference       *string       `json:"reference"`
	InvoiceDate     *time.Time    `gorm:"type:date" json:"invoice_date"`
	DueDate         *time.Time    `gorm:"type:date" json:"due_date"`
	CurrencyCode    *string       `gorm:"type:char(3)" json:"currency_code"`
	JournalID       *uint64       `json:"journal_id"`
	PaymentTermID   *uint64       `json:"payment_term_id"`
	TaxRuleID       *uint64       `json:"tax_rule_id"`
	State           string        `gorm:"type:text" json:"state"`
	PaymentState    string        `gorm:"type:text" json:"payment_state"`
	AmountUntaxed   amount.Amount `gorm:"type:numeric(18,4);default:0" json:"amount_untaxed"`
	AmountTax       amount.Amount `gorm:"type:numeric(18,4);default:0" json:"amount_tax"`
	AmountTotal     amount.Amount `gorm:"type:numeric(18,4);default:0" json:"amount_total"`
	AmountResidual  amount.Amount `gorm:"type:numeric(18,4);default:0" json:"amount_residual"`
	TaxRule         *string       `json:"tax_rule"`
	OriginInvoiceID *uint64       `json:"origin_invoice_id"`
}

func (Invoice) TableName() string {
	return "invoices"
}

type InvoiceLine struct {
	model.Base
	InvoiceID      uint64            `gorm:"not null" json:"invoice_id"`
	Sequence       int               `gorm:"default:10" json:"sequence"`
	ItemID         *uint64           `json:"item_id"`
	Description    *string           `json:"description"`
	Qty            float64           `gorm:"type:numeric(18,4)" json:"qty"`
	UnitID         *uint64           `json:"unit_id"`
	UnitPrice      amount.Amount     `gorm:"type:numeric(18,4)" json:"unit_price"`
	DiscountPct    float64           `gorm:"type:numeric(8,4);default:0" json:"discount_pct"`
	TaxIDs         helper.Int64Array `gorm:"type:bigint[]" json:"tax_ids"`
	AccountID      *uint64           `json:"account_id"`
	DimensionID    *uint64           `json:"dimension_id"`
	PriceSubtotal  amount.Amount     `gorm:"type:numeric(18,4)" json:"price_subtotal"`
	SaleLineID     *uint64           `json:"sale_line_id"`
	PurchaseLineID *uint64           `json:"purchase_line_id"`
}

func (InvoiceLine) TableName() string {
	return "invoice_lines"
}

type InvoiceTax struct {
	model.Base
	InvoiceID  uint64        `json:"invoice_id"`
	TaxID      *uint64       `json:"tax_id"`
	BaseAmount amount.Amount `gorm:"column:base;type:numeric(18,4)" json:"base"`
	Amount     amount.Amount `gorm:"type:numeric(18,4)" json:"amount"`
	AccountID  *uint64       `json:"account_id"`
}

func (InvoiceTax) TableName() string {
	return "invoice_taxes"
}

const (
	InstallmentStatePending   = "pending"
	InstallmentStatePaid      = "paid"
	InstallmentStateCancelled = "cancelled"
)

type InvoiceInstallment struct {
	model.Base
	OrganizationID *uint64       `json:"organization_id"`
	InvoiceID      uint64        `gorm:"not null" json:"invoice_id"`
	Sequence       int           `gorm:"default:10" json:"sequence"`
	DueDate        *time.Time    `gorm:"type:date" json:"due_date"`
	Amount         amount.Amount `gorm:"type:numeric(18,4)" json:"amount"`
	State          string        `gorm:"type:text" json:"state"`
}

func (InvoiceInstallment) TableName() string {
	return "invoice_installments"
}

type InvoiceCreditApplication struct {
	model.Base
	OrganizationID *uint64       `json:"organization_id"`
	InvoiceID      uint64        `gorm:"not null" json:"invoice_id"`
	CreditNoteID   uint64        `gorm:"not null" json:"credit_note_id"`
	Amount         amount.Amount `gorm:"type:numeric(18,4)" json:"amount"`
}

func (InvoiceCreditApplication) TableName() string {
	return "invoice_credit_applications"
}

type InvoiceContraSettlement struct {
	model.Base
	OrganizationID    *uint64       `json:"organization_id"`
	CustomerInvoiceID uint64        `gorm:"not null" json:"customer_invoice_id"`
	SupplierBillID    uint64        `gorm:"not null" json:"supplier_bill_id"`
	Amount            amount.Amount `gorm:"type:numeric(18,4)" json:"amount"`
	EntryID           *uint64       `json:"entry_id"`
	Date              *time.Time    `gorm:"type:date" json:"date"`
}

func (InvoiceContraSettlement) TableName() string {
	return "invoice_contra_settlements"
}

type DownPaymentLink struct {
	model.Base
	OrganizationID   *uint64       `json:"organization_id"`
	AdvanceInvoiceID uint64        `gorm:"not null" json:"advance_invoice_id"`
	FinalInvoiceID   uint64        `gorm:"not null" json:"final_invoice_id"`
	Amount           amount.Amount `gorm:"type:numeric(18,4)" json:"amount"`
}

func (DownPaymentLink) TableName() string {
	return "down_payment_links"
}
