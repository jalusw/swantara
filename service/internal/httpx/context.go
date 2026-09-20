package httpx

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	LocalOrganizationID = "organization_id"
)

var LocalUserID = model.ActorKey

func CallerID(c fiber.Ctx) (uint64, bool) {
	id, ok := c.Locals(LocalUserID).(uint64)
	return id, ok
}

func CallerOrganizationID(c fiber.Ctx) (uint64, bool) {
	id, ok := c.Locals(LocalOrganizationID).(uint64)
	return id, ok
}
