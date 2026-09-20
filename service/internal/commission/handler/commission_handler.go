package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/commission"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type CommissionHandler struct {
	svc commission.CommissionService
}

func NewCommissionHandler(
	svc commission.CommissionService,
) CommissionHandler {
	return CommissionHandler{svc: svc}
}

type CommissionPlanResponse struct {
	ID             uint64  `json:"id"`
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	Basis          string  `json:"basis"`
	Active         bool    `json:"active"`
}

func newCommissionPlanResponse(plan *commission.CommissionPlan) CommissionPlanResponse {
	return CommissionPlanResponse{ID: plan.ID, OrganizationID: plan.OrganizationID, Name: plan.Name, Basis: plan.Basis, Active: plan.Active}
}

var commissionPlanQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"basis":           {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}

type ListCommissionPlansResponse struct {
	CommissionPlans []CommissionPlanResponse `json:"commission_plans"`
}

type ListCommissionPlansResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCommissionPlansResponse `json:"data"`
}

// @Summary List commission plans
// @Description Lists commission plans with pagination, sorting, and filtering. Results are scoped to the caller's organization.
// @Tags Commissions
// @Accept json
// @Produce json
// @Success 200 {object} ListCommissionPlansResponseEnvelope "Commission plans retrieved successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-plans [get]
func (h CommissionHandler) ListPlans(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, commissionPlanQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListPlans(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("commission plan list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission plans.", err)
	}
	items := make([]CommissionPlanResponse, len(page.Items))
	for i, plan := range page.Items {
		items[i] = newCommissionPlanResponse(plan)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Commission plans retrieved successfully.", ListCommissionPlansResponse{
		CommissionPlans: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetCommissionPlanResponse struct {
	CommissionPlan CommissionPlanResponse `json:"commission_plan"`
}

type GetCommissionPlanResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetCommissionPlanResponse `json:"data"`
}

// @Summary Get commission plan
// @Description Gets a single commission plan by id. The plan must belong to the caller's organization; a request for a plan owned by another tenant returns 404.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param id path integer true "Commission plan ID"
// @Success 200 {object} GetCommissionPlanResponseEnvelope "Commission plan retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Commission plan not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-plans/{id} [get]
func (h CommissionHandler) GetPlan(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	plan, err := h.svc.FindPlan(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("commission plan get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission plan.", err)
	}
	if plan == nil || !httpx.OwnsTenant(c, plan.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Commission plan not found.")
	}
	return httpx.CreateSuccessResponse(c, "Commission plan retrieved successfully.", GetCommissionPlanResponse{
		CommissionPlan: newCommissionPlanResponse(plan),
	})
}

type CreateCommissionPlanRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name" validate:"required"`
	Basis          string  `json:"basis" validate:"required,oneof=revenue margin collected"`
}

type CreateCommissionPlanResponse struct {
	CommissionPlan CommissionPlanResponse `json:"commission_plan"`
}

type CreateCommissionPlanResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateCommissionPlanResponse `json:"data"`
}

// @Summary Create commission plan
// @Description Creates a commission plan with a basis (revenue, margin, or collected).
// @Tags Commissions
// @Accept json
// @Produce json
// @Param body body CreateCommissionPlanRequest true "Commission plan details"
// @Success 201 {object} CreateCommissionPlanResponseEnvelope "Commission plan created successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-plans [post]
func (h CommissionHandler) CreatePlan(c fiber.Ctx) error {
	var request CreateCommissionPlanRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	created, err := h.svc.CreatePlan(c, commission.CreatePlanRequest{
		OrganizationID: *organizationID,
		Name:           request.Name,
		Basis:          request.Basis,
	})
	if err != nil {
		return writeCommissionError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Commission plan created successfully.", CreateCommissionPlanResponse{
		CommissionPlan: newCommissionPlanResponse(created),
	})
}

type UpdateCommissionPlanRequest struct {
	Active *bool `json:"active"`
}

// @Summary Update commission plan
// @Description Activates or deactivates a commission plan. Inactive plans cannot accrue new entries.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param id path integer true "Commission plan ID"
// @Param body body UpdateCommissionPlanRequest true "Plan update"
// @Success 200 {object} GetCommissionPlanResponseEnvelope "Commission plan updated successfully."
// @Failure 404 {object} httpx.ErrorResponse "Commission plan not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-plans/{id} [put]
func (h CommissionHandler) UpdatePlan(c fiber.Ctx) error {
	var request UpdateCommissionPlanRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	plan, err := h.svc.FindPlan(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("commission plan get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission plan.", err)
	}
	if plan == nil || !httpx.OwnsTenant(c, plan.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Commission plan not found.")
	}
	updated, err := h.svc.UpdatePlan(c, commission.UpdatePlanRequest{PlanID: plan.ID, Active: request.Active})
	if err != nil {
		return writeCommissionError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Commission plan updated successfully.", GetCommissionPlanResponse{
		CommissionPlan: newCommissionPlanResponse(updated),
	})
}

