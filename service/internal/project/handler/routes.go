package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h ProjectHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	projects := api.Group("/projects", guards.AuthN)

	projects.Get("/", guards.Guard("project", "view"), h.List)
	projects.Post("/", guards.Guard("project", "create"), httpx.IdempotencyGuard(guards.Idempotency, "project-create"), h.Create)
	projects.Get("/:id", guards.Guard("project", "view"), h.Get)
	projects.Put("/:id", guards.Guard("project", "update"), httpx.IdempotencyGuard(guards.Idempotency, "project-update"), h.Update)
	projects.Put("/:id/state", guards.Guard("project", "update"), httpx.IdempotencyGuard(guards.Idempotency, "project-set-state"), h.SetState)
	projects.Get("/:id/summary", guards.Guard("project", "view"), h.Summary)
	projects.Post("/:id/bill", guards.Guard("project", "update"), httpx.IdempotencyGuard(guards.Idempotency, "project-bill"), h.BillTimeMaterial)

	projects.Get("/:id/tasks", guards.Guard("project_task", "view"), h.ListTasks)
	projects.Post("/:id/tasks", guards.Guard("project_task", "create"), httpx.IdempotencyGuard(guards.Idempotency, "project-task-create"), h.CreateTask)
	projects.Put("/:id/tasks/:task_id", guards.Guard("project_task", "update"), httpx.IdempotencyGuard(guards.Idempotency, "project-task-update"), h.UpdateTask)

	projects.Get("/:id/milestones", guards.Guard("project_milestone", "view"), h.ListMilestones)
	projects.Post("/:id/milestones", guards.Guard("project_milestone", "create"), httpx.IdempotencyGuard(guards.Idempotency, "project-milestone-create"), h.CreateMilestone)
	projects.Put("/:id/milestones/:milestone_id/reached", guards.Guard("project_milestone", "update"), httpx.IdempotencyGuard(guards.Idempotency, "project-milestone-set-reached"), h.SetMilestoneReached)
}
