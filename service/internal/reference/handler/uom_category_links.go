package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func buildUnitGroupLinks(c fiber.Ctx, categoryID uint64) []httpx.Link {
	base := fmt.Sprintf("%s/unit-categories/%d", httpx.LinkBaseFor(c), categoryID)
	return []httpx.Link{
		{Rel: "self", Method: fiber.MethodGet, Href: base},
		{Rel: "update", Method: fiber.MethodPut, Href: base},
		{Rel: "delete", Method: fiber.MethodDelete, Href: base},
	}
}
