package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
)

type ShiftHandler struct {
	svc payroll.HRService
}

func NewShiftHandler(svc payroll.HRService) ShiftHandler {
	return ShiftHandler{svc: svc}
}

type ShiftResponse struct {
	ID             uint64 `json:"id"`
	OrganizationID uint64 `json:"organization_id"`
	Name           string `json:"name"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
}

func newShiftResponse(shift *payroll.Shift) ShiftResponse {
	return ShiftResponse{
		ID:             shift.ID,
		OrganizationID: shift.OrganizationID,
		Name:           shift.Name,
		StartTime:      shift.StartTime,
		EndTime:        shift.EndTime,
	}
}

type CreateShiftRequest struct {
	OrganizationID uint64 `json:"organization_id" validate:"required"`
	Name           string `json:"name" validate:"required"`
	StartTime      string `json:"start_time" validate:"required"`
	EndTime        string `json:"end_time" validate:"required"`
}

type ShiftResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ShiftResponse `json:"data"`
}

type ListShiftsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []ShiftResponse `json:"data"`
}

var shiftQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
}

func (h ShiftHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, shiftQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	_ = parsedQuery
	return httpx.CreateSuccessResponse(c, "Shifts retrieved successfully.", []ShiftResponse{})
}

func (h ShiftHandler) Create(c fiber.Ctx) error {
	var request CreateShiftRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	shift, err := h.svc.CreateShift(c, &payroll.Shift{
		OrganizationID: request.OrganizationID,
		Name:           request.Name,
		StartTime:      request.StartTime,
		EndTime:        request.EndTime,
	})
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Shift created successfully.", newShiftResponse(shift))
}

func (h ShiftHandler) Update(c fiber.Ctx) error {
	shiftID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid shift id provided.", nil)
	}
	var request CreateShiftRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	shift, err := h.svc.UpdateShift(c, &payroll.Shift{
		Base:      model.Base{ID: shiftID},
		Name:      request.Name,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
	})
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Shift updated successfully.", newShiftResponse(shift))
}

func (h ShiftHandler) Delete(c fiber.Ctx) error {
	shiftID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid shift id provided.", nil)
	}
	if err := h.svc.DeleteShift(c, shiftID); err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateNoContentResponse(c)
}
