package handler

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type GetMeResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetMeResponse `json:"data"`
}
type GetMeResponse struct {
	User Profile `json:"user"`
}

func (h UserHandler) GetMe(c fiber.Ctx) error {
	userID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingUserInContext)
	}

	user, err := h.userSvc.Find(c, userID)
	if err != nil {
		httpx.RequestLog(c).Error("failed to load current user profile", "user_id", userID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve profile.", err)
	}

	if user == nil || user.ID == 0 {
		return httpx.CreateNotFoundResponse(c, "User not found.")
	}

	httpx.RequestLog(c).Info("current user profile retrieved", "user_id", userID)

	return httpx.CreateSuccessResponse(c, "Profile retrieved successfully.", GetMeResponse{
		User: newProfile(user),
	})
}

type UpdateMeAvatarResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateMeAvatarResponse `json:"data"`
}
type UpdateMeAvatarResponse struct {
	Avatar string `json:"avatar"`
}

func (h UserHandler) UpdateMeAvatar(c fiber.Ctx) error {
	userID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingUserInContext)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A file field is required.", nil)
	}
	src, err := file.Open()
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Failed to read uploaded file.", nil)
	}
	defer func() { _ = src.Close() }()

	avatarID, err := h.avatarSvc.Upload(c, userID, src)
	if err != nil {
		switch {
		case errors.Is(err, iam.ErrAvatarEmpty):
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Avatar content cannot be empty.", nil)
		case errors.Is(err, iam.ErrAvatarInvalidType):
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Avatar must be an image.", nil)
		case errors.Is(err, iam.ErrUserNotFound):
			return httpx.CreateNotFoundResponse(c, "User not found.")
		default:
			httpx.RequestLog(c).Error("avatar upload failed", "user_id", userID, "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to update avatar.", err)
		}
	}

	httpx.RequestLog(c).Info("avatar updated", "user_id", userID)

	return httpx.CreateSuccessResponse(c, "Avatar updated successfully.", UpdateMeAvatarResponse{
		Avatar: avatarID,
	})
}

func (h UserHandler) GetMeAvatar(c fiber.Ctx) error {
	userID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingUserInContext)
	}
	return h.serveAvatar(c, userID)
}

func (h UserHandler) serveAvatar(c fiber.Ctx, userID uint64) error {
	reader, err := h.avatarSvc.Download(c, userID)
	if err != nil {
		switch {
		case errors.Is(err, iam.ErrAvatarNotFound):
			return httpx.CreateNotFoundResponse(c, "Avatar not found.")
		default:
			httpx.RequestLog(c).Error("avatar download failed", "user_id", userID, "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve avatar.", err)
		}
	}
	defer func() { _ = reader.Close() }()

	buffered := bufio.NewReader(reader)
	head, err := buffered.Peek(512)
	if err != nil && !errors.Is(err, io.EOF) {
		httpx.RequestLog(c).Error("avatar download failed", "user_id", userID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve avatar.", err)
	}
	c.Set(fiber.HeaderContentType, http.DetectContentType(head))
	c.Set(fiber.HeaderCacheControl, "private, max-age=300")
	return c.SendStream(buffered)
}

type MeOrganizationsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data MeOrganizationsResponse `json:"data"`
}
type MeOrganizationsResponse struct {
	Organizations []*reference.Organization `json:"organizations"`
}

func (h UserHandler) MeOrganizations(c fiber.Ctx) error {
	userID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingUserInContext)
	}

	orgs, err := h.memberSvc.ListOrganizationsByUser(c, userID)
	if err != nil {
		httpx.RequestLog(c).Error("failed to list current user organizations", "user_id", userID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve organizations.", err)
	}

	httpx.RequestLog(c).Info("current user organizations retrieved", "user_id", userID)

	return httpx.CreateSuccessResponse(c, "Organizations retrieved successfully.", MeOrganizationsResponse{
		Organizations: orgs,
	})
}

type MeOrganizationPermissionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data MeOrganizationPermissionsResponse `json:"data"`
}
type MeOrganizationPermissionsResponse struct {
	Permissions []*iam.Permission `json:"permissions"`
}

func (h UserHandler) MeOrganizationPermissions(c fiber.Ctx) error {
	userID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingUserInContext)
	}

	organizationID, err := strconv.ParseUint(c.Params("organization_id"), 10, 64)
	if err != nil || organizationID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid organization_id.", nil)
	}

	member, err := h.memberSvc.FindByUserAndOrganization(c, userID, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("organization membership lookup failed", "user_id", userID, "organization_id", organizationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve permissions.", err)
	}
	if member == nil {
		return httpx.CreateNotFoundResponse(c, "Organization not found.")
	}

	permissions, err := h.authzSvc.EffectiveOrganizationPermissions(c, userID, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("effective permissions failed", "user_id", userID, "organization_id", organizationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve permissions.", err)
	}

	httpx.RequestLog(c).Info("current user organization permissions retrieved", "user_id", userID, "organization_id", organizationID)

	return httpx.CreateSuccessResponse(c, "Permissions retrieved successfully.", MeOrganizationPermissionsResponse{
		Permissions: permissions,
	})
}
