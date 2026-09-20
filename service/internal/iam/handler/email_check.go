package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type CheckEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type CheckEmailResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CheckEmailResponse `json:"data"`
}

type CheckEmailResponse struct {
	Available bool `json:"available"`
}

// @Summary Check email availability
// @Description Checks whether an email address is available for registration. Always returns 200 with an availability flag to prevent account enumeration.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body CheckEmailRequest true "Email to check"
// @Success 200 {object} CheckEmailResponseEnvelope "Email availability check result."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 429 {object} httpx.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Router /auth/email/check [post]
func (h AuthHandler) CheckEmailHandler(c fiber.Ctx) error {
	var request CheckEmailRequest

	if !bindAndValidate(c, &request) {
		return nil
	}

	existing, err := h.authSvc.IsEmailTaken(c, request.Email)
	if err != nil {
		httpx.RequestLog(c).Error("email check failed", "email", helper.MaskEmail(request.Email), "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to check email availability.", err)
	}

	return httpx.CreateSuccessResponse(c, "Email availability checked.", CheckEmailResponse{
		Available: !existing,
	})
}
