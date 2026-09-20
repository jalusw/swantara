package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginResponseEnvelope struct {
	httpx.EnvelopeBase
	Data LoginResponse `json:"data"`
}
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// @Summary Authenticate user
// @Description Logs in an active user by email and password, authenticating the credentials and issuing an access and refresh token pair. A new session is recorded with the client's IP address and the device, browser, and OS parsed from the User-Agent header, and invalid credentials are rejected with 401.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponseEnvelope "User authenticated."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Invalid credentials"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h AuthHandler) LoginHandler(c fiber.Ctx) error {
	var request LoginRequest

	if !bindAndValidate(c, &request) {
		return nil
	}

	tokens, err := h.authSvc.Login(c, request.Email, request.Password, iam.SessionClientInfo{
		IPAddress:  httpx.ClientIP(c),
		DeviceName: httpx.DeviceName(c),
		Browser:    httpx.Browser(c),
		OS:         httpx.OS(c),
	})
	if err != nil {
		switch {
		case errors.Is(err, iam.ErrInvalidCredentials):
			httpx.RequestLog(c).Warn("login failed", "email", helper.MaskEmail(request.Email), "reason", "invalid_credentials")
			return httpx.CreateUnauthorizedErrorResponse(c, "Invalid credentials provided.", err)
		default:
			httpx.RequestLog(c).Error("login failed", "email", helper.MaskEmail(request.Email), "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "", err)
		}
	}

	httpx.RequestLog(c).Info("login succeeded", "user_id", tokens.UserID)

	return httpx.CreateSuccessResponse(c, "User authenticated.", LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}
