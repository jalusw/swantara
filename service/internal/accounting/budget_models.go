package accounting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

const (
	BudgetStateDraft     = "draft"
	BudgetStateApproved  = "approved"
	BudgetStateConfirmed = "confirmed"
	BudgetStateCancelled = "cancelled"
)

type Budget struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	DateStart      *time.Time `gorm:"type:date" json:"date_start"`
	DateEnd        *time.Time `gorm:"type:date" json:"date_end"`
	State          string     `gorm:"type:text" json:"state"`
}

func (Budget) TableName() string {
	return "budgets"
}

type BudgetLine struct {
	model.Base
	BudgetID        uint64  `json:"budget_id"`
	AccountID       uint64  `json:"account_id"`
	DimensionID     *uint64 `json:"dimension_id"`
	PlannedAmount   float64 `gorm:"type:numeric(18,4)" json:"planned_amount"`
	PracticalAmount float64 `gorm:"type:numeric(18,4);default:0" json:"practical_amount"`
}

func (BudgetLine) TableName() string {
	return "budget_lines"
}

const (
	TaxRuleStateActive   = "active"
	TaxRuleStateInactive = "inactive"
)

type TaxRule struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           *string `json:"name"`
	CountryCode    *string `gorm:"type:char(2)" json:"country_code"`
	AutoApply      bool    `gorm:"default:false" json:"auto_apply"`
	Active         bool    `gorm:"default:true" json:"active"`
}

func (TaxRule) TableName() string {
	return "tax_rules"
}

type TaxRuleTaxMap struct {
	model.Base
	TaxRuleID uint64  `json:"tax_rule_id"`
	SrcTaxID  uint64  `json:"src_tax_id"`
	DestTaxID *uint64 `json:"dest_tax_id"`
}

func (TaxRuleTaxMap) TableName() string {
	return "tax_rule_tax_maps"
}

type TaxRuleAccountMap struct {
	model.Base
	TaxRuleID     uint64 `json:"tax_rule_id"`
	SrcAccountID  uint64 `json:"src_account_id"`
	DestAccountID uint64 `json:"dest_account_id"`
}

func (TaxRuleAccountMap) TableName() string {
	return "tax_rule_account_maps"
}

const (
	WithholdingScopeSale     = reference.WithholdingScopeSale
	WithholdingScopePurchase = reference.WithholdingScopePurchase
)

type WithholdingTax = reference.WithholdingTax

const (
	TaxReturnTypeSale     = "sale"
	TaxReturnTypePurchase = "purchase"
	TaxReturnStateDraft   = "draft"
	TaxReturnStateFiled   = "filed"
	TaxReturnStatePaid    = "paid"
)

type TaxReturn struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	PeriodID       uint64     `json:"period_id"`
	Type           string     `gorm:"type:text" json:"type"`
	OutputTax      float64    `gorm:"type:numeric(18,4);default:0" json:"output_tax"`
	InputTax       float64    `gorm:"type:numeric(18,4);default:0" json:"input_tax"`
	NetPayable     float64    `gorm:"type:numeric(18,4);default:0" json:"net_payable"`
	State          string     `gorm:"type:text" json:"state"`
	FiledAt        *time.Time `json:"filed_at"`
}

func (TaxReturn) TableName() string {
	return "tax_returns"
}
