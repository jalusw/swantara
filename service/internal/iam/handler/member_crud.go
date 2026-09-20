package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type ListMembersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListMembersResponse `json:"data"`
}
type ListMembersResponse struct {
	Members []*iam.Member `json:"members"`
}

// @Summary List organization members
// @Description Lists the members of the caller's organization together with the member roles assigned to each, scoped to the organization resolved from the caller's auth context.
// @Tags Members
// @Accept json
// @Produce json
// @Success 200 {object} ListMembersResponseEnvelope "Members retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/members [get]
func (h MemberHandler) ListMembers(c fiber.Ctx) error {
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingOrganizationInContext)
	}

	members, err := h.memberSvc.ListByOrganization(c, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("member list failed", "organization_id", organizationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve members.", err)
	}

	return httpx.CreateSuccessResponse(c, "Members retrieved successfully.", ListMembersResponse{
		Members: members,
	})
}

type InviteMemberRequest struct {
	UserID   uint64  `json:"user_id" validate:"required"`
	Position *string `json:"position"`
}

type InviteMemberResponseEnvelope struct {
	httpx.EnvelopeBase
	Data InviteMemberResponse `json:"data"`
}
type InviteMemberResponse struct {
	Member *iam.Member `json:"member"`
}

// @Summary Invite member
// @Description Adds an existing user to the caller's organization with the default member role. Only the organization owner can invite members; users already in the organization are rejected with 409 and a missing default role with 422.
// @Tags Members
// @Accept json
// @Produce json
// @Param body body InviteMemberRequest true "User to invite"
// @Success 201 {object} InviteMemberResponseEnvelope "Member invited successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 409 {object} httpx.ErrorResponse "User is already a member"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or default member role not configured"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/members [post]
func (h MemberHandler) InviteMember(c fiber.Ctx) error {
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingOrganizationInContext)
	}

	var request InviteMemberRequest
	if !bindAndValidate(c, &request) {
		return nil
	}

	member, err := h.memberSvc.Invite(c, organizationID, request.UserID)
	if err != nil {
		switch {
		case errors.Is(err, iam.ErrAlreadyMember):
			return httpx.CreateConflictResponse(c, "User is already a member.", err)
		case errors.Is(err, iam.ErrMemberRoleNotFound):
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Default member role is not configured for this organization.", nil)
		default:
			httpx.RequestLog(c).Error("member invite failed", "organization_id", organizationID, "user_id", request.UserID, "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to invite member.", err)
		}
	}

	httpx.RequestLog(c).Info("member invited", "organization_id", organizationID, "user_id", request.UserID)

	return httpx.CreateCreatedResponse(c, "Member invited successfully.", InviteMemberResponse{
		Member: member,
	})
}

// @Summary Remove member
// @Description Permanently removes a member from the organization, revoking their access. Only the organization owner can remove members, and unknown members return 404.
// @Tags Members
// @Accept json
// @Produce json
// @Param id path integer true "Member ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 404 {object} httpx.ErrorResponse "Member not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid member id"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/members/{id} [delete]
func (h MemberHandler) RemoveMember(c fiber.Ctx) error {
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingOrganizationInContext)
	}

	memberID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid member id provided.", nil)
	}

	if err := h.memberSvc.RemoveMember(c, organizationID, memberID); err != nil {
		switch {
		case errors.Is(err, iam.ErrMemberNotFound):
			return httpx.CreateNotFoundResponse(c, "Member not found.")
		default:
			httpx.RequestLog(c).Error("member removal failed", "organization_id", organizationID, "member_id", memberID, "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to remove member.", err)
		}
	}

	httpx.RequestLog(c).Info("member removed", "organization_id", organizationID, "member_id", memberID)

	return httpx.CreateNoContentResponse(c)
}
