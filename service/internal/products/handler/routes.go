package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h ItemCategoryHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	categories := api.Group("/item-categories", guards.AuthN)

	categories.Get("/", guards.Guard("item", "view"), h.List)
	categories.Post("/", guards.Guard("item", "create"), h.Create)
	categories.Get("/:id", guards.Guard("item", "view"), h.Get)
	categories.Put("/:id", guards.Guard("item", "update"), h.Update)
	categories.Delete("/:id", guards.Guard("item", "delete"), h.Delete)
}

func (h ProductHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	products := api.Group("/products", guards.AuthN)

	products.Get("/", guards.Guard("item", "view"), h.List)
	products.Post("/", guards.Guard("item", "create"), h.Create)
	products.Get("/:id", guards.Guard("item", "view"), h.Get)
	products.Put("/:id", guards.Guard("item", "update"), h.Update)
	products.Delete("/:id", guards.Guard("item", "delete"), h.Delete)
	products.Get("/:id/variants", guards.Guard("item", "view"), h.ListVariants)
	products.Post("/:id/variants", guards.Guard("item", "create"), h.CreateVariant)
}

func (h PriceBookHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	price_books := api.Group("/price_books", guards.AuthN)

	price_books.Get("/", guards.Guard("price_book", "view"), h.List)
	price_books.Post("/", guards.Guard("price_book", "create"), h.Create)
	price_books.Get("/:id", guards.Guard("price_book", "view"), h.Get)
	price_books.Put("/:id", guards.Guard("price_book", "update"), h.Update)
	price_books.Delete("/:id", guards.Guard("price_book", "delete"), h.Delete)
	price_books.Get("/:id/rules", guards.Guard("price_book", "view"), h.ListRules)
	price_books.Post("/:id/rules", guards.Guard("price_book", "create"), h.CreateRule)
	price_books.Post("/:id/resolve", guards.Guard("price_book", "view"), h.ResolvePrice)
}

func (h SupplierProductHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	supplierProducts := api.Group("/supplier-products", guards.AuthN)

	supplierProducts.Get("/", guards.Guard("supplier_product", "view"), h.List)
	supplierProducts.Post("/", guards.Guard("supplier_product", "create"), h.Create)
	supplierProducts.Get("/:id", guards.Guard("supplier_product", "view"), h.Get)
	supplierProducts.Put("/:id", guards.Guard("supplier_product", "update"), h.Update)
	supplierProducts.Delete("/:id", guards.Guard("supplier_product", "delete"), h.Delete)
	supplierProducts.Post("/best-offer", guards.Guard("supplier_product", "view"), h.BestOffer)
}
