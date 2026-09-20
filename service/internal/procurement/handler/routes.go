package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h PurchaseRequestHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	requisitions := api.Group("/purchase-requests", guards.AuthN)

	requisitions.Get("/", guards.Guard("purchase_request", "view"), h.List)
	requisitions.Post("/", guards.Guard("purchase_request", "create"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-request-create"), h.Create)
	requisitions.Get("/:id", guards.Guard("purchase_request", "view"), h.Get)
	requisitions.Post("/:id/confirm", guards.Guard("purchase_request", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-request-confirm"), h.Confirm)
	requisitions.Post("/:id/approve", guards.Guard("purchase_request", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-request-approve"), h.Approve)
	requisitions.Post("/:id/cancel", guards.Guard("purchase_request", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-request-cancel"), h.Cancel)
}

func (h PurchaseOrderHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	orders := api.Group("/purchase-orders", guards.AuthN)

	orders.Get("/", guards.Guard("purchase_order", "view"), h.List)
	orders.Post("/", guards.Guard("purchase_order", "create"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-order-create"), h.Create)
	orders.Get("/:id", guards.Guard("purchase_order", "view"), h.Get)
	orders.Put("/:id", guards.Guard("purchase_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-order-update"), h.UpdateDraft)
	orders.Post("/:id/confirm", guards.Guard("purchase_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-order-confirm"), h.Confirm)
	orders.Post("/:id/cancel", guards.Guard("purchase_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-order-cancel"), h.Cancel)
	orders.Post("/:id/receive", guards.Guard("purchase_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-order-receive"), h.Receive)
	orders.Post("/:id/supplier-bill", guards.Guard("purchase_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-order-supplier-bill"), h.CreateSupplierBill)
	orders.Post("/:id/pay", guards.Guard("purchase_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-order-pay"), h.Pay)
	orders.Post("/:id/credit-memo", guards.Guard("purchase_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-order-credit-memo"), h.CreateCreditMemo)
	orders.Post("/:id/debit-memo", guards.Guard("purchase_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-order-debit-memo"), h.CreateDebitMemo)
}

func (h SupplierQuoteRequestHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	quoteRequests := api.Group("/supplier-quote-requests", guards.AuthN)

	quoteRequests.Get("/", guards.Guard("supplier_quoteRequest", "view"), h.List)
	quoteRequests.Post("/", guards.Guard("supplier_quoteRequest", "create"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-quote-request-create"), h.Create)
	quoteRequests.Post("/from-request", guards.Guard("supplier_quoteRequest", "create"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-quote-request-create-from-request"), h.CreateFromRequest)
	quoteRequests.Get("/:id", guards.Guard("supplier_quoteRequest", "view"), h.Get)
	quoteRequests.Post("/:id/send", guards.Guard("supplier_quoteRequest", "update"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-quote-request-send"), h.Send)
	quoteRequests.Post("/:id/cancel", guards.Guard("supplier_quoteRequest", "update"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-quote-request-cancel"), h.Cancel)
	quoteRequests.Get("/:id/lines", guards.Guard("supplier_quoteRequest", "view"), h.ListLines)
	quoteRequests.Post("/:id/quotes", guards.Guard("supplier_quoteRequest", "update"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-quote-request-submit-quote"), h.SubmitQuote)
	quoteRequests.Get("/:id/quotes", guards.Guard("supplier_quoteRequest", "view"), h.ListQuotes)
	quoteRequests.Post("/:id/quotes/:supplier_quote_id/accept", guards.Guard("supplier_quoteRequest", "update"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-quote-request-accept-quote"), h.AcceptQuote)
	quoteRequests.Post("/:id/purchase-order", guards.Guard("supplier_quoteRequest", "update"), httpx.IdempotencyGuard(guards.Idempotency, "supplier-quote-request-create-purchase-order"), h.CreatePurchaseOrder)
}

func (h CurrencyRateHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	rates := api.Group("/currency-rates", guards.AuthN)

	rates.Get("/", guards.Guard("currency_rate", "view"), h.List)
	rates.Post("/", guards.Guard("currency_rate", "create"), httpx.IdempotencyGuard(guards.Idempotency, "currency-rate-create"), h.Create)
	rates.Get("/:id", guards.Guard("currency_rate", "view"), h.Get)
	rates.Put("/:id", guards.Guard("currency_rate", "update"), httpx.IdempotencyGuard(guards.Idempotency, "currency-rate-update"), h.Update)
	rates.Delete("/:id", guards.Guard("currency_rate", "delete"), h.Delete)
}

func (h SupplyAgreementHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	agreements := api.Group("/supply-agreements", guards.AuthN)

	agreements.Get("/", guards.Guard("supply_agreement", "view"), h.List)
	agreements.Post("/", guards.Guard("supply_agreement", "create"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-agreement-create"), h.Create)
	agreements.Get("/:id", guards.Guard("supply_agreement", "view"), h.Get)
	agreements.Post("/:id/activate", guards.Guard("supply_agreement", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-agreement-activate"), h.Activate)
	agreements.Post("/:id/cancel", guards.Guard("supply_agreement", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-agreement-cancel"), h.Cancel)
	agreements.Post("/:id/create-order", guards.Guard("supply_agreement", "update"), httpx.IdempotencyGuard(guards.Idempotency, "purchase-agreement-create-order"), h.CreateFromAgreement)
}

func (h SupplierScorecardHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	scorecards := api.Group("/supplier-scorecards", guards.AuthN)

	scorecards.Get("/", guards.Guard("supplier_scorecard", "view"), h.List)
	scorecards.Get("/:id", guards.Guard("supplier_scorecard", "view"), h.Get)
	scorecards.Get("/supplier/:supplier_id", guards.Guard("supplier_scorecard", "view"), h.ListBySupplier)
}

func (h CostCenterHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	costCenters := api.Group("/cost-centers", guards.AuthN)

	costCenters.Get("/", guards.Guard("cost_center", "view"), h.List)
	costCenters.Post("/", guards.Guard("cost_center", "create"), httpx.IdempotencyGuard(guards.Idempotency, "cost-center-create"), h.Create)
	costCenters.Get("/:id", guards.Guard("cost_center", "view"), h.Get)
	costCenters.Put("/:id", guards.Guard("cost_center", "update"), httpx.IdempotencyGuard(guards.Idempotency, "cost-center-update"), h.Update)
	costCenters.Delete("/:id", guards.Guard("cost_center", "delete"), h.Delete)
}

func (h PaymentBatchHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	batches := api.Group("/payment-batches", guards.AuthN)

	batches.Get("/", guards.Guard("payment_batch", "view"), h.List)
	batches.Post("/", guards.Guard("payment_batch", "create"), httpx.IdempotencyGuard(guards.Idempotency, "payment-batch-create"), h.Create)
	batches.Get("/:id", guards.Guard("payment_batch", "view"), h.Get)
	batches.Post("/:id/confirm", guards.Guard("payment_batch", "update"), httpx.IdempotencyGuard(guards.Idempotency, "payment-batch-confirm"), h.Confirm)
}
