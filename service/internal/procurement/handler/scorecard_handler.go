package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type SupplierScorecardHandler struct {
	svc procurement.SupplierScorecardService
}

func NewSupplierScorecardHandler(svc procurement.SupplierScorecardService) SupplierScorecardHandler {
	return SupplierScorecardHandler{svc: svc}
}

type SupplierScorecardResponse struct {
	ID               uint64     `json:"id"`
	SupplierID       uint64     `json:"supplier_id"`
	PeriodStart      *time.Time `json:"period_start"`
	PeriodEnd        *time.Time `json:"period_end"`
	QualityScore     float64    `json:"quality_score"`
	DeliveryScore    float64    `json:"delivery_score"`
	PriceScore       float64    `json:"price_score"`
	OverallScore     float64    `json:"overall_score"`
	TotalOrders      int        `json:"total_orders"`
	OnTimeDeliveries int        `json:"on_time_deliveries"`
	QualityFailures  int        `json:"quality_failures"`
}

func newSupplierScorecardResponse(sc *procurement.SupplierScorecard) SupplierScorecardResponse {
	return SupplierScorecardResponse{
		ID:               sc.ID,
		SupplierID:       sc.SupplierID,
		PeriodStart:      sc.PeriodStart,
		PeriodEnd:        sc.PeriodEnd,
		QualityScore:     sc.QualityScore,
		DeliveryScore:    sc.DeliveryScore,
		PriceScore:       sc.PriceScore,
		OverallScore:     sc.OverallScore,
		TotalOrders:      sc.TotalOrders,
		OnTimeDeliveries: sc.OnTimeDeliveries,
		QualityFailures:  sc.QualityFailures,
	}
}

var vendorScorecardQueryAllowlist = map[string]struct{}{
	"supplier_id":    {},
	"period_start":   {},
	"period_end":     {},
	"quality_score":  {},
	"delivery_score": {},
	"price_score":    {},
	"overall_score":  {},
	"created_at":     {},
	"updated_at":     {},
}

type ListSupplierScorecardsResponse struct {
	Scorecards []SupplierScorecardResponse `json:"scorecards"`
}

type ListSupplierScorecardsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSupplierScorecardsResponse `json:"data"`
}

// @Summary List supplier scorecards
// @Description Lists supplier scorecards across the caller's organization with pagination, sorting, and filtering.
// @Tags Supplier Scorecards
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListSupplierScorecardsResponseEnvelope "Supplier scorecards retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-scorecards [get]
func (h SupplierScorecardHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, vendorScorecardQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("supplier scorecard list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve supplier scorecards.", err)
	}
	items := make([]SupplierScorecardResponse, len(page.Items))
	for i, sc := range page.Items {
		items[i] = newSupplierScorecardResponse(sc)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Supplier scorecards retrieved successfully.", ListSupplierScorecardsResponse{
		Scorecards: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetSupplierScorecardResponse struct {
	Scorecard SupplierScorecardResponse `json:"scorecard"`
}

type GetSupplierScorecardResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetSupplierScorecardResponse `json:"data"`
}

// @Summary Get supplier scorecard
// @Description Gets a single supplier scorecard by id.
// @Tags Supplier Scorecards
// @Accept json
// @Produce json
// @Param id path integer true "Supplier scorecard ID"
// @Success 200 {object} GetSupplierScorecardResponseEnvelope "Supplier scorecard retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Supplier scorecard not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-scorecards/{id} [get]
func (h SupplierScorecardHandler) Get(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid supplier scorecard id provided.", nil)
	}
	sc, err := h.svc.Find(c, id)
	if err != nil {
		return writeScorecardError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Supplier scorecard retrieved successfully.", GetSupplierScorecardResponse{
		Scorecard: newSupplierScorecardResponse(sc),
	})
}

type ListSupplierScorecardsByVendorResponse struct {
	Scorecards []SupplierScorecardResponse `json:"scorecards"`
}

type ListSupplierScorecardsByVendorResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSupplierScorecardsByVendorResponse `json:"data"`
}

// @Summary List supplier scorecards by supplier
// @Description Lists all scorecards for a specific supplier.
// @Tags Supplier Scorecards
// @Accept json
// @Produce json
// @Param supplier_id path integer true "Supplier ID"
// @Success 200 {object} ListSupplierScorecardsByVendorResponseEnvelope "Supplier scorecards retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-scorecards/supplier/{supplier_id} [get]
func (h SupplierScorecardHandler) ListBySupplier(c fiber.Ctx) error {
	supplierID, err := strconv.ParseUint(c.Params("supplier_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid supplier id provided.", nil)
	}
	scorecards, err := h.svc.ListBySupplier(c, supplierID)
	if err != nil {
		httpx.RequestLog(c).Error("supplier scorecard list by supplier failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve supplier scorecards.", err)
	}
	items := make([]SupplierScorecardResponse, len(scorecards))
	for i, sc := range scorecards {
		items[i] = newSupplierScorecardResponse(sc)
	}
	return httpx.CreateSuccessResponse(c, "Supplier scorecards retrieved successfully.", ListSupplierScorecardsByVendorResponse{
		Scorecards: items,
	})
}

func writeScorecardError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, procurement.ErrScorecardNotFound):
		return httpx.CreateNotFoundResponse(c, "Supplier scorecard not found.")
	default:
		httpx.RequestLog(c).Error("supplier scorecard write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process supplier scorecard.", err)
	}
}
