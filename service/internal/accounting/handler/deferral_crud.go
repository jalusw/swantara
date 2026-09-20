package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary List deferral schedules
// @Description Lists the deferral schedules belonging to the caller's organization.
// @Tags Deferrals
// @Accept json
// @Produce json
// @Success 200 {object} ListDeferralSchedulesResponseEnvelope "Deferral schedules retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Organization is required"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/deferrals [get]
func (h DeferralHandler) List(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	schedules, err := h.svc.List(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("deferral schedule list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve deferral schedules.", err)
	}

	items := make([]DeferralScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		items[i] = newDeferralScheduleResponse(schedule)
	}

	return httpx.CreateSuccessResponse(c, "Deferral schedules retrieved successfully.", ListDeferralSchedulesResponse{
		Schedules: items,
	})
}

// @Summary Get deferral schedule
// @Description Gets a single deferral schedule by id for the caller's organization. A 404 is returned when it does not exist or belongs to another organization.
// @Tags Deferrals
// @Accept json
// @Produce json
// @Param id path integer true "Deferral schedule ID"
// @Success 200 {object} GetDeferralScheduleResponseEnvelope "Deferral schedule retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Deferral schedule not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/deferrals/{id} [get]
func (h DeferralHandler) Get(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	scheduleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid deferral schedule id provided.", nil)
	}

	schedule, err := h.svc.Get(c, *organizationID, scheduleID)
	if err != nil {
		return writeDeferralError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Deferral schedule retrieved successfully.", GetDeferralScheduleResponseEnvelope{
		Data: newDeferralScheduleResponse(schedule),
	})
}

// @Summary List deferral schedule lines
// @Description Lists the recognition lines of a deferral schedule for the caller's organization.
// @Tags Deferrals
// @Accept json
// @Produce json
// @Param id path integer true "Deferral schedule ID"
// @Success 200 {object} ListDeferralLinesResponseEnvelope "Deferral schedule lines retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Deferral schedule not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/deferrals/{id}/lines [get]
func (h DeferralHandler) ListLines(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	scheduleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid deferral schedule id provided.", nil)
	}

	if _, err := h.svc.Get(c, *organizationID, scheduleID); err != nil {
		return writeDeferralError(c, err)
	}

	lines, err := h.svc.ListLines(c, scheduleID)
	if err != nil {
		httpx.RequestLog(c).Error("deferral schedule line list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve deferral schedule lines.", err)
	}

	items := make([]DeferralLineResponse, len(lines))
	for i, line := range lines {
		items[i] = DeferralLineResponse{
			ID:              line.ID,
			ScheduleID:      line.ScheduleID,
			Sequence:        line.Sequence,
			RecognitionDate: line.RecognitionDate,
			Amount:          line.Amount,
			Posted:          line.Posted,
			EntryID:         line.EntryID,
		}
	}

	return httpx.CreateSuccessResponse(c, "Deferral schedule lines retrieved successfully.", ListDeferralLinesResponse{
		Lines: items,
	})
}

// @Summary Create a deferral schedule
// @Description Creates a deferral schedule (deferred revenue, deferred expense or prepaid) from an invoice/bill line source with a linear, manual or milestone method.
// @Tags Deferrals
// @Accept json
// @Produce json
// @Param body body CreateDeferralScheduleRequest true "Deferral schedule details"
// @Success 201 {object} CreateDeferralScheduleResponseEnvelope "Deferral schedule created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/deferrals [post]
func (h DeferralHandler) Create(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	var request CreateDeferralScheduleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	lines := make([]accounting.ScheduleLineRequest, len(request.Lines))
	for i, line := range request.Lines {
		if line.RecognitionDate == nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Each schedule line requires a recognition date.", nil)
		}
		lines[i] = accounting.ScheduleLineRequest{
			RecognitionDate: *line.RecognitionDate,
			Amount:          line.Amount,
		}
	}

	var dateEnd time.Time
	if request.DateEnd != nil {
		dateEnd = *request.DateEnd
	}
	schedule, err := h.svc.Create(c, accounting.CreateScheduleRequest{
		OrganizationID:        *organizationID,
		Type:                  request.Type,
		SourceType:            request.SourceType,
		SourceID:              request.SourceID,
		ContactID:             request.ContactID,
		ItemID:                request.ItemID,
		TotalAmount:           request.TotalAmount,
		BalanceSheetAccountID: request.BalanceSheetAccountID,
		PLAccountID:           request.PLAccountID,
		DimensionID:           request.DimensionID,
		Method:                request.Method,
		DateStart:             *request.DateStart,
		DateEnd:               dateEnd,
		Periods:               request.Periods,
		Lines:                 lines,
	})
	if err != nil {
		return writeDeferralError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Deferral schedule created successfully.", CreateDeferralScheduleResponseEnvelope{
		Data: newDeferralScheduleResponse(schedule),
	})
}
