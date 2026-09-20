package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h OrganizationHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	organizations := api.Group("/organizations", guards.AuthN)

	organizations.Post("/", h.Create)
	organizations.Post("/quick", h.QuickCreate)

	organizations.Get("/", guards.Guard("organization", "view"), h.List)
	organizations.Get("/:id", guards.Guard("organization", "view"), h.Get)
	organizations.Put("/:id", guards.Guard("organization", "update"), h.Update)
	organizations.Delete("/:id", guards.Guard("organization", "delete"), h.Delete)

	organizations.Get("/:id/modules", guards.Guard("organization", "view"), h.ListModules)
	organizations.Put("/:id/modules", guards.Guard("organization", "update"), h.UpdateModule)
}
