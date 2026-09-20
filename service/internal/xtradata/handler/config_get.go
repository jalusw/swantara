package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetSystemConfigResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetSystemConfigResponse `json:"data"`
}

type GetSystemConfigResponse struct {
	SystemConfig SystemConfigResponse `json:"system_config"`
}

// @Summary Get system config
// @Description Returns a single system config by id; a non-numeric id yields a 422.
// @Tags System Configs
// @Accept json
// @Produce json
// @Param id path integer true "System config ID"
// @Success 200 {object} GetSystemConfigResponseEnvelope "System config retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "System config not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/system-configs/{id} [get]
func (h SystemConfigHandler) Get(c fiber.Ctx) error {
	configID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid system config id provided.", nil)
	}

	config, err := h.configSvc.Find(c, configID)
	if err != nil {
		httpx.RequestLog(c).Error("system config lookup failed", "config_id", configID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve system config.", err)
	}
	if config == nil {
		return httpx.CreateNotFoundResponse(c, "System config not found.")
	}

	return httpx.CreateSuccessResponse(c, "System config retrieved successfully.", GetSystemConfigResponse{
		SystemConfig: newSystemConfigResponse(config),
	})
}
