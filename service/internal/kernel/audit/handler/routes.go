package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h AuditLogHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	auditLogs := api.Group("/audit-logs", guards.AuthN)

	auditLogs.Get("/", guards.Guard("audit-log", "view"), h.List)
	auditLogs.Get("/:id", guards.Guard("audit-log", "view"), h.Get)
}
