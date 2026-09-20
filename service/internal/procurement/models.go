package procurement

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	RequestStateDraft     = "draft"
	RequestStateConfirmed = "confirmed"
	RequestStateApproved  = "approved"
	RequestStateDone      = "done"
	RequestStateCancelled = "cancelled"

	PurchaseOrderStateDraft     = "draft"
	PurchaseOrderStateSent      = "sent"
	PurchaseOrderStateConfirmed = "confirmed"
	PurchaseOrderStateDone      = "done"
	PurchaseOrderStateCancelled = "cancelled"

	QuoteRequestStateDraft     = "draft"
	QuoteRequestStateSent      = "sent"
	QuoteRequestStateDone      = "done"
	QuoteRequestStateCancelled = "cancelled"

	SupplierQuoteStateDraft     = "draft"
	SupplierQuoteStateSubmitted = "submitted"
	SupplierQuoteStateAccepted  = "accepted"
	SupplierQuoteStateRejected  = "rejected"

	PurchaseOrderInvoiceStatusNo        = "no"
	PurchaseOrderInvoiceStatusToInvoice = "to_invoice"
	PurchaseOrderInvoiceStatusInvoiced  = "invoiced"

	PurchaseOrderReceiptStatusPending = "pending"
	PurchaseOrderReceiptStatusPartial = "partial"
	PurchaseOrderReceiptStatusDone    = "done"

	SupplyAgreementStateDraft     = "draft"
	SupplyAgreementStateActive    = "active"
	SupplyAgreementStateExpired   = "expired"
	SupplyAgreementStateCancelled = "cancelled"

	PaymentBatchStateDraft  = "draft"
	PaymentBatchStatePosted = "posted"

	CreditMemoStatePosted = "posted"
	DebitMemoStatePosted  = "posted"

	SequencePurchaseOrderCode        = "purchase_order"
	SequencePurchaseRequestCode      = "purchase_request"
	SequenceSupplierQuoteRequestCode = "purchase_rfq"
	SequenceSupplyAgreementCode      = "supply_agreement"

	ApprovalOwnerType = "purchase_order"
)

type PurchaseRequest struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	RequesterID    uint64     `gorm:"not null" json:"requester_id"`
	DepartmentID   *uint64    `json:"department_id"`
	State          string     `gorm:"type:text" json:"state"`
	NeededBy       *time.Time `gorm:"type:date" json:"needed_by"`
}

func (PurchaseRequest) TableName() string {
	return "purchase_requests"
}

type PurchaseRequestLine struct {
	model.Base
	RequestID    uint64     `json:"request_id"`
	ItemID       *uint64    `json:"item_id"`
	Description  *string    `json:"description"`
	Qty          float64    `gorm:"type:numeric(18,4)" json:"qty"`
	UnitID       *uint64    `json:"unit_id"`
	NeededBy     *time.Time `gorm:"type:date" json:"needed_by"`
	CostCenterID *uint64    `json:"cost_center_id"`
}

func (PurchaseRequestLine) TableName() string {
	return "purchase_request_lines"
}

type PurchaseOrder struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	SupplierID     uint64     `gorm:"not null" json:"supplier_id"`
	SupplierRef    *string    `json:"supplier_ref"`
	CurrencyCode   *string    `gorm:"type:char(3)" json:"currency_code"`
	WarehouseID    *uint64    `json:"warehouse_id"`
	DestLocationID *uint64    `json:"dest_location_id"`
	State          string     `gorm:"type:text" json:"state"`
	OrderDate      *time.Time `gorm:"type:date" json:"order_date"`
	ExpectedDate   *time.Time `gorm:"type:date" json:"expected_date"`
	PaymentTermID  *uint64    `json:"payment_term_id"`
	Incoterm       *string    `json:"incoterm"`
	AmountUntaxed  float64    `gorm:"type:numeric(18,4);default:0" json:"amount_untaxed"`
	AmountTax      float64    `gorm:"type:numeric(18,4);default:0" json:"amount_tax"`
	AmountTotal    float64    `gorm:"type:numeric(18,4);default:0" json:"amount_total"`
	InvoiceStatus  string     `gorm:"type:text" json:"invoice_status"`
	ReceiptStatus  string     `gorm:"type:text" json:"receipt_status"`
}

func (PurchaseOrder) TableName() string {
	return "purchase_orders"
}

