package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/quality"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type QualityPointHandler struct {
	svc quality.QualityPointService
}

func NewQualityPointHandler(svc quality.QualityPointService) QualityPointHandler {
	return QualityPointHandler{svc: svc}
}

type QualityPointResponse struct {
	ID             uint64   `json:"id"`
	OrganizationID *uint64  `json:"organization_id"`
	ItemID         *uint64  `json:"item_id"`
	Operation      *string  `json:"operation"`
	TestType       string   `json:"test_type"`
	NormMin        *float64 `json:"norm_min"`
	NormMax        *float64 `json:"norm_max"`
	UnitID         *uint64  `json:"unit_id"`
}

func newQualityPointResponse(point *reference.QualityPoint) QualityPointResponse {
	return QualityPointResponse{
		ID:             point.ID,
		OrganizationID: point.OrganizationID,
		ItemID:         point.ItemID,
		Operation:      point.Operation,
		TestType:       point.TestType,
		NormMin:        point.NormMin,
		NormMax:        point.NormMax,
		UnitID:         point.UnitID,
	}
}

var qualityPointQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"item_id":         {},
	"operation":       {},
	"test_type":       {},
	"created_at":      {},
	"updated_at":      {},
}

type ListQualityPointsResponse struct {
	Points []QualityPointResponse `json:"points"`
}

type ListQualityPointsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListQualityPointsResponse `json:"data"`
}

// @Summary List quality points
// @Description Lists quality control points with pagination, sorting, and filtering. Results are scoped to the caller's organization and can be narrowed by item, operation, and test type so inspection points can be managed per item line.
// @Tags Quality Points
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListQualityPointsResponseEnvelope "Quality points retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-points [get]
func (h QualityPointHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, qualityPointQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("quality point list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve quality points.", err)
	}
	items := make([]QualityPointResponse, len(page.Items))
	for i, point := range page.Items {
		items[i] = newQualityPointResponse(point)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Quality points retrieved successfully.", ListQualityPointsResponse{
		Points: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetQualityPointResponse struct {
	Point QualityPointResponse `json:"point"`
}

type GetQualityPointResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetQualityPointResponse `json:"data"`
}

// @Summary Get quality point
// @Description Gets a single quality control point by id, including its test type, optional norm bounds, and unit of measure. Points that do not belong to the caller's organization are treated as not found.
// @Tags Quality Points
// @Accept json
// @Produce json
// @Param id path integer true "Quality point ID"
// @Success 200 {object} GetQualityPointResponseEnvelope "Quality point retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Quality point not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-points/{id} [get]
func (h QualityPointHandler) Get(c fiber.Ctx) error {
	pointID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid quality point id provided.", nil)
	}
	point, err := h.svc.Find(c, pointID)
	if err != nil {
		httpx.RequestLog(c).Error("quality point lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get quality point.", err)
	}
	if point == nil || !httpx.OwnsTenant(c, point.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Quality point not found.")
	}
	return httpx.CreateSuccessResponse(c, "Quality point retrieved successfully.", GetQualityPointResponse{
		Point: newQualityPointResponse(point),
	})
}

type CreateQualityPointRequest struct {
	OrganizationID *uint64  `json:"organization_id"`
	ItemID         *uint64  `json:"item_id"`
	Operation      *string  `json:"operation"`
	TestType       string   `json:"test_type" validate:"required,oneof=pass_fail measure instruction"`
	NormMin        *float64 `json:"norm_min"`
	NormMax        *float64 `json:"norm_max"`
	UnitID         *uint64  `json:"unit_id"`
}

type CreateQualityPointResponse struct {
	Point QualityPointResponse `json:"point"`
}

type CreateQualityPointResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateQualityPointResponse `json:"data"`
}

// @Summary Create quality point
// @Description Creates a new quality control point for a item within the caller's organization. The test type must be pass/fail, measurement, or instruction, and an optional norm minimum and maximum with unit of measure can be supplied to support automated pass or fail determination for measurement checks.
// @Tags Quality Points
// @Accept json
// @Produce json
// @Param body body CreateQualityPointRequest true "Quality point details"
// @Success 201 {object} CreateQualityPointResponseEnvelope "Quality point created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-points [post]
func (h QualityPointHandler) Create(c fiber.Ctx) error {
	var request CreateQualityPointRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	point, err := h.svc.Create(c, &reference.QualityPoint{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		ItemID:         request.ItemID,
		Operation:      request.Operation,
		TestType:       request.TestType,
		NormMin:        request.NormMin,
		NormMax:        request.NormMax,
		UnitID:         request.UnitID,
	})
	if err != nil {
		httpx.RequestLog(c).Error("quality point create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create quality point.", err)
	}
	return httpx.CreateCreatedResponse(c, "Quality point created successfully.", CreateQualityPointResponse{
		Point: newQualityPointResponse(point),
	})
}

