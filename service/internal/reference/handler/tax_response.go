package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type TaxResponse struct {
	ID                 uint64    `json:"id"`
	OrganizationID     *uint64   `json:"organization_id"`
	Name               string    `json:"name"`
	Amount             *float64  `json:"amount"`
	Type               string    `json:"type"`
	Scope              string    `json:"scope"`
	PriceInclude       bool      `json:"price_include"`
	TaxAccountID       *uint64   `json:"tax_account_id"`
	RefundTaxAccountID *uint64   `json:"refund_tax_account_id"`
	Active             bool      `json:"active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func newTaxResponse(tax *reference.Tax) TaxResponse {
	return TaxResponse{
		ID:                 tax.ID,
		OrganizationID:     tax.OrganizationID,
		Name:               tax.Name,
		Amount:             tax.Amount,
		Type:               tax.Type,
		Scope:              tax.Scope,
		PriceInclude:       tax.PriceInclude,
		TaxAccountID:       tax.TaxAccountID,
		RefundTaxAccountID: tax.RefundTaxAccountID,
		Active:             tax.Active,
		CreatedAt:          tax.CreatedAt,
		UpdatedAt:          tax.UpdatedAt,
	}
}

type ListTaxesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListTaxesResponse `json:"data"`
}
type ListTaxesResponse struct {
	Taxes []TaxResponse `json:"taxes"`
}

type GetTaxResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetTaxResponse `json:"data"`
}
type GetTaxResponse struct {
	Tax TaxResponse `json:"tax"`
}

type CreateTaxResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateTaxResponse `json:"data"`
}
type CreateTaxResponse struct {
	Tax TaxResponse `json:"tax"`
}

type UpdateTaxResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateTaxResponse `json:"data"`
}
type UpdateTaxResponse struct {
	Tax TaxResponse `json:"tax"`
}
