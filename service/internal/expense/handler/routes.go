package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h ExpenseCategoryHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	categories := api.Group("/expense-categories", guards.AuthN)

	categories.Get("/", guards.Guard("expense_category", "view"), h.List)
	categories.Post("/", guards.Guard("expense_category", "create"), httpx.IdempotencyGuard(guards.Idempotency, "expense-category-create"), h.Create)
	categories.Get("/:id", guards.Guard("expense_category", "view"), h.Get)
	categories.Put("/:id", guards.Guard("expense_category", "update"), httpx.IdempotencyGuard(guards.Idempotency, "expense-category-update"), h.Update)
	categories.Delete("/:id", guards.Guard("expense_category", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "expense-category-delete"), h.Delete)
}

func (h ExpenseReportHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	reports := api.Group("/expense-reports", guards.AuthN)

	reports.Get("/", guards.Guard("expense_report", "view"), h.List)
	reports.Post("/", guards.Guard("expense_report", "create"), httpx.IdempotencyGuard(guards.Idempotency, "expense-report-create"), h.Create)
	reports.Get("/:id", guards.Guard("expense_report", "view"), h.Get)
	reports.Post("/:id/submit", guards.Guard("expense_report", "update"), httpx.IdempotencyGuard(guards.Idempotency, "expense-report-submit"), h.Submit)
	reports.Post("/:id/approve", guards.Guard("expense_report", "approve"), httpx.IdempotencyGuard(guards.Idempotency, "expense-report-approve"), h.Approve)
	reports.Post("/:id/refuse", guards.Guard("expense_report", "approve"), httpx.IdempotencyGuard(guards.Idempotency, "expense-report-refuse"), h.Refuse)
	reports.Post("/:id/post", guards.Guard("expense_report", "post"), httpx.IdempotencyGuard(guards.Idempotency, "expense-report-post"), h.Post)
	reports.Post("/:id/reimburse", guards.Guard("expense_report", "post"), httpx.IdempotencyGuard(guards.Idempotency, "expense-report-reimburse"), h.Reimburse)
	reports.Post("/:id/bill", guards.Guard("expense_report", "post"), httpx.IdempotencyGuard(guards.Idempotency, "expense-report-bill"), h.Bill)
}
