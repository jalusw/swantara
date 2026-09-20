package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type SplitPaymentTermRequest struct {
	Total string `json:"total" validate:"required"`
	Date  string `json:"date" validate:"required"`
}

type SplitPaymentTermResponseEnvelope struct {
	httpx.EnvelopeBase
	Data SplitPaymentTermResponse `json:"data"`
}

type SplitPaymentTermResponse struct {
	Splits []reference.PaymentSplit `json:"splits"`
}

// @Summary Split payment term
// @Description Computes payment splits for a total decimal amount and a date in YYYY-MM-DD format based on the term's lines. Returns 404 if the term does not exist or a 422 if the total or date is invalid.
// @Tags Payment Terms
// @Accept json
// @Produce json
// @Param id path integer true "Payment term ID"
// @Param body body SplitPaymentTermRequest true "Split criteria"
// @Success 200 {object} SplitPaymentTermResponseEnvelope "Payment splits computed successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payment term not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or invalid lines"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-terms/{id}/splits [post]
func (h PaymentTermHandler) Splits(c fiber.Ctx) error {
	termID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payment term id provided.", nil)
	}

	var request SplitPaymentTermRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	total, err := amount.FromString(request.Total)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Total must be a valid decimal.", nil)
	}

	date, err := helper.ParseDate(&request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	termCheck, err := h.svc.Find(c, termID)
	if err != nil {
		httpx.RequestLog(c).Error("payment term lookup failed", "term_id", termID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to compute payment splits.", err)
	}
	if termCheck == nil || !httpx.OwnsTenant(c, helper.Ptr(termCheck.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Payment term not found.")
	}

	splits, err := h.svc.Splits(c, termID, total, *date)
	if err != nil {
		return writePaymentTermError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Payment splits computed successfully.", SplitPaymentTermResponse{
		Splits: splits,
	})
}
