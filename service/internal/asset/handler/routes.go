package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h AssetCategoryHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	categories := api.Group("/asset-categories", guards.AuthN)
	categories.Get("/", guards.Guard("asset_category", "view"), h.ListCategories)
	categories.Post("/", guards.Guard("asset_category", "create"), httpx.IdempotencyGuard(guards.Idempotency, "asset-category-create"), h.CreateCategory)
	categories.Get("/:id", guards.Guard("asset_category", "view"), h.GetCategory)

	assets := api.Group("/fixed-assets", guards.AuthN)
	assets.Get("/", guards.Guard("fixed_asset", "view"), h.ListAssets)
	assets.Post("/", guards.Guard("fixed_asset", "create"), httpx.IdempotencyGuard(guards.Idempotency, "fixed-asset-create"), h.RegisterAsset)
	assets.Get("/:id", guards.Guard("fixed_asset", "view"), h.GetAsset)
	assets.Post("/:id/schedule", guards.Guard("fixed_asset", "update"), httpx.IdempotencyGuard(guards.Idempotency, "fixed-asset-schedule"), h.GenerateSchedule)
	assets.Post("/:id/post-depreciation", guards.Guard("fixed_asset", "update"), httpx.IdempotencyGuard(guards.Idempotency, "fixed-asset-post-depreciation"), h.PostDepreciation)
	assets.Post("/:id/dispose", guards.Guard("fixed_asset", "update"), httpx.IdempotencyGuard(guards.Idempotency, "fixed-asset-dispose"), h.DisposeAsset)
}
