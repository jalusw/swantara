package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type PdcHandler struct {
	svc accounting.PdcService
}

func NewPdcHandler(svc accounting.PdcService) PdcHandler {
	return PdcHandler{svc: svc}
}

type RegisterPdcRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	ContactID      uint64  `json:"contact_id" validate:"required,gt=0"`
	Direction      string  `json:"direction" validate:"required,oneof=inbound outbound"`
	Number         string  `json:"number"`
	BankName       string  `json:"bank_name"`
	Amount         float64 `json:"amount" validate:"required,gt=0"`
	CurrencyCode   *string `json:"currency_code" validate:"omitempty,len=3"`
	DueDate        *string `json:"due_date" validate:"required"`
	InvoiceID      *uint64 `json:"invoice_id"`
	JournalID      uint64  `json:"journal_id" validate:"required,gt=0"`
}

type PdcInstrumentResponse struct {
	ID           uint64        `json:"id"`
	ContactID    uint64        `json:"contact_id"`
	Direction    string        `json:"direction"`
	Number       *string       `json:"number"`
	BankName     *string       `json:"bank_name"`
	Amount       amount.Amount `json:"amount"`
	CurrencyCode *string       `json:"currency_code"`
	DueDate      *time.Time    `json:"due_date"`
	State        string        `json:"state"`
	InvoiceID    *uint64       `json:"invoice_id"`
	PaymentID    *uint64       `json:"payment_id"`
	JournalID    *uint64       `json:"journal_id"`
}

func newPdcInstrumentResponse(instrument *accounting.PdcInstrument) PdcInstrumentResponse {
	return PdcInstrumentResponse{
		ID:           instrument.ID,
		ContactID:    instrument.ContactID,
		Direction:    instrument.Direction,
		Number:       instrument.Number,
		BankName:     instrument.BankName,
		Amount:       instrument.Amount,
		CurrencyCode: instrument.CurrencyCode,
		DueDate:      instrument.DueDate,
		State:        instrument.State,
		InvoiceID:    instrument.InvoiceID,
		PaymentID:    instrument.PaymentID,
		JournalID:    instrument.JournalID,
	}
}

// @Summary Register PDC instrument
// @Description Registers a post-dated cheque as a held instrument linked to an optional invoice and bank journal.
// @Tags Payments
// @Accept json
// @Produce json
// @Param body body RegisterPdcRequest true "PDC details"
// @Success 201 {object} RegisterPdcResponseEnvelope "PDC registered successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pdc-instruments [post]
func (h PdcHandler) Create(c fiber.Ctx) error {
	var request RegisterPdcRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	dueDate, err := helper.ParseDate(request.DueDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Due date must be in YYYY-MM-DD format.", nil)
	}

	instrument, err := h.svc.Register(c, accounting.RegisterPdcRequest{
		OrganizationID: *organizationID,
		ContactID:      request.ContactID,
		Direction:      request.Direction,
		Number:         request.Number,
		BankName:       request.BankName,
		Amount:         request.Amount,
		CurrencyCode:   request.CurrencyCode,
		DueDate:        *dueDate,
		InvoiceID:      request.InvoiceID,
		JournalID:      request.JournalID,
	})
	if err != nil {
		return writePdcError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "PDC registered successfully.", newPdcInstrumentResponse(instrument))
}

// @Summary Transition PDC instrument
// @Description Moves a PDC through held to deposited to cleared or bounced, posting the bank payment on clear.
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path integer true "PDC instrument ID"
// @Param action path string true "deposit, clear, bounce, or cancel"
// @Success 200 {object} RegisterPdcResponseEnvelope "PDC transitioned successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "PDC not found"
// @Failure 422 {object} httpx.ErrorResponse "Transition not allowed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pdc-instruments/{id}/{action} [post]
func (h PdcHandler) Transition(c fiber.Ctx) error {
	instrumentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid PDC id provided.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	var instrument *accounting.PdcInstrument
	switch c.Params("action") {
	case "deposit":
		instrument, err = h.svc.Deposit(c, instrumentID, *organizationID)
	case "clear":
		instrument, err = h.svc.Clear(c, instrumentID, *organizationID)
	case "bounce":
		instrument, err = h.svc.Bounce(c, instrumentID, *organizationID)
	case "cancel":
		instrument, err = h.svc.Cancel(c, instrumentID, *organizationID)
	default:
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unknown PDC action.", nil)
	}
	if err != nil {
		return writePdcError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "PDC transitioned successfully.", newPdcInstrumentResponse(instrument))
}

func writePdcError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrPdcNotFound):
		return httpx.CreateNotFoundResponse(c, "PDC instrument not found.")
	case errors.Is(err, accounting.ErrPdcInvalidState):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "PDC state transition is not allowed.", nil)
	case errors.Is(err, accounting.ErrPaymentAmount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "PDC amount must be positive.", nil)
	case errors.Is(err, accounting.ErrNoBankAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No bank account configured for the journal.", nil)
	default:
		httpx.RequestLog(c).Error("pdc write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save PDC instrument.", err)
	}
}

type RegisterPdcResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PdcInstrumentResponse `json:"data"`
}
