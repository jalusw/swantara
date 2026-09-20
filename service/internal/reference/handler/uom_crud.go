package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

// @Summary List UoMs
// @Description Lists units of measure with pagination, sorting, and filtering, supporting filters such as category and UoM type; the response can be exported as JSON, XML, or CSV.
// @Tags UoM
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. category_id:eq:1)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListUnitsResponseEnvelope "UoMs retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /units [get]
func (h UnitHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, uomQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	page, err := h.svc.ListUnits(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("unit list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve UoMs.", err)
	}

	items := make([]UnitResponse, len(page.Items))
	for i, unit := range page.Items {
		items[i] = newUnitResponse(unit)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "units.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "UoMs retrieved successfully.", ListUnitsResponse{
		Units: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get UoM
// @Description Returns a single unit of measure by id with its category, conversion factor, type, and rounding; a 404 is returned if the UoM does not exist.
// @Tags UoM
// @Accept json
// @Produce json
// @Param id path integer true "UoM ID"
// @Success 200 {object} GetUnitResponseEnvelope "UoM retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "UoM not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /units/{id} [get]
func (h UnitHandler) Get(c fiber.Ctx) error {
	uomID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid unit id provided.", nil)
	}

	unit, err := h.svc.FindUnit(c, uomID)
	if err != nil {
		httpx.RequestLog(c).Error("unit lookup failed", "unit_id", uomID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get UoM.", err)
	}
	if unit == nil {
		return httpx.CreateNotFoundResponse(c, "UoM not found.")
	}

	return httpx.CreateSuccessResponse(c, "UoM retrieved successfully.", GetUnitResponse{
		Unit: newUnitResponse(unit),
	})
}

// @Summary Create UoM
// @Description Creates a unit of measure within a category, requiring a positive conversion factor and an optional rounding value. The category must exist and the UoM name must be unique within that category, otherwise a 422 is returned.
// @Tags UoM
// @Accept json
// @Produce json
// @Param body body CreateUnitRequest true "UoM details"
// @Success 201 {object} CreateUnitResponseEnvelope "UoM created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error, unknown category, or duplicate name"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /units [post]
func (h UnitHandler) Create(c fiber.Ctx) error {
	var request CreateUnitRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	unit, err := h.svc.CreateUnit(c, &reference.Unit{
		CategoryID: request.CategoryID,
		Name:       request.Name,
		Factor:     request.Factor,
		UnitType:   request.UnitType,
		Rounding:   request.Rounding,
	})
	if err != nil {
		return writeUnitError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "UoM created successfully.", CreateUnitResponse{
		Unit: newUnitResponse(unit),
	})
}

// @Summary Update UoM
// @Description Updates a unit of measure's name, factor, type, and rounding. The factor must remain positive and the name must stay unique within its category; a 404 is returned if the UoM does not exist.
// @Tags UoM
// @Accept json
// @Produce json
// @Param id path integer true "UoM ID"
// @Param body body UpdateUnitRequest true "UoM details"
// @Success 200 {object} UpdateUnitResponseEnvelope "UoM updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "UoM not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or duplicate name"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /units/{id} [put]
func (h UnitHandler) Update(c fiber.Ctx) error {
	uomID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid unit id provided.", nil)
	}

	var request UpdateUnitRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	unit, err := h.svc.FindUnit(c, uomID)
	if err != nil {
		httpx.RequestLog(c).Error("unit lookup failed", "unit_id", uomID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update UoM.", err)
	}
	if unit == nil {
		return httpx.CreateNotFoundResponse(c, "UoM not found.")
	}

	unit.Name = request.Name
	unit.Factor = request.Factor
	unit.UnitType = request.UnitType
	unit.Rounding = request.Rounding

	updated, err := h.svc.UpdateUnit(c, unit)
	if err != nil {
		return writeUnitError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "UoM updated successfully.", UpdateUnitResponse{
		Unit: newUnitResponse(updated),
	})
}

// @Summary Delete UoM
// @Description Deletes a unit of measure by id; returns 404 if the UoM does not exist and 204 No Content on success.
// @Tags UoM
// @Accept json
// @Produce json
// @Param id path integer true "UoM ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "UoM not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /units/{id} [delete]
func (h UnitHandler) Delete(c fiber.Ctx) error {
	uomID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid unit id provided.", nil)
	}

	unit, err := h.svc.FindUnit(c, uomID)
	if err != nil {
		httpx.RequestLog(c).Error("unit lookup failed", "unit_id", uomID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete UoM.", err)
	}
	if unit == nil {
		return httpx.CreateNotFoundResponse(c, "UoM not found.")
	}

	if err := h.svc.DeleteUnit(c, uomID); err != nil {
		httpx.RequestLog(c).Error("unit deletion failed", "unit_id", uomID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete UoM.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
