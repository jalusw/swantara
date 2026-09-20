package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListPaymentBatchesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPaymentBatchesResponse `json:"data"`
}
type ListPaymentBatchesResponse struct {
	Batches []PaymentBatchResponse `json:"batches"`
}

// @Summary List payment batches
// @Description Lists payment batches with pagination, sorting, and filtering, automatically scoping results to the caller's organization
// @Tags Accounting Payment Batches
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. created_at:asc)"
// @Param filter query string false "Filters (repeatable, e.g. state:eq:posted)"
// @Success 200 {object} ListPaymentBatchesResponseEnvelope "Payment batches retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /payment-batches [get]
func (h PaymentBatchHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, paymentBatchQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("payment batch list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve payment batches.", err)
	}

	items := make([]PaymentBatchResponse, 0, len(page.Items))
	for _, batch := range page.Items {
		items = append(items, newPaymentBatchResponse(batch))
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Payment batches retrieved successfully.", ListPaymentBatchesResponse{
		Batches: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type CreatePaymentBatchRequest struct {
	OrganizationID uint64   `json:"organization_id"`
	JournalID      uint64   `json:"journal_id" validate:"required,gt=0"`
	Name           string   `json:"name" validate:"required"`
	Date           string   `json:"date"`
	PaymentIDs     []uint64 `json:"payment_ids" validate:"required,min=1,dive,gt=0"`
}

type CreatePaymentBatchResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreatePaymentBatchResponse `json:"data"`
}
type CreatePaymentBatchResponse struct {
	Batch PaymentBatchResponse `json:"batch"`
}

// @Summary Create payment batch
// @Description Creates a new payment batch from selected payments
// @Tags Accounting Payment Batches
// @Accept json
// @Produce json
// @Param body body CreatePaymentBatchRequest true "Payment batch details"
// @Success 201 {object} CreatePaymentBatchResponseEnvelope "Payment batch created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid request or payment batch state"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /payment-batches [post]
func (h PaymentBatchHandler) Create(c fiber.Ctx) error {
	var request CreatePaymentBatchRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, &request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	var date time.Time
	if request.Date != "" {
		parsed, err := helper.ParseDate(&request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
		}
		date = *parsed
	}

	batch, err := h.svc.Create(c, accounting.CreatePaymentBatchRequest{
		OrganizationID: *organizationID,
		JournalID:      request.JournalID,
		Name:           request.Name,
		Date:           date,
		PaymentIDs:     request.PaymentIDs,
	})
	if err != nil {
		return writePaymentBatchError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Payment batch created successfully.", CreatePaymentBatchResponse{
		Batch: newPaymentBatchResponse(batch),
	})
}

type GetPaymentBatchResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPaymentBatchResponse `json:"data"`
}
type GetPaymentBatchResponse struct {
	Batch PaymentBatchResponse       `json:"batch"`
	Lines []PaymentBatchLineResponse `json:"lines"`
}

// @Summary Get payment batch
// @Description Returns a payment batch and its line items by id
// @Tags Accounting Payment Batches
// @Accept json
// @Produce json
// @Param id path integer true "Payment batch ID"
// @Success 200 {object} GetPaymentBatchResponseEnvelope "Payment batch retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payment batch not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /payment-batches/{id} [get]
func (h PaymentBatchHandler) Get(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid batch id provided.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	batch, err := h.svc.Get(c, id, *organizationID)
	if err != nil {
		return writePaymentBatchError(c, err)
	}

	lines, err := h.svc.ListPayments(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("payment batch lines lookup failed", "batch_id", id, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get batch lines.", err)
	}

	lineItems := make([]PaymentBatchLineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newPaymentBatchLineResponse(line)
	}

	return httpx.CreateSuccessResponse(c, "Payment batch retrieved successfully.", GetPaymentBatchResponse{
		Batch: newPaymentBatchResponse(batch),
		Lines: lineItems,
	})
}
