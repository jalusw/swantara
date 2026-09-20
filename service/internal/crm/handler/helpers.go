package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type PipelineStageLister interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.PipelineStage], error)
}

type ProspectResponse struct {
	ID              uint64     `json:"id"`
	OrganizationID  *uint64    `json:"organization_id"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	ContactID       *uint64    `json:"contact_id"`
	ContactName     *string    `json:"contact_name"`
	Email           *string    `json:"email"`
	Phone           *string    `json:"phone"`
	JobPosition     *string    `json:"job_position"`
	StageID         *uint64    `json:"stage_id"`
	StageName       *string    `json:"stage_name,omitempty"`
	ExpectedRevenue float64    `json:"expected_revenue"`
	Probability     float64    `json:"probability"`
	Priority        int16      `json:"priority"`
	SalespersonID   *uint64    `json:"salesperson_id"`
	SalesGroupID    *uint64    `json:"sales_group_id"`
	Source          *string    `json:"source"`
	Medium          *string    `json:"medium"`
	Campaign        *string    `json:"campaign"`
	LostReason      *string    `json:"lost_reason"`
	ExpectedClose   *time.Time `json:"expected_close"`
	ClosedAt        *time.Time `json:"closed_at"`
	IsWon           bool       `json:"is_won"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func newProspectResponse(prospect *crm.Prospect, stages map[uint64]*reference.PipelineStage) ProspectResponse {
	response := ProspectResponse{
		ID:              prospect.ID,
		OrganizationID:  prospect.OrganizationID,
		Name:            prospect.Name,
		Type:            prospect.Type,
		ContactID:       prospect.ContactID,
		ContactName:     prospect.ContactName,
		Email:           prospect.Email,
		Phone:           prospect.Phone,
		JobPosition:     prospect.JobPosition,
		StageID:         prospect.StageID,
		ExpectedRevenue: prospect.ExpectedRevenue,
		Probability:     prospect.Probability,
		Priority:        prospect.Priority,
		SalespersonID:   prospect.SalespersonID,
		SalesGroupID:    prospect.SalesGroupID,
		Source:          prospect.Source,
		Medium:          prospect.Medium,
		Campaign:        prospect.Campaign,
		LostReason:      prospect.LostReason,
		ExpectedClose:   prospect.ExpectedClose,
		ClosedAt:        prospect.ClosedAt,
		CreatedAt:       prospect.CreatedAt,
		UpdatedAt:       prospect.UpdatedAt,
	}
	if prospect.StageID != nil {
		if stage, ok := stages[*prospect.StageID]; ok {
			response.StageName = &stage.Name
			response.IsWon = stage.IsWon
		}
	}
	return response
}

func loadStageMap(c fiber.Ctx, stages PipelineStageLister) (map[uint64]*reference.PipelineStage, error) {
	parsedQuery := &query.Query{}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return nil, err
	}
	page, err := stages.List(c, parsedQuery)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint64]*reference.PipelineStage, len(page.Items))
	for _, stage := range page.Items {
		byID[stage.ID] = stage
	}
	return byID, nil
}
