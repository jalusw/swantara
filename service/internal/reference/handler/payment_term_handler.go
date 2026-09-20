package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type PaymentTermHandler struct {
	svc reference.PaymentTermService
}

func NewPaymentTermHandler(svc reference.PaymentTermService) PaymentTermHandler {
	return PaymentTermHandler{svc: svc}
}

func (h PaymentTermHandler) paymentTermResponse(c fiber.Ctx, term *reference.PaymentTerm) (PaymentTermResponse, error) {
	lines, err := h.svc.ListLines(c, term.ID)
	if err != nil {
		return PaymentTermResponse{}, err
	}
	lineItems := make([]PaymentTermLineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newPaymentTermLineResponse(line)
	}
	return PaymentTermResponse{
		ID:             term.ID,
		OrganizationID: term.OrganizationID,
		Name:           term.Name,
		Note:           term.Note,
		Code:           term.Code,
		IsActive:       term.IsActive,
		TemplateKey:    term.TemplateKey,
		Lines:          lineItems,
		CreatedAt:      term.CreatedAt,
		UpdatedAt:      term.UpdatedAt,
	}, nil
}

var paymentTermQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"code":            {},
	"is_active":       {},
	"created_at":      {},
	"updated_at":      {},
}

// @Summary List payment terms
// @Description Lists payment terms with pagination, sorting, and filtering, each including its schedule lines; the response can be exported as JSON, XML, or CSV.
// @Tags Payment Terms
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. name:like:Net)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListPaymentTermsResponseEnvelope "Payment terms retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-terms [get]
func (h PaymentTermHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, paymentTermQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("payment term list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve payment terms.", err)
	}

	items := make([]PaymentTermResponse, len(page.Items))
	for i, term := range page.Items {
		response, err := h.paymentTermResponse(c, term)
		if err != nil {
			httpx.RequestLog(c).Error("payment term lines lookup failed", "term_id", term.ID, "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve payment terms.", err)
		}
		items[i] = response
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "payment-terms.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Payment terms retrieved successfully.", ListPaymentTermsResponse{
		PaymentTerms: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get payment term
// @Description Returns a single payment term by id, including its schedule lines with value type, days, and discount details; a 404 is returned if the term does not exist.
// @Tags Payment Terms
// @Accept json
// @Produce json
// @Param id path integer true "Payment term ID"
// @Success 200 {object} GetPaymentTermResponseEnvelope "Payment term retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payment term not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-terms/{id} [get]
func (h PaymentTermHandler) Get(c fiber.Ctx) error {
	termID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payment term id provided.", nil)
	}

	term, err := h.svc.Find(c, termID)
	if err != nil {
		httpx.RequestLog(c).Error("payment term lookup failed", "term_id", termID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get payment term.", err)
	}
	if term == nil || !httpx.OwnsTenant(c, helper.Ptr(term.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Payment term not found.")
	}

	response, err := h.paymentTermResponse(c, term)
	if err != nil {
		httpx.RequestLog(c).Error("payment term lines lookup failed", "term_id", termID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get payment term.", err)
	}

	return httpx.CreateSuccessResponse(c, "Payment term retrieved successfully.", GetPaymentTermResponse{
		PaymentTerm: response,
	})
}

// @Summary Create payment term
// @Description Creates a payment term with a required name, optional note, and at least one schedule line. Line value types must be percent, fixed, or balance; percent lines must sum to 100 unless a single balance line absorbs the remainder.
// @Tags Payment Terms
// @Accept json
// @Produce json
// @Param body body CreatePaymentTermRequest true "Payment term details"
// @Success 201 {object} CreatePaymentTermResponseEnvelope "Payment term created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or invalid lines"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-terms [post]
func (h PaymentTermHandler) Create(c fiber.Ctx) error {
	var request CreatePaymentTermRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingOrganizationInContext)
	}

	lines := make([]*reference.PaymentTermLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &reference.PaymentTermLine{
			Sequence:     line.Sequence,
			ValueType:    line.ValueType,
			Value:        line.Value,
			DaysAfter:    line.DaysAfter,
			DayOfMonth:   line.DayOfMonth,
			DiscountPct:  line.DiscountPct,
			DiscountDays: line.DiscountDays,
		}
	}

	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}
	term, err := h.svc.Create(c, &reference.PaymentTerm{
		OrganizationID: organizationID,
		Name:           request.Name,
		Note:           request.Note,
		Code:           request.Code,
		IsActive:       isActive,
		TemplateKey:    request.TemplateKey,
	}, lines)
	if err != nil {
		return writePaymentTermError(c, err)
	}

	response, err := h.paymentTermResponse(c, term)
	if err != nil {
		httpx.RequestLog(c).Error("payment term lines lookup failed", "term_id", term.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create payment term.", err)
	}

	return httpx.CreateCreatedResponse(c, "Payment term created successfully.", CreatePaymentTermResponse{
		PaymentTerm: response,
	})
}

// @Summary Update payment term
// @Description Updates a payment term's name and note, replacing its schedule lines. Percent lines must sum to 100 unless a single balance line absorbs the remainder; a 404 is returned if the term does not exist.
// @Tags Payment Terms
// @Accept json
// @Produce json
// @Param id path integer true "Payment term ID"
// @Param body body UpdatePaymentTermRequest true "Payment term details"
// @Success 200 {object} UpdatePaymentTermResponseEnvelope "Payment term updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payment term not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or invalid lines"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-terms/{id} [put]
func (h PaymentTermHandler) Update(c fiber.Ctx) error {
	termID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payment term id provided.", nil)
	}

	var request UpdatePaymentTermRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	existing, err := h.svc.Find(c, termID)
	if err != nil {
		httpx.RequestLog(c).Error("payment term lookup failed", "term_id", termID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update payment term.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, helper.Ptr(existing.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Payment term not found.")
	}

	lines := make([]*reference.PaymentTermLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &reference.PaymentTermLine{
			Sequence:     line.Sequence,
			ValueType:    line.ValueType,
			Value:        line.Value,
			DaysAfter:    line.DaysAfter,
			DayOfMonth:   line.DayOfMonth,
			DiscountPct:  line.DiscountPct,
			DiscountDays: line.DiscountDays,
		}
	}

	isActive := existing.IsActive
	if request.IsActive != nil {
		isActive = *request.IsActive
	}
	updated, err := h.svc.Update(c, termID, &reference.PaymentTerm{
		Name:           request.Name,
		Note:           request.Note,
		Code:           request.Code,
		IsActive:       isActive,
		TemplateKey:    request.TemplateKey,
		OrganizationID: existing.OrganizationID,
	}, lines)
	if err != nil {
		return writePaymentTermError(c, err)
	}

	response, err := h.paymentTermResponse(c, updated)
	if err != nil {
		httpx.RequestLog(c).Error("payment term lines lookup failed", "term_id", updated.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update payment term.", err)
	}

	return httpx.CreateSuccessResponse(c, "Payment term updated successfully.", UpdatePaymentTermResponse{
		PaymentTerm: response,
	})
}

// @Summary Delete payment term
// @Description Deletes a payment term and its schedule lines by id; returns 404 if the term does not exist and 204 No Content on success.
// @Tags Payment Terms
// @Accept json
// @Produce json
// @Param id path integer true "Payment term ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payment term not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payment-terms/{id} [delete]
func (h PaymentTermHandler) Delete(c fiber.Ctx) error {
	termID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payment term id provided.", nil)
	}

	term, err := h.svc.Find(c, termID)
	if err != nil {
		httpx.RequestLog(c).Error("payment term lookup failed", "term_id", termID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete payment term.", err)
	}
	if term == nil || !httpx.OwnsTenant(c, helper.Ptr(term.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Payment term not found.")
	}

	if err := h.svc.Delete(c, termID); err != nil {
		return writePaymentTermError(c, err)
	}

	return httpx.CreateNoContentResponse(c)
}

func writePaymentTermError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reference.ErrTermNotFound):
		return httpx.CreateNotFoundResponse(c, "Payment term not found.")
	case errors.Is(err, reference.ErrInvalidTermLines), errors.Is(err, reference.ErrMultipleBalanceLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payment term lines are invalid: percent must sum to 100 with at most one balance line.", nil)
	default:
		httpx.RequestLog(c).Error("payment term write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save payment term.", err)
	}
}
