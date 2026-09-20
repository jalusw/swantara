package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateSystemConfigRequest struct {
	Value json.RawMessage `json:"value" validate:"required" swaggertype:"object"`
}

type UpdateSystemConfigResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateSystemConfigResponse `json:"data"`
}

type UpdateSystemConfigResponse struct {
	SystemConfig SystemConfigResponse `json:"system_config"`
}

// @Summary Update system config
// @Description Updates a system config's JSON value by id; a non-numeric id yields a 422.
// @Tags System Configs
// @Accept json
// @Produce json
// @Param id path integer true "System config ID"
// @Param body body UpdateSystemConfigRequest true "System config value"
// @Success 200 {object} UpdateSystemConfigResponseEnvelope "System config updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "System config not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or invalid id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/system-configs/{id} [put]
func (h SystemConfigHandler) Update(c fiber.Ctx) error {
	configID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid system config id provided.", nil)
	}

	var request UpdateSystemConfigRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	config, err := h.configSvc.Find(c, configID)
	if err != nil {
		httpx.RequestLog(c).Error("system config lookup failed", "config_id", configID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update system config.", err)
	}
	if config == nil {
		return httpx.CreateNotFoundResponse(c, "System config not found.")
	}

	config.Value = request.Value

	updated, err := h.configSvc.Update(c, config)
	if err != nil {
		return writeSystemConfigError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "System config updated successfully.", UpdateSystemConfigResponse{
		SystemConfig: newSystemConfigResponse(updated),
	})
}
