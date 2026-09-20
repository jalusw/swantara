package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type ListPermissionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPermissionsResponse `json:"data"`
}
type ListPermissionsResponse struct {
	Permissions []*iam.Permission `json:"permissions"`
}

// @Summary List permissions
// @Description Lists the full permission catalog available to bind to organization roles, including each permission's resource, action, name, and code.
// @Tags Members
// @Accept json
// @Produce json
// @Success 200 {object} ListPermissionsResponseEnvelope "Permissions retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/permissions [get]
func (h MemberHandler) ListPermissions(c fiber.Ctx) error {
	page, err := h.memberSvc.ListPermissions(c, nil)
	if err != nil {
		httpx.RequestLog(c).Error("permission list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve permissions.", err)
	}

	return httpx.CreateSuccessResponse(c, "Permissions retrieved successfully.", ListPermissionsResponse{
		Permissions: page.Items,
	})
}