type CommissionRuleResponse struct {
	ID             uint64  `json:"id"`
	PlanID         uint64  `json:"plan_id"`
	ItemCategoryID *uint64 `json:"item_category_id"`
	MinAmount      float64 `json:"min_amount"`
	MaxAmount      float64 `json:"max_amount"`
	RatePct        float64 `json:"rate_pct"`
	FixedAmount    float64 `json:"fixed_amount"`
}

func newCommissionRuleResponse(rule *commission.CommissionRule) CommissionRuleResponse {
	return CommissionRuleResponse{
		ID: rule.ID, PlanID: rule.PlanID, ItemCategoryID: rule.ItemCategoryID,
		MinAmount: rule.MinAmount, MaxAmount: rule.MaxAmount, RatePct: rule.RatePct, FixedAmount: rule.FixedAmount,
	}
}

type ListCommissionRulesResponse struct {
	CommissionRules []CommissionRuleResponse `json:"commission_rules"`
}

type ListCommissionRulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCommissionRulesResponse `json:"data"`
}

// @Summary List commission rules
// @Description Lists the rules of a commission plan. The plan must belong to the caller's organization.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param id path integer true "Commission plan ID"
// @Success 200 {object} ListCommissionRulesResponseEnvelope "Commission rules retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Commission plan not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-plans/{id}/rules [get]
func (h CommissionHandler) ListRules(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	plan, err := h.svc.FindPlan(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("commission plan get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission plan.", err)
	}
	if plan == nil || !httpx.OwnsTenant(c, plan.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Commission plan not found.")
	}
	rules, err := h.svc.ListRulesByPlan(c, plan.ID)
	if err != nil {
		httpx.RequestLog(c).Error("commission rules list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission rules.", err)
	}
	items := make([]CommissionRuleResponse, len(rules))
	for i, rule := range rules {
		items[i] = newCommissionRuleResponse(rule)
	}
	return httpx.CreateSuccessResponse(c, "Commission rules retrieved successfully.", ListCommissionRulesResponse{
		CommissionRules: items,
	})
}

type CreateCommissionRuleRequest struct {
	ItemCategoryID *uint64 `json:"item_category_id"`
	MinAmount      float64 `json:"min_amount"`
	MaxAmount      float64 `json:"max_amount"`
	RatePct        float64 `json:"rate_pct"`
	FixedAmount    float64 `json:"fixed_amount"`
}

type CreateCommissionRuleResponse struct {
	CommissionRule CommissionRuleResponse `json:"commission_rule"`
}

type CreateCommissionRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateCommissionRuleResponse `json:"data"`
}

// @Summary Create commission rule
// @Description Adds a commission rule to a plan. A rule pays either a percentage of the base amount or a fixed amount, optionally limited by item category and amount range.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param id path integer true "Commission plan ID"
// @Param body body CreateCommissionRuleRequest true "Commission rule details"
// @Success 201 {object} CreateCommissionRuleResponseEnvelope "Commission rule created successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-plans/{id}/rules [post]
func (h CommissionHandler) CreateRule(c fiber.Ctx) error {
	var request CreateCommissionRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	plan, err := h.svc.FindPlan(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("commission plan get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission plan.", err)
	}
	if plan == nil || !httpx.OwnsTenant(c, plan.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Commission plan not found.")
	}
	created, err := h.svc.CreateRule(c, commission.CreateRuleRequest{
		PlanID:         plan.ID,
		ItemCategoryID: request.ItemCategoryID,
		MinAmount:      request.MinAmount,
		MaxAmount:      request.MaxAmount,
		RatePct:        request.RatePct,
		FixedAmount:    request.FixedAmount,
	})
	if err != nil {
		return writeCommissionError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Commission rule created successfully.", CreateCommissionRuleResponse{
		CommissionRule: newCommissionRuleResponse(created),
	})
}

type CommissionAssignmentResponse struct {
	ID            uint64     `json:"id"`
	PlanID        uint64     `json:"plan_id"`
	SalespersonID uint64     `json:"salesperson_id"`
	DateStart     *time.Time `json:"date_start"`
	DateEnd       *time.Time `json:"date_end"`
}

