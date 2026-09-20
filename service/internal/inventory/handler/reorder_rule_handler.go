package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

// @Summary List reorder rules
// @Description Lists reorder rules with pagination, sorting, and filtering over the allowlisted query fields, such as the item, warehouse, location, and active flag. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Reorder Rules
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListReorderRulesResponseEnvelope "Reorder rules retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /reorder-rules [get]
func (h ReorderHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, reorderRuleQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	page, err := h.svc.ListInOrg(c, parsedQuery, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("reorder rule list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve reorder rules.", err)
	}

	items := make([]ReorderRuleResponse, len(page.Items))
	for i, rule := range page.Items {
		items[i] = newReorderRuleResponse(rule)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "reorder-rules.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Reorder rules retrieved successfully.", ListReorderRulesResponse{
		Rules: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get reorder rule
// @Description Gets a single reorder rule by its id, including its minimum and maximum thresholds, quantity multiple, lead time, and active flag. A missing rule is answered with 404 Not Found. A non-numeric or malformed id is rejected with 422 Unprocessable Entity.
// @Tags Reorder Rules
// @Accept json
// @Produce json
// @Param id path integer true "Reorder rule ID"
// @Success 200 {object} GetReorderRuleResponseEnvelope "Reorder rule retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Reorder rule not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /reorder-rules/{id} [get]
func (h ReorderHandler) Get(c fiber.Ctx) error {
	ruleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid reorder rule id provided.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	rule, err := h.svc.FindInOrg(c, ruleID, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("reorder rule lookup failed", "rule_id", ruleID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get reorder rule.", err)
	}
	if rule == nil {
		return httpx.CreateNotFoundResponse(c, "Reorder rule not found.")
	}

	return httpx.CreateSuccessResponse(c, "Reorder rule retrieved successfully.", GetReorderRuleResponse{
		Rule: newReorderRuleResponse(rule),
	})
}

// @Summary Create reorder rule
// @Description Creates a reorder rule for a item at a stock location, defaulting the rule to active when not specified. The minimum quantity must not be negative, the maximum must not be below the minimum, the quantity multiple must be positive, and a stock location is required. Returns the created rule with a 201 status.
// @Tags Reorder Rules
// @Accept json
// @Produce json
// @Param body body ReorderRuleRequest true "Reorder rule details"
// @Success 201 {object} ReorderRuleResponseBodyEnvelope "Reorder rule created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /reorder-rules [post]
func (h ReorderHandler) Create(c fiber.Ctx) error {
	var request ReorderRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	active := true
	if request.Active != nil {
		active = *request.Active
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	rule, err := h.svc.Create(c, *organizationID, &inventory.ReorderRule{
		ItemID:       request.ItemID,
		WarehouseID:  request.WarehouseID,
		LocationID:   request.LocationID,
		MinQty:       request.MinQty,
		MaxQty:       request.MaxQty,
		QtyMultiple:  request.QtyMultiple,
		LeadTimeDays: request.LeadTimeDays,
		Active:       active,
	})
	if err != nil {
		return writeReorderError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Reorder rule created successfully.", ReorderRuleResponseBody{
		Rule: newReorderRuleResponse(rule),
	})
}

// @Summary Update reorder rule
// @Description Updates an existing reorder rule by its id, returning 404 Not Found if the rule does not exist. The quantity thresholds, quantity multiple, and required stock location are re-validated before saving. Returns the updated rule.
// @Tags Reorder Rules
// @Accept json
// @Produce json
// @Param id path integer true "Reorder rule ID"
// @Param body body ReorderRuleRequest true "Reorder rule details"
// @Success 200 {object} ReorderRuleResponseBodyEnvelope "Reorder rule updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Reorder rule not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /reorder-rules/{id} [put]
func (h ReorderHandler) Update(c fiber.Ctx) error {
	ruleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid reorder rule id provided.", nil)
	}

	var request ReorderRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	rule, err := h.svc.FindInOrg(c, ruleID, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("reorder rule lookup failed", "rule_id", ruleID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update reorder rule.", err)
	}
	if rule == nil {
		return httpx.CreateNotFoundResponse(c, "Reorder rule not found.")
	}

	rule.ItemID = request.ItemID
	rule.WarehouseID = request.WarehouseID
	rule.LocationID = request.LocationID
	rule.MinQty = request.MinQty
	rule.MaxQty = request.MaxQty
	rule.QtyMultiple = request.QtyMultiple
	rule.LeadTimeDays = request.LeadTimeDays
	if request.Active != nil {
		rule.Active = *request.Active
	}

	updated, err := h.svc.Update(c, *organizationID, rule)
	if err != nil {
		return writeReorderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Reorder rule updated successfully.", ReorderRuleResponseBody{
		Rule: newReorderRuleResponse(updated),
	})
}

// @Summary Delete reorder rule
// @Description Deletes a reorder rule by its id after confirming it exists, otherwise a 404 Not Found is returned. Responds with 204 No Content on success.
// @Tags Reorder Rules
// @Accept json
// @Produce json
// @Param id path integer true "Reorder rule ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Reorder rule not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /reorder-rules/{id} [delete]
func (h ReorderHandler) Delete(c fiber.Ctx) error {
	ruleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid reorder rule id provided.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	rule, err := h.svc.FindInOrg(c, ruleID, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("reorder rule lookup failed", "rule_id", ruleID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete reorder rule.", err)
	}
	if rule == nil {
		return httpx.CreateNotFoundResponse(c, "Reorder rule not found.")
	}

	if err := h.svc.Delete(c, ruleID); err != nil {
		httpx.RequestLog(c).Error("reorder rule deletion failed", "rule_id", ruleID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete reorder rule.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
