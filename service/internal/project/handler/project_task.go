package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/project"
)

type ProjectTaskResponse struct {
	ID             uint64     `json:"id"`
	ProjectID      uint64     `json:"project_id"`
	Name           string     `json:"name"`
	AssigneeID     *uint64    `json:"assignee_id"`
	Stage          string     `json:"stage"`
	PlannedHours   float64    `json:"planned_hours"`
	EffectiveHours float64    `json:"effective_hours"`
	ParentTaskID   *uint64    `json:"parent_task_id"`
	Deadline       *time.Time `json:"deadline"`
	Priority       *int       `json:"priority"`
}

func newProjectTaskResponse(t *project.ProjectTask) ProjectTaskResponse {
	return ProjectTaskResponse{
		ID:             t.ID,
		ProjectID:      t.ProjectID,
		Name:           t.Name,
		AssigneeID:     t.AssigneeID,
		Stage:          t.Stage,
		PlannedHours:   t.PlannedHours,
		EffectiveHours: t.EffectiveHours,
		ParentTaskID:   t.ParentTaskID,
		Deadline:       t.Deadline,
		Priority:       t.Priority,
	}
}

type CreateProjectTaskRequest struct {
	Name         string     `json:"name" validate:"required"`
	AssigneeID   *uint64    `json:"assignee_id"`
	Stage        string     `json:"stage" validate:"omitempty,oneof=backlog todo in_progress done"`
	PlannedHours float64    `json:"planned_hours"`
	ParentTaskID *uint64    `json:"parent_task_id"`
	Deadline     *time.Time `json:"deadline"`
	Priority     *int       `json:"priority"`
}

type ListProjectTasksResponse struct {
	Tasks []ProjectTaskResponse `json:"tasks"`
}

type ListProjectTasksResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListProjectTasksResponse `json:"data"`
}

// @Summary List project tasks
// @Description Lists all tasks for a project, including each task's assignee, stage, planned and effective hours, parent task, deadline, and priority.
// @Tags Projects
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {object} ListProjectTasksResponseEnvelope "Tasks retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{project_id}/tasks [get]
func (h ProjectHandler) ListTasks(c fiber.Ctx) error {
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid project_id.", nil)
	}
	tasks, err := h.svc.ListTasksByProject(c, projectID)
	if err != nil {
		httpx.RequestLog(c).Error("project task list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve project tasks.", err)
	}
	items := make([]ProjectTaskResponse, len(tasks))
	for i, task := range tasks {
		items[i] = newProjectTaskResponse(task)
	}
	return httpx.CreateSuccessResponse(c, "Project tasks retrieved successfully.", ListProjectTasksResponse{
		Tasks: items,
	})
}

type CreateProjectTaskResponse struct {
	Task ProjectTaskResponse `json:"task"`
}

type CreateProjectTaskResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateProjectTaskResponse `json:"data"`
}

// @Summary Create project task
// @Description Creates a new task on a project, defaulting its stage to backlog when omitted. Only projects in the open state can receive new tasks.
// @Tags Projects
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param body body CreateProjectTaskRequest true "Task details"
// @Success 201 {object} CreateProjectTaskResponseEnvelope "Task created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{project_id}/tasks [post]
func (h ProjectHandler) CreateTask(c fiber.Ctx) error {
	var request CreateProjectTaskRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid project_id.", nil)
	}
	created, err := h.svc.CreateTask(c, &project.ProjectTask{
		ProjectID:    projectID,
		Name:         request.Name,
		AssigneeID:   request.AssigneeID,
		Stage:        request.Stage,
		PlannedHours: request.PlannedHours,
		ParentTaskID: request.ParentTaskID,
		Deadline:     request.Deadline,
		Priority:     request.Priority,
	})
	if err != nil {
		return writeProjectError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Project task created successfully.", CreateProjectTaskResponse{
		Task: newProjectTaskResponse(created),
	})
}

type UpdateProjectTaskRequest struct {
	Name         string     `json:"name" validate:"required"`
	AssigneeID   *uint64    `json:"assignee_id"`
	Stage        string     `json:"stage" validate:"omitempty,oneof=backlog todo in_progress done"`
	PlannedHours float64    `json:"planned_hours"`
	ParentTaskID *uint64    `json:"parent_task_id"`
	Deadline     *time.Time `json:"deadline"`
	Priority     *int       `json:"priority"`
}

// @Summary Update project task
// @Description Updates an existing task's fields, including name, stage, assignee, planned hours, parent task, deadline, and priority. Planned hours must not be negative.
// @Tags Projects
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param task_id path integer true "Task ID"
// @Param body body UpdateProjectTaskRequest true "Task details"
// @Success 200 {object} CreateProjectTaskResponseEnvelope "Task updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Task not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{project_id}/tasks/{task_id} [put]
func (h ProjectHandler) UpdateTask(c fiber.Ctx) error {
	var request UpdateProjectTaskRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid project_id.", nil)
	}
	taskID, err := strconv.ParseUint(c.Params("task_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid task_id.", nil)
	}
	updated, err := h.svc.UpdateTask(c, &project.ProjectTask{
		Base:         model.Base{ID: taskID},
		ProjectID:    projectID,
		Name:         request.Name,
		AssigneeID:   request.AssigneeID,
		Stage:        request.Stage,
		PlannedHours: request.PlannedHours,
		ParentTaskID: request.ParentTaskID,
		Deadline:     request.Deadline,
		Priority:     request.Priority,
	})
	if err != nil {
		return writeProjectError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Project task updated successfully.", CreateProjectTaskResponse{
		Task: newProjectTaskResponse(updated),
	})
}
