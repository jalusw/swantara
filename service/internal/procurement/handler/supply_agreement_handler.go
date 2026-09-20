package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type SupplyAgreementHandler struct {
	svc    procurement.SupplyAgreementService
	orders procurement.CreateFromAgreementEngine
}

func NewSupplyAgreementHandler(svc procurement.SupplyAgreementService, orders procurement.CreateFromAgreementEngine) SupplyAgreementHandler {
	return SupplyAgreementHandler{svc: svc, orders: orders}
}

type SupplyAgreementResponse struct {
	ID             uint64     `json:"id"`
	Name           *string    `json:"name"`
	SupplierID     uint64     `json:"supplier_id"`
	CurrencyCode   *string    `json:"currency_code"`
	State          string     `json:"state"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	AmountLimit    float64    `json:"amount_limit"`
	QtyLimit       float64    `json:"qty_limit"`
	ConsumedAmount float64    `json:"consumed_amount"`
	ConsumedQty    float64    `json:"consumed_qty"`
}

func newSupplyAgreementResponse(agreement *procurement.SupplyAgreement) SupplyAgreementResponse {
	return SupplyAgreementResponse{
		ID:             agreement.ID,
		Name:           agreement.Name,
		SupplierID:     agreement.SupplierID,
		CurrencyCode:   agreement.CurrencyCode,
		State:          agreement.State,
		StartDate:      agreement.StartDate,
		EndDate:        agreement.EndDate,
		AmountLimit:    agreement.AmountLimit,
		QtyLimit:       agreement.QtyLimit,
		ConsumedAmount: agreement.ConsumedAmount,
		ConsumedQty:    agreement.ConsumedQty,
	}
}

type SupplyAgreementLineResponse struct {
	ID          uint64     `json:"id"`
	ItemID      *uint64    `json:"item_id"`
	Description *string    `json:"description"`
	Qty         float64    `json:"qty"`
	UnitPrice   float64    `json:"unit_price"`
	UnitID      *uint64    `json:"unit_id"`
	NeededBy    *time.Time `json:"needed_by"`
}

func newSupplyAgreementLineResponse(line *procurement.SupplyAgreementLine) SupplyAgreementLineResponse {
	return SupplyAgreementLineResponse{
		ID:          line.ID,
		ItemID:      line.ItemID,
		Description: line.Description,
		Qty:         line.Qty,
		UnitPrice:   line.UnitPrice,
		UnitID:      line.UnitID,
		NeededBy:    line.NeededBy,
	}
}

var supplyAgreementQueryAllowlist = map[string]struct{}{
	"name":          {},
	"supplier_id":   {},
	"currency_code": {},
	"state":         {},
	"start_date":    {},
	"end_date":      {},
	"created_at":    {},
	"updated_at":    {},
}

type ListSupplyAgreementsResponse struct {
	Agreements []SupplyAgreementResponse `json:"agreements"`
}

type ListSupplyAgreementsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSupplyAgreementsResponse `json:"data"`
}

// @Summary List purchase agreements
// @Description Lists purchase agreements across the caller's organization with pagination, sorting, and filtering on attributes such as supplier, state, and date range.
// @Tags Purchase Agreements
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListSupplyAgreementsResponseEnvelope "Purchase agreements retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-agreements [get]
func (h SupplyAgreementHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, supplyAgreementQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	svc := h.svc
	page, err := svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("purchase agreement list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve purchase agreements.", err)
	}
	items := make([]SupplyAgreementResponse, len(page.Items))
	for i, agreement := range page.Items {
		items[i] = newSupplyAgreementResponse(agreement)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Purchase agreements retrieved successfully.", ListSupplyAgreementsResponse{
		Agreements: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetSupplyAgreementResponse struct {
	Agreement SupplyAgreementResponse       `json:"agreement"`
	Lines     []SupplyAgreementLineResponse `json:"lines"`
}

type GetSupplyAgreementResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetSupplyAgreementResponse `json:"data"`
}

// @Summary Get purchase agreement
// @Description Gets a single purchase agreement by id, including its lines.
// @Tags Purchase Agreements
// @Accept json
// @Produce json
// @Param id path integer true "Purchase agreement ID"
// @Success 200 {object} GetSupplyAgreementResponseEnvelope "Purchase agreement retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase agreement not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-agreements/{id} [get]
func (h SupplyAgreementHandler) Get(c fiber.Ctx) error {
	agreementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase agreement id provided.", nil)
	}
	agreement, err := h.svc.Find(c, agreementID)
	if err != nil {
		httpx.RequestLog(c).Error("purchase agreement lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get purchase agreement.", err)
	}
	if agreement == nil || !httpx.OwnsTenant(c, agreement.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Purchase agreement not found.")
	}
	lines, err := h.svc.ListLines(c, agreementID)
	if err != nil {
		httpx.RequestLog(c).Error("purchase agreement lines lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get purchase agreement lines.", err)
	}
	lineItems := make([]SupplyAgreementLineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newSupplyAgreementLineResponse(line)
	}
	return httpx.CreateSuccessResponse(c, "Purchase agreement retrieved successfully.", GetSupplyAgreementResponse{
		Agreement: newSupplyAgreementResponse(agreement),
		Lines:     lineItems,
	})
}

type CreateSupplyAgreementRequest struct {
	SupplierID   uint64                             `json:"supplier_id" validate:"required"`
	CurrencyCode *string                            `json:"currency_code"`
	StartDate    *string                            `json:"start_date"`
	EndDate      *string                            `json:"end_date"`
	AmountLimit  float64                            `json:"amount_limit"`
	QtyLimit     float64                            `json:"qty_limit"`
	Lines        []CreateSupplyAgreementLineRequest `json:"lines" validate:"required,min=1"`
}

type CreateSupplyAgreementLineRequest struct {
	ItemID      *uint64 `json:"item_id"`
	Description *string `json:"description"`
	Qty         float64 `json:"qty" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gte=0"`
	UnitID      *uint64 `json:"unit_id"`
	NeededBy    *string `json:"needed_by"`
}

