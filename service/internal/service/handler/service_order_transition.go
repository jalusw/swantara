package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/service"
)

type AddServiceOrderLineRequest struct {
	Type              string  `json:"type" validate:"required,oneof=part labor expense"`
	ItemID            *uint64 `json:"item_id"`
	Description       string  `json:"description"`
	Qty               float64 `json:"qty" validate:"required,gt=0"`
	UnitID            *uint64 `json:"unit_id"`
	UnitCost          float64 `json:"unit_cost"`
	UnitPrice         float64 `json:"unit_price"`
	Billable          bool    `json:"billable"`
	CoveredByWarranty bool    `json:"covered_by_warranty"`
}

type CreateServiceOrderLineResponse struct {
	ServiceOrderLine ServiceOrderLineResponse `json:"service_order_line"`
}

type CreateServiceOrderLineResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateServiceOrderLineResponse `json:"data"`
}

// @Summary Add service order line
// @Description Adds a part/labor/expense line to an open service order.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service order ID"
// @Param body body AddServiceOrderLineRequest true "Line details"
// @Success 201 {object} CreateServiceOrderLineResponseEnvelope "Service order line added successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders/{id}/lines [post]
func (h ServiceHandler) AddOrderLine(c fiber.Ctx) error {
	var request AddServiceOrderLineRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	order, err := h.svc.FindOrder(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service order get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service order.", err)
	}
	if order == nil || !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service order not found.")
	}
	created, err := h.svc.AddLine(c, service.AddLineRequest{
		OrderID: order.ID,
		Line: service.LineRequest{
			Type: request.Type, ItemID: request.ItemID, Description: request.Description,
			Qty: request.Qty, UnitID: request.UnitID, UnitCost: request.UnitCost, UnitPrice: request.UnitPrice,
			Billable: request.Billable, CoveredByWarranty: request.CoveredByWarranty,
		},
	})
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Service order line added successfully.", CreateServiceOrderLineResponse{
		ServiceOrderLine: newServiceOrderLineResponse(created),
	})
}

type ScheduleServiceOrderRequest struct {
	ScheduledDate string  `json:"scheduled_date" validate:"required"`
	TechnicianID  *uint64 `json:"technician_id"`
}

// @Summary Schedule service order
// @Description Moves a service order from new to scheduled with a technician and scheduled date.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service order ID"
// @Param body body ScheduleServiceOrderRequest true "Schedule details"
// @Success 200 {object} GetServiceOrderResponseEnvelope "Service order scheduled successfully."
// @Failure 422 {object} httpx.ErrorResponse "Invalid state or date"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders/{id}/schedule [post]
func (h ServiceHandler) ScheduleOrder(c fiber.Ctx) error {
	var request ScheduleServiceOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := time.Parse(time.RFC3339, request.ScheduledDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid scheduled_date.", nil)
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	order, err := h.svc.FindOrder(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service order get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service order.", err)
	}
	if order == nil || !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service order not found.")
	}
	updated, err := h.svc.Schedule(c, service.ScheduleRequest{OrderID: order.ID, ScheduledDate: date, TechnicianID: request.TechnicianID})
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Service order scheduled successfully.", GetServiceOrderResponse{
		ServiceOrder: newServiceOrderResponse(updated),
	})
}

// @Summary Start service order
// @Description Moves a scheduled service order to in_progress.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service order ID"
// @Success 200 {object} GetServiceOrderResponseEnvelope "Service order started successfully."
// @Failure 422 {object} httpx.ErrorResponse "Invalid state"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders/{id}/start [post]
func (h ServiceHandler) StartOrder(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	order, err := h.svc.FindOrder(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service order get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service order.", err)
	}
	if order == nil || !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service order not found.")
	}
	updated, err := h.svc.Start(c, order.ID)
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Service order started successfully.", GetServiceOrderResponse{
		ServiceOrder: newServiceOrderResponse(updated),
	})
}

