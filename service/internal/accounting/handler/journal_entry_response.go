package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type JournalEntryResponse struct {
	ID              uint64     `json:"id"`
	OrganizationID  uint64     `json:"organization_id"`
	JournalID       uint64     `json:"journal_id"`
	Name            *string    `json:"name"`
	Date            time.Time  `json:"date"`
	Ref             *string    `json:"ref"`
	State           string     `json:"state"`
	CurrencyCode    *string    `json:"currency_code"`
	OriginType      *string    `json:"origin_type"`
	OriginID        *uint64    `json:"origin_id"`
	ReversedEntryID *uint64    `json:"reversed_entry_id"`
	PostedAt        *time.Time `json:"posted_at"`
	PostedBy        *uint64    `json:"posted_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func newJournalEntryResponse(entry *accounting.JournalEntry) JournalEntryResponse {
	return JournalEntryResponse{
		ID:              entry.ID,
		OrganizationID:  entry.OrganizationID,
		JournalID:       entry.JournalID,
		Name:            entry.Name,
		Date:            entry.Date,
		Ref:             entry.Ref,
		State:           entry.State,
		CurrencyCode:    entry.CurrencyCode,
		OriginType:      entry.OriginType,
		OriginID:        entry.OriginID,
		ReversedEntryID: entry.ReversedEntryID,
		PostedAt:        entry.PostedAt,
		PostedBy:        entry.PostedBy,
		CreatedAt:       entry.CreatedAt,
		UpdatedAt:       entry.UpdatedAt,
	}
}

type JournalLineResponse struct {
	ID              uint64        `json:"id"`
	EntryID         uint64        `json:"entry_id"`
	AccountID       uint64        `json:"account_id"`
	ContactID       *uint64       `json:"contact_id"`
	Name            *string       `json:"name"`
	Debit           amount.Amount `json:"debit"`
	Credit          amount.Amount `json:"credit"`
	CurrencyCode    *string       `json:"currency_code"`
	AmountCurrency  float64       `json:"amount_currency"`
	DimensionID     *uint64       `json:"dimension_id"`
	TaxID           *uint64       `json:"tax_id"`
	Reconciled      bool          `json:"reconciled"`
	FullReconcileID *uint64       `json:"full_reconcile_id"`
	DueDate         *time.Time    `json:"due_date"`
}

func newJournalLineResponse(line *accounting.JournalLine) JournalLineResponse {
	return JournalLineResponse{
		ID:              line.ID,
		EntryID:         line.EntryID,
		AccountID:       line.AccountID,
		ContactID:       line.ContactID,
		Name:            line.Name,
		Debit:           line.Debit,
		Credit:          line.Credit,
		CurrencyCode:    line.CurrencyCode,
		AmountCurrency:  line.AmountCurrency,
		DimensionID:     line.DimensionID,
		TaxID:           line.TaxID,
		Reconciled:      line.Reconciled,
		FullReconcileID: line.FullReconcileID,
		DueDate:         line.DueDate,
	}
}
