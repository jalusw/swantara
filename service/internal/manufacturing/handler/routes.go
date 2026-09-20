package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h RecipeHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	recipes := api.Group("/recipes", guards.AuthN)

	recipes.Get("/", guards.Guard("recipe", "view"), h.List)
	recipes.Post("/", guards.Guard("recipe", "create"), httpx.IdempotencyGuard(guards.Idempotency, "recipe-create"), h.Create)
	recipes.Get("/:id", guards.Guard("recipe", "view"), h.Get)
	recipes.Put("/:id", guards.Guard("recipe", "update"), httpx.IdempotencyGuard(guards.Idempotency, "recipe-update"), h.Update)
	recipes.Delete("/:id", guards.Guard("recipe", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "recipe-delete"), h.Delete)
	recipes.Get("/:id/lines", guards.Guard("recipe", "view"), h.ListLines)
	recipes.Post("/:id/explode", guards.Guard("recipe", "view"), httpx.IdempotencyGuard(guards.Idempotency, "recipe-explode"), h.Explode)
}

func (h ProductionOrderHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	orders := api.Group("/production-orders", guards.AuthN)

	orders.Get("/", guards.Guard("production_order", "view"), h.List)
	orders.Post("/", guards.Guard("production_order", "create"), httpx.IdempotencyGuard(guards.Idempotency, "production-order-create"), h.Create)
	orders.Get("/:id", guards.Guard("production_order", "view"), h.Get)
	orders.Get("/:id/components", guards.Guard("production_order", "view"), h.ListComponents)
	orders.Post("/:id/confirm", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "production-order-confirm"), h.Confirm)
	orders.Post("/:id/plan", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "production-order-plan"), h.Plan)
	orders.Post("/:id/cancel", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "production-order-cancel"), h.Cancel)
}

func (h ProductionHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	orders := api.Group("/production-orders", guards.AuthN)

	orders.Post("/:id/start", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "production-order-start"), h.Start)
	orders.Post("/:id/shop-tasks", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "production-order-generate-shop-tasks"), h.GenerateShopTasks)
	orders.Get("/:id/shop-tasks", guards.Guard("shop_task", "view"), h.ListShopTasks)
	orders.Post("/:id/consume", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "production-order-consume"), h.Consume)
	orders.Post("/:id/produce", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "production-order-produce"), h.Produce)
	orders.Post("/:id/settle", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "production-order-settle"), h.SettleVariance)
	orders.Post("/:production_order_id/shop-tasks/:id/labor", guards.Guard("shop_task", "update"), httpx.IdempotencyGuard(guards.Idempotency, "shop-task-record-labor"), h.RecordLabor)
}

func (h PlanningHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	planning := api.Group("/planning", guards.AuthN)

	planning.Post("/runs", guards.Guard("planning", "create"), httpx.IdempotencyGuard(guards.Idempotency, "planning-run"), h.Run)
	planning.Get("/runs", guards.Guard("planning", "view"), h.List)
	planning.Get("/runs/:id", guards.Guard("planning", "view"), h.Get)
	planning.Post("/planned-orders/:id/confirm", guards.Guard("planning", "update"), httpx.IdempotencyGuard(guards.Idempotency, "planning-planned-order-confirm"), h.Confirm)
	planning.Post("/forecasts", guards.Guard("planning", "create"), httpx.IdempotencyGuard(guards.Idempotency, "planning-create-forecast"), h.CreateForecast)
	planning.Get("/forecasts", guards.Guard("planning", "view"), h.ListForecasts)
}

func (h OutsideProcessingHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	subs := api.Group("/outside-processing-orders", guards.AuthN)

	subs.Post("/", guards.Guard("production_order", "create"), httpx.IdempotencyGuard(guards.Idempotency, "outside-processing-order-create"), h.Create)
	subs.Post("/:id/send", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "outside-processing-order-send"), h.Send)
	subs.Post("/:id/receive", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "outside-processing-order-receive"), h.Receive)
	subs.Post("/:id/done", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "outside-processing-order-done"), h.Done)
	subs.Post("/:id/cancel", guards.Guard("production_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "outside-processing-order-cancel"), h.Cancel)
}
