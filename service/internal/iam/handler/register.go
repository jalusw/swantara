package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=72"`
}

type RegisterResponseEnvelope struct {
	httpx.EnvelopeBase
	Data RegisterResponse `json:"data"`
}
type RegisterResponse struct {
	User PublicUser `json:"user"`
}

// @Summary Register user
// @Description Registers a new user account assigned the default user role and returns the public profile. The username is derived from the email address and suffixed when already taken, the account is active immediately, and email verification remains a separate optional step.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "Registration details"
// @Success 201 {object} RegisterResponseEnvelope "User registered successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or email already registered"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h AuthHandler) RegisterHandler(c fiber.Ctx) error {
	var request RegisterRequest

	if !bindAndValidate(c, &request) {
		return nil
	}

	user, err := h.authSvc.Register(c, iam.RegisterData{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
		Password:  request.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, iam.ErrEmailRegistered):
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Email is already registered.", nil)
		default:
			httpx.RequestLog(c).Error("user registration failed", "email", helper.MaskEmail(request.Email), "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to create user.", err)
		}
	}

	httpx.RequestLog(c).Info("user registered", "user_id", user.ID)

	return httpx.CreateCreatedResponse(c, "User registered successfully.", RegisterResponse{
		User: newPublicUser(user),
	})
}