type CompleteServiceOrderRequest struct {
	JournalID               uint64 `json:"journal_id" validate:"required,gt=0"`
	Date                    string `json:"date" validate:"required"`
	COGSAccountID           uint64 `json:"cogs_account_id" validate:"required,gt=0"`
	StockValuationAccountID uint64 `json:"stock_cost_account_id" validate:"required,gt=0"`
	Resolution              string `json:"resolution"`
}

// @Summary Complete service order
// @Description Completes an in-progress service order, posting COGS for part lines (Dr COGS / Cr Stock Valuation) and recording the resolution.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service order ID"
// @Param body body CompleteServiceOrderRequest true "Completion details"
// @Success 200 {object} GetServiceOrderResponseEnvelope "Service order completed successfully."
// @Failure 422 {object} httpx.ErrorResponse "Invalid state or accounts"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders/{id}/complete [post]
func (h ServiceHandler) CompleteOrder(c fiber.Ctx) error {
	var request CompleteServiceOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDateStr(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	order, err := h.svc.FindOrder(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service order get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service order.", err)
	}
	if order == nil || !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service order not found.")
	}
	updated, err := h.svc.Complete(c, service.CompleteRequest{
		OrderID:                 order.ID,
		JournalID:               request.JournalID,
		Date:                    date,
		COGSAccountID:           request.COGSAccountID,
		StockValuationAccountID: request.StockValuationAccountID,
		Resolution:              request.Resolution,
	})
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Service order completed successfully.", GetServiceOrderResponse{
		ServiceOrder: newServiceOrderResponse(updated),
	})
}

type BillServiceOrderRequest struct {
	JournalID        uint64  `json:"journal_id" validate:"required,gt=0"`
	Date             string  `json:"date"`
	RevenueAccountID uint64  `json:"revenue_account_id" validate:"required,gt=0"`
	DeferredAccount  *uint64 `json:"deferred_account"`
}

// @Summary Bill service order
// @Description Creates a customer invoice from the billable lines of a done service order and marks it invoiced.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service order ID"
// @Param body body BillServiceOrderRequest true "Billing details"
// @Success 200 {object} GetServiceOrderResponseEnvelope "Service order billed successfully."
// @Failure 422 {object} httpx.ErrorResponse "Order must be done with billable lines"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders/{id}/bill [post]
func (h ServiceHandler) BillOrder(c fiber.Ctx) error {
	var request BillServiceOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := parseOptionalDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	order, err := h.svc.FindOrder(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service order get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service order.", err)
	}
	if order == nil || !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service order not found.")
	}
	updated, err := h.svc.Bill(c, service.BillRequest{
		OrderID: order.ID, JournalID: request.JournalID, Date: date,
		RevenueAccountID: request.RevenueAccountID, DeferredAccount: request.DeferredAccount,
	})
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Service order billed successfully.", GetServiceOrderResponse{
		ServiceOrder: newServiceOrderResponse(updated),
	})
}

// @Summary Cancel service order
// @Description Cancels a service order that is not yet done or invoiced.
// @Tags Service & Maintenance
// @Accept json
// @Produce json
// @Param id path integer true "Service order ID"
// @Success 200 {object} GetServiceOrderResponseEnvelope "Service order cancelled successfully."
// @Failure 422 {object} httpx.ErrorResponse "Order is done or invoiced"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/service-orders/{id}/cancel [post]
func (h ServiceHandler) CancelOrder(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	order, err := h.svc.FindOrder(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("service order get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve service order.", err)
	}
	if order == nil || !httpx.OwnsTenant(c, order.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Service order not found.")
	}
	updated, err := h.svc.CancelOrder(c, order.ID)
	if err != nil {
		return writeServiceError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Service order cancelled successfully.", GetServiceOrderResponse{
		ServiceOrder: newServiceOrderResponse(updated),
	})
}
