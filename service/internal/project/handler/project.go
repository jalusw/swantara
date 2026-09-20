package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/project"
)

type ProjectResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID uint64     `json:"organization_id"`
	Name           string     `json:"name"`
	ContactID      uint64     `json:"contact_id"`
	ManagerID      *uint64    `json:"manager_id"`
	DimensionID    *uint64    `json:"dimension_id"`
	SaleOrderID    *uint64    `json:"sale_order_id"`
	BillingType    string     `json:"billing_type"`
	BillableRate   float64    `json:"billable_rate"`
	DateStart      *time.Time `json:"date_start"`
	DateEnd        *time.Time `json:"date_end"`
	State          string     `json:"state"`
}

func newProjectResponse(p *project.Project) ProjectResponse {
	return ProjectResponse{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		Name:           p.Name,
		ContactID:      p.ContactID,
		ManagerID:      p.ManagerID,
		DimensionID:    p.DimensionID,
		SaleOrderID:    p.SaleOrderID,
		BillingType:    p.BillingType,
		BillableRate:   p.BillableRate,
		DateStart:      p.DateStart,
		DateEnd:        p.DateEnd,
		State:          p.State,
	}
}

type ProjectHandler struct {
	svc project.ProjectService
}

func NewProjectHandler(
	svc project.ProjectService,
) ProjectHandler {
	return ProjectHandler{svc: svc}
}

var projectQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"contact_id":      {},
	"manager_id":      {},
	"dimension_id":    {},
	"billing_type":    {},
	"state":           {},
	"created_at":      {},
	"updated_at":      {},
}

func writeProjectError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, project.ErrProjectNotFound),
		errors.Is(err, project.ErrProjectContactNotFound),
		errors.Is(err, project.ErrProjectDimensionNotFound),
		errors.Is(err, project.ErrProjectTaskNotFound),
		errors.Is(err, project.ErrProjectMilestoneNotFound):
		return httpx.CreateNotFoundResponse(c, "Resource not found.")
	case errors.Is(err, project.ErrProjectInvalidState),
		errors.Is(err, project.ErrProjectNotTimeMaterial),
		errors.Is(err, project.ErrProjectNoBillableRate),
		errors.Is(err, project.ErrProjectNothingToBill),
		errors.Is(err, project.ErrProjectNoRevenueAccount),
		errors.Is(err, project.ErrProjectCostNotAttributable),
		errors.Is(err, project.ErrProjectNameRequired),
		errors.Is(err, project.ErrProjectInvalidBillingType),
		errors.Is(err, project.ErrProjectInvalidDates),
		errors.Is(err, project.ErrProjectTaskNameRequired),
		errors.Is(err, project.ErrProjectInvalidHours),
		errors.Is(err, project.ErrProjectMilestoneNameRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Request cannot be processed.", nil)
	default:
		httpx.RequestLog(c).Error("project write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process request.", err)
	}
}
