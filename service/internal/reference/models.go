package reference

import (
	"encoding/json"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type Currency struct {
	model.Base
	Code          string  `gorm:"type:char(3);uniqueIndex;not null" json:"code"`
	Name          string  `gorm:"not null" json:"name"`
	Symbol        *string `json:"symbol"`
	DecimalPlaces int16   `json:"decimal_places"`
	Rounding      float64 `gorm:"type:numeric(12,6);default:0.01" json:"rounding"`
}

type Organization struct {
	model.Base
	Name              string  `json:"name"`
	LegalName         *string `json:"legal_name"`
	ParentID          *uint64 `json:"parent_id"`
	BaseCurrency      string  `gorm:"type:char(3);not null" json:"base_currency"`
	CountryCode       *string `gorm:"type:char(2)" json:"country_code"`
	TaxID             *string `json:"tax_id" audit:"redact"`
	Logo              *string `json:"logo"`
	Timezone          string  `gorm:"default:UTC" json:"timezone"`
	TaxYearStartMonth int16   `gorm:"default:1" json:"tax_year_start_month"`
	AutoCheckoutHour  *int    `json:"auto_checkout_hour"`
	RoundingMinutes   int     `gorm:"default:0" json:"rounding_minutes"`
}

type OrganizationModule struct {
	model.Base
	OrganizationID uint64 `gorm:"not null" json:"organization_id"`
	ModuleID       string `gorm:"type:text;not null" json:"module_id"`
	Active         bool   `gorm:"default:true" json:"active"`
}

type FxRate struct {
	model.Base
	CurrencyCode   string    `gorm:"type:char(3);not null" json:"currency_code"`
	OrganizationID *uint64   `json:"organization_id"`
	Rate           float64   `gorm:"type:numeric(18,8);not null" json:"rate"`
	RateType       string    `gorm:"default:spot" json:"rate_type"`
	ValidFrom      time.Time `gorm:"type:date;not null" json:"valid_from"`
}

type UnitGroup struct {
	model.Base
	Name string `json:"name"`
}

func (UnitGroup) TableName() string {
	return "unit_groups"
}

type Unit struct {
	model.Base
	CategoryID uint64  `json:"category_id"`
	Name       string  `json:"name"`
	Factor     float64 `gorm:"type:numeric(28,8);not null" json:"factor"`
	UnitType   string  `gorm:"type:text" json:"unit_type"`
	Rounding   float64 `gorm:"type:numeric(12,6)" json:"rounding"`
}

func (Unit) TableName() string {
	return "units"
}

type PaymentTerm struct {
	model.Base
	OrganizationID uint64  `gorm:"not null" json:"organization_id"`
	Name           string  `json:"name"`
	Note           *string `json:"note"`
	Code           *string `json:"code"`
	IsActive       bool    `gorm:"default:true" json:"is_active"`
	TemplateKey    *string `json:"template_key"`
}

type PaymentTermLine struct {
	model.Base
	PaymentTermID uint64   `json:"payment_term_id"`
	Sequence      int      `gorm:"default:10" json:"sequence"`
	ValueType     string   `gorm:"type:text" json:"value_type"`
	Value         float64  `gorm:"type:numeric(12,4)" json:"value"`
	DaysAfter     int      `gorm:"default:0" json:"days_after"`
	DayOfMonth    *int     `json:"day_of_month"`
	DiscountPct   *float64 `gorm:"type:numeric(8,4)" json:"discount_pct"`
	DiscountDays  *int     `json:"discount_days"`
}

type Account struct {
	model.Base
	OrganizationID uint64  `gorm:"not null" json:"organization_id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Type           string  `gorm:"type:text" json:"type"`
	Reconcilable   bool    `json:"reconcilable"`
	CurrencyCode   *string `gorm:"type:char(3)" json:"currency_code"`
	ParentID       *uint64 `json:"parent_id"`
	Active         bool    `gorm:"default:true" json:"active"`
}

type Journal struct {
	model.Base
	OrganizationID   uint64  `gorm:"not null" json:"organization_id"`
	Name             string  `json:"name"`
	Code             *string `json:"code"`
	Type             string  `gorm:"type:text" json:"type"`
	DefaultAccountID *uint64 `json:"default_account_id"`
	CurrencyCode     *string `gorm:"type:char(3)" json:"currency_code"`
	BankAccountID    *uint64 `json:"bank_account_id"`
	SequenceID       *uint64 `json:"sequence_id"`
}

type Dimension struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	Code           *string `json:"code"`
	Kind           *string `json:"kind"`
	ParentID       *uint64 `json:"parent_id"`
	Active         *bool   `gorm:"default:true" json:"active"`
}

const (
	TaxTypePercent = "percent"
	TaxTypeFixed   = "fixed"
	TaxTypeGroup   = "group"

	TaxScopeSale     = "sale"
	TaxScopePurchase = "purchase"
	TaxScopeNone     = "none"
)

type Tax struct {
	model.Base
	OrganizationID     *uint64  `json:"organization_id"`
	Name               string   `json:"name"`
	Amount             *float64 `gorm:"type:numeric(8,4)" json:"amount"`
	Type               string   `gorm:"type:text" json:"type"`
	Scope              string   `gorm:"type:text" json:"scope"`
	PriceInclude       bool     `json:"price_include"`
	TaxAccountID       *uint64  `json:"tax_account_id"`
	RefundTaxAccountID *uint64  `json:"refund_tax_account_id"`
	Active             bool     `gorm:"default:true" json:"active"`
}

type TaxYear struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           string     `json:"name"`
	DateStart      *time.Time `gorm:"type:date" json:"date_start"`
	DateEnd        *time.Time `gorm:"type:date" json:"date_end"`
	State          *string    `json:"state"`
}

type ItemCategory struct {
	model.Base
	OrganizationID          *uint64 `json:"organization_id"`
	Name                    string  `json:"name"`
	ParentID                *uint64 `json:"parent_id"`
	IncomeAccountID         *uint64 `json:"income_account_id"`
	ExpenseAccountID        *uint64 `json:"expense_account_id"`
	StockValuationAccountID *uint64 `json:"stock_cost_account_id"`
	StockInputAccountID     *uint64 `json:"stock_input_account_id"`
	StockOutputAccountID    *uint64 `json:"stock_output_account_id"`
	CogsAccountID           *uint64 `json:"cogs_account_id"`
	CostMethod              *string `gorm:"type:text" json:"cost_method"`
	Valuation               *string `gorm:"type:text" json:"valuation"`
}

type ItemAttribute struct {
	model.Base
	Name string `json:"name"`
}

type ItemAttributeValue struct {
	model.Base
	AttributeID uint64 `json:"attribute_id"`
	Value       string `json:"value"`
}

type PipelineStage struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	Sequence       int     `json:"sequence"`
	IsWon          bool    `json:"is_won"`
	Probability    float64 `gorm:"type:numeric(5,2)" json:"probability"`
}

func (PipelineStage) TableName() string {
	return "pipeline_stages"
}

type SalesGroup struct {
	model.Base
	Name           string  `json:"name"`
	LeaderID       *uint64 `json:"leader_id"`
	OrganizationID *uint64 `json:"organization_id"`
}

type Carrier struct {
	model.Base
	Name           string  `json:"name"`
	TrackingURLTpl *string `json:"tracking_url_tpl"`
	DeliveryItemID *uint64 `json:"delivery_item_id"`
}

type ReminderLevel struct {
	model.Base
	Name        string `json:"name"`
	DaysOverdue int    `json:"days_overdue"`
	Sequence    int    `json:"sequence"`
}

type Warehouse struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	Code           *string `json:"code"`
	Line1          *string `json:"line1"`
	Line2          *string `json:"line2"`
	City           *string `json:"city"`
	State          *string `json:"state"`
	PostalCode     *string `json:"postal_code"`
	CountryCode    *string `json:"country_code"`
}

type StockLocation struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	WarehouseID    *uint64 `json:"warehouse_id"`
	Name           string  `json:"name"`
	Code           *string `json:"code"`
	ParentID       *uint64 `json:"parent_id"`
	Usage          string  `gorm:"type:text" json:"usage"`
	Barcode        *string `json:"barcode"`
}

type Department struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	ParentID       *uint64 `json:"parent_id"`
	ManagerID      *uint64 `json:"manager_id"`
	DimensionID    *uint64 `json:"dimension_id"`
}

type JobPosition struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	DepartmentID   *uint64 `json:"department_id"`
}

type WorkCenter struct {
	model.Base
	OrganizationID  *uint64  `json:"organization_id"`
	Name            string   `json:"name"`
	Code            *string  `json:"code"`
	CostPerHour     *float64 `gorm:"type:numeric(18,4)" json:"cost_per_hour"`
	CapacityPerHour *float64 `gorm:"type:numeric(18,4)" json:"capacity_per_hour"`
	EfficiencyPct   float64  `gorm:"type:numeric(6,2);default:100" json:"efficiency_pct"`
	OeeTarget       *float64 `gorm:"type:numeric(6,2)" json:"oee_target"`
	DimensionID     *uint64  `json:"dimension_id"`
}

type LeaveType struct {
	model.Base
	OrganizationID *uint64  `json:"organization_id"`
	Name           string   `json:"name"`
	Paid           bool     `json:"paid"`
	AllocationDays *float64 `gorm:"type:numeric(6,2)" json:"allocation_days"`
}

type SalaryRule struct {
	model.Base
	OrganizationID  *uint64  `json:"organization_id"`
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Category        *string  `json:"category"`
	ComputeType     *string  `json:"compute_type"`
	Amount          *float64 `gorm:"type:numeric(18,4)" json:"amount"`
	Formula         *string  `json:"formula"`
	AccountDebitID  *uint64  `json:"account_debit_id"`
	AccountCreditID *uint64  `json:"account_credit_id"`
}

type AssetCategory struct {
	model.Base
	Name                  string  `json:"name"`
	OrganizationID        *uint64 `json:"organization_id"`
	AssetAccountID        *uint64 `json:"asset_account_id"`
	DepreciationAccountID *uint64 `json:"depreciation_account_id"`
	ExpenseAccountID      *uint64 `json:"expense_account_id"`
	GainAccountID         *uint64 `json:"gain_account_id"`
	LossAccountID         *uint64 `json:"loss_account_id"`
	Method                *string `gorm:"type:text" json:"method"`
	MethodNumber          *int    `json:"method_number"`
	MethodPeriod          *string `json:"method_period"`
}

type QualityPoint struct {
	model.Base
	OrganizationID *uint64  `json:"organization_id"`
	ItemID         *uint64  `json:"item_id"`
	Operation      *string  `json:"operation"`
	TestType       string   `gorm:"type:text" json:"test_type"`
	NormMin        *float64 `gorm:"type:numeric(18,4)" json:"norm_min"`
	NormMax        *float64 `gorm:"type:numeric(18,4)" json:"norm_max"`
	UnitID         *uint64  `json:"unit_id"`
}

type POSConfig struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	WarehouseID    *uint64 `json:"warehouse_id"`
	JournalID      *uint64 `json:"journal_id"`
	PriceBookID    *uint64 `json:"price_book_id"`
}

func (POSConfig) TableName() string {
	return "pos_configs"
}

type POSPaymentAccount struct {
	model.Base
	OrganizationID uint64 `gorm:"not null" json:"organization_id"`
	Method         string `gorm:"type:text;not null" json:"method"`
	AccountID      uint64 `gorm:"not null" json:"account_id"`
}

func (POSPaymentAccount) TableName() string {
	return "pos_payment_accounts"
}

const (
	WithholdingScopeSale     = "sale"
	WithholdingScopePurchase = "purchase"
)

type WithholdingTax struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           *string `json:"name"`
	RatePct        float64 `gorm:"type:numeric(8,4)" json:"rate_pct"`
	AccountID      *uint64 `json:"account_id"`
	Scope          string  `gorm:"type:text" json:"scope"`
	Active         bool    `gorm:"default:true" json:"active"`
}

func (WithholdingTax) TableName() string {
	return "withholding_taxes"
}

type ExpenseCategory struct {
	model.Base
	OrganizationID   *uint64           `json:"organization_id"`
	Name             string            `json:"name"`
	ExpenseAccountID *uint64           `json:"expense_account_id"`
	DefaultTaxIDs    helper.Int64Array `gorm:"type:bigint[]" json:"default_tax_ids"`
}

type SubscriptionPlan struct {
	model.Base
	OrganizationID    *uint64 `json:"organization_id"`
	Name              string  `json:"name"`
	RecurringInterval string  `gorm:"type:text" json:"recurring_interval"`
	RecurringCount    int     `gorm:"default:1" json:"recurring_count"`
}

type SystemConfig struct {
	model.Base
	OrganizationID *uint64         `json:"organization_id"`
	Key            string          `json:"key"`
	Value          json.RawMessage `gorm:"type:jsonb" json:"value" swaggertype:"object"`
}
