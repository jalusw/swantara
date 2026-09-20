package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResponse = LoginResponse

type RefreshResponseEnvelope struct {
	httpx.EnvelopeBase
	Data RefreshResponse `json:"data"`
}

// @Summary Refresh tokens
// @Description Refreshes the session by exchanging a refresh token for a fresh access and refresh token pair, rotating the session so the presented token cannot be reused. Requires an active account; invalid, expired, or already-reused tokens are rejected with 401.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body RefreshRequest true "Refresh token"
// @Success 200 {object} RefreshResponseEnvelope "Tokens refreshed."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Invalid, expired, reused (ERR_TOKEN_REUSED), or missing refresh token"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h AuthHandler) RefreshHandler(c fiber.Ctx) error {
	var request RefreshRequest

	if !bindAndValidate(c, &request) {
		return nil
	}

	tokens, err := h.authSvc.Refresh(c, request.RefreshToken, iam.SessionClientInfo{
		IPAddress:  httpx.ClientIP(c),
		DeviceName: httpx.DeviceName(c),
		Browser:    httpx.Browser(c),
		OS:         httpx.OS(c),
	})
	if err != nil {
		httpx.RequestLog(c).Warn("token refresh failed", "error", err)
		switch {
		case errors.Is(err, iam.ErrTokenReused):
			return httpx.CreateTokenReusedErrorResponse(c, "Refresh token already used.", err)
		case errors.Is(err, iam.ErrTokenNotFound):
			return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
		case errors.Is(err, iam.ErrInvalidRefreshToken):
			return httpx.CreateUnauthorizedErrorResponse(c, "Invalid or expired refresh token.", err)
		default:
			return httpx.CreateInternalServerErrorResponse(c, "", err)
		}
	}

	httpx.RequestLog(c).Info("token refreshed", "user_id", tokens.UserID)

	return httpx.CreateSuccessResponse(c, "", RefreshResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}
