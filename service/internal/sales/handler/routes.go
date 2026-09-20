package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h SaleOrderHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	orders := api.Group("/sale-orders", guards.AuthN)

	orders.Get("/", guards.Guard("sale_order", "view"), h.List)
	orders.Post("/", guards.Guard("sale_order", "create"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-create"), h.Create)
	orders.Get("/:id", guards.Guard("sale_order", "view"), h.Get)
	orders.Put("/:id", guards.Guard("sale_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-update"), h.Update)
	orders.Delete("/:id", guards.Guard("sale_order", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-delete"), h.Delete)
	orders.Post("/:id/send", guards.Guard("sale_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-send"), h.Send)
	orders.Post("/:id/confirm", guards.Guard("sale_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-confirm"), h.Confirm)
	orders.Post("/:id/cancel", guards.Guard("sale_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-cancel"), h.Cancel)
	orders.Post("/:id/done", guards.Guard("sale_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-done"), h.Done)
	orders.Post("/:id/recompute-statuses", guards.Guard("sale_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-recompute-statuses"), h.RecomputeStatuses)
	orders.Post("/:id/deliver", guards.Guard("sale_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-deliver"), h.Deliver)
	orders.Post("/:id/invoice", guards.Guard("sale_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-invoice"), h.Invoice)
	orders.Post("/:id/pay", guards.Guard("sale_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "sale-order-pay"), h.Pay)
}
