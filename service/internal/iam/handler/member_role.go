package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type ListMemberRolesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListMemberRolesResponse `json:"data"`
}
type ListMemberRolesResponse struct {
	Roles []*iam.MemberRole `json:"member_roles"`
}

// @Summary List member roles
// @Description Lists the member roles defined for the caller's organization together with the permissions bound to each role, scoped to the caller's organization.
// @Tags Members
// @Accept json
// @Produce json
// @Success 200 {object} ListMemberRolesResponseEnvelope "Member roles retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/member-roles [get]
func (h MemberHandler) ListMemberRoles(c fiber.Ctx) error {
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingOrganizationInContext)
	}

	roles, err := h.memberSvc.ListRolesByOrganization(c, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("member role list failed", "organization_id", organizationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve member roles.", err)
	}

	return httpx.CreateSuccessResponse(c, "Member roles retrieved successfully.", ListMemberRolesResponse{
		Roles: roles,
	})
}

type CreateMemberRoleRequest struct {
	Name          string   `json:"name" validate:"required"`
	Code          string   `json:"code" validate:"required"`
	Description   *string  `json:"description"`
	PermissionIDs []uint64 `json:"permission_ids"`
}

type CreateMemberRoleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateMemberRoleResponse `json:"data"`
}
type CreateMemberRoleResponse struct {
	Role *iam.MemberRole `json:"member_role"`
}

// @Summary Create member role
// @Description Creates a new member role for the organization with the given code, name, and description, and binds it to the supplied permission ids. Only the organization owner can create roles, and duplicate role codes are rejected with 409.
// @Tags Members
// @Accept json
// @Produce json
// @Param body body CreateMemberRoleRequest true "Member role details"
// @Success 201 {object} CreateMemberRoleResponseEnvelope "Member role created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 409 {object} httpx.ErrorResponse "Member role with this code already exists"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/member-roles [post]
func (h MemberHandler) CreateMemberRole(c fiber.Ctx) error {
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingOrganizationInContext)
	}

	var request CreateMemberRoleRequest
	if !bindAndValidate(c, &request) {
		return nil
	}

	role, err := h.memberSvc.CreateRole(c, organizationID, &iam.MemberRole{
		Name:        request.Name,
		Code:        request.Code,
		Description: request.Description,
	}, request.PermissionIDs)
	if err != nil {
		switch {
		case errors.Is(err, iam.ErrMemberRoleExists):
			return httpx.CreateConflictResponse(c, "Member role with this code already exists.", err)
		default:
			httpx.RequestLog(c).Error("member role creation failed", "organization_id", organizationID, "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to create member role.", err)
		}
	}

	httpx.RequestLog(c).Info("member role created", "organization_id", organizationID, "role_id", role.ID)

	return httpx.CreateCreatedResponse(c, "Member role created successfully.", CreateMemberRoleResponse{
		Role: role,
	})
}
