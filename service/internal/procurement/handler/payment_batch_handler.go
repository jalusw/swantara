package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type PaymentBatchHandler struct {
	svc procurement.PaymentBatchService
}

func NewPaymentBatchHandler(svc procurement.PaymentBatchService) PaymentBatchHandler {
	return PaymentBatchHandler{svc: svc}
}

type PaymentBatchResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	JournalID      uint64     `json:"journal_id"`
	ContactID      uint64     `json:"contact_id"`
	Date           *time.Time `json:"date"`
	State          string     `json:"state"`
	TotalAmount    float64    `json:"total_amount"`
	PaymentCount   int        `json:"payment_count"`
}

func newPaymentBatchResponse(batch *procurement.PaymentBatch) PaymentBatchResponse {
	return PaymentBatchResponse{
		ID:             batch.ID,
		OrganizationID: batch.OrganizationID,
		JournalID:      batch.JournalID,
		ContactID:      batch.ContactID,
		Date:           batch.Date,
		State:          batch.State,
		TotalAmount:    batch.TotalAmount,
		PaymentCount:   batch.PaymentCount,
	}
}

type PaymentBatchLineResponse struct {
	ID      uint64  `json:"id"`
	OrderID uint64  `json:"order_id"`
	Amount  float64 `json:"amount"`
}

func newPaymentBatchLineResponse(line *procurement.PaymentBatchLine) PaymentBatchLineResponse {
	return PaymentBatchLineResponse{
		ID:      line.ID,
		OrderID: line.OrderID,
		Amount:  line.Amount,
	}
}

var paymentBatchQueryAllowlist = map[string]struct{}{
	"journal_id":    {},
	"contact_id":    {},
	"date":          {},
	"state":         {},
	"total_amount":  {},
	"payment_count": {},
	"created_at":    {},
	"updated_at":    {},
}

type ListPaymentBatchesResponse struct {
	Batches []PaymentBatchResponse `json:"batches"`
}

type ListPaymentBatchesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPaymentBatchesResponse `json:"data"`
}

// @Summary List payment batches
// @Description Lists payment batches across the caller's organization with pagination, sorting, and filtering.
// @Tags Payment Batches
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListPaymentBatchesResponseEnvelope "Payment batches retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-batches [get]
func (h PaymentBatchHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, paymentBatchQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListPaymentBatches(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("payment batch list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve payment batches.", err)
	}
	items := make([]PaymentBatchResponse, len(page.Items))
	for i, batch := range page.Items {
		items[i] = newPaymentBatchResponse(batch)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Payment batches retrieved successfully.", ListPaymentBatchesResponse{
		Batches: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetPaymentBatchResponse struct {
	Batch PaymentBatchResponse       `json:"batch"`
	Lines []PaymentBatchLineResponse `json:"lines"`
}

type GetPaymentBatchResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPaymentBatchResponse `json:"data"`
}

// @Summary Get payment batch
// @Description Gets a single payment batch by id, including its lines.
// @Tags Payment Batches
// @Accept json
// @Produce json
// @Param id path integer true "Payment batch ID"
// @Success 200 {object} GetPaymentBatchResponseEnvelope "Payment batch retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payment batch not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-batches/{id} [get]
func (h PaymentBatchHandler) Get(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payment batch id provided.", nil)
	}
	batch, lines, err := h.svc.GetPaymentBatch(c, id)
	if err != nil {
		return writePaymentBatchError(c, err)
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

type CreatePaymentBatchRequest struct {
	SupplierID uint64                          `json:"supplier_id" validate:"required"`
	JournalID  uint64                          `json:"journal_id" validate:"required"`
	Date       *string                         `json:"date"`
	Orders     []CreatePaymentBatchRequestLine `json:"orders" validate:"required,min=1"`
}

type CreatePaymentBatchRequestLine struct {
	OrderID uint64  `json:"order_id" validate:"required"`
	Amount  float64 `json:"amount" validate:"required,gt=0"`
}

type CreatePaymentBatchResponse struct {
	Batch PaymentBatchResponse       `json:"batch"`
	Lines []PaymentBatchLineResponse `json:"lines"`
}

type CreatePaymentBatchResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreatePaymentBatchResponse `json:"data"`
}

// @Summary Create payment batch
// @Description Creates a new payment batch for multiple purchase orders from the same supplier.
// @Tags Payment Batches
// @Accept json
// @Produce json
// @Param body body CreatePaymentBatchRequest true "Payment batch details"
// @Success 201 {object} CreatePaymentBatchResponseEnvelope "Payment batch created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-batches [post]
func (h PaymentBatchHandler) Create(c fiber.Ctx) error {
	var request CreatePaymentBatchRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	batchDate := helper.Deref(date, time.Now().UTC())
	orgID := httpx.TenantOrganizationID(c, nil)
	if orgID == nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}
	orders := make([]procurement.PaymentBatchRequestLine, len(request.Orders))
	for i, o := range request.Orders {
		orders[i] = procurement.PaymentBatchRequestLine{
			OrderID: o.OrderID,
			Amount:  o.Amount,
		}
	}
	batch, lines, err := h.svc.CreatePaymentBatch(c, *orgID, request.SupplierID, procurement.PaymentBatchRequest{
		JournalID: request.JournalID,
		Date:      batchDate,
		Orders:    orders,
	})
	if err != nil {
		return writePaymentBatchError(c, err)
	}
	lineItems := make([]PaymentBatchLineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newPaymentBatchLineResponse(line)
	}
	return httpx.CreateCreatedResponse(c, "Payment batch created successfully.", CreatePaymentBatchResponse{
		Batch: newPaymentBatchResponse(batch),
		Lines: lineItems,
	})
}

type ConfirmPaymentBatchResponse struct {
	Batch PaymentBatchResponse `json:"batch"`
}

type ConfirmPaymentBatchResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ConfirmPaymentBatchResponse `json:"data"`
}

// @Summary Confirm payment batch
// @Description Confirms a payment batch, posting payments for all included orders.
// @Tags Payment Batches
// @Accept json
// @Produce json
// @Param id path integer true "Payment batch ID"
// @Success 200 {object} ConfirmPaymentBatchResponseEnvelope "Payment batch confirmed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payment batch not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid state transition"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-batches/{id}/confirm [post]
func (h PaymentBatchHandler) Confirm(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payment batch id provided.", nil)
	}
	batch, err := h.svc.ConfirmBatch(c, id)
	if err != nil {
		return writePaymentBatchError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Payment batch confirmed successfully.", ConfirmPaymentBatchResponse{
		Batch: newPaymentBatchResponse(batch),
	})
}
