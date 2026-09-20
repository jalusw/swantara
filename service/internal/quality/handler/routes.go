package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h QualityPointHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	points := api.Group("/quality-points", guards.AuthN)

	points.Get("/", guards.Guard("quality_point", "view"), h.List)
	points.Post("/", guards.Guard("quality_point", "create"), httpx.IdempotencyGuard(guards.Idempotency, "quality-point-create"), h.Create)
	points.Get("/:id", guards.Guard("quality_point", "view"), h.Get)
}

func (h QualityCheckHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	checks := api.Group("/quality-checks", guards.AuthN)

	checks.Get("/", guards.Guard("quality_check", "view"), h.List)
	checks.Get("/:id", guards.Guard("quality_check", "view"), h.Get)
	checks.Post("/:id/result", guards.Guard("quality_check", "update"), httpx.IdempotencyGuard(guards.Idempotency, "quality-check-result"), h.RecordResult)
	checks.Post("/shipments/:id/scrap", guards.Guard("quality_check", "update"), httpx.IdempotencyGuard(guards.Idempotency, "quality-check-scrap"), h.RouteToScrap)
}

func (h QualityAlertHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	alerts := api.Group("/quality-alerts", guards.AuthN)

	alerts.Get("/", guards.Guard("quality_alert", "view"), h.List)
	alerts.Get("/:id", guards.Guard("quality_alert", "view"), h.Get)
	alerts.Put("/:id/state", guards.Guard("quality_alert", "update"), httpx.IdempotencyGuard(guards.Idempotency, "quality-alert-update-state"), h.UpdateState)
}
