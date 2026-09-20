package httpx

import "github.com/gofiber/fiber/v3"

type RouteGuards struct {
	AuthN       fiber.Handler
	Guard       func(resource, action string) fiber.Handler
	Idempotency IdempotencyReader
}
