package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h CommissionHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	plans := api.Group("/commission-plans", guards.AuthN)
	plans.Get("/", guards.Guard("commission_plan", "view"), h.ListPlans)
	plans.Post("/", guards.Guard("commission_plan", "create"), httpx.IdempotencyGuard(guards.Idempotency, "commission-plan-create"), h.CreatePlan)
	plans.Get("/:id", guards.Guard("commission_plan", "view"), h.GetPlan)
	plans.Put("/:id", guards.Guard("commission_plan", "update"), httpx.IdempotencyGuard(guards.Idempotency, "commission-plan-update"), h.UpdatePlan)
	plans.Get("/:id/rules", guards.Guard("commission_rule", "view"), h.ListRules)
	plans.Post("/:id/rules", guards.Guard("commission_rule", "create"), httpx.IdempotencyGuard(guards.Idempotency, "commission-rule-create"), h.CreateRule)
	plans.Get("/:id/assignments", guards.Guard("commission_assignment", "view"), h.ListAssignments)
	plans.Post("/:id/assignments", guards.Guard("commission_assignment", "create"), httpx.IdempotencyGuard(guards.Idempotency, "commission-assignment-create"), h.CreateAssignment)

	entries := api.Group("/commission-entries", guards.AuthN)
	entries.Get("/", guards.Guard("commission_entry", "view"), h.ListEntries)
	entries.Post("/accrue", guards.Guard("commission_entry", "create"), httpx.IdempotencyGuard(guards.Idempotency, "commission-entry-accrue"), h.Accrue)
	entries.Post("/accrue-from-invoice", guards.Guard("commission_entry", "create"), httpx.IdempotencyGuard(guards.Idempotency, "commission-entry-accrue-from-invoice"), h.AccrueFromInvoice)
	entries.Post("/:id/pay", guards.Guard("commission_entry", "update"), httpx.IdempotencyGuard(guards.Idempotency, "commission-entry-pay"), h.Pay)
	entries.Post("/:id/cancel", guards.Guard("commission_entry", "update"), httpx.IdempotencyGuard(guards.Idempotency, "commission-entry-cancel"), h.Cancel)
}
