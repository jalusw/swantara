package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type TaxRuleTaxMapResponse struct {
	ID        uint64  `json:"id"`
	TaxRuleID uint64  `json:"tax_rule_id"`
	SrcTaxID  uint64  `json:"src_tax_id"`
	DestTaxID *uint64 `json:"dest_tax_id"`
}

func newTaxRuleTaxMapResponse(taxMap *accounting.TaxRuleTaxMap) TaxRuleTaxMapResponse {
	return TaxRuleTaxMapResponse{
		ID:        taxMap.ID,
		TaxRuleID: taxMap.TaxRuleID,
		SrcTaxID:  taxMap.SrcTaxID,
		DestTaxID: taxMap.DestTaxID,
	}
}

type TaxRuleAccountMapResponse struct {
	ID            uint64 `json:"id"`
	TaxRuleID     uint64 `json:"tax_rule_id"`
	SrcAccountID  uint64 `json:"src_account_id"`
	DestAccountID uint64 `json:"dest_account_id"`
}

func newTaxRuleAccountMapResponse(accountMap *accounting.TaxRuleAccountMap) TaxRuleAccountMapResponse {
	return TaxRuleAccountMapResponse{
		ID:            accountMap.ID,
		TaxRuleID:     accountMap.TaxRuleID,
		SrcAccountID:  accountMap.SrcAccountID,
		DestAccountID: accountMap.DestAccountID,
	}
}

type TaxRuleResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID *uint64   `json:"organization_id"`
	Name           *string   `json:"name"`
	CountryCode    *string   `json:"country_code"`
	AutoApply      bool      `json:"auto_apply"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newTaxRuleResponse(position *accounting.TaxRule) TaxRuleResponse {
	return TaxRuleResponse{
		ID:             position.ID,
		OrganizationID: position.OrganizationID,
		Name:           position.Name,
		CountryCode:    position.CountryCode,
		AutoApply:      position.AutoApply,
		Active:         position.Active,
		CreatedAt:      position.CreatedAt,
		UpdatedAt:      position.UpdatedAt,
	}
}

type ListTaxRulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListTaxRulesResponse `json:"data"`
}
type ListTaxRulesResponse struct {
	TaxRules []TaxRuleResponse `json:"tax_rules"`
}

type GetTaxRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetTaxRuleResponse `json:"data"`
}
type GetTaxRuleResponse struct {
	TaxRule     TaxRuleResponse             `json:"tax_rule"`
	TaxMaps     []TaxRuleTaxMapResponse     `json:"tax_maps"`
	AccountMaps []TaxRuleAccountMapResponse `json:"account_maps"`
}

type CreateTaxRuleTaxMapRequest struct {
	SrcTaxID  uint64  `json:"src_tax_id" validate:"required,gt=0"`
	DestTaxID *uint64 `json:"dest_tax_id"`
}

type CreateTaxRuleAccountMapRequest struct {
	SrcAccountID  uint64 `json:"src_account_id" validate:"required,gt=0"`
	DestAccountID uint64 `json:"dest_account_id" validate:"required,gt=0"`
}

type CreateTaxRuleRequest struct {
	OrganizationID *uint64                          `json:"organization_id"`
	Name           string                           `json:"name" validate:"required"`
	CountryCode    string                           `json:"country_code" validate:"omitempty,len=2"`
	AutoApply      bool                             `json:"auto_apply"`
	TaxMaps        []CreateTaxRuleTaxMapRequest     `json:"tax_maps" validate:"dive"`
	AccountMaps    []CreateTaxRuleAccountMapRequest `json:"account_maps" validate:"dive"`
}

type CreateTaxRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateTaxRuleResponse `json:"data"`
}
type CreateTaxRuleResponse struct {
	TaxRule TaxRuleResponse `json:"tax_rule"`
}

type ResolveTaxRuleRequest struct {
	TaxID     *uint64 `json:"tax_id"`
	AccountID uint64  `json:"account_id" validate:"required,gt=0"`
}

type ResolveTaxRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ResolveTaxRuleResponse `json:"data"`
}
type ResolveTaxRuleResponse struct {
	TaxID     *uint64 `json:"tax_id"`
	AccountID *uint64 `json:"account_id"`
}
