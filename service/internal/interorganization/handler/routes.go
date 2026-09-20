package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h DropShipHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	orders := api.Group("/dropship-orders", guards.AuthN)
	orders.Post("/", guards.Guard("dropship_order", "create"), httpx.IdempotencyGuard(guards.Idempotency, "dropship-order-create"), h.Create)
	orders.Post("/:id/receive", guards.Guard("dropship_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "dropship-order-receive"), h.Receive)
}

func (h InterorganizationHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	rules := api.Group("/interorganization-rules", guards.AuthN)
	rules.Get("/", guards.Guard("interorganization_rule", "view"), h.ListRules)
	rules.Post("/", guards.Guard("interorganization_rule", "create"), httpx.IdempotencyGuard(guards.Idempotency, "interorg-rule-create"), h.CreateRule)
	rules.Put("/:id", guards.Guard("interorganization_rule", "update"), httpx.IdempotencyGuard(guards.Idempotency, "interorg-rule-update"), h.UpdateRule)
	rules.Delete("/:id", guards.Guard("interorganization_rule", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "interorg-rule-delete"), h.DeleteRule)

	transactions := api.Group("/interorganization-transactions", guards.AuthN)
	transactions.Get("/", guards.Guard("interorganization_transaction", "view"), h.ListTransactions)
	transactions.Post("/mirror-sale-order/:id", guards.Guard("interorganization_transaction", "create"), httpx.IdempotencyGuard(guards.Idempotency, "interorg-transaction-mirror-sale-order"), h.MirrorSaleOrder)
	transactions.Post("/mirror-invoice/:id", guards.Guard("interorganization_transaction", "create"), httpx.IdempotencyGuard(guards.Idempotency, "interorg-transaction-mirror-invoice"), h.MirrorInvoice)
}

func (h ConsolidationHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	runs := api.Group("/consolidation-runs", guards.AuthN)
	runs.Get("/", guards.Guard("consolidation_run", "view"), h.ListRuns)
	runs.Post("/", guards.Guard("consolidation_run", "create"), httpx.IdempotencyGuard(guards.Idempotency, "consolidation-run-create"), h.CreateRun)
	runs.Get("/:id", guards.Guard("consolidation_run", "view"), h.GetRun)
	runs.Post("/:id/run", guards.Guard("consolidation_run", "update"), httpx.IdempotencyGuard(guards.Idempotency, "consolidation-run-run"), h.Run)
}
