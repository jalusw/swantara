package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type PaymentTermLineResponse struct {
	ID           uint64   `json:"id"`
	Sequence     int      `json:"sequence"`
	ValueType    string   `json:"value_type"`
	Value        float64  `json:"value"`
	DaysAfter    int      `json:"days_after"`
	DayOfMonth   *int     `json:"day_of_month"`
	DiscountPct  *float64 `json:"discount_pct"`
	DiscountDays *int     `json:"discount_days"`
}

func newPaymentTermLineResponse(line *reference.PaymentTermLine) PaymentTermLineResponse {
	return PaymentTermLineResponse{
		ID:           line.ID,
		Sequence:     line.Sequence,
		ValueType:    line.ValueType,
		Value:        line.Value,
		DaysAfter:    line.DaysAfter,
		DayOfMonth:   line.DayOfMonth,
		DiscountPct:  line.DiscountPct,
		DiscountDays: line.DiscountDays,
	}
}

type PaymentTermResponse struct {
	ID             uint64                    `json:"id"`
	OrganizationID uint64                    `json:"organization_id"`
	Name           string                    `json:"name"`
	Note           *string                   `json:"note"`
	Code           *string                   `json:"code"`
	IsActive       bool                      `json:"is_active"`
	TemplateKey    *string                   `json:"template_key"`
	Lines          []PaymentTermLineResponse `json:"lines"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

type ListPaymentTermsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPaymentTermsResponse `json:"data"`
}

type ListPaymentTermsResponse struct {
	PaymentTerms []PaymentTermResponse `json:"payment_terms"`
}

type GetPaymentTermResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPaymentTermResponse `json:"data"`
}

type GetPaymentTermResponse struct {
	PaymentTerm PaymentTermResponse `json:"payment_term"`
}

type CreatePaymentTermLineRequest struct {
	Sequence     int      `json:"sequence"`
	ValueType    string   `json:"value_type" validate:"required,oneof=percent fixed balance"`
	Value        float64  `json:"value"`
	DaysAfter    int      `json:"days_after"`
	DayOfMonth   *int     `json:"day_of_month"`
	DiscountPct  *float64 `json:"discount_pct"`
	DiscountDays *int     `json:"discount_days"`
}

type CreatePaymentTermRequest struct {
	Name        string                         `json:"name" validate:"required"`
	Note        *string                        `json:"note"`
	Code        *string                        `json:"code"`
	IsActive    *bool                          `json:"is_active"`
	TemplateKey *string                        `json:"template_key"`
	Lines       []CreatePaymentTermLineRequest `json:"lines" validate:"required,min=1,dive"`
}

type CreatePaymentTermResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreatePaymentTermResponse `json:"data"`
}

type CreatePaymentTermResponse struct {
	PaymentTerm PaymentTermResponse `json:"payment_term"`
}

type UpdatePaymentTermRequest struct {
	Name        string                         `json:"name" validate:"required"`
	Note        *string                        `json:"note"`
	Code        *string                        `json:"code"`
	IsActive    *bool                          `json:"is_active"`
	TemplateKey *string                        `json:"template_key"`
	Lines       []CreatePaymentTermLineRequest `json:"lines" validate:"required,min=1,dive"`
}

type UpdatePaymentTermResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdatePaymentTermResponse `json:"data"`
}

type UpdatePaymentTermResponse struct {
	PaymentTerm PaymentTermResponse `json:"payment_term"`
}
