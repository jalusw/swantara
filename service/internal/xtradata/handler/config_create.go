package handler

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CreateSystemConfigRequest struct {
	Key   string          `json:"key" validate:"required"`
	Value json.RawMessage `json:"value" validate:"required" swaggertype:"object"`
}

type CreateSystemConfigResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateSystemConfigResponse `json:"data"`
}

type CreateSystemConfigResponse struct {
	SystemConfig SystemConfigResponse `json:"system_config"`
}

// @Summary Create system config
// @Description Creates a system config with a unique key and a JSON value.
// @Tags System Configs
// @Accept json
// @Produce json
// @Param body body CreateSystemConfigRequest true "System config details"
// @Success 201 {object} CreateSystemConfigResponseEnvelope "System config created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 409 {object} httpx.ErrorResponse "System config key already exists"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/system-configs [post]
func (h SystemConfigHandler) Create(c fiber.Ctx) error {
	var request CreateSystemConfigRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	config, err := h.configSvc.Create(c, &reference.SystemConfig{
		Key:            request.Key,
		Value:          request.Value,
		OrganizationID: httpx.TenantOrganizationID(c, nil),
	})
	if err != nil {
		return writeSystemConfigError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "System config created successfully.", CreateSystemConfigResponse{
		SystemConfig: newSystemConfigResponse(config),
	})
}