type QualityCheckResponse struct {
	ID                uint64     `json:"id"`
	PointID           *uint64    `json:"point_id"`
	ItemID            *uint64    `json:"item_id"`
	BatchID           *uint64    `json:"batch_id"`
	ShipmentID        *uint64    `json:"shipment_id"`
	ProductionOrderID *uint64    `json:"production_order_id"`
	MeasuredValue     *float64   `json:"measured_value"`
	Result            string     `json:"result"`
	CheckedBy         *uint64    `json:"checked_by"`
	CheckedAt         *time.Time `json:"checked_at"`
}

type QualityCheckHandler struct {
	svc quality.QualityCheckService
}

func NewQualityCheckHandler(svc quality.QualityCheckService) QualityCheckHandler {
	return QualityCheckHandler{svc: svc}
}

func newQualityCheckResponse(check *quality.QualityCheck) QualityCheckResponse {
	return QualityCheckResponse{
		ID:                check.ID,
		PointID:           check.PointID,
		ItemID:            check.ItemID,
		BatchID:           check.BatchID,
		ShipmentID:        check.ShipmentID,
		ProductionOrderID: check.ProductionOrderID,
		MeasuredValue:     check.MeasuredValue,
		Result:            check.Result,
		CheckedBy:         check.CheckedBy,
		CheckedAt:         check.CheckedAt,
	}
}

var qualityCheckQueryAllowlist = map[string]struct{}{
	"point_id":            {},
	"item_id":             {},
	"batch_id":            {},
	"shipment_id":         {},
	"production_order_id": {},
	"result":              {},
	"checked_by":          {},
	"created_at":          {},
	"updated_at":          {},
}

type ListQualityChecksResponse struct {
	Checks []QualityCheckResponse `json:"checks"`
}

type ListQualityChecksResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListQualityChecksResponse `json:"data"`
}

// @Summary List quality checks
// @Description Lists quality checks with pagination, sorting, and filtering. Filters by quality point, item, lot, shipment, production order, and result are supported so pending inspections and their outcomes can be reviewed across the warehouse.
// @Tags Quality Checks
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListQualityChecksResponseEnvelope "Quality checks retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-checks [get]
func (h QualityCheckHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, qualityCheckQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	page, err := h.svc.ListChecksInOrg(c, parsedQuery, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("quality check list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve quality checks.", err)
	}
	items := make([]QualityCheckResponse, len(page.Items))
	for i, check := range page.Items {
		items[i] = newQualityCheckResponse(check)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Quality checks retrieved successfully.", ListQualityChecksResponse{
		Checks: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetQualityCheckResponse struct {
	Check QualityCheckResponse `json:"check"`
}

type GetQualityCheckResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetQualityCheckResponse `json:"data"`
}

// @Summary Get quality check
// @Description Gets a single quality check by id, including its quality point, item and lot context, measured value, pass or fail result, and the checker who recorded it along with the check time.
// @Tags Quality Checks
// @Accept json
// @Produce json
// @Param id path integer true "Quality check ID"
// @Success 200 {object} GetQualityCheckResponseEnvelope "Quality check retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Quality check not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-checks/{id} [get]
func (h QualityCheckHandler) Get(c fiber.Ctx) error {
	checkID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid quality check id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	check, err := h.svc.FindCheckInOrg(c, checkID, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("quality check lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get quality check.", err)
	}
	if check == nil {
		return httpx.CreateNotFoundResponse(c, "Quality check not found.")
	}
	return httpx.CreateSuccessResponse(c, "Quality check retrieved successfully.", GetQualityCheckResponse{
		Check: newQualityCheckResponse(check),
	})
}

type RecordQualityCheckRequest struct {
	Pass          bool     `json:"pass"`
	MeasuredValue *float64 `json:"measured_value"`
	CheckedBy     *uint64  `json:"checked_by"`
}

// @Summary Record quality check result
// @Description Records the pass or fail outcome of a pending quality check, storing the measured value and the checking user. For measurement-type points the measured value is mandatory and is compared against the point's norm bounds to derive the result; a failed check automatically raises a quality alert, and already-resolved checks are rejected with a conflict.
// @Tags Quality Checks
// @Accept json
// @Produce json
// @Param id path integer true "Quality check ID"
// @Param body body RecordQualityCheckRequest true "Check result"
// @Success 200 {object} GetQualityCheckResponseEnvelope "Quality check recorded successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Quality check not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-checks/{id}/result [post]
func (h QualityCheckHandler) RecordResult(c fiber.Ctx) error {
	checkID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid quality check id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	if check, err := h.svc.FindCheckInOrg(c, checkID, organizationID); err != nil {
		httpx.RequestLog(c).Error("quality check lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get quality check.", err)
	} else if check == nil {
		return httpx.CreateNotFoundResponse(c, "Quality check not found.")
	}
	var request RecordQualityCheckRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	checkedBy, ok := httpx.CallerID(c)
	if !ok || request.CheckedBy != nil {
		if request.CheckedBy != nil {
			checkedBy = *request.CheckedBy
		}
	}
	check, err := h.svc.RecordResult(c, checkID, request.Pass, request.MeasuredValue, checkedBy)
	if err != nil {
		return writeQualityError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Quality check recorded successfully.", GetQualityCheckResponse{
		Check: newQualityCheckResponse(check),
	})
}

