package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type SupplierQuoteRequestHandler struct {
	svc procurement.SupplierQuoteRequestService
}

func NewSupplierQuoteRequestHandler(
	svc procurement.SupplierQuoteRequestService,
) SupplierQuoteRequestHandler {
	return SupplierQuoteRequestHandler{svc: svc}
}

type SupplierQuoteRequestResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	RequesterID    uint64     `json:"requester_id"`
	SupplierID     *uint64    `json:"supplier_id"`
	CurrencyCode   *string    `json:"currency_code"`
	State          string     `json:"state"`
	OrderDate      *time.Time `json:"order_date"`
	QuoteDeadline  *time.Time `json:"quote_deadline"`
	Notes          *string    `json:"notes"`
}

func newSupplierQuoteRequestResponse(quoteRequest *procurement.SupplierQuoteRequest) SupplierQuoteRequestResponse {
	return SupplierQuoteRequestResponse{
		ID:             quoteRequest.ID,
		OrganizationID: quoteRequest.OrganizationID,
		Name:           quoteRequest.Name,
		RequesterID:    quoteRequest.RequesterID,
		SupplierID:     quoteRequest.SupplierID,
		CurrencyCode:   quoteRequest.CurrencyCode,
		State:          quoteRequest.State,
		OrderDate:      quoteRequest.OrderDate,
		QuoteDeadline:  quoteRequest.QuoteDeadline,
		Notes:          quoteRequest.Notes,
	}
}

var purchaseRFQQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"requester_id":    {},
	"supplier_id":     {},
	"state":           {},
	"order_date":      {},
	"created_at":      {},
	"updated_at":      {},
}

type ListSupplierQuoteRequestsResponse struct {
	QuoteRequests []SupplierQuoteRequestResponse `json:"quote_requests"`
}

type ListSupplierQuoteRequestsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSupplierQuoteRequestsResponse `json:"data"`
}

// @Summary List purchase QuoteRequests
// @Description Lists purchase requests for quotation across the caller's organization with pagination, sorting, and filtering on attributes such as requester, supplier, and state. The result set is always scoped to the caller's organization even when no organization_id is supplied. Returns the matching QuoteRequests together with pagination metadata.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListSupplierQuoteRequestsResponseEnvelope "Purchase QuoteRequests retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests [get]
func (h SupplierQuoteRequestHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, purchaseRFQQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("purchase quoteRequest list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve purchase QuoteRequests.", err)
	}
	items := make([]SupplierQuoteRequestResponse, len(page.Items))
	for i, quoteRequest := range page.Items {
		items[i] = newSupplierQuoteRequestResponse(quoteRequest)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Purchase QuoteRequests retrieved successfully.", ListSupplierQuoteRequestsResponse{
		QuoteRequests: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetSupplierQuoteRequestResponse struct {
	QuoteRequest SupplierQuoteRequestResponse `json:"quoteRequest"`
}

type GetSupplierQuoteRequestResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetSupplierQuoteRequestResponse `json:"data"`
}

// @Summary Get purchase QuoteRequest
// @Description Gets a single purchase request for quotation by id, scoped to the caller's organization. A 404 is returned when the QuoteRequest does not exist or belongs to another tenant.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param id path integer true "Purchase QuoteRequest ID"
// @Success 200 {object} GetSupplierQuoteRequestResponseEnvelope "Purchase QuoteRequest retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase QuoteRequest not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests/{id} [get]
func (h SupplierQuoteRequestHandler) Get(c fiber.Ctx) error {
	rfqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase QuoteRequest id provided.", nil)
	}
	quoteRequest, err := h.svc.Find(c, rfqID)
	if err != nil {
		httpx.RequestLog(c).Error("purchase quoteRequest lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get purchase QuoteRequest.", err)
	}
	if quoteRequest == nil || !httpx.OwnsTenant(c, quoteRequest.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Purchase QuoteRequest not found.")
	}
	return httpx.CreateSuccessResponse(c, "Purchase QuoteRequest retrieved successfully.", GetSupplierQuoteRequestResponse{
		QuoteRequest: newSupplierQuoteRequestResponse(quoteRequest),
	})
}

type SupplierQuoteRequestLineRequest struct {
	ItemID      *uint64    `json:"item_id"`
	Description *string    `json:"description"`
	Qty         float64    `json:"qty" validate:"required,gt=0"`
	UnitID      *uint64    `json:"unit_id"`
	NeededBy    *time.Time `json:"needed_by"`
}

type CreateSupplierQuoteRequestRequest struct {
	OrganizationID *uint64                           `json:"organization_id"`
	RequesterID    uint64                            `json:"requester_id" validate:"required,gt=0"`
	SupplierID     *uint64                           `json:"supplier_id"`
	CurrencyCode   *string                           `json:"currency_code"`
	OrderDate      *time.Time                        `json:"order_date"`
	QuoteDeadline  *time.Time                        `json:"quote_deadline"`
	Notes          *string                           `json:"notes"`
	Lines          []SupplierQuoteRequestLineRequest `json:"lines" validate:"required,min=1"`
}