type PurchaseOrderLine struct {
	model.Base
	OrderID       uint64            `gorm:"not null" json:"order_id"`
	Sequence      int               `gorm:"default:10" json:"sequence"`
	ItemID        *uint64           `json:"item_id"`
	Description   *string           `json:"description"`
	QtyOrdered    float64           `gorm:"type:numeric(18,4);not null" json:"qty_ordered"`
	QtyReceived   float64           `gorm:"type:numeric(18,4);default:0" json:"qty_received"`
	QtyBilled     float64           `gorm:"type:numeric(18,4);default:0" json:"qty_billed"`
	QtyReturns    float64           `gorm:"type:numeric(18,4);default:0" json:"qty_returns"`
	UnitID        *uint64           `json:"unit_id"`
	UnitPrice     float64           `gorm:"type:numeric(18,4)" json:"unit_price"`
	DiscountPct   float64           `gorm:"type:numeric(8,4);default:0" json:"discount_pct"`
	TaxIDs        helper.Int64Array `gorm:"type:bigint[]" json:"tax_ids"`
	DimensionID   *uint64           `json:"dimension_id"`
	CostCenterID  *uint64           `json:"cost_center_id"`
	PriceSubtotal float64           `gorm:"type:numeric(18,4)" json:"price_subtotal"`
}

func (PurchaseOrderLine) TableName() string {
	return "purchase_order_lines"
}

type SupplierQuoteRequest struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	RequesterID    uint64     `gorm:"not null" json:"requester_id"`
	SupplierID     *uint64    `json:"supplier_id"`
	CurrencyCode   *string    `gorm:"type:char(3)" json:"currency_code"`
	State          string     `gorm:"type:text" json:"state"`
	OrderDate      *time.Time `gorm:"type:date" json:"order_date"`
	QuoteDeadline  *time.Time `gorm:"type:date" json:"quote_deadline"`
	Notes          *string    `json:"notes"`
}

func (SupplierQuoteRequest) TableName() string {
	return "supplier_quoteRequests"
}

type SupplierQuoteRequestLine struct {
	model.Base
	QuoteRequestID uint64     `json:"quoteRequest_id"`
	ItemID         *uint64    `json:"item_id"`
	Description    *string    `json:"description"`
	Qty            float64    `gorm:"type:numeric(18,4)" json:"qty"`
	UnitID         *uint64    `json:"unit_id"`
	NeededBy       *time.Time `gorm:"type:date" json:"needed_by"`
}

func (SupplierQuoteRequestLine) TableName() string {
	return "supplier_quoteRequest_lines"
}

type SupplierQuote struct {
	model.Base
	QuoteRequestID uint64     `gorm:"not null" json:"quoteRequest_id"`
	SupplierID     uint64     `gorm:"not null" json:"supplier_id"`
	CurrencyCode   *string    `gorm:"type:char(3)" json:"currency_code"`
	State          string     `gorm:"type:text" json:"state"`
	QuoteDate      *time.Time `gorm:"type:date" json:"quote_date"`
	ValidUntil     *time.Time `gorm:"type:date" json:"valid_until"`
	Notes          *string    `json:"notes"`
	AmountUntaxed  float64    `gorm:"type:numeric(18,4);default:0" json:"amount_untaxed"`
	AmountTax      float64    `gorm:"type:numeric(18,4);default:0" json:"amount_tax"`
	AmountTotal    float64    `gorm:"type:numeric(18,4);default:0" json:"amount_total"`
}

func (SupplierQuote) TableName() string {
	return "supplier_quotes"
}

type SupplierQuoteLine struct {
	model.Base
	SupplierQuoteID    uint64  `gorm:"not null" json:"supplier_quote_id"`
	QuoteRequestLineID uint64  `gorm:"not null" json:"quoteRequest_line_id"`
	ItemID             *uint64 `json:"item_id"`
	Description        *string `json:"description"`
	Qty                float64 `gorm:"type:numeric(18,4)" json:"qty"`
	UnitPrice          float64 `gorm:"type:numeric(18,4)" json:"unit_price"`
	DiscountPct        float64 `gorm:"type:numeric(8,4);default:0" json:"discount_pct"`
	PriceSubtotal      float64 `gorm:"type:numeric(18,4);default:0" json:"price_subtotal"`
}

func (SupplierQuoteLine) TableName() string {
	return "supplier_quote_lines"
}

type CurrencyRate struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	FromCurrency   string     `gorm:"type:char(3);not null" json:"from_currency"`
	ToCurrency     string     `gorm:"type:char(3);not null" json:"to_currency"`
	Rate           float64    `gorm:"type:numeric(18,8);not null" json:"rate"`
	RateDate       *time.Time `gorm:"type:date;not null" json:"rate_date"`
}

func (CurrencyRate) TableName() string {
	return "currency_rates"
}

type SupplyAgreement struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	SupplierID     uint64     `gorm:"not null" json:"supplier_id"`
	CurrencyCode   *string    `gorm:"type:char(3)" json:"currency_code"`
	State          string     `gorm:"type:text" json:"state"`
	StartDate      *time.Time `gorm:"type:date" json:"start_date"`
	EndDate        *time.Time `gorm:"type:date" json:"end_date"`
	AmountLimit    float64    `gorm:"type:numeric(18,4);default:0" json:"amount_limit"`
	QtyLimit       float64    `gorm:"type:numeric(18,4);default:0" json:"qty_limit"`
	ConsumedAmount float64    `gorm:"type:numeric(18,4);default:0" json:"consumed_amount"`
	ConsumedQty    float64    `gorm:"type:numeric(18,4);default:0" json:"consumed_qty"`
}

