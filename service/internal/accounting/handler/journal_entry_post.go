package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type PostingLineRequest struct {
	AccountID      uint64  `json:"account_id" validate:"required,gt=0"`
	DimensionID    *uint64 `json:"dimension_id"`
	Name           string  `json:"name"`
	Debit          float64 `json:"debit" validate:"gte=0"`
	Credit         float64 `json:"credit" validate:"gte=0"`
	ContactID      *uint64 `json:"contact_id"`
	TaxID          *uint64 `json:"tax_id"`
	CurrencyCode   *string `json:"currency_code"`
	AmountCurrency float64 `json:"amount_currency" validate:"gte=0"`
	DueDate        *string `json:"due_date"`
}

type PostJournalEntryRequest struct {
	OrganizationID  *uint64              `json:"organization_id"`
	JournalID       uint64               `json:"journal_id" validate:"required,gt=0"`
	Date            string               `json:"date" validate:"required"`
	Ref             string               `json:"ref"`
	OriginType      string               `json:"origin_type"`
	OriginID        uint64               `json:"origin_id"`
	Description     string               `json:"description"`
	CurrencyCode    *string              `json:"currency_code"`
	Draft           bool                 `json:"draft"`
	AutoReverseDate *string              `json:"auto_reverse_date"`
	Lines           []PostingLineRequest `json:"lines" validate:"required,min=1,dive"`
}

type PostJournalEntryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PostJournalEntryResponse `json:"data"`
}
type PostJournalEntryResponse struct {
	Entry JournalEntryResponse `json:"entry"`
}

// @Summary Post journal entry
// @Description Posts a journal entry to the specified journal after resolving the tenant organization and parsing the entry date. The posting engine validates that the entry has at least one valid line with a non-zero debit or credit, that debits equal credits, and that the target tax period is open, then records the entry as posted with its posted-at timestamp.
// @Tags Journal Entries
// @Accept json
// @Produce json
// @Param body body PostJournalEntryRequest true "Journal entry details"
// @Success 201 {object} PostJournalEntryResponseEnvelope "Journal entry posted successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unbalanced entry"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/account-movements [post]
func (h JournalEntryHandler) Post(c fiber.Ctx) error {
	var request PostJournalEntryRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	date, err := helper.ParseDate(&request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	var autoReverseDate *time.Time
	if request.AutoReverseDate != nil && *request.AutoReverseDate != "" {
		autoReverseDate, err = helper.ParseDate(request.AutoReverseDate)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Auto reverse date must be in YYYY-MM-DD format.", nil)
		}
	}

	lines := make([]accounting.PostingLine, len(request.Lines))
	for i, line := range request.Lines {
		var dueDate *time.Time
		if line.DueDate != nil && *line.DueDate != "" {
			dueDate, err = helper.ParseDate(line.DueDate)
			if err != nil {
				return httpx.CreateUnprocessableEntityErrorResponse(c, "Due date must be in YYYY-MM-DD format.", nil)
			}
		}
		lines[i] = accounting.PostingLine{
			AccountID:      line.AccountID,
			DimensionID:    line.DimensionID,
			Name:           line.Name,
			Debit:          amount.FromFloat64(line.Debit),
			Credit:         amount.FromFloat64(line.Credit),
			ContactID:      line.ContactID,
			TaxID:          line.TaxID,
			CurrencyCode:   line.CurrencyCode,
			AmountCurrency: line.AmountCurrency,
			DueDate:        dueDate,
		}
	}

	entry, err := h.poster.Post(c, accounting.PostRequest{
		OrganizationID:  *organizationID,
		JournalID:       request.JournalID,
		Date:            *date,
		Ref:             request.Ref,
		OriginType:      request.OriginType,
		OriginID:        request.OriginID,
		Description:     request.Description,
		CurrencyCode:    request.CurrencyCode,
		Draft:           request.Draft,
		AutoReverseDate: autoReverseDate,
		Lines:           lines,
	})
	if err != nil {
		return writeJournalEntryError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Journal entry posted successfully.", PostJournalEntryResponse{
		Entry: newJournalEntryResponse(entry),
	})
}
