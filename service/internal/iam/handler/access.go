package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
)

func ensureSelf(c fiber.Ctx, targetUserID uint64) error {
	userID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.ErrMissingUserInContext
	}

	if userID == targetUserID {
		return nil
	}

	return iam.ErrForbidden
}

func writeAccessError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, iam.ErrForbidden):
		return httpx.CreateForbiddenErrorResponse(c, "Forbidden.", iam.ErrForbidden)
	case errors.Is(err, httpx.ErrMissingUserInContext):
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	default:
		httpx.RequestLog(c).Error("access verification failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to verify access.", err)
	}
}
