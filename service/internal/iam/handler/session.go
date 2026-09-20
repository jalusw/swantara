package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type GetSessionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetSessionsResponse `json:"data"`
}
type GetSessionsResponse struct {
	Sessions []*iam.UserSession `json:"sessions"`
}

// @Summary List sessions
// @Description Lists the authenticated user's active sessions with client device, IP, and expiry information, so individual sessions can be reviewed and managed, for example by revoking a specific device.
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} GetSessionsResponseEnvelope "Sessions retrieved."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /auth/sessions [get]
func (h AuthHandler) GetSessionsHandler(c fiber.Ctx) error {
	userID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingUserInContext)
	}

	sessions, err := h.authSvc.ListUserSessions(c, userID)
	if err != nil {
		httpx.RequestLog(c).Error("user session list failed", "user_id", userID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve sessions.", err)
	}

	return httpx.CreateSuccessResponse(c, "Sessions retrieved.", GetSessionsResponse{
		Sessions: sessions,
	})
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// @Summary Log out user
// @Description Logs out the user by revoking the session associated with the given refresh token, invalidating it server-side so it can no longer be used. Returns 204 regardless of whether the token was still valid.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body LogoutRequest true "Refresh token"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h AuthHandler) LogoutHandler(c fiber.Ctx) error {
	var request LogoutRequest

	if !bindAndValidate(c, &request) {
		return nil
	}

	if err := h.authSvc.Logout(c, request.RefreshToken); err != nil {
		httpx.RequestLog(c).Error("logout failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "", err)
	}

	httpx.RequestLog(c).Info("user logged out")

	return httpx.CreateNoContentResponse(c)
}

// @Summary Revoke session
// @Description Revokes one of the authenticated user's sessions by id, immediately invalidating that device's tokens. Sessions not owned by the caller are rejected with 401, and unknown sessions return 404.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param session_id path integer true "Session ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Invalid session id"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Session not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /auth/sessions/{session_id} [delete]
func (h AuthHandler) RevokeSessionHandler(c fiber.Ctx) error {
	sessionID, err := strconv.ParseUint(c.Params("session_id"), 10, 64)
	if err != nil {
		return httpx.CreateBadRequestResponse(c, "", httpx.ErrSessionIDRequired)
	}

	userID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingUserInContext)
	}

	if err := h.authSvc.RevokeSession(c, userID, sessionID); err != nil {
		switch {
		case errors.Is(err, iam.ErrSessionNotFound):
			return httpx.CreateNotFoundResponse(c, "Session not found.")
		case errors.Is(err, iam.ErrSessionNotOwnedByUser):
			httpx.RequestLog(c).Warn("session revoke denied", "user_id", userID, "session_id", sessionID)
			return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
		default:
			httpx.RequestLog(c).Error("session revoke failed", "user_id", userID, "session_id", sessionID, "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "", err)
		}
	}

	httpx.RequestLog(c).Info("session revoked", "user_id", userID, "session_id", sessionID)

	return httpx.CreateNoContentResponse(c)
}
