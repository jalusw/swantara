package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type RequestEmailVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// @Summary Request email verification
// @Description Sends a one-time email verification token to the given address for a registered, unverified user. The response is identical whether or not the user exists or is already verified, which prevents account enumeration.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body RequestEmailVerificationRequest true "Email address"
// @Success 202 {object} httpx.EmptyEnvelope "Verification email sent."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Router /auth/email-verification/request [post]
func (h AuthHandler) RequestEmailVerificationHandler(c fiber.Ctx) error {
	var request RequestEmailVerificationRequest

	if !bindAndValidate(c, &request) {
		return nil
	}

	if err := h.authSvc.RequestEmailVerification(c, request.Email); err != nil {
		httpx.RequestLog(c).Error("email verification request failed", "email", helper.MaskEmail(request.Email), "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "", err)
	}

	httpx.RequestLog(c).Info("email verification requested", "email", helper.MaskEmail(request.Email))

	return httpx.CreateAcceptedResponse(c, "Verification email sent.", struct{}{})
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// @Summary Verify email address
// @Description Marks the user's email as verified using a one-time token delivered by email. Tokens expire after a configured TTL, and invalid, expired, or already-used tokens are rejected with 400.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body VerifyEmailRequest true "Verification token"
// @Success 200 {object} httpx.EmptyEnvelope "Email verified."
// @Failure 400 {object} httpx.ErrorResponse "Invalid or expired verification token"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Router /auth/email-verification/verify [post]
func (h AuthHandler) VerifyEmailHandler(c fiber.Ctx) error {
	var request VerifyEmailRequest

	if !bindAndValidate(c, &request) {
		return nil
	}

	if err := h.authSvc.VerifyEmail(c, request.Token); err != nil {
		switch {
		case errors.Is(err, iam.ErrInvalidActionToken), errors.Is(err, iam.ErrActionTokenExpired):
			httpx.RequestLog(c).Warn("email verification failed", "error", err)
			return httpx.CreateBadRequestResponse(c, "Invalid or expired verification token.", err)
		default:
			httpx.RequestLog(c).Error("email verification failed", "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "", err)
		}
	}

	httpx.RequestLog(c).Info("email verified")

	return httpx.CreateSuccessResponse(c, "Email verified.", struct{}{})
}