func newCommissionAssignmentResponse(a *commission.CommissionAssignment) CommissionAssignmentResponse {
	return CommissionAssignmentResponse{ID: a.ID, PlanID: a.PlanID, SalespersonID: a.SalespersonID, DateStart: a.DateStart, DateEnd: a.DateEnd}
}

type ListCommissionAssignmentsResponse struct {
	CommissionAssignments []CommissionAssignmentResponse `json:"commission_assignments"`
}

type ListCommissionAssignmentsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCommissionAssignmentsResponse `json:"data"`
}

// @Summary List commission assignments
// @Description Lists the salesperson assignments of a commission plan. The plan must belong to the caller's organization.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param id path integer true "Commission plan ID"
// @Success 200 {object} ListCommissionAssignmentsResponseEnvelope "Commission assignments retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Commission plan not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-plans/{id}/assignments [get]
func (h CommissionHandler) ListAssignments(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	plan, err := h.svc.FindPlan(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("commission plan get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission plan.", err)
	}
	if plan == nil || !httpx.OwnsTenant(c, plan.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Commission plan not found.")
	}
	assigns, err := h.svc.ListAssignments(c, &query.Query{Filters: []query.Filter{{Field: "plan_id", Operator: query.Equal, Value: plan.ID}}})
	if err != nil {
		httpx.RequestLog(c).Error("commission assignments list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission assignments.", err)
	}
	items := make([]CommissionAssignmentResponse, len(assigns.Items))
	for i, a := range assigns.Items {
		items[i] = newCommissionAssignmentResponse(a)
	}
	return httpx.CreateSuccessResponse(c, "Commission assignments retrieved successfully.", ListCommissionAssignmentsResponse{
		CommissionAssignments: items,
	})
}

type CreateCommissionAssignmentRequest struct {
	SalespersonID uint64  `json:"salesperson_id" validate:"required,gt=0"`
	DateStart     string  `json:"date_start" validate:"required"`
	DateEnd       *string `json:"date_end"`
}

type CreateCommissionAssignmentResponse struct {
	CommissionAssignment CommissionAssignmentResponse `json:"commission_assignment"`
}

type CreateCommissionAssignmentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateCommissionAssignmentResponse `json:"data"`
}

// @Summary Assign salesperson to plan
// @Description Assigns a salesperson to a commission plan for a period, rejecting overlapping assignments for the same salesperson.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param id path integer true "Commission plan ID"
// @Param body body CreateCommissionAssignmentRequest true "Assignment details"
// @Success 201 {object} CreateCommissionAssignmentResponseEnvelope "Commission assignment created successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-plans/{id}/assignments [post]
func (h CommissionHandler) CreateAssignment(c fiber.Ctx) error {
	var request CreateCommissionAssignmentRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	plan, err := h.svc.FindPlan(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("commission plan get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission plan.", err)
	}
	if plan == nil || !httpx.OwnsTenant(c, plan.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Commission plan not found.")
	}
	start, err := helper.ParseDateStr(request.DateStart)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date_start.", nil)
	}
	var end *time.Time
	if request.DateEnd != nil {
		parsed, err := helper.ParseDateStr(*request.DateEnd)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date_end.", nil)
		}
		end = &parsed
	}
	created, err := h.svc.Assign(c, commission.AssignRequest{
		PlanID:        plan.ID,
		SalespersonID: request.SalespersonID,
		DateStart:     start,
		DateEnd:       end,
	})
	if err != nil {
		return writeCommissionError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Commission assignment created successfully.", CreateCommissionAssignmentResponse{
		CommissionAssignment: newCommissionAssignmentResponse(created),
	})
}

type CommissionEntryResponse struct {
	ID               uint64  `json:"id"`
	SalespersonID    uint64  `json:"salesperson_id"`
	PlanID           uint64  `json:"plan_id"`
	SourceType       string  `json:"source_type"`
	SourceID         uint64  `json:"source_id"`
	BaseAmount       float64 `json:"base_amount"`
	CommissionAmount float64 `json:"commission_amount"`
	State            string  `json:"state"`
	PeriodID         *uint64 `json:"period_id"`
	PayslipID        *uint64 `json:"payslip_id"`
}

func newCommissionEntryResponse(e *commission.CommissionEntry) CommissionEntryResponse {
	return CommissionEntryResponse{
		ID: e.ID, SalespersonID: e.SalespersonID, PlanID: e.PlanID,
		SourceType: e.SourceType, SourceID: e.SourceID, BaseAmount: e.BaseAmount,
		CommissionAmount: e.CommissionAmount, State: e.State, PeriodID: e.PeriodID, PayslipID: e.PayslipID,
	}
}

