package accounting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	EntryStateDraft     = "draft"
	EntryStatePosted    = "posted"
	EntryStateCancelled = "cancelled"
)

const (
	OriginTypeStockMovement          = "stock_movement"
	OriginTypeInboundCost            = "inbound_cost"
	OriginTypeStockCount             = "stock_count"
	OriginTypeReversal               = "reversal"
	OriginTypeWithholding            = "withholding"
	OriginTypeAccrual                = "accrual"
	OriginTypeFxRevaluation          = "fx_revaluation"
	OriginTypeProductionOrder        = "production_order"
	OriginTypeOutsideProcessingOrder = "outside_processing_order"
	OriginTypeSaleOrder              = "sale_order"
	OriginTypePurchaseOrder          = "purchase_order"
	OriginTypeExpenseReport          = "expense_report"
	OriginTypeCommission             = "commission"
	OriginTypeServicePart            = "service_part"
	OriginTypeDropship               = "dropship"
	OriginTypeQualityScrap           = "quality_scrap"
	OriginTypeBankFee                = "bank_fee"
	OriginTypeBankInterest           = "bank_interest"
	OriginTypeBadDebt                = "bad_debt"
	OriginTypeAsset                  = "asset"
	OriginTypeAssetDisposal          = "asset_disposal"
	OriginTypeFinancing              = "financing"
	OriginTypeCapitalContribution    = "capital_contribution"
	OriginTypeDividend               = "dividend"
	OriginTypeLoan                   = "loan"
	OriginTypeEquity                 = "equity"
	OriginTypeTaxPayment             = "tax_payment"
)

const (
	AccountTypeIncome       = "income"
	AccountTypeCOGS         = "cogs"
	AccountTypeExpense      = "expense"
	AccountTypeCash         = "cash"
	AccountTypeBank         = "bank"
	AccountTypeReceivable   = "receivable"
	AccountTypePayable      = "payable"
	AccountTypeLiability    = "liability"
	AccountTypeTax          = "tax"
	AccountTypeEquity       = "equity"
	AccountTypeCurrentAsset = "current_asset"
	AccountTypeFixedAsset   = "fixed_asset"
	AccountTypeDepreciation = "depreciation"
)

type JournalEntry struct {
	model.Base
	OrganizationID  uint64     `gorm:"not null" json:"organization_id"`
	JournalID       uint64     `gorm:"not null" json:"journal_id"`
	Name            *string    `json:"name"`
	Date            time.Time  `gorm:"type:date;not null" json:"date"`
	Ref             *string    `json:"ref"`
	State           string     `gorm:"type:text" json:"state"`
	CurrencyCode    *string    `gorm:"type:char(3)" json:"currency_code"`
	OriginType      *string    `json:"origin_type"`
	OriginID        *uint64    `json:"origin_id"`
	ReversedEntryID *uint64    `json:"reversed_entry_id"`
	PostedAt        *time.Time `json:"posted_at"`
	PostedBy        *uint64    `json:"posted_by"`
}

func (JournalEntry) TableName() string {
	return "journal_entrys"
}

func (m JournalEntry) IsPosted() bool {
	return m.State == EntryStatePosted
}

type JournalLine struct {
	model.Base
	EntryID         uint64        `gorm:"not null" json:"entry_id"`
	Date            time.Time     `gorm:"type:date;not null" json:"date"`
	AccountID       uint64        `gorm:"not null" json:"account_id"`
	ContactID       *uint64       `json:"contact_id"`
	Name            *string       `json:"name"`
	Debit           amount.Amount `gorm:"type:numeric(18,4);default:0" json:"debit"`
	Credit          amount.Amount `gorm:"type:numeric(18,4);default:0" json:"credit"`
	CurrencyCode    *string       `gorm:"type:char(3)" json:"currency_code"`
	AmountCurrency  float64       `gorm:"type:numeric(18,4)" json:"amount_currency"`
	DimensionID     *uint64       `json:"dimension_id"`
	TaxID           *uint64       `json:"tax_id"`
	Reconciled      bool          `gorm:"default:false" json:"reconciled"`
	FullReconcileID *uint64       `json:"full_reconcile_id"`
	DueDate         *time.Time    `json:"due_date"`
}

func (JournalLine) TableName() string {
	return "journal_lines"
}
