package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

type AccrualHandler struct {
	svc reporting.AccrualService
}

func NewAccrualHandler(svc reporting.AccrualService) AccrualHandler {
	return AccrualHandler{svc: svc}
}

type AccrualResponse struct {
	ID              uint64     `json:"id"`
	OrganizationID  uint64     `json:"organization_id"`
	PeriodID        uint64     `json:"period_id"`
	Name            *string    `json:"name"`
	Description     *string    `json:"description"`
	ReversalDate    *time.Time `json:"reversal_date"`
	State           string     `json:"state"`
	EntryID         *uint64    `json:"entry_id"`
	ReversalEntryID *uint64    `json:"reversal_entry_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func newAccrualResponse(accrual *reporting.Accrual) AccrualResponse {
	return AccrualResponse{
		ID:              accrual.ID,
		OrganizationID:  accrual.OrganizationID,
		PeriodID:        accrual.PeriodID,
		Name:            accrual.Name,
		Description:     accrual.Description,
		ReversalDate:    accrual.ReversalDate,
		State:           accrual.State,
		EntryID:         accrual.EntryID,
		ReversalEntryID: accrual.ReversalEntryID,
		CreatedAt:       accrual.CreatedAt,
		UpdatedAt:       accrual.UpdatedAt,
	}
}
