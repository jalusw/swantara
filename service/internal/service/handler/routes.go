package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h ServiceHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	equipments := api.Group("/equipments", guards.AuthN)
	equipments.Get("/", guards.Guard("equipment", "view"), h.ListEquipments)
	equipments.Post("/", guards.Guard("equipment", "create"), httpx.IdempotencyGuard(guards.Idempotency, "equipment-create"), h.CreateEquipment)
	equipments.Get("/:id", guards.Guard("equipment", "view"), h.GetEquipment)

	contracts := api.Group("/service-contracts", guards.AuthN)
	contracts.Get("/", guards.Guard("service_contract", "view"), h.ListContracts)
	contracts.Post("/", guards.Guard("service_contract", "create"), httpx.IdempotencyGuard(guards.Idempotency, "service-contract-create"), h.CreateContract)
	contracts.Get("/:id", guards.Guard("service_contract", "view"), h.GetContract)
	contracts.Post("/:id/activate", guards.Guard("service_contract", "update"), httpx.IdempotencyGuard(guards.Idempotency, "service-contract-activate"), h.ActivateContract)
	contracts.Post("/:id/cancel", guards.Guard("service_contract", "update"), httpx.IdempotencyGuard(guards.Idempotency, "service-contract-cancel"), h.CancelContract)

	orders := api.Group("/service-orders", guards.AuthN)
	orders.Get("/", guards.Guard("service_order", "view"), h.ListOrders)
	orders.Post("/", guards.Guard("service_order", "create"), httpx.IdempotencyGuard(guards.Idempotency, "service-order-create"), h.CreateOrder)
	orders.Get("/:id", guards.Guard("service_order", "view"), h.GetOrder)
	orders.Get("/:id/lines", guards.Guard("service_order", "view"), h.ListOrderLines)
	orders.Post("/:id/lines", guards.Guard("service_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "service-order-add-line"), h.AddOrderLine)
	orders.Post("/:id/schedule", guards.Guard("service_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "service-order-schedule"), h.ScheduleOrder)
	orders.Post("/:id/start", guards.Guard("service_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "service-order-start"), h.StartOrder)
	orders.Post("/:id/complete", guards.Guard("service_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "service-order-complete"), h.CompleteOrder)
	orders.Post("/:id/bill", guards.Guard("service_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "service-order-bill"), h.BillOrder)
	orders.Post("/:id/cancel", guards.Guard("service_order", "update"), httpx.IdempotencyGuard(guards.Idempotency, "service-order-cancel"), h.CancelOrder)

	plans := api.Group("/maintenance-plans", guards.AuthN)
	plans.Get("/", guards.Guard("maintenance_plan", "view"), h.ListMaintenancePlans)
	plans.Post("/", guards.Guard("maintenance_plan", "create"), httpx.IdempotencyGuard(guards.Idempotency, "maintenance-plan-create"), h.CreateMaintenancePlan)
	plans.Post("/generate-orders", guards.Guard("maintenance_plan", "update"), httpx.IdempotencyGuard(guards.Idempotency, "maintenance-plan-generate-orders"), h.GenerateMaintenanceOrders)
	plans.Get("/:id", guards.Guard("maintenance_plan", "view"), h.GetMaintenancePlan)
}
