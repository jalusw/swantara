package handler

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
)

type IntegrationEventHandler struct {
	eventSvc xtradata.EventService
}

func NewIntegrationEventHandler(eventSvc xtradata.EventService) IntegrationEventHandler {
	return IntegrationEventHandler{eventSvc: eventSvc}
}

type IntegrationEventResponse struct {
	ID        uint64          `json:"id"`
	Topic     string          `json:"topic"`
	Payload   json.RawMessage `json:"payload" swaggertype:"object"`
	Status    string          `json:"status"`
	Retries   int             `json:"retries"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func newIntegrationEventResponse(event *xtradata.IntegrationEvent) IntegrationEventResponse {
	return IntegrationEventResponse{
		ID:        event.ID,
		Topic:     event.Topic,
		Payload:   event.Payload,
		Status:    event.Status,
		Retries:   event.Retries,
		CreatedAt: event.CreatedAt,
		UpdatedAt: event.UpdatedAt,
	}
}

var integrationEventQueryAllowlist = map[string]struct{}{
	"topic":      {},
	"status":     {},
	"created_at": {},
	"updated_at": {},
}

type ListIntegrationEventsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListIntegrationEventsResponse `json:"data"`
}
type ListIntegrationEventsResponse struct {
	IntegrationEvents []IntegrationEventResponse `json:"integration_events"`
}

// @Summary List integration events
// @Description Lists outbound integration events with pagination, sorting, and filtering; the response can be exported as JSON, XML, or CSV.
// @Tags Integration Events
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. created_at:desc)"
// @Param filter query string false "Filters (repeatable, e.g. status:eq:pending)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListIntegrationEventsResponseEnvelope "Integration events retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/integration-events [get]
func (h IntegrationEventHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, integrationEventQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	page, err := h.eventSvc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("integration event list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve integration events.", err)
	}

	items := make([]IntegrationEventResponse, len(page.Items))
	for i, event := range page.Items {
		items[i] = newIntegrationEventResponse(event)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "integration-events.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Integration events retrieved successfully.", ListIntegrationEventsResponse{
		IntegrationEvents: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetIntegrationEventResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetIntegrationEventResponse `json:"data"`
}
type GetIntegrationEventResponse struct {
	IntegrationEvent IntegrationEventResponse `json:"integration_event"`
}

// @Summary Get integration event
// @Description Returns a single integration event by id; a non-numeric id yields a 422.
// @Tags Integration Events
// @Accept json
// @Produce json
// @Param id path integer true "Integration event ID"
// @Success 200 {object} GetIntegrationEventResponseEnvelope "Integration event retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Integration event not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/integration-events/{id} [get]
func (h IntegrationEventHandler) Get(c fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid integration event id provided.", nil)
	}

	event, err := h.eventSvc.Find(c, eventID)
	if err != nil {
		httpx.RequestLog(c).Error("integration event lookup failed", "event_id", eventID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve integration event.", err)
	}
	if event == nil {
		return httpx.CreateNotFoundResponse(c, "Integration event not found.")
	}

	return httpx.CreateSuccessResponse(c, "Integration event retrieved successfully.", GetIntegrationEventResponse{
		IntegrationEvent: newIntegrationEventResponse(event),
	})
}

type DispatchIntegrationEventResponseEnvelope struct {
	httpx.EnvelopeBase
	Data DispatchIntegrationEventResponse `json:"data"`
}
type DispatchIntegrationEventResponse struct {
	IntegrationEvent IntegrationEventResponse `json:"integration_event"`
}

// @Summary Dispatch/retry integration event
// @Description Re-dispatches a failed integration event by id, resetting retries and attempting delivery.
// @Tags Integration Events
// @Accept json
// @Produce json
// @Param id path integer true "Integration event ID"
// @Success 200 {object} DispatchIntegrationEventResponseEnvelope "Integration event dispatched successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Integration event not found"
// @Failure 409 {object} httpx.ErrorResponse "Event already sent"
// @Failure 422 {object} httpx.ErrorResponse "Invalid id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/integration-events/{id}/dispatch [post]
func (h IntegrationEventHandler) Dispatch(c fiber.Ctx) error {
	eventID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid integration event id provided.", nil)
	}

	if err := h.eventSvc.Dispatch(c, eventID); err != nil {
		if errors.Is(err, xtradata.ErrEventNotFound) {
			return httpx.CreateNotFoundResponse(c, "Integration event not found.")
		}
		if errors.Is(err, xtradata.ErrEventAlreadySent) {
			return httpx.CreateConflictResponse(c, "Event already sent.", err)
		}
		httpx.RequestLog(c).Error("integration event dispatch failed", "event_id", eventID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to dispatch integration event.", err)
	}

	event, err := h.eventSvc.Find(c, eventID)
	if err != nil {
		httpx.RequestLog(c).Error("integration event fetch failed", "event_id", eventID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve integration event.", err)
	}

	return httpx.CreateSuccessResponse(c, "Integration event dispatched successfully.", DispatchIntegrationEventResponse{
		IntegrationEvent: newIntegrationEventResponse(event),
	})
}
