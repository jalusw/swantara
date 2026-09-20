package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h WarehouseHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	warehouses := api.Group("/warehouses", guards.AuthN)

	warehouses.Get("/", guards.Guard("warehouse", "view"), h.List)
	warehouses.Post("/", guards.Guard("warehouse", "create"), httpx.IdempotencyGuard(guards.Idempotency, "warehouse-create"), h.Create)
	warehouses.Get("/:id", guards.Guard("warehouse", "view"), h.Get)
	warehouses.Put("/:id", guards.Guard("warehouse", "update"), httpx.IdempotencyGuard(guards.Idempotency, "warehouse-update"), h.Update)
	warehouses.Delete("/:id", guards.Guard("warehouse", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "warehouse-delete"), h.Delete)

	locations := api.Group("/stock-locations", guards.AuthN)

	locations.Get("/", guards.Guard("stock_location", "view"), h.ListLocations)
	locations.Post("/", guards.Guard("stock_location", "create"), httpx.IdempotencyGuard(guards.Idempotency, "stock-location-create"), h.CreateLocation)
	locations.Get("/:id", guards.Guard("stock_location", "view"), h.GetLocation)
	locations.Put("/:id", guards.Guard("stock_location", "update"), httpx.IdempotencyGuard(guards.Idempotency, "stock-location-update"), h.UpdateLocation)
	locations.Delete("/:id", guards.Guard("stock_location", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "stock-location-delete"), h.DeleteLocation)
}

func (h StockHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	stock := api.Group("/stock", guards.AuthN)

	stock.Get("/on-hand", guards.Guard("stock_balance", "view"), h.OnHand)
	stock.Get("/available-to-promise", guards.Guard("stock_balance", "view"), h.AvailableToPromise)
	stock.Get("/balances", guards.Guard("stock_balance", "view"), h.ListBalances)
	stock.Post("/rebuild", guards.Guard("stock_balance", "update"), h.RebuildBalances)

	movements := api.Group("/stock-movements", guards.AuthN)

	movements.Get("/", guards.Guard("stock_movement", "view"), h.ListMovements)
	movements.Post("/", guards.Guard("stock_movement", "create"), httpx.IdempotencyGuard(guards.Idempotency, "stock-movement-create"), h.CreateMovement)
	movements.Get("/:id", guards.Guard("stock_movement", "view"), h.GetMovement)
	movements.Delete("/:id", guards.Guard("stock_movement", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "stock-movement-delete"), h.DeleteMovement)
	movements.Post("/:id/receive", guards.Guard("stock_movement", "update"), httpx.IdempotencyGuard(guards.Idempotency, "stock-movement-receive"), h.Receive)
	movements.Post("/:id/ship", guards.Guard("stock_movement", "update"), httpx.IdempotencyGuard(guards.Idempotency, "stock-movement-ship"), h.Ship)

	shipments := api.Group("/shipments", guards.AuthN)

	shipments.Get("/", guards.Guard("shipment", "view"), h.ListShipments)
	shipments.Get("/:id", guards.Guard("shipment", "view"), h.GetShipment)
}

func (h HoldHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	holds := api.Group("/stock-holds", guards.AuthN)

	holds.Get("/", guards.Guard("stock_hold", "view"), h.List)
	holds.Post("/", guards.Guard("stock_hold", "create"), httpx.IdempotencyGuard(guards.Idempotency, "stock-hold-create"), h.Reserve)
	holds.Post("/release-by-movement", guards.Guard("stock_hold", "update"), httpx.IdempotencyGuard(guards.Idempotency, "stock-hold-release-by-movement"), h.ReleaseByMovement)
	holds.Delete("/:id", guards.Guard("stock_hold", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "stock-hold-release"), h.Release)
}

func (h BatchHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	batches := api.Group("/batches", guards.AuthN)

	batches.Get("/", guards.Guard("batch", "view"), h.List)
	batches.Post("/", guards.Guard("batch", "create"), httpx.IdempotencyGuard(guards.Idempotency, "batch-create"), h.Create)
	batches.Get("/:id", guards.Guard("batch", "view"), h.Get)
	batches.Put("/:id", guards.Guard("batch", "update"), httpx.IdempotencyGuard(guards.Idempotency, "batch-update"), h.Update)
}

func (h ReorderHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	rules := api.Group("/reorder-rules", guards.AuthN)

	rules.Get("/candidates", guards.Guard("reorder_rule", "view"), h.Candidates)
	rules.Get("/", guards.Guard("reorder_rule", "view"), h.List)
	rules.Post("/", guards.Guard("reorder_rule", "create"), httpx.IdempotencyGuard(guards.Idempotency, "reorder-rule-create"), h.Create)
	rules.Get("/:id", guards.Guard("reorder_rule", "view"), h.Get)
	rules.Put("/:id", guards.Guard("reorder_rule", "update"), httpx.IdempotencyGuard(guards.Idempotency, "reorder-rule-update"), h.Update)
	rules.Delete("/:id", guards.Guard("reorder_rule", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "reorder-rule-delete"), h.Delete)
}

func (h StockCountHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	stockCounts := api.Group("/stock-counts", guards.AuthN)

	stockCounts.Get("/", guards.Guard("stock_count", "view"), h.List)
	stockCounts.Post("/", guards.Guard("stock_count", "create"), httpx.IdempotencyGuard(guards.Idempotency, "stock-count-create"), h.Create)
	stockCounts.Get("/:id", guards.Guard("stock_count", "view"), h.Get)
	stockCounts.Get("/:id/lines", guards.Guard("stock_count", "view"), h.ListLines)
	stockCounts.Post("/:id/post", guards.Guard("stock_count", "update"), httpx.IdempotencyGuard(guards.Idempotency, "stock-count-post"), h.Post)
	stockCounts.Delete("/:id", guards.Guard("stock_count", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "stock-count-delete"), h.Delete)
}

func (h TransferHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	warehouseTransfers := api.Group("/warehouse-transfers", guards.AuthN)

	warehouseTransfers.Get("/", guards.Guard("warehouse_transfer", "view"), h.List)
	warehouseTransfers.Post("/", guards.Guard("warehouse_transfer", "create"), httpx.IdempotencyGuard(guards.Idempotency, "warehouse-transfer-create"), h.Create)
	warehouseTransfers.Get("/:id", guards.Guard("warehouse_transfer", "view"), h.Get)
	warehouseTransfers.Post("/:id/send", guards.Guard("warehouse_transfer", "update"), httpx.IdempotencyGuard(guards.Idempotency, "warehouse-transfer-send"), h.Send)
	warehouseTransfers.Post("/:id/receive", guards.Guard("warehouse_transfer", "update"), httpx.IdempotencyGuard(guards.Idempotency, "warehouse-transfer-receive"), h.Receive)
}

func (h InboundCostHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	inboundCosts := api.Group("/inbound-costs", guards.AuthN)

	inboundCosts.Get("/", guards.Guard("inbound_cost", "view"), h.List)
	inboundCosts.Post("/", guards.Guard("inbound_cost", "create"), httpx.IdempotencyGuard(guards.Idempotency, "inbound-cost-create"), h.Create)
	inboundCosts.Get("/:id", guards.Guard("inbound_cost", "view"), h.Get)
	inboundCosts.Get("/:id/lines", guards.Guard("inbound_cost", "view"), h.ListLines)
	inboundCosts.Get("/:id/adjustments", guards.Guard("inbound_cost", "view"), h.ListAdjustments)
	inboundCosts.Post("/:id/post", guards.Guard("inbound_cost", "post"), httpx.IdempotencyGuard(guards.Idempotency, "inbound-cost-post"), h.Post)
}
