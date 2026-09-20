package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ReminderHandler struct {
	svc accounting.ReminderService
}

func NewReminderHandler(svc accounting.ReminderService) ReminderHandler {
	return ReminderHandler{svc: svc}
}

type ReminderActionResponse struct {
	ID        uint64  `json:"id"`
	ContactID uint64  `json:"contact_id"`
	InvoiceID uint64  `json:"invoice_id"`
	LevelID   uint64  `json:"level_id"`
	SentAt    *string `json:"sent_at"`
	Channel   *string `json:"channel"`
}

func newReminderActionResponse(action *accounting.ReminderAction) ReminderActionResponse {
	return ReminderActionResponse{
		ID:        action.ID,
		ContactID: action.ContactID,
		InvoiceID: action.InvoiceID,
		LevelID:   action.LevelID,
		SentAt:    helper.FormatDatePtr(action.SentAt),
		Channel:   action.Channel,
	}
}

var reminderActionQueryAllowlist = map[string]struct{}{
	"contact_id": {},
	"invoice_id": {},
	"level_id":   {},
	"created_at": {},
	"updated_at": {},
}

type ListReminderActionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListReminderActionsResponse `json:"data"`
}
type ListReminderActionsResponse struct {
	Actions []ReminderActionResponse `json:"actions"`
}

// @Summary List reminder actions
// @Description Lists reminder actions that have been sent for the caller's tenant, applying pagination, sorting, and filtering over fields such as contact, invoice, and level. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Reminder
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. created_at:desc)"
// @Param filter query string false "Filters (repeatable, e.g. contact_id:eq:5)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListReminderActionsResponseEnvelope "Reminder actions retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reminder/actions [get]
func (h ReminderHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, reminderActionQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("reminder action list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve reminder actions.", err)
	}

	items := make([]ReminderActionResponse, len(page.Items))
	for i, action := range page.Items {
		items[i] = newReminderActionResponse(action)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "reminder-actions.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Reminder actions retrieved successfully.", ListReminderActionsResponse{
		Actions: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GenerateReminderRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	AsOf           *string `json:"as_of"`
}

type GenerateReminderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GenerateReminderResponse `json:"data"`
}
type GenerateReminderResponse struct {
	Actions []ReminderActionResponse `json:"actions"`
}

// @Summary Generate reminder actions
// @Description Generates reminder actions for overdue open invoices at the applicable reminder level as of a given date, escalating unpaid receivables through their configured levels. Invoices that already have an action at the applicable level are skipped, and an error is returned when no reminder level applies to the overdue invoices.
// @Tags Reminder
// @Accept json
// @Produce json
// @Param body body GenerateReminderRequest true "Reminder generation options"
// @Success 200 {object} GenerateReminderResponseEnvelope "Reminder actions generated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "No reminder level applies"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reminder/actions/generate [post]
func (h ReminderHandler) Generate(c fiber.Ctx) error {
	var request GenerateReminderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	asOf, err := helper.ParseDate(request.AsOf)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "As-of date must be in YYYY-MM-DD format.", nil)
	}

	actions, err := h.svc.Generate(c, accounting.GenerateReminderRequest{OrganizationID: *organizationID, AsOf: asOf})
	if err != nil {
		return writeReminderError(c, err)
	}

	items := make([]ReminderActionResponse, len(actions))
	for i, action := range actions {
		items[i] = newReminderActionResponse(action)
	}

	return httpx.CreateSuccessResponse(c, "Reminder actions generated successfully.", GenerateReminderResponse{
		Actions: items,
	})
}

func writeReminderError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrReminderLevelNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No reminder level applies to the invoices.", nil)
	case errors.Is(err, accounting.ErrReminderAlreadySent):
		return httpx.CreateConflictResponse(c, "Reminder action already sent for the invoice at this level.", err)
	default:
		httpx.RequestLog(c).Error("reminder generate failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to generate reminder actions.", err)
	}
}
