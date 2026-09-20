package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/organization"
)

type ModuleResponse struct {
	ModuleID string `json:"module_id"`
	Active   bool   `json:"active"`
}

type ListModulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListModulesResponse `json:"data"`
}
type ListModulesResponse struct {
	Modules []ModuleResponse `json:"modules"`
}

type UpdateModuleRequest struct {
	ModuleID string `json:"module_id" validate:"required"`
	Active   *bool  `json:"active" validate:"required"`
}

type UpdateModuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateModuleResponse `json:"data"`
}
type UpdateModuleResponse struct {
	Module ModuleResponse `json:"module"`
}

// @Summary List organization modules
// @Description Lists every available module for an organization with its active flag. Modules without a stored row default to active.
// @Tags Organizations
// @Produce json
// @Param id path int true "Organization ID"
// @Success 200 {object} ListModulesResponseEnvelope "Modules retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 404 {object} httpx.ErrorResponse "Organization not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{id}/modules [get]
func (h OrganizationHandler) ListModules(c fiber.Ctx) error {
	org, ok := h.findOrganization(c)
	if !ok {
		return nil
	}

	states, err := h.svc.ListModules(c.Context(), org.ID)
	if err != nil {
		httpx.RequestLog(c).Error("organization modules list failed", "organization_id", org.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get modules.", err)
	}

	modules := make([]ModuleResponse, 0, len(states))
	for _, state := range states {
		modules = append(modules, ModuleResponse{ModuleID: state.ModuleID, Active: state.Active})
	}
	return httpx.CreateSuccessResponse(c, "Modules retrieved successfully.", ListModulesResponse{Modules: modules})
}

// @Summary Update organization module
// @Description Activates or deactivates a module for an organization.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param id path int true "Organization ID"
// @Param body body UpdateModuleRequest true "Module activation"
// @Success 200 {object} UpdateModuleResponseEnvelope "Module updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 404 {object} httpx.ErrorResponse "Organization not found"
// @Failure 422 {object} httpx.ErrorResponse "Unknown module"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{id}/modules [put]
func (h OrganizationHandler) UpdateModule(c fiber.Ctx) error {
	org, ok := h.findOrganization(c)
	if !ok {
		return nil
	}

	var request UpdateModuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	if err := h.svc.SetModuleActive(c.Context(), org.ID, request.ModuleID, *request.Active); err != nil {
		if errors.Is(err, organization.ErrModuleUnknown) {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Unknown module.", nil)
		}
		httpx.RequestLog(c).Error("organization module update failed", "organization_id", org.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update module.", err)
	}

	return httpx.CreateSuccessResponse(c, "Module updated successfully.", UpdateModuleResponse{
		Module: ModuleResponse{ModuleID: request.ModuleID, Active: *request.Active},
	})
}
