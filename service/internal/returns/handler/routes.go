package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h RMAHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	rmas := api.Group("/rmas", guards.AuthN)

	rmas.Get("/", guards.Guard("rma", "view"), h.List)
	rmas.Post("/", guards.Guard("rma", "create"), httpx.IdempotencyGuard(guards.Idempotency, "rma-create"), h.Create)
	rmas.Get("/:id", guards.Guard("rma", "view"), h.Get)
	rmas.Post("/:id/confirm", guards.Guard("rma", "update"), httpx.IdempotencyGuard(guards.Idempotency, "rma-confirm"), h.Confirm)
	rmas.Post("/:id/receive", guards.Guard("rma", "update"), httpx.IdempotencyGuard(guards.Idempotency, "rma-receive"), h.Receive)
	rmas.Post("/:id/refund", guards.Guard("rma", "update"), httpx.IdempotencyGuard(guards.Idempotency, "rma-refund"), h.Refund)
	rmas.Post("/:id/done", guards.Guard("rma", "update"), httpx.IdempotencyGuard(guards.Idempotency, "rma-done"), h.Done)
	rmas.Post("/:id/cancel", guards.Guard("rma", "update"), httpx.IdempotencyGuard(guards.Idempotency, "rma-cancel"), h.Cancel)
}