type CreateSupplierQuoteRequestResponse struct {
	QuoteRequest SupplierQuoteRequestResponse `json:"quoteRequest"`
}

type CreateSupplierQuoteRequestResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateSupplierQuoteRequestResponse `json:"data"`
}

// @Summary Create purchase QuoteRequest
// @Description Creates a new draft purchase QuoteRequest with its request lines, validating that the requester exists and that every line quantity is positive. A sequence number is generated for the QuoteRequest, the currency defaults to IDR, and the QuoteRequest is saved in draft state ready to be sent to vendors.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param body body CreateSupplierQuoteRequestRequest true "Purchase QuoteRequest details"
// @Success 201 {object} CreateSupplierQuoteRequestResponseEnvelope "Purchase QuoteRequest created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests [post]
func (h SupplierQuoteRequestHandler) Create(c fiber.Ctx) error {
	var request CreateSupplierQuoteRequestRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	lines := make([]*procurement.SupplierQuoteRequestLine, 0, len(request.Lines))
	for _, line := range request.Lines {
		lines = append(lines, &procurement.SupplierQuoteRequestLine{
			ItemID:      line.ItemID,
			Description: line.Description,
			Qty:         line.Qty,
			UnitID:      line.UnitID,
			NeededBy:    line.NeededBy,
		})
	}
	quoteRequest, err := h.svc.Create(c, &procurement.SupplierQuoteRequest{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		RequesterID:    request.RequesterID,
		SupplierID:     request.SupplierID,
		CurrencyCode:   request.CurrencyCode,
		OrderDate:      request.OrderDate,
		QuoteDeadline:  request.QuoteDeadline,
		Notes:          request.Notes,
	}, lines)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Purchase QuoteRequest created successfully.", CreateSupplierQuoteRequestResponse{
		QuoteRequest: newSupplierQuoteRequestResponse(quoteRequest),
	})
}

type CreateSupplierQuoteRequestFromRequisitionRequest struct {
	RequestID uint64 `json:"request_id" validate:"required,gt=0"`
}

// @Summary Create purchase QuoteRequest from requisition
// @Description Creates a new draft purchase QuoteRequest from the lines of an approved purchase requisition, carrying over the requested products, quantities, and needed-by dates. Only requisitions in the approved state can be converted; otherwise the request is rejected with a conflict.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param body body CreateSupplierQuoteRequestFromRequisitionRequest true "Requisition reference"
// @Success 201 {object} CreateSupplierQuoteRequestResponseEnvelope "Purchase QuoteRequest created from requisition successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests/from-request [post]
func (h SupplierQuoteRequestHandler) CreateFromRequest(c fiber.Ctx) error {
	var request CreateSupplierQuoteRequestFromRequisitionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	quoteRequest, err := h.svc.CreateFromRequest(c, request.RequestID, &procurement.SupplierQuoteRequest{})
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Purchase QuoteRequest created from requisition successfully.", CreateSupplierQuoteRequestResponse{
		QuoteRequest: newSupplierQuoteRequestResponse(quoteRequest),
	})
}

// @Summary Send purchase QuoteRequest
// @Description Sends a draft purchase QuoteRequest, transitioning it to the sent state and opening it for supplier quotations. The state machine rejects the transition from any state other than draft.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param id path integer true "Purchase QuoteRequest ID"
// @Success 200 {object} GetSupplierQuoteRequestResponseEnvelope "Purchase QuoteRequest sent successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase QuoteRequest not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests/{id}/send [post]
func (h SupplierQuoteRequestHandler) Send(c fiber.Ctx) error {
	rfqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase QuoteRequest id provided.", nil)
	}
	quoteRequest, err := h.svc.Send(c, rfqID)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase QuoteRequest sent successfully.", GetSupplierQuoteRequestResponse{
		QuoteRequest: newSupplierQuoteRequestResponse(quoteRequest),
	})
}

// @Summary Cancel purchase QuoteRequest
// @Description Cancels a draft or sent purchase QuoteRequest, moving it to cancelled before it is converted into a purchase order. The state machine rejects cancellation from any state other than draft or sent.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param id path integer true "Purchase QuoteRequest ID"
// @Success 200 {object} GetSupplierQuoteRequestResponseEnvelope "Purchase QuoteRequest cancelled successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase QuoteRequest not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests/{id}/cancel [post]
func (h SupplierQuoteRequestHandler) Cancel(c fiber.Ctx) error {
	rfqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase QuoteRequest id provided.", nil)
	}
	quoteRequest, err := h.svc.Cancel(c, rfqID)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase QuoteRequest cancelled successfully.", GetSupplierQuoteRequestResponse{
		QuoteRequest: newSupplierQuoteRequestResponse(quoteRequest),
	})
}
