package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h ContactHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	contacts := api.Group("/contacts", guards.AuthN)

	contacts.Get("/", guards.Guard("contact", "view"), h.List)
	contacts.Post("/", guards.Guard("contact", "create"), h.Create)
	contacts.Get("/:id", guards.Guard("contact", "view"), h.Get)
	contacts.Put("/:id", guards.Guard("contact", "update"), h.Update)
	contacts.Delete("/:id", guards.Guard("contact", "delete"), h.Delete)
}

func (h ContactRelationHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	contacts := api.Group("/contacts", guards.AuthN)

	addresses := contacts.Group("/:id/addresses")
	addresses.Get("/", guards.Guard("contact", "view"), h.ListAddresses)
	addresses.Post("/", guards.Guard("contact", "update"), h.CreateAddress)
	addresses.Put("/:address_id", guards.Guard("contact", "update"), h.UpdateAddress)
	addresses.Delete("/:address_id", guards.Guard("contact", "update"), h.DeleteAddress)
	addresses.Post("/:address_id/default", guards.Guard("contact", "update"), h.SetDefaultAddress)

	banks := contacts.Group("/:id/bank-accounts")
	banks.Get("/", guards.Guard("contact", "view"), h.ListBankAccounts)
	banks.Post("/", guards.Guard("contact", "update"), h.CreateBankAccount)
	banks.Put("/:account_id", guards.Guard("contact", "update"), h.UpdateBankAccount)
	banks.Delete("/:account_id", guards.Guard("contact", "update"), h.DeleteBankAccount)

	contacts.Put("/:id/customer", guards.Guard("contact", "update"), h.EnableCustomer)
	contacts.Delete("/:id/customer", guards.Guard("contact", "update"), h.DisableCustomer)
	contacts.Put("/:id/supplier", guards.Guard("contact", "update"), h.EnableSupplier)
	contacts.Delete("/:id/supplier", guards.Guard("contact", "update"), h.DisableSupplier)
}
