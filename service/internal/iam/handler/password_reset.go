package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type RequestPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// @Summary Request password reset
// @Description Sends a one-time password reset token to the given address when it belongs to a registered user. The response is identical whether or not the user exists, which prevents account enumeration.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body RequestPasswordResetRequest true "Email address"
// @Success 202 {object} httpx.EmptyEnvelope "Password reset email sent."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Router /auth/password-reset/request [post]
func (h AuthHandler) RequestPasswordResetHandler(c fiber.Ctx) error {
	var request RequestPasswordResetRequest

	if !bindAndValidate(c, &request) {
		return nil
	}

	if err := h.authSvc.RequestPasswordReset(c, request.Email); err != nil {
		httpx.RequestLog(c).Error("password reset request failed", "email", helper.MaskEmail(request.Email), "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "", err)
	}

	httpx.RequestLog(c).Info("password reset requested", "email", helper.MaskEmail(request.Email))

	return httpx.CreateAcceptedResponse(c, "Password reset email sent.", struct{}{})
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// @Summary Reset password
// @Description Sets a new password using a one-time reset token, marks the token as used, and revokes all existing sessions for the user, forcing a fresh sign-in. Invalid or expired tokens are rejected with 400.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} httpx.EmptyEnvelope "Password reset."
// @Failure 400 {object} httpx.ErrorResponse "Invalid or expired reset token"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Router /auth/password-reset [post]
func (h AuthHandler) ResetPasswordHandler(c fiber.Ctx) error {
	var request ResetPasswordRequest

	if !bindAndValidate(c, &request) {
		return nil
	}

	if err := h.authSvc.ResetPassword(c, request.Token, request.Password); err != nil {
		switch {
		case errors.Is(err, iam.ErrInvalidActionToken), errors.Is(err, iam.ErrActionTokenExpired):
			httpx.RequestLog(c).Warn("password reset failed", "error", err)
			return httpx.CreateBadRequestResponse(c, "Invalid or expired reset token.", err)
		default:
			httpx.RequestLog(c).Error("password reset failed", "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "", err)
		}
	}

	httpx.RequestLog(c).Info("password reset completed")

	return httpx.CreateSuccessResponse(c, "Password reset.", struct{}{})
}
