package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h JournalEntryHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	entries := api.Group("/journal-entries", guards.AuthN)

	entries.Get("/", guards.Guard("journal_entry", "view"), h.List)
	entries.Post("/", guards.Guard("journal_entry", "create"), httpx.IdempotencyGuard(guards.Idempotency, "journal-entry-create"), h.Post)
	entries.Get("/:id", guards.Guard("journal_entry", "view"), h.Get)
	entries.Post("/:id/confirm", guards.Guard("journal_entry", "create"), httpx.IdempotencyGuard(guards.Idempotency, "journal-entry-confirm"), h.Confirm)
	entries.Post("/:id/reverse", guards.Guard("journal_entry", "create"), httpx.IdempotencyGuard(guards.Idempotency, "journal-entry-reverse"), h.Reverse)
}

func (h InvoiceHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	invoices := api.Group("/invoices", guards.AuthN)

	invoices.Get("/", guards.Guard("invoice", "view"), h.List)
	invoices.Post("/", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "invoice-create"), h.Create)
	invoices.Get("/aging", guards.Guard("invoice", "view"), h.Aging)
	invoices.Get("/:id", guards.Guard("invoice", "view"), h.Get)
	invoices.Post("/:id/post", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "invoice-post"), h.Post)
	invoices.Post("/:id/write-off", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "invoice-write-off"), h.WriteOff)
	invoices.Post("/:id/credit-note", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "invoice-create-credit-note"), h.CreateCreditNote)
	invoices.Post("/:id/apply-credit", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "invoice-apply-credit"), h.ApplyCredit)
	invoices.Post("/:id/deduct-down-payment", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "invoice-deduct-down-payment"), h.DeductDownPayment)

	contra := api.Group("/contra-settlements", guards.AuthN)
	contra.Post("/", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "contra-settle"), h.ContraSettle)
}

