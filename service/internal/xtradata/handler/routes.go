package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h SystemConfigHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	configs := api.Group("/system-configs", guards.AuthN)

	configs.Get("/", guards.Guard("system-config", "view"), h.List)
	configs.Post("/", guards.Guard("system-config", "create"), httpx.IdempotencyGuard(guards.Idempotency, "system-config-create"), h.Create)
	configs.Get("/:id", guards.Guard("system-config", "view"), h.Get)
	configs.Put("/:id", guards.Guard("system-config", "update"), httpx.IdempotencyGuard(guards.Idempotency, "system-config-update"), h.Update)
	configs.Delete("/:id", guards.Guard("system-config", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "system-config-delete"), h.Delete)
}

func (h IntegrationEventHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	events := api.Group("/integration-events", guards.AuthN)

	events.Get("/", guards.Guard("integration-event", "view"), h.List)
	events.Get("/:id", guards.Guard("integration-event", "view"), h.Get)
	events.Post("/:id/dispatch", guards.Guard("integration-event", "update"), httpx.IdempotencyGuard(guards.Idempotency, "integration-event-dispatch"), h.Dispatch)
}
