package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h SubscriptionPlanHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	plans := api.Group("/subscription-plans", guards.AuthN)

	plans.Get("/", guards.Guard("subscription_plan", "view"), h.List)
	plans.Post("/", guards.Guard("subscription_plan", "create"), httpx.IdempotencyGuard(guards.Idempotency, "subscription-plan-create"), h.Create)
	plans.Get("/:id", guards.Guard("subscription_plan", "view"), h.Get)
	plans.Put("/:id", guards.Guard("subscription_plan", "update"), httpx.IdempotencyGuard(guards.Idempotency, "subscription-plan-update"), h.Update)
	plans.Delete("/:id", guards.Guard("subscription_plan", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "subscription-plan-delete"), h.Delete)
}

func (h SubscriptionHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	subscriptions := api.Group("/subscriptions", guards.AuthN)

	subscriptions.Get("/", guards.Guard("subscription", "view"), h.List)
	subscriptions.Post("/", guards.Guard("subscription", "create"), httpx.IdempotencyGuard(guards.Idempotency, "subscription-create"), h.Create)
	subscriptions.Get("/metrics", guards.Guard("subscription", "view"), h.Metrics)
	subscriptions.Get("/:id", guards.Guard("subscription", "view"), h.Get)
	subscriptions.Post("/:id/activate", guards.Guard("subscription", "update"), httpx.IdempotencyGuard(guards.Idempotency, "subscription-activate"), h.Activate)
	subscriptions.Post("/:id/pause", guards.Guard("subscription", "update"), httpx.IdempotencyGuard(guards.Idempotency, "subscription-pause"), h.Pause)
	subscriptions.Post("/:id/resume", guards.Guard("subscription", "update"), httpx.IdempotencyGuard(guards.Idempotency, "subscription-resume"), h.Resume)
	subscriptions.Post("/:id/churn", guards.Guard("subscription", "update"), httpx.IdempotencyGuard(guards.Idempotency, "subscription-churn"), h.Churn)
	subscriptions.Post("/:id/close", guards.Guard("subscription", "update"), httpx.IdempotencyGuard(guards.Idempotency, "subscription-close"), h.Close)
}
