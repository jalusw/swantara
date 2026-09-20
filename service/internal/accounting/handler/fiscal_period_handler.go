package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

type TaxPeriodHandler struct {
	svc accounting.TaxPeriodService
}

func NewTaxPeriodHandler(svc accounting.TaxPeriodService) TaxPeriodHandler {
	return TaxPeriodHandler{svc: svc}
}

type TaxPeriodResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID uint64    `json:"organization_id"`
	TaxYearID      uint64    `json:"tax_year_id"`
	Name           string    `json:"name"`
	DateStart      *string   `json:"date_start"`
	DateEnd        *string   `json:"date_end"`
	State          string    `json:"state"`
	PeriodType     string    `json:"period_type"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newTaxPeriodResponse(period *accounting.TaxPeriod) TaxPeriodResponse {
	return TaxPeriodResponse{
		ID:             period.ID,
		OrganizationID: period.OrganizationID,
		TaxYearID:      period.TaxYearID,
		Name:           period.Name,
		DateStart:      helper.FormatDatePtr(period.DateStart),
		DateEnd:        helper.FormatDatePtr(period.DateEnd),
		State:          period.State,
		PeriodType:     period.PeriodType,
		CreatedAt:      period.CreatedAt,
		UpdatedAt:      period.UpdatedAt,
	}
}

var taxPeriodQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"tax_year_id":     {},
	"name":            {},
	"state":           {},
	"period_type":     {},
	"created_at":      {},
	"updated_at":      {},
}
