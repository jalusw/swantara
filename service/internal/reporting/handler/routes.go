package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h ReportHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	reports := api.Group("/reports", guards.AuthN)

	reports.Get("/trial-balance", guards.Guard("reporting", "view"), h.TrialBalance)
	reports.Get("/aging", guards.Guard("reporting", "view"), h.Aging)
	reports.Get("/inventory-valuation", guards.Guard("reporting", "view"), h.InventoryValuation)
}

func (h StatementHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	reports := api.Group("/reports", guards.AuthN)

	reports.Get("/profit-and-loss", guards.Guard("reporting", "view"), h.ProfitAndLoss)
	reports.Get("/balance-sheet", guards.Guard("reporting", "view"), h.BalanceSheet)
	reports.Get("/cash-flow", guards.Guard("reporting", "view"), h.CashFlow)

	close := api.Group("/period-close", guards.AuthN)
	close.Post("/year-end-roll", guards.Guard("reporting", "create"), httpx.IdempotencyGuard(guards.Idempotency, "reporting-year-end-roll"), h.YearEndRoll)
}

func (h KpiHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	kpis := api.Group("/kpis", guards.AuthN)

	kpis.Get("/sales", guards.Guard("reporting", "view"), h.Sales)
	kpis.Get("/pipeline", guards.Guard("reporting", "view"), h.Pipeline)
	kpis.Get("/inventory", guards.Guard("reporting", "view"), h.Inventory)
	kpis.Get("/subscription", guards.Guard("reporting", "view"), h.Subscription)
	kpis.Get("/projects", guards.Guard("reporting", "view"), h.Projects)
	kpis.Get("/payroll", guards.Guard("reporting", "view"), h.Payroll)
	kpis.Get("/finance", guards.Guard("reporting", "view"), h.Finance)
	kpis.Get("/procurement", guards.Guard("reporting", "view"), h.Procurement)
	kpis.Get("/manufacturing", guards.Guard("reporting", "view"), h.Manufacturing)
	kpis.Get("/ar-ap", guards.Guard("reporting", "view"), h.ArAp)
	kpis.Get("/cash", guards.Guard("reporting", "view"), h.Cash)
	kpis.Get("/inventory-ratio", guards.Guard("reporting", "view"), h.InventoryRatio)
}

func (h FxRevaluationHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	revaluations := api.Group("/fx-revaluations", guards.AuthN)

	revaluations.Get("/", guards.Guard("reporting", "view"), h.List)
	revaluations.Post("/", guards.Guard("reporting", "create"), httpx.IdempotencyGuard(guards.Idempotency, "fx-revaluation-create"), h.Revalue)
}

func (h AccrualHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	accruals := api.Group("/accruals", guards.AuthN)

	accruals.Get("/", guards.Guard("reporting", "view"), h.List)
	accruals.Post("/", guards.Guard("reporting", "create"), httpx.IdempotencyGuard(guards.Idempotency, "accrual-create"), h.Create)
	accruals.Post("/reverse-due", guards.Guard("reporting", "create"), httpx.IdempotencyGuard(guards.Idempotency, "accrual-reverse-due"), h.ReverseDue)
}

func (h PeriodCloseHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	close := api.Group("/period-close", guards.AuthN)

	close.Post("/", guards.Guard("reporting", "create"), httpx.IdempotencyGuard(guards.Idempotency, "period-close-create"), h.Close)
}
