package accounting

import (
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	ReconcileRuleMatchContact    = "contact"
	ReconcileRuleMatchAmount     = "amount"
	ReconcileRuleMatchRef        = "reference"
	ReconcileRuleMatchAccount    = "account"
	ReconcileRuleMatchInvoiceRef = "invoice_ref"
)

type ReconcileRule struct {
	model.Base
	OrganizationID  uint64  `gorm:"not null" json:"organization_id"`
	Name            *string `json:"name"`
	AccountID       *uint64 `json:"account_id"`
	MatchContact    bool    `gorm:"default:false" json:"match_contact"`
	MatchAmount     bool    `gorm:"default:false" json:"match_amount"`
	MatchRef        bool    `gorm:"default:false" json:"match_ref"`
	AmountTolerance float64 `gorm:"type:numeric(18,4);default:0" json:"amount_tolerance"`
	Active          bool    `gorm:"default:true" json:"active"`
	Sequence        int     `gorm:"default:0" json:"sequence"`
}

func (ReconcileRule) TableName() string {
	return "reconcile_rules"
}

type ReconcileRuleMatch struct {
	model.Base
	RuleID       uint64 `gorm:"not null" json:"rule_id"`
	DebitLineID  uint64 `gorm:"not null" json:"debit_line_id"`
	CreditLineID uint64 `gorm:"not null" json:"credit_line_id"`
	Score        int    `gorm:"default:0" json:"score"`
}

func (ReconcileRuleMatch) TableName() string {
	return "reconcile_rule_matches"
}
