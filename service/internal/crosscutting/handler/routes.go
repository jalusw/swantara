package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h ApprovalRequestHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	requests := api.Group("/approval-requests", guards.AuthN)

	requests.Get("/", guards.Guard("approval_request", "view"), h.List)
	requests.Post("/", guards.Guard("approval_request", "create"), httpx.IdempotencyGuard(guards.Idempotency, "approval-request-create"), h.Create)
	requests.Get("/:id", guards.Guard("approval_request", "view"), h.Get)
	requests.Post("/:id/decide", guards.Guard("approval_request", "update"), httpx.IdempotencyGuard(guards.Idempotency, "approval-request-decide"), h.Decide)
}

func (h AttachmentHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	attachments := api.Group("/attachments", guards.AuthN)

	attachments.Get("/", guards.Guard("attachment", "view"), h.List)
	attachments.Post("/", guards.Guard("attachment", "create"), httpx.IdempotencyGuard(guards.Idempotency, "attachment-create"), h.Upload)
	attachments.Get("/:id", guards.Guard("attachment", "view"), h.Get)
	attachments.Get("/:id/download", guards.Guard("attachment", "view"), h.Download)
	attachments.Delete("/:id", guards.Guard("attachment", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "attachment-delete"), h.Delete)
}

func (h MessageHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	messages := api.Group("/messages", guards.AuthN)

	messages.Get("/", guards.Guard("message", "view"), h.List)
	messages.Post("/", guards.Guard("message", "create"), httpx.IdempotencyGuard(guards.Idempotency, "message-create"), h.Create)
	messages.Get("/:id", guards.Guard("message", "view"), h.Get)
}
