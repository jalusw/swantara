package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type PaymentResponse struct {
	ID                   uint64    `json:"id"`
	OrganizationID       *uint64   `json:"organization_id"`
	Name                 *string   `json:"name"`
	ContactID            uint64    `json:"contact_id"`
	Type                 string    `json:"type"`
	JournalID            *uint64   `json:"journal_id"`
	PaymentMethod        *string   `json:"payment_method"`
	Amount               float64   `json:"amount"`
	CurrencyCode         *string   `json:"currency_code"`
	Date                 time.Time `json:"date"`
	Reference            *string   `json:"reference"`
	EntryID              *uint64   `json:"entry_id"`
	ContactBankAccountID *uint64   `json:"contact_bank_account_id"`
	State                string    `json:"state"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func newPaymentResponse(payment *accounting.Payment) PaymentResponse {
	return PaymentResponse{
		ID:                   payment.ID,
		OrganizationID:       payment.OrganizationID,
		Name:                 payment.Name,
		ContactID:            payment.ContactID,
		Type:                 payment.Type,
		JournalID:            payment.JournalID,
		PaymentMethod:        payment.PaymentMethod,
		Amount:               payment.Amount,
		CurrencyCode:         payment.CurrencyCode,
		Date:                 payment.Date,
		Reference:            payment.Reference,
		EntryID:              payment.EntryID,
		ContactBankAccountID: payment.ContactBankAccountID,
		State:                payment.State,
		CreatedAt:            payment.CreatedAt,
		UpdatedAt:            payment.UpdatedAt,
	}
}

type PaymentAllocationResponse struct {
	ID        uint64  `json:"id"`
	PaymentID uint64  `json:"payment_id"`
	InvoiceID uint64  `json:"invoice_id"`
	Amount    float64 `json:"amount"`
}

func newPaymentAllocationResponse(allocation *accounting.PaymentAllocation) PaymentAllocationResponse {
	return PaymentAllocationResponse{
		ID:        allocation.ID,
		PaymentID: allocation.PaymentID,
		InvoiceID: allocation.InvoiceID,
		Amount:    allocation.Amount,
	}
}

type ListPaymentsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPaymentsResponse `json:"data"`
}
type ListPaymentsResponse struct {
	Payments []PaymentResponse `json:"payments"`
}

type GetPaymentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPaymentResponse `json:"data"`
}
type GetPaymentResponse struct {
	Payment     PaymentResponse             `json:"payment"`
	Allocations []PaymentAllocationResponse `json:"allocations"`
}

type CreatePaymentRequest struct {
	OrganizationID    *uint64  `json:"organization_id"`
	ContactID         uint64   `json:"contact_id" validate:"required,gt=0"`
	JournalID         uint64   `json:"journal_id" validate:"required,gt=0"`
	Amount            float64  `json:"amount" validate:"required,gt=0"`
	CurrencyCode      *string  `json:"currency_code" validate:"omitempty,len=3"`
	Date              *string  `json:"date"`
	Reference         string   `json:"reference"`
	InvoiceIDs        []uint64 `json:"invoice_ids" validate:"required,min=1,dive,gt=0"`
	Tolerance         float64  `json:"tolerance" validate:"gte=0"`
	AllowAdvance      bool     `json:"allow_advance"`
	DiscountAmount    float64  `json:"discount_amount" validate:"gte=0"`
	WithholdingAmount float64  `json:"withholding_amount" validate:"gte=0"`
	WriteOffAccountID *uint64  `json:"write_off_account_id"`
}

type CreatePaymentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreatePaymentResponse `json:"data"`
}
type CreatePaymentResponse struct {
	Payment PaymentResponse `json:"payment"`
}
