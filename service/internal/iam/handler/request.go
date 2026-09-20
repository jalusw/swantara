package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func bindAndValidate(c fiber.Ctx, request any) bool {
	return httpx.BindAndValidate(c, request)
}

func parseBirthday(c fiber.Ctx, value *string) (*time.Time, bool) {
	birthday, err := helper.ParseBirthday(value)
	if err != nil {
		_ = httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid birthday format, expected YYYY-MM-DD.", nil)
		return nil, false
	}
	return birthday, true
}
