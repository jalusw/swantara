package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ReconcileHandler struct {
	svc accounting.ReconcileService
}

func NewReconcileHandler(svc accounting.ReconcileService) ReconcileHandler {
	return ReconcileHandler{svc: svc}
}

type ReconcileRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	DebitLineID    uint64  `json:"debit_line_id" validate:"required,gt=0"`
	CreditLineID   uint64  `json:"credit_line_id" validate:"required,gt=0"`
	Amount         float64 `json:"amount" validate:"required,gt=0"`
}

type AccountPartialReconcileResponse struct {
	ID              uint64  `json:"id"`
	DebitLineID     uint64  `json:"debit_line_id"`
	CreditLineID    uint64  `json:"credit_line_id"`
	Amount          float64 `json:"amount"`
	FullReconcileID *uint64 `json:"full_reconcile_id"`
}

func newPartialReconcileResponse(partial *accounting.AccountPartialReconcile) AccountPartialReconcileResponse {
	return AccountPartialReconcileResponse{
		ID:              partial.ID,
		DebitLineID:     partial.DebitLineID,
		CreditLineID:    partial.CreditLineID,
		Amount:          partial.Amount,
		FullReconcileID: partial.FullReconcileID,
	}
}

type ReconcileResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ReconcileResponse `json:"data"`
}
type ReconcileResponse struct {
	Partial    *AccountPartialReconcileResponse `json:"partial_reconcile"`
	Reconciled bool                             `json:"reconciled"`
}

// @Summary Reconcile account entry lines
// @Description Reconciles a debit and a credit line on the same account by recording a partial reconciliation for the given amount. Both lines must belong to the organization and share an account, and the amount must be positive and not exceed the remaining balance; lines are flagged fully reconciled when the amount settles their balance.
// @Tags Reconciliations
// @Accept json
// @Produce json
// @Param body body ReconcileRequest true "Reconciliation details"
// @Success 200 {object} ReconcileResponseEnvelope "Lines reconciled successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reconciliations [post]
func (h ReconcileHandler) Reconcile(c fiber.Ctx) error {
	var request ReconcileRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	result, err := h.svc.Reconcile(c, accounting.ReconcileRequest{
		OrganizationID: *organizationID,
		DebitLineID:    request.DebitLineID,
		CreditLineID:   request.CreditLineID,
		Amount:         request.Amount,
	})
	if err != nil {
		return writeReconcileError(c, err)
	}

	partial := newPartialReconcileResponse(result.Partial)
	return httpx.CreateSuccessResponse(c, "Lines reconciled successfully.", ReconcileResponse{
		Partial:    &partial,
		Reconciled: result.Reconciled,
	})
}

func writeReconcileError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrLineNotFound):
		return httpx.CreateNotFoundResponse(c, "Account entry line not found.")
	case errors.Is(err, accounting.ErrLinesDifferentAccount), errors.Is(err, accounting.ErrReconcileAmount), errors.Is(err, accounting.ErrReconcileExceedsBalance), errors.Is(err, accounting.ErrInvalidLine):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Lines cannot be reconciled: they must share an account, the amount must be positive, and not exceed the remaining balance.", nil)
	case errors.Is(err, accounting.ErrLineNotInOrganization):
		return httpx.CreateNotFoundResponse(c, "Account entry line not found.")
	default:
		httpx.RequestLog(c).Error("reconcile failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to reconcile lines.", err)
	}
}