type CreateSupplyAgreementResponse struct {
	Agreement SupplyAgreementResponse `json:"agreement"`
}

type CreateSupplyAgreementResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateSupplyAgreementResponse `json:"data"`
}

// @Summary Create purchase agreement
// @Description Creates a new purchase agreement with lines for a supplier. The agreement starts in draft state.
// @Tags Purchase Agreements
// @Accept json
// @Produce json
// @Param body body CreateSupplyAgreementRequest true "Purchase agreement details"
// @Success 201 {object} CreateSupplyAgreementResponseEnvelope "Purchase agreement created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-agreements [post]
func (h SupplyAgreementHandler) Create(c fiber.Ctx) error {
	var request CreateSupplyAgreementRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	startDate, err := helper.ParseDate(request.StartDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid start date provided.", nil)
	}
	endDate, err := helper.ParseDate(request.EndDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid end date provided.", nil)
	}
	agreement := &procurement.SupplyAgreement{
		OrganizationID: httpx.TenantOrganizationID(c, nil),
		SupplierID:     request.SupplierID,
		CurrencyCode:   request.CurrencyCode,
		StartDate:      startDate,
		EndDate:        endDate,
		AmountLimit:    request.AmountLimit,
		QtyLimit:       request.QtyLimit,
	}
	lines := make([]*procurement.SupplyAgreementLine, 0, len(request.Lines))
	for _, rl := range request.Lines {
		neededBy, err := helper.ParseDate(rl.NeededBy)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid line needed by date provided.", nil)
		}
		lines = append(lines, &procurement.SupplyAgreementLine{
			ItemID:      rl.ItemID,
			Description: rl.Description,
			Qty:         rl.Qty,
			UnitPrice:   rl.UnitPrice,
			UnitID:      rl.UnitID,
			NeededBy:    neededBy,
		})
	}
	created, err := h.svc.Create(c, agreement, lines)
	if err != nil {
		return writeSupplyAgreementError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Purchase agreement created successfully.", CreateSupplyAgreementResponse{
		Agreement: newSupplyAgreementResponse(created),
	})
}

type SupplyAgreementStateResponse struct {
	Agreement SupplyAgreementResponse `json:"agreement"`
}

type SupplyAgreementStateResponseEnvelope struct {
	httpx.EnvelopeBase
	Data SupplyAgreementStateResponse `json:"data"`
}