var commissionEntryQueryAllowlist = map[string]struct{}{
	"salesperson_id": {},
	"plan_id":        {},
	"state":          {},
	"source_type":    {},
	"source_id":      {},
	"period_id":      {},
	"created_at":     {},
	"updated_at":     {},
}

type ListCommissionEntriesResponse struct {
	CommissionEntries []CommissionEntryResponse `json:"commission_entries"`
}

type ListCommissionEntriesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCommissionEntriesResponse `json:"data"`
}

// @Summary List commission entries
// @Description Lists accrued commission entries with pagination, sorting, and filtering.
// @Tags Commissions
// @Accept json
// @Produce json
// @Success 200 {object} ListCommissionEntriesResponseEnvelope "Commission entries retrieved successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-entries [get]
func (h CommissionHandler) ListEntries(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, commissionEntryQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	page, err := h.svc.ListEntries(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("commission entries list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve commission entries.", err)
	}
	items := make([]CommissionEntryResponse, len(page.Items))
	for i, entry := range page.Items {
		items[i] = newCommissionEntryResponse(entry)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Commission entries retrieved successfully.", ListCommissionEntriesResponse{
		CommissionEntries: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type AccrueCommissionRequest struct {
	SalespersonID    uint64  `json:"salesperson_id" validate:"required,gt=0"`
	PlanID           uint64  `json:"plan_id" validate:"required,gt=0"`
	SourceType       string  `json:"source_type" validate:"required"`
	SourceID         uint64  `json:"source_id" validate:"required,gt=0"`
	ItemCategoryID   *uint64 `json:"item_category_id"`
	BaseAmount       float64 `json:"base_amount" validate:"required,gt=0"`
	JournalID        uint64  `json:"journal_id" validate:"required,gt=0"`
	ExpenseAccountID uint64  `json:"expense_account_id" validate:"required,gt=0"`
	PayableAccountID uint64  `json:"payable_account_id" validate:"required,gt=0"`
	Date             string  `json:"date"`
	PeriodID         *uint64 `json:"period_id"`
}

type AccrueCommissionResponse struct {
	CommissionEntry CommissionEntryResponse `json:"commission_entry"`
}

type AccrueCommissionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data AccrueCommissionResponse `json:"data"`
}

// @Summary Accrue commission entry
// @Description Computes the commission from the plan's matching rule and posts Dr Commission Expense / Cr Commission Payable, recording a confirmed entry.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param body body AccrueCommissionRequest true "Accrual details"
// @Success 201 {object} AccrueCommissionResponseEnvelope "Commission entry accrued successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-entries/accrue [post]
func (h CommissionHandler) Accrue(c fiber.Ctx) error {
	var request AccrueCommissionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date := time.Time{}
	if request.Date != "" {
		parsed, err := helper.ParseDateStr(request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
		}
		date = parsed
	}
	created, err := h.svc.Accrue(c, commission.AccrueRequest{
		SalespersonID:    request.SalespersonID,
		PlanID:           request.PlanID,
		SourceType:       request.SourceType,
		SourceID:         request.SourceID,
		ItemCategoryID:   request.ItemCategoryID,
		BaseAmount:       request.BaseAmount,
		JournalID:        request.JournalID,
		ExpenseAccountID: request.ExpenseAccountID,
		PayableAccountID: request.PayableAccountID,
		Date:             date,
		PeriodID:         request.PeriodID,
	})
	if err != nil {
		return writeCommissionError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Commission entry accrued successfully.", AccrueCommissionResponse{
		CommissionEntry: newCommissionEntryResponse(created),
	})
}

type AccrueFromInvoiceCommissionRequest struct {
	InvoiceID        uint64  `json:"invoice_id" validate:"required,gt=0"`
	SalespersonID    uint64  `json:"salesperson_id" validate:"required,gt=0"`
	JournalID        uint64  `json:"journal_id" validate:"required,gt=0"`
	ExpenseAccountID uint64  `json:"expense_account_id" validate:"required,gt=0"`
	PayableAccountID uint64  `json:"payable_account_id" validate:"required,gt=0"`
	Date             string  `json:"date"`
	PeriodID         *uint64 `json:"period_id"`
}

type AccrueFromInvoiceCommissionResponse struct {
	CommissionEntry CommissionEntryResponse `json:"commission_entry"`
}

type AccrueFromInvoiceCommissionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data AccrueFromInvoiceCommissionResponse `json:"data"`
}

// @Summary Accrue commission from invoice
// @Description Derives the commission base from a live invoice per the plan basis (revenue, margin, or collected), then accrues Dr Commission Expense / Cr Commission Payable.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param body body AccrueFromInvoiceCommissionRequest true "Accrual details"
// @Success 201 {object} AccrueFromInvoiceCommissionResponseEnvelope "Commission entry accrued successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-entries/accrue-from-invoice [post]
func (h CommissionHandler) AccrueFromInvoice(c fiber.Ctx) error {
	var request AccrueFromInvoiceCommissionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date := time.Time{}
	if request.Date != "" {
		parsed, err := helper.ParseDateStr(request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
		}
		date = parsed
	}
	created, err := h.svc.AccrueFromInvoice(c, commission.AccrueFromInvoiceRequest{
		InvoiceID:        request.InvoiceID,
		SalespersonID:    request.SalespersonID,
		JournalID:        request.JournalID,
		ExpenseAccountID: request.ExpenseAccountID,
		PayableAccountID: request.PayableAccountID,
		Date:             date,
		PeriodID:         request.PeriodID,
	})
	if err != nil {
		return writeCommissionError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Commission entry accrued successfully.", AccrueFromInvoiceCommissionResponse{
		CommissionEntry: newCommissionEntryResponse(created),
	})
}

type PayCommissionRequest struct {
	PayslipID *uint64 `json:"payslip_id"`
}

type GetCommissionEntryResponse struct {
	CommissionEntry CommissionEntryResponse `json:"commission_entry"`
}

type GetCommissionEntryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetCommissionEntryResponse `json:"data"`
}

// @Summary Settle commission entry
// @Description Marks a confirmed commission entry as paid, optionally linking it to a payslip.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param id path integer true "Commission entry ID"
// @Param body body PayCommissionRequest true "Settlement details"
// @Success 200 {object} GetCommissionEntryResponseEnvelope "Commission entry paid successfully."
// @Failure 422 {object} httpx.ErrorResponse "Entry must be confirmed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-entries/{id}/pay [post]
func (h CommissionHandler) Pay(c fiber.Ctx) error {
	var request PayCommissionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	entryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	updated, err := h.svc.Pay(c, commission.PayRequest{EntryID: entryID, PayslipID: request.PayslipID})
	if err != nil {
		return writeCommissionError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Commission entry paid successfully.", GetCommissionEntryResponse{
		CommissionEntry: newCommissionEntryResponse(updated),
	})
}

// @Summary Cancel commission entry
// @Description Cancels a draft or confirmed commission entry.
// @Tags Commissions
// @Accept json
// @Produce json
// @Param id path integer true "Commission entry ID"
// @Success 200 {object} GetCommissionEntryResponseEnvelope "Commission entry cancelled successfully."
// @Failure 422 {object} httpx.ErrorResponse "Entry is not open"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/commission-entries/{id}/cancel [post]
func (h CommissionHandler) Cancel(c fiber.Ctx) error {
	entryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	updated, err := h.svc.Cancel(c, entryID)
	if err != nil {
		return writeCommissionError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Commission entry cancelled successfully.", GetCommissionEntryResponse{
		CommissionEntry: newCommissionEntryResponse(updated),
	})
}

func writeCommissionError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, commission.ErrPlanNotFound),
		errors.Is(err, commission.ErrRuleNotFound),
		errors.Is(err, commission.ErrAssignmentNotFound),
		errors.Is(err, commission.ErrEntryNotFound),
		errors.Is(err, commission.ErrInvoiceNotFound):
		return httpx.CreateNotFoundResponse(c, "Resource not found.")
	case errors.Is(err, commission.ErrPlanInactive),
		errors.Is(err, commission.ErrPlanInvalidBasis),
		errors.Is(err, commission.ErrRuleInvalid),
		errors.Is(err, commission.ErrRuleRange),
		errors.Is(err, commission.ErrAssignmentOverlap),
		errors.Is(err, commission.ErrEntryNotConfirmed),
		errors.Is(err, commission.ErrEntryNotDraft),
		errors.Is(err, commission.ErrEntryNotOpen),
		errors.Is(err, commission.ErrNoRuleMatches),
		errors.Is(err, commission.ErrAccountsRequired),
		errors.Is(err, commission.ErrSourceMissing),
		errors.Is(err, commission.ErrEntryDuplicate):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Request cannot be processed.", nil)
	default:
		httpx.RequestLog(c).Error("commission write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process request.", err)
	}
}
