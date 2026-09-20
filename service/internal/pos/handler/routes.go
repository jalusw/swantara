package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h POSConfigHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	configs := api.Group("/pos/configs", guards.AuthN)

	configs.Get("/", guards.Guard("pos_config", "view"), h.List)
	configs.Post("/", guards.Guard("pos_config", "create"), httpx.IdempotencyGuard(guards.Idempotency, "pos-config-create"), h.Create)
	configs.Get("/:id", guards.Guard("pos_config", "view"), h.Get)
}

func (h POSSessionHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	sessions := api.Group("/pos/sessions", guards.AuthN)

	sessions.Get("/", guards.Guard("pos_session", "view"), h.List)
	sessions.Post("/", guards.Guard("pos_session", "create"), httpx.IdempotencyGuard(guards.Idempotency, "pos-session-create"), h.Open)
	sessions.Get("/:id", guards.Guard("pos_session", "view"), h.Get)
	sessions.Post("/:id/closing", guards.Guard("pos_session", "update"), httpx.IdempotencyGuard(guards.Idempotency, "pos-session-closing"), h.StartClosing)
	sessions.Post("/:id/close", guards.Guard("pos_session", "update"), httpx.IdempotencyGuard(guards.Idempotency, "pos-session-close"), h.Close)
}

func (h POSOrderHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	orders := api.Group("/pos/orders", guards.AuthN)

	orders.Get("/", guards.Guard("pos_order", "view"), h.List)
	orders.Post("/", guards.Guard("pos_order", "create"), httpx.IdempotencyGuard(guards.Idempotency, "pos-order-sell"), h.Sell)
	orders.Get("/:id", guards.Guard("pos_order", "view"), h.Get)
	orders.Post("/:id/invoice", guards.Guard("pos_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "pos-order-invoice"), h.Invoice)
	orders.Post("/:id/refund", guards.Guard("pos_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "pos-order-refund"), h.Refund)
}

func (h PaymentAccountHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	paymentAccounts := api.Group("/pos/payment-accounts", guards.AuthN)

	paymentAccounts.Get("/", guards.Guard("pos_payment_account", "view"), h.List)
	paymentAccounts.Put("/:method", guards.Guard("pos_payment_account", "update"), httpx.IdempotencyGuard(guards.Idempotency, "pos-payment-account-update"), h.Upsert)
	paymentAccounts.Delete("/:method", guards.Guard("pos_payment_account", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "pos-payment-account-delete"), h.Delete)
}
