package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

type PaymentBatchHandler struct {
	svc accounting.PaymentBatchService
}

func NewPaymentBatchHandler(svc accounting.PaymentBatchService) PaymentBatchHandler {
	return PaymentBatchHandler{svc: svc}
}

type PaymentBatchResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID uint64     `json:"organization_id"`
	Name           *string    `json:"name"`
	JournalID      uint64     `json:"journal_id"`
	TotalAmount    float64    `json:"total_amount"`
	PaymentCount   int        `json:"payment_count"`
	State          string     `json:"state"`
	BatchDate      *string    `json:"batch_date"`
	GeneratedAt    *time.Time `json:"generated_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newPaymentBatchResponse(b *accounting.PaymentBatch) PaymentBatchResponse {
	return PaymentBatchResponse{
		ID:             b.ID,
		OrganizationID: b.OrganizationID,
		Name:           b.Name,
		JournalID:      b.JournalID,
		TotalAmount:    b.TotalAmount,
		PaymentCount:   b.PaymentCount,
		State:          b.State,
		BatchDate:      helper.FormatDatePtr(b.BatchDate),
		GeneratedAt:    b.GeneratedAt,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
	}
}

type PaymentBatchLineResponse struct {
	ID        uint64 `json:"id"`
	BatchID   uint64 `json:"batch_id"`
	PaymentID uint64 `json:"payment_id"`
}

func newPaymentBatchLineResponse(l *accounting.PaymentBatchLine) PaymentBatchLineResponse {
	return PaymentBatchLineResponse{
		ID:        l.ID,
		BatchID:   l.BatchID,
		PaymentID: l.PaymentID,
	}
}

var paymentBatchQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"journal_id":      {},
	"name":            {},
	"state":           {},
	"created_at":      {},
	"updated_at":      {},
}