func (SupplyAgreement) TableName() string {
	return "supply_agreements"
}

type SupplyAgreementLine struct {
	model.Base
	AgreementID uint64     `gorm:"not null" json:"agreement_id"`
	ItemID      *uint64    `json:"item_id"`
	Description *string    `json:"description"`
	Qty         float64    `gorm:"type:numeric(18,4)" json:"qty"`
	UnitPrice   float64    `gorm:"type:numeric(18,4)" json:"unit_price"`
	UnitID      *uint64    `json:"unit_id"`
	NeededBy    *time.Time `gorm:"type:date" json:"needed_by"`
}

func (SupplyAgreementLine) TableName() string {
	return "supply_agreement_lines"
}

type SupplierScorecard struct {
	model.Base
	OrganizationID   *uint64    `json:"organization_id"`
	SupplierID       uint64     `gorm:"not null" json:"supplier_id"`
	PeriodStart      *time.Time `gorm:"type:date" json:"period_start"`
	PeriodEnd        *time.Time `gorm:"type:date" json:"period_end"`
	QualityScore     float64    `gorm:"type:numeric(5,2);default:0" json:"quality_score"`
	DeliveryScore    float64    `gorm:"type:numeric(5,2);default:0" json:"delivery_score"`
	PriceScore       float64    `gorm:"type:numeric(5,2);default:0" json:"price_score"`
	OverallScore     float64    `gorm:"type:numeric(5,2);default:0" json:"overall_score"`
	TotalOrders      int        `gorm:"default:0" json:"total_orders"`
	OnTimeDeliveries int        `gorm:"default:0" json:"on_time_deliveries"`
	QualityFailures  int        `gorm:"default:0" json:"quality_failures"`
}

func (SupplierScorecard) TableName() string {
	return "supplier_scorecards"
}

type CostCenter struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           *string `json:"name"`
	Code           *string `json:"code"`
	Active         bool    `gorm:"default:true" json:"active"`
}

func (CostCenter) TableName() string {
	return "cost_centers"
}

type PurchaseCreditMemo struct {
	model.Base
	OrderID       uint64     `gorm:"not null" json:"order_id"`
	JournalID     uint64     `gorm:"not null" json:"journal_id"`
	ContactID     uint64     `gorm:"not null" json:"contact_id"`
	Date          *time.Time `gorm:"type:date" json:"date"`
	State         string     `gorm:"type:text" json:"state"`
	AmountUntaxed float64    `gorm:"type:numeric(18,4);default:0" json:"amount_untaxed"`
	AmountTax     float64    `gorm:"type:numeric(18,4);default:0" json:"amount_tax"`
	AmountTotal   float64    `gorm:"type:numeric(18,4);default:0" json:"amount_total"`
	Reason        *string    `json:"reason"`
}

func (PurchaseCreditMemo) TableName() string {
	return "purchase_credit_memos"
}

type PurchaseDebitMemo struct {
	model.Base
	OrderID       uint64     `gorm:"not null" json:"order_id"`
	JournalID     uint64     `gorm:"not null" json:"journal_id"`
	ContactID     uint64     `gorm:"not null" json:"contact_id"`
	Date          *time.Time `gorm:"type:date" json:"date"`
	State         string     `gorm:"type:text" json:"state"`
	AmountUntaxed float64    `gorm:"type:numeric(18,4);default:0" json:"amount_untaxed"`
	AmountTax     float64    `gorm:"type:numeric(18,4);default:0" json:"amount_tax"`
	AmountTotal   float64    `gorm:"type:numeric(18,4);default:0" json:"amount_total"`
	Reason        *string    `json:"reason"`
}

func (PurchaseDebitMemo) TableName() string {
	return "purchase_debit_memos"
}

type PaymentBatch struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	JournalID      uint64     `gorm:"not null" json:"journal_id"`
	ContactID      uint64     `gorm:"not null" json:"contact_id"`
	Date           *time.Time `gorm:"type:date" json:"date"`
	State          string     `gorm:"type:text" json:"state"`
	TotalAmount    float64    `gorm:"type:numeric(18,4);default:0" json:"total_amount"`
	PaymentCount   int        `gorm:"default:0" json:"payment_count"`
}

func (PaymentBatch) TableName() string {
	return "payment_batches"
}

type PaymentBatchLine struct {
	model.Base
	BatchID uint64  `gorm:"not null" json:"batch_id"`
	OrderID uint64  `gorm:"not null" json:"order_id"`
	Amount  float64 `gorm:"type:numeric(18,4)" json:"amount"`
}

func (PaymentBatchLine) TableName() string {
	return "payment_batch_lines"
}
