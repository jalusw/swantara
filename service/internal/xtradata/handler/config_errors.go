package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
)

func writeSystemConfigError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, xtradata.ErrConfigKeyExists):
		return httpx.CreateConflictResponse(c, "System config key already exists.", nil)
	default:
		httpx.RequestLog(c).Error("system config write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save system config.", err)
	}
}