// @Summary Activate purchase agreement
// @Description Activates a purchase agreement, transitioning it from draft to active state.
// @Tags Purchase Agreements
// @Accept json
// @Produce json
// @Param id path integer true "Purchase agreement ID"
// @Success 200 {object} SupplyAgreementStateResponseEnvelope "Purchase agreement activated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase agreement not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid state transition"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-agreements/{id}/activate [post]
func (h SupplyAgreementHandler) Activate(c fiber.Ctx) error {
	agreementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase agreement id provided.", nil)
	}
	agreement, err := h.svc.Activate(c, agreementID)
	if err != nil {
		return writeSupplyAgreementError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase agreement activated successfully.", SupplyAgreementStateResponse{
		Agreement: newSupplyAgreementResponse(agreement),
	})
}

// @Summary Cancel purchase agreement
// @Description Cancels a purchase agreement, transitioning it to cancelled state.
// @Tags Purchase Agreements
// @Accept json
// @Produce json
// @Param id path integer true "Purchase agreement ID"
// @Success 200 {object} SupplyAgreementStateResponseEnvelope "Purchase agreement cancelled successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase agreement not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid state transition"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-agreements/{id}/cancel [post]
func (h SupplyAgreementHandler) Cancel(c fiber.Ctx) error {
	agreementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase agreement id provided.", nil)
	}
	agreement, err := h.svc.Cancel(c, agreementID)
	if err != nil {
		return writeSupplyAgreementError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase agreement cancelled successfully.", SupplyAgreementStateResponse{
		Agreement: newSupplyAgreementResponse(agreement),
	})
}

type CreateFromAgreementRequest struct {
	WarehouseID  *uint64            `json:"warehouse_id"`
	SupplierID   *uint64            `json:"supplier_id"`
	QtyOverrides map[uint64]float64 `json:"qty_overrides"`
}

type CreateFromAgreementResponse struct {
	Order PurchaseOrderResponse `json:"order"`
}

type CreateFromAgreementResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateFromAgreementResponse `json:"data"`
}

// @Summary Create purchase order from agreement
// @Description Creates a purchase order from an active purchase agreement, consuming agreed quantities.
// @Tags Purchase Agreements
// @Accept json
// @Produce json
// @Param id path integer true "Purchase agreement ID"
// @Param body body CreateFromAgreementRequest true "Override options"
// @Success 201 {object} CreateFromAgreementResponseEnvelope "Purchase order created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase agreement not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-agreements/{id}/create-order [post]
func (h SupplyAgreementHandler) CreateFromAgreement(c fiber.Ctx) error {
	agreementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase agreement id provided.", nil)
	}
	var request CreateFromAgreementRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	overrides := &procurement.CreateFromAgreementOverrides{
		SupplierID:   request.SupplierID,
		WarehouseID:  request.WarehouseID,
		QtyOverrides: request.QtyOverrides,
	}
	order, err := h.orders.CreateFromAgreement(c, agreementID, overrides)
	if err != nil {
		return writeSupplyAgreementError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Purchase order created successfully.", CreateFromAgreementResponse{
		Order: newPurchaseOrderResponse(order),
	})
}

func writeSupplyAgreementError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, procurement.ErrAgreementNotFound):
		return httpx.CreateNotFoundResponse(c, "Purchase agreement not found.")
	case errors.Is(err, procurement.ErrSupplyAgreementState), errors.Is(err, procurement.ErrAgreementNotActive):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid state transition.", nil)
	case errors.Is(err, procurement.ErrAgreementNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Agreement must have at least one line.", nil)
	case errors.Is(err, procurement.ErrAgreementLineQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Line quantity must be positive.", nil)
	case errors.Is(err, procurement.ErrAgreementVendor):
		return httpx.CreateNotFoundResponse(c, "Supplier not found.")
	case errors.Is(err, procurement.ErrAgreementExpired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Agreement has expired.", nil)
	case errors.Is(err, procurement.ErrAgreementQtyExceeded):
		return httpx.CreateConflictResponse(c, "Agreement quantity limit exceeded.", err)
	case errors.Is(err, procurement.ErrAgreementAmountExceeded):
		return httpx.CreateConflictResponse(c, "Agreement amount limit exceeded.", err)
	default:
		httpx.RequestLog(c).Error("purchase agreement write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process purchase agreement.", err)
	}
}
