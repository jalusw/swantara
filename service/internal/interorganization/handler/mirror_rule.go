package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
)

type InterorganizationRuleResponse struct {
	ID                 uint64  `json:"id"`
	FromOrganizationID *uint64 `json:"from_organization_id"`
	ToOrganizationID   *uint64 `json:"to_organization_id"`
	AutoMirror         bool    `json:"auto_mirror"`
	SupplierContactID  *uint64 `json:"supplier_contact_id"`
	CustomerContactID  *uint64 `json:"customer_contact_id"`
}

func newInterorganizationRuleResponse(rule *interorganization.InterorganizationRule) InterorganizationRuleResponse {
	return InterorganizationRuleResponse{
		ID: rule.ID, FromOrganizationID: rule.FromOrganizationID, ToOrganizationID: rule.ToOrganizationID,
		AutoMirror: rule.AutoMirror, SupplierContactID: rule.SupplierContactID, CustomerContactID: rule.CustomerContactID,
	}
}

type UpsertInterorganizationRuleRequest struct {
	FromOrganizationID *uint64 `json:"from_organization_id"`
	ToOrganizationID   *uint64 `json:"to_organization_id"`
	AutoMirror         *bool   `json:"auto_mirror"`
	SupplierContactID  *uint64 `json:"supplier_contact_id"`
	CustomerContactID  *uint64 `json:"customer_contact_id"`
}

type CreateInterorganizationRuleResponse struct {
	InterorganizationRule InterorganizationRuleResponse `json:"interorganization_rule"`
}

type CreateInterorganizationRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateInterorganizationRuleResponse `json:"data"`
}

// @Summary Create interorganization rule
// @Description Defines a mirroring rule between two organizations, including the supplier and customer contact used to represent each side.
// @Tags Inter-organization
// @Accept json
// @Produce json
// @Param body body UpsertInterorganizationRuleRequest true "Rule details"
// @Success 201 {object} CreateInterorganizationRuleResponseEnvelope "Interorganization rule created successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/interorganization-rules [post]
func (h InterorganizationHandler) CreateRule(c fiber.Ctx) error {
	var request UpsertInterorganizationRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.FromOrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "From organization ID is required.", nil)
	}
	autoMirror := true
	if request.AutoMirror != nil {
		autoMirror = *request.AutoMirror
	}
	created, err := h.svc.CreateRule(c, interorganization.UpsertInterorganizationRuleRequest{
		FromOrganizationID: organizationID,
		ToOrganizationID:   request.ToOrganizationID,
		AutoMirror:         autoMirror,
		SupplierContactID:  request.SupplierContactID,
		CustomerContactID:  request.CustomerContactID,
	})
	if err != nil {
		return writeInterorganizationError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Interorganization rule created successfully.", CreateInterorganizationRuleResponse{
		InterorganizationRule: newInterorganizationRuleResponse(created),
	})
}

type ListInterorganizationRulesResponse struct {
	InterorganizationRules []InterorganizationRuleResponse `json:"interorganization_rules"`
}

type ListInterorganizationRulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListInterorganizationRulesResponse `json:"data"`
}

// @Summary List interorganization rules
// @Description Lists the outgoing mirroring rules of the caller's organization.
// @Tags Inter-organization
// @Accept json
// @Produce json
// @Success 200 {object} ListInterorganizationRulesResponseEnvelope "Interorganization rules retrieved successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/interorganization-rules [get]
func (h InterorganizationHandler) ListRules(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	rules, err := h.svc.ListRules(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("interorganization rules list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve interorganization rules.", err)
	}
	items := make([]InterorganizationRuleResponse, len(rules))
	for i, rule := range rules {
		items[i] = newInterorganizationRuleResponse(rule)
	}
	return httpx.CreateSuccessResponse(c, "Interorganization rules retrieved successfully.", ListInterorganizationRulesResponse{
		InterorganizationRules: items,
	})
}

type GetInterorganizationRuleResponse struct {
	InterorganizationRule InterorganizationRuleResponse `json:"interorganization_rule"`
}

type GetInterorganizationRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetInterorganizationRuleResponse `json:"data"`
}

// @Summary Update interorganization rule
// @Description Updates an outgoing mirroring rule of the caller's organization.
// @Tags Inter-organization
// @Accept json
// @Produce json
// @Param id path integer true "Interorganization rule ID"
// @Param body body UpsertInterorganizationRuleRequest true "Rule update"
// @Success 200 {object} GetInterorganizationRuleResponseEnvelope "Interorganization rule updated successfully."
// @Failure 404 {object} httpx.ErrorResponse "Interorganization rule not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/interorganization-rules/{id} [put]
func (h InterorganizationHandler) UpdateRule(c fiber.Ctx) error {
	var request UpsertInterorganizationRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	ruleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	existing, err := h.svc.FindRule(c, ruleID)
	if err != nil {
		httpx.RequestLog(c).Error("interorganization rule get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve interorganization rule.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.FromOrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Interorganization rule not found.")
	}
	autoMirror := existing.AutoMirror
	if request.AutoMirror != nil {
		autoMirror = *request.AutoMirror
	}
	updated, err := h.svc.UpdateRule(c, interorganization.UpsertInterorganizationRuleRequest{
		ID:                 &ruleID,
		FromOrganizationID: request.FromOrganizationID,
		ToOrganizationID:   request.ToOrganizationID,
		AutoMirror:         autoMirror,
		SupplierContactID:  request.SupplierContactID,
		CustomerContactID:  request.CustomerContactID,
	})
	if err != nil {
		return writeInterorganizationError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Interorganization rule updated successfully.", GetInterorganizationRuleResponse{
		InterorganizationRule: newInterorganizationRuleResponse(updated),
	})
}

// @Summary Delete interorganization rule
// @Description Deletes an outgoing mirroring rule of the caller's organization.
// @Tags Inter-organization
// @Accept json
// @Produce json
// @Param id path integer true "Interorganization rule ID"
// @Success 200 {object} httpx.EnvelopeBase "Interorganization rule deleted successfully."
// @Failure 404 {object} httpx.ErrorResponse "Interorganization rule not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/interorganization-rules/{id} [delete]
func (h InterorganizationHandler) DeleteRule(c fiber.Ctx) error {
	ruleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	existing, err := h.svc.FindRule(c, ruleID)
	if err != nil {
		httpx.RequestLog(c).Error("interorganization rule get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve interorganization rule.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.FromOrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Interorganization rule not found.")
	}
	if err := h.svc.DeleteRule(c, ruleID); err != nil {
		return writeInterorganizationError(c, err)
	}
	return httpx.CreateNoContentResponse(c)
}
