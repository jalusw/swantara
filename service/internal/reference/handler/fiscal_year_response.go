package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type TaxYearResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID *uint64   `json:"organization_id"`
	Name           string    `json:"name"`
	DateStart      *string   `json:"date_start"`
	DateEnd        *string   `json:"date_end"`
	State          *string   `json:"state"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newTaxYearResponse(year *reference.TaxYear) TaxYearResponse {
	return TaxYearResponse{
		ID:             year.ID,
		OrganizationID: year.OrganizationID,
		Name:           year.Name,
		DateStart:      helper.FormatDatePtr(year.DateStart),
		DateEnd:        helper.FormatDatePtr(year.DateEnd),
		State:          year.State,
		CreatedAt:      year.CreatedAt,
		UpdatedAt:      year.UpdatedAt,
	}
}

type ListTaxYearsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListTaxYearsResponse `json:"data"`
}
type ListTaxYearsResponse struct {
	TaxYears []TaxYearResponse `json:"tax_years"`
}

type GetTaxYearResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetTaxYearResponse `json:"data"`
}
type GetTaxYearResponse struct {
	TaxYear TaxYearResponse `json:"tax_year"`
}

type CreateTaxYearResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateTaxYearResponse `json:"data"`
}
type CreateTaxYearResponse struct {
	TaxYear TaxYearResponse `json:"tax_year"`
}

type UpdateTaxYearResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateTaxYearResponse `json:"data"`
}
type UpdateTaxYearResponse struct {
	TaxYear TaxYearResponse `json:"tax_year"`
}
