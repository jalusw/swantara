package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/project"
)

type ProjectMilestoneResponse struct {
	ID         uint64     `json:"id"`
	ProjectID  uint64     `json:"project_id"`
	Name       string     `json:"name"`
	Deadline   *time.Time `json:"deadline"`
	Reached    bool       `json:"reached"`
	SaleLineID *uint64    `json:"sale_line_id"`
}

func newProjectMilestoneResponse(m *project.ProjectMilestone) ProjectMilestoneResponse {
	return ProjectMilestoneResponse{
		ID:         m.ID,
		ProjectID:  m.ProjectID,
		Name:       m.Name,
		Deadline:   m.Deadline,
		Reached:    m.Reached,
		SaleLineID: m.SaleLineID,
	}
}

type CreateProjectMilestoneRequest struct {
	Name       string     `json:"name" validate:"required"`
	Deadline   *time.Time `json:"deadline"`
	SaleLineID *uint64    `json:"sale_line_id"`
}

type ListProjectMilestonesResponse struct {
	Milestones []ProjectMilestoneResponse `json:"milestones"`
}

type ListProjectMilestonesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListProjectMilestonesResponse `json:"data"`
}

// @Summary List project milestones
// @Description Lists all milestones for a project, including each milestone's deadline, whether it has been reached, and the optional linked sale line.
// @Tags Projects
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Success 200 {object} ListProjectMilestonesResponseEnvelope "Milestones retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{project_id}/milestones [get]
func (h ProjectHandler) ListMilestones(c fiber.Ctx) error {
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid project_id.", nil)
	}
	milestones, err := h.svc.ListMilestonesByProject(c, projectID)
	if err != nil {
		httpx.RequestLog(c).Error("project milestone list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve project milestones.", err)
	}
	items := make([]ProjectMilestoneResponse, len(milestones))
	for i, milestone := range milestones {
		items[i] = newProjectMilestoneResponse(milestone)
	}
	return httpx.CreateSuccessResponse(c, "Project milestones retrieved successfully.", ListProjectMilestonesResponse{
		Milestones: items,
	})
}

type CreateProjectMilestoneResponse struct {
	Milestone ProjectMilestoneResponse `json:"milestone"`
}

type CreateProjectMilestoneResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateProjectMilestoneResponse `json:"data"`
}

// @Summary Create project milestone
// @Description Creates a new milestone on a project with a name, optional deadline, and optional linked sale line for milestone-based billing.
// @Tags Projects
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param body body CreateProjectMilestoneRequest true "Milestone details"
// @Success 201 {object} CreateProjectMilestoneResponseEnvelope "Milestone created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{project_id}/milestones [post]
func (h ProjectHandler) CreateMilestone(c fiber.Ctx) error {
	var request CreateProjectMilestoneRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	projectID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid project_id.", nil)
	}
	created, err := h.svc.CreateMilestone(c, &project.ProjectMilestone{
		ProjectID:  projectID,
		Name:       request.Name,
		Deadline:   request.Deadline,
		SaleLineID: request.SaleLineID,
	})
	if err != nil {
		return writeProjectError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Project milestone created successfully.", CreateProjectMilestoneResponse{
		Milestone: newProjectMilestoneResponse(created),
	})
}

type SetMilestoneReachedRequest struct {
	Reached bool `json:"reached"`
}

// @Summary Set milestone reached
// @Description Updates a project milestone's reached flag to record whether the milestone has been achieved, for example to unlock milestone-based billing.
// @Tags Projects
// @Accept json
// @Produce json
// @Param project_id path integer true "Project ID"
// @Param milestone_id path integer true "Milestone ID"
// @Param body body SetMilestoneReachedRequest true "Reached flag"
// @Success 200 {object} CreateProjectMilestoneResponseEnvelope "Milestone updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Milestone not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{project_id}/milestones/{milestone_id}/reached [put]
func (h ProjectHandler) SetMilestoneReached(c fiber.Ctx) error {
	var request SetMilestoneReachedRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	milestoneID, err := strconv.ParseUint(c.Params("milestone_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid milestone_id.", nil)
	}
	updated, err := h.svc.SetMilestoneReached(c, milestoneID, request.Reached)
	if err != nil {
		return writeProjectError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Project milestone updated successfully.", CreateProjectMilestoneResponse{
		Milestone: newProjectMilestoneResponse(updated),
	})
}