type RouteToScrapRequest struct {
	JournalID        uint64 `json:"journal_id" validate:"required,gt=0"`
	ExpenseAccountID uint64 `json:"expense_account_id" validate:"required,gt=0"`
}

type RouteToScrapResponse struct {
	Scrapped []uint64 `json:"scrapped_item_ids"`
}

type RouteToScrapResponseEnvelope struct {
	httpx.EnvelopeBase
	Data RouteToScrapResponse `json:"data"`
}

// @Summary Route failed checks to scrap
// @Description Routes the failed quantities of a shipment's quality checks to scrap by creating scrap stock movements posted to the given journal and expense account within the caller's organization. Requires a configured scrap location and sufficient stock, and returns the ids of the products that were scrapped.
// @Tags Quality Checks
// @Accept json
// @Produce json
// @Param shipment_id path integer true "Stock shipment ID"
// @Param body body RouteToScrapRequest true "Scrap routing details"
// @Success 200 {object} RouteToScrapResponseEnvelope "Failed quality checks routed to scrap."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-checks/shipments/{shipment_id}/scrap [post]
func (h QualityCheckHandler) RouteToScrap(c fiber.Ctx) error {
	shipmentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid shipment id provided.", nil)
	}
	var request RouteToScrapRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	scrapped, err := h.svc.RouteFailedToScrap(c, organizationID, shipmentID, request.JournalID, request.ExpenseAccountID, time.Now())
	if err != nil {
		return writeQualityError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Failed quality checks routed to scrap.", RouteToScrapResponse{
		Scrapped: scrapped,
	})
}