func (h InvoiceHandler) RegisterSupplierBills(api fiber.Router, guards httpx.RouteGuards) {
	bills := api.Group("/supplier-bills", guards.AuthN)

	bills.Post("/", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-bill-create"), h.CreateSupplierBill)
	bills.Post("/:id/post", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-bill-post"), h.Post)
	bills.Post("/:id/credit-note", guards.Guard("invoice", "create"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-bill-create-credit-note"), h.CreateVendorCreditNote)
}

func (h PaymentHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	payments := api.Group("/payments", guards.AuthN)

	payments.Get("/", guards.Guard("payment", "view"), h.List)
	payments.Post("/", guards.Guard("payment", "create"), httpx.IdempotencyGuard(guards.Idempotency, "payment-create"), h.Create)
	payments.Post("/outbound", guards.Guard("payment", "create"), httpx.IdempotencyGuard(guards.Idempotency, "payment-create-outbound"), h.CreateOutbound)
	payments.Get("/:id", guards.Guard("payment", "view"), h.Get)
}

func (h PdcHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	instruments := api.Group("/pdc-instruments", guards.AuthN)

	instruments.Post("/", guards.Guard("payment", "create"), httpx.IdempotencyGuard(guards.Idempotency, "pdc-create"), h.Create)
	instruments.Post("/:id/:action", guards.Guard("payment", "create"), httpx.IdempotencyGuard(guards.Idempotency, "pdc-transition"), h.Transition)
}

func (h TaxPeriodHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	periods := api.Group("/tax-periods", guards.AuthN)

	periods.Get("/", guards.Guard("journal_entry", "view"), h.List)
	periods.Post("/", guards.Guard("journal_entry", "create"), httpx.IdempotencyGuard(guards.Idempotency, "tax-period-create"), h.Create)
	periods.Get("/:id", guards.Guard("journal_entry", "view"), h.Get)
	periods.Post("/:id/close", guards.Guard("journal_entry", "update"), httpx.IdempotencyGuard(guards.Idempotency, "tax-period-close"), h.Close)
	periods.Post("/:id/lock", guards.Guard("journal_entry", "update"), httpx.IdempotencyGuard(guards.Idempotency, "tax-period-lock"), h.Lock)
	periods.Post("/:id/open", guards.Guard("journal_entry", "update"), httpx.IdempotencyGuard(guards.Idempotency, "tax-period-open"), h.Open)
}

func (h BankStatementHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	statements := api.Group("/bank-statements", guards.AuthN)

	statements.Get("/", guards.Guard("bank_statement", "view"), h.List)
	statements.Post("/", guards.Guard("bank_statement", "create"), httpx.IdempotencyGuard(guards.Idempotency, "bank-statement-create"), h.Create)
	statements.Get("/:id", guards.Guard("bank_statement", "view"), h.Get)
	statements.Post("/:id/match", guards.Guard("bank_statement", "update"), httpx.IdempotencyGuard(guards.Idempotency, "bank-statement-match"), h.Match)
}

func (h ReconcileHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	reconciles := api.Group("/reconciliations", guards.AuthN)

	reconciles.Post("/", guards.Guard("journal_entry", "update"), httpx.IdempotencyGuard(guards.Idempotency, "reconcile-create"), h.Reconcile)
}

func (h ReminderHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	reminder := api.Group("/reminder", guards.AuthN)

	reminder.Get("/actions", guards.Guard("reminder", "view"), h.List)
	reminder.Post("/actions/generate", guards.Guard("reminder", "create"), httpx.IdempotencyGuard(guards.Idempotency, "reminder-generate"), h.Generate)
}

func (h BudgetHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	budgets := api.Group("/budgets", guards.AuthN)

	budgets.Get("/", guards.Guard("budget", "view"), h.List)
	budgets.Post("/", guards.Guard("budget", "create"), httpx.IdempotencyGuard(guards.Idempotency, "budget-create"), h.Create)
	budgets.Get("/:id", guards.Guard("budget", "view"), h.Get)
	budgets.Get("/:id/variance", guards.Guard("budget", "view"), h.Variance)
}

func (h TaxRuleHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	positions := api.Group("/tax-rules", guards.AuthN)

	positions.Get("/", guards.Guard("tax_rule", "view"), h.List)
	positions.Post("/", guards.Guard("tax_rule", "create"), httpx.IdempotencyGuard(guards.Idempotency, "tax-rule-create"), h.Create)
	positions.Get("/:id", guards.Guard("tax_rule", "view"), h.Get)
	positions.Post("/:id/resolve", guards.Guard("tax_rule", "view"), httpx.IdempotencyGuard(guards.Idempotency, "tax-rule-resolve"), h.Resolve)
}

func (h WithholdingTaxHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	withholdings := api.Group("/withholding-taxes", guards.AuthN)

	withholdings.Get("/", guards.Guard("withholding_tax", "view"), h.List)
	withholdings.Post("/", guards.Guard("withholding_tax", "create"), httpx.IdempotencyGuard(guards.Idempotency, "withholding-tax-create"), h.Create)
	withholdings.Post("/apply", guards.Guard("withholding_tax", "create"), httpx.IdempotencyGuard(guards.Idempotency, "withholding-tax-apply"), h.Apply)
}

func (h TaxReturnHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	taxReturns := api.Group("/tax-returns", guards.AuthN)

	taxReturns.Get("/", guards.Guard("tax_return", "view"), h.List)
	taxReturns.Post("/", guards.Guard("tax_return", "create"), httpx.IdempotencyGuard(guards.Idempotency, "tax-return-create"), h.Create)
	taxReturns.Get("/:id", guards.Guard("tax_return", "view"), h.Get)
	taxReturns.Post("/:id/file", guards.Guard("tax_return", "update"), httpx.IdempotencyGuard(guards.Idempotency, "tax-return-file"), h.File)
	taxReturns.Post("/:id/pay", guards.Guard("tax_return", "update"), httpx.IdempotencyGuard(guards.Idempotency, "tax-return-pay"), h.Pay)
	taxReturns.Post("/:id/open", guards.Guard("tax_return", "update"), httpx.IdempotencyGuard(guards.Idempotency, "tax-return-open"), h.Open)
	taxReturns.Get("/:id/export", guards.Guard("tax_return", "view"), h.Export)
}

func (h DeferralHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	deferrals := api.Group("/deferrals", guards.AuthN)

	deferrals.Post("/recognize", guards.Guard("journal_entry", "create"), httpx.IdempotencyGuard(guards.Idempotency, "deferral-recognize"), h.Recognize)
	deferrals.Get("/", guards.Guard("journal_entry", "view"), h.List)
	deferrals.Post("/", guards.Guard("journal_entry", "create"), httpx.IdempotencyGuard(guards.Idempotency, "deferral-create"), h.Create)
	deferrals.Get("/:id", guards.Guard("journal_entry", "view"), h.Get)
	deferrals.Get("/:id/lines", guards.Guard("journal_entry", "view"), h.ListLines)
}

func (h ReconcileRuleHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	rules := api.Group("/reconcile-rules", guards.AuthN)

	rules.Get("/", guards.Guard("reconcile_rule", "view"), h.List)
	rules.Post("/", guards.Guard("reconcile_rule", "create"), httpx.IdempotencyGuard(guards.Idempotency, "reconcile-rule-create"), h.Create)
	rules.Post("/suggest", guards.Guard("reconcile_rule", "view"), h.Suggest)
	rules.Post("/apply", guards.Guard("reconcile_rule", "create"), httpx.IdempotencyGuard(guards.Idempotency, "reconcile-rule-apply"), h.Apply)
}

func (h PaymentBatchHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	batches := api.Group("/payment-batches", guards.AuthN)

	batches.Get("/", guards.Guard("payment_batch", "view"), h.List)
	batches.Post("/", guards.Guard("payment_batch", "create"), httpx.IdempotencyGuard(guards.Idempotency, "payment-batch-create"), h.Create)
	batches.Get("/:id", guards.Guard("payment_batch", "view"), h.Get)
	batches.Post("/:id/generate-sepa", guards.Guard("payment_batch", "update"), httpx.IdempotencyGuard(guards.Idempotency, "payment-batch-generate-sepa"), h.GenerateSEPA)
}

func (h GeneralLedgerHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	reports := api.Group("/general-ledger", guards.AuthN)

	reports.Get("/", guards.Guard("journal_entry", "view"), h.List)
}

func (h TrialBalanceHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	periods := api.Group("/tax-periods", guards.AuthN)

	periods.Get("/:id/trial-balance", guards.Guard("journal_entry", "view"), h.Generate)
	reports := api.Group("/reports", guards.AuthN)
	reports.Get("/trial-balance", guards.Guard("journal_entry", "view"), h.Generate)
}

func (h CashFlowHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	reports := api.Group("/reports", guards.AuthN)

	reports.Get("/cash-flow", guards.Guard("journal_entry", "view"), h.Generate)
}

func (h EquityHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	periods := api.Group("/tax-periods", guards.AuthN)

	periods.Get("/:id/equity", guards.Guard("journal_entry", "view"), h.Generate)
}

func (h IntegrityHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	reports := api.Group("/reports", guards.AuthN)

	reports.Get("/integrity", guards.Guard("journal_entry", "view"), h.Check)
}
