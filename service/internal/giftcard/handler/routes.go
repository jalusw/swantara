package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h GiftCardHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	cards := api.Group("/gift-cards", guards.AuthN)
	cards.Get("/", guards.Guard("gift_card", "view"), h.List)
	cards.Post("/", guards.Guard("gift_card", "create"), httpx.IdempotencyGuard(guards.Idempotency, "gift-card-create"), h.Issue)
	cards.Post("/forfeit-expired", guards.Guard("gift_card", "update"), httpx.IdempotencyGuard(guards.Idempotency, "gift-card-forfeit-expired"), h.ForfeitExpired)
	cards.Get("/:id", guards.Guard("gift_card", "view"), h.Get)
	cards.Post("/:id/redeem", guards.Guard("gift_card", "update"), httpx.IdempotencyGuard(guards.Idempotency, "gift-card-redeem"), h.Redeem)
	cards.Post("/:id/refund", guards.Guard("gift_card", "update"), httpx.IdempotencyGuard(guards.Idempotency, "gift-card-refund"), h.Refund)
	cards.Get("/:id/transactions", guards.Guard("gift_card_transaction", "view"), h.ListTransactions)
}

func (h CouponHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	coupons := api.Group("/coupons", guards.AuthN)
	coupons.Get("/", guards.Guard("coupon", "view"), h.List)
	coupons.Post("/", guards.Guard("coupon", "create"), httpx.IdempotencyGuard(guards.Idempotency, "coupon-create"), h.Create)
	coupons.Post("/redeem", guards.Guard("coupon", "update"), httpx.IdempotencyGuard(guards.Idempotency, "coupon-redeem"), h.Redeem)
	coupons.Get("/:id", guards.Guard("coupon", "view"), h.Get)
	coupons.Put("/:id", guards.Guard("coupon", "update"), httpx.IdempotencyGuard(guards.Idempotency, "coupon-update"), h.Update)
	coupons.Delete("/:id", guards.Guard("coupon", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "coupon-delete"), h.Delete)
}