type QualityAlertResponse struct {
	ID          uint64  `json:"id"`
	ItemID      *uint64 `json:"item_id"`
	BatchID     *uint64 `json:"batch_id"`
	CheckID     *uint64 `json:"check_id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Severity    *string `json:"severity"`
	State       string  `json:"state"`
	AssignedTo  *uint64 `json:"assigned_to"`
}

type QualityAlertHandler struct {
	svc quality.QualityCheckService
}

func NewQualityAlertHandler(svc quality.QualityCheckService) QualityAlertHandler {
	return QualityAlertHandler{svc: svc}
}

func newQualityAlertResponse(alert *quality.QualityAlert) QualityAlertResponse {
	return QualityAlertResponse{
		ID:          alert.ID,
		ItemID:      alert.ItemID,
		BatchID:     alert.BatchID,
		CheckID:     alert.CheckID,
		Title:       alert.Title,
		Description: alert.Description,
		Severity:    alert.Severity,
		State:       alert.State,
		AssignedTo:  alert.AssignedTo,
	}
}

var qualityAlertQueryAllowlist = map[string]struct{}{
	"item_id":     {},
	"batch_id":    {},
	"check_id":    {},
	"severity":    {},
	"state":       {},
	"assigned_to": {},
	"created_at":  {},
	"updated_at":  {},
}

type ListQualityAlertsResponse struct {
	Alerts []QualityAlertResponse `json:"alerts"`
}

type ListQualityAlertsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListQualityAlertsResponse `json:"data"`
}

// @Summary List quality alerts
// @Description Lists quality alerts with pagination, sorting, and filtering. Filters by item, lot, originating check, severity, state, and assignee are supported so the quality team can triage and track open issues.
// @Tags Quality Alerts
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListQualityAlertsResponseEnvelope "Quality alerts retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-alerts [get]
func (h QualityAlertHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, qualityAlertQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	page, err := h.svc.ListAlertsInOrg(c, parsedQuery, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("quality alert list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve quality alerts.", err)
	}
	items := make([]QualityAlertResponse, len(page.Items))
	for i, alert := range page.Items {
		items[i] = newQualityAlertResponse(alert)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Quality alerts retrieved successfully.", ListQualityAlertsResponse{
		Alerts: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetQualityAlertResponse struct {
	Alert QualityAlertResponse `json:"alert"`
}

type GetQualityAlertResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetQualityAlertResponse `json:"data"`
}

// @Summary Get quality alert
// @Description Gets a single quality alert by id, including its linked item, lot, and originating quality check, along with severity, current state, and assignee.
// @Tags Quality Alerts
// @Accept json
// @Produce json
// @Param id path integer true "Quality alert ID"
// @Success 200 {object} GetQualityAlertResponseEnvelope "Quality alert retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Quality alert not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-alerts/{id} [get]
func (h QualityAlertHandler) Get(c fiber.Ctx) error {
	alertID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid quality alert id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	alert, err := h.svc.FindAlertInOrg(c, alertID, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("quality alert lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get quality alert.", err)
	}
	if alert == nil {
		return httpx.CreateNotFoundResponse(c, "Quality alert not found.")
	}
	return httpx.CreateSuccessResponse(c, "Quality alert retrieved successfully.", GetQualityAlertResponse{
		Alert: newQualityAlertResponse(alert),
	})
}

type UpdateQualityAlertRequest struct {
	State string `json:"state" validate:"required,oneof=open in_progress solved cancelled"`
}

// @Summary Update quality alert state
// @Description Transitions a quality alert through its lifecycle: open to in_progress or cancelled, and in_progress to solved or cancelled. Invalid transitions are rejected with a conflict so the resolution workflow stays consistent.
// @Tags Quality Alerts
// @Accept json
// @Produce json
// @Param id path integer true "Quality alert ID"
// @Param body body UpdateQualityAlertRequest true "Alert state"
// @Success 200 {object} GetQualityAlertResponseEnvelope "Quality alert updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Quality alert not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /quality-alerts/{id}/state [put]
func (h QualityAlertHandler) UpdateState(c fiber.Ctx) error {
	alertID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid quality alert id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	if alert, err := h.svc.FindAlertInOrg(c, alertID, organizationID); err != nil {
		httpx.RequestLog(c).Error("quality alert lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get quality alert.", err)
	} else if alert == nil {
		return httpx.CreateNotFoundResponse(c, "Quality alert not found.")
	}
	var request UpdateQualityAlertRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	alert, err := h.svc.SetAlertState(c, alertID, request.State)
	if err != nil {
		return writeQualityError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Quality alert updated successfully.", GetQualityAlertResponse{
		Alert: newQualityAlertResponse(alert),
	})
}

func writeQualityError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, quality.ErrQualityCheckNotFound):
		return httpx.CreateNotFoundResponse(c, "Quality check not found.")
	case errors.Is(err, quality.ErrQualityCheckDone):
		return httpx.CreateConflictResponse(c, "Quality check is already resolved.", err)
	case errors.Is(err, quality.ErrQualityCheckNoValue):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Measured value required for measurement check.", nil)
	case errors.Is(err, quality.ErrQualityCheckNoProduct):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Quality check requires a item.", nil)
	case errors.Is(err, quality.ErrQualityPointNotFound):
		return httpx.CreateNotFoundResponse(c, "Quality point not found.")
	case errors.Is(err, quality.ErrQualityAlertNotFound):
		return httpx.CreateNotFoundResponse(c, "Quality alert not found.")
	case errors.Is(err, quality.ErrQualityAlertState):
		return httpx.CreateConflictResponse(c, "Quality alert cannot transition to that state.", err)
	case errors.Is(err, quality.ErrQualityScrapRouter):
		return httpx.CreateInternalServerErrorResponse(c, "Scrap routing is not configured.", err)
	case errors.Is(err, quality.ErrScrapLocation):
		return httpx.CreateNotFoundResponse(c, "No scrap location found for this organization.")
	case errors.Is(err, quality.ErrScrapAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Scrap expense account is required.", nil)
	case errors.Is(err, quality.ErrMovementNotFound):
		return httpx.CreateNotFoundResponse(c, "Stock movement or shipment not found.")
	case errors.Is(err, quality.ErrInsufficientStock):
		return httpx.CreateConflictResponse(c, "Insufficient stock for scrapping.", err)
	default:
		httpx.RequestLog(c).Error("quality write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process quality operation.", err)
	}
}
