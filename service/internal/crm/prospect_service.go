package crm

import (
	"context"
	"strings"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ProspectService struct {
	leads    ProspectDAO
	stages   dao.CRUD[reference.PipelineStage]
	teams    dao.CRUD[reference.SalesGroup]
	contacts contacts.ContactDAO
}

func NewProspectService(
	leads ProspectDAO,
	stages dao.CRUD[reference.PipelineStage],
	teams dao.CRUD[reference.SalesGroup],
	contacts contacts.ContactDAO,
) ProspectService {
	return ProspectService{leads: leads, stages: stages, teams: teams, contacts: contacts}
}

func (s ProspectService) List(ctx context.Context, q *query.Query) (*query.Page[Prospect], error) {
	return s.leads.List(ctx, q)
}

func (s ProspectService) Find(ctx context.Context, id uint64) (*Prospect, error) {
	return s.leads.Find(ctx, id)
}

func (s ProspectService) Delete(ctx context.Context, id uint64) error {
	return s.leads.Delete(ctx, id)
}

func (s ProspectService) CreateLead(ctx context.Context, prospect *Prospect) (*Prospect, error) {
	if err := s.validateLead(ctx, prospect); err != nil {
		return nil, err
	}
	if prospect.Type == "" {
		prospect.Type = ProspectKindLead
	}
	if prospect.Type == ProspectKindOpportunity && prospect.StageID == nil {
		return nil, ErrStageRequired
	}
	stage, err := s.resolveStage(ctx, prospect.OrganizationID, prospect.StageID)
	if err != nil {
		return nil, err
	}
	if stage != nil {
		prospect.Probability = stage.Probability
	}
	return s.leads.Create(ctx, prospect)
}

func (s ProspectService) UpdateLead(ctx context.Context, prospect *Prospect) (*Prospect, error) {
	if err := s.validateLead(ctx, prospect); err != nil {
		return nil, err
	}
	if prospect.Type == ProspectKindOpportunity && prospect.StageID == nil {
		return nil, ErrStageRequired
	}
	stage, err := s.resolveStage(ctx, prospect.OrganizationID, prospect.StageID)
	if err != nil {
		return nil, err
	}
	if stage != nil {
		prospect.Probability = stage.Probability
	}
	return s.leads.Update(ctx, prospect)
}

type PromotionInput struct {
	StageID         uint64     `json:"stage_id"`
	SalespersonID   *uint64    `json:"salesperson_id"`
	SalesGroupID    *uint64    `json:"sales_group_id"`
	ExpectedRevenue float64    `json:"expected_revenue"`
	Probability     *float64   `json:"probability"`
	Priority        int16      `json:"priority"`
	ExpectedClose   *time.Time `json:"expected_close"`
}

func (s ProspectService) Promote(ctx context.Context, prospectID uint64, input PromotionInput) (*Prospect, error) {
	prospect, err := s.leads.Find(ctx, prospectID)
	if err != nil {
		return nil, err
	}
	if prospect == nil {
		return nil, ErrLeadNotFound
	}
	if prospect.Type != ProspectKindLead {
		return nil, ErrNotLead
	}

	stage, err := s.resolveStage(ctx, prospect.OrganizationID, &input.StageID)
	if err != nil {
		return nil, err
	}
	if stage == nil {
		return nil, ErrStageRequired
	}
	if err := s.validatePromotionRefs(ctx, input); err != nil {
		return nil, err
	}
	if input.ExpectedRevenue < 0 {
		return nil, ErrInvalidRevenue
	}
	if input.Priority < 0 {
		return nil, ErrInvalidPriority
	}

	prospect.Type = ProspectKindOpportunity
	prospect.StageID = &stage.ID
	prospect.SalespersonID = input.SalespersonID
	prospect.SalesGroupID = input.SalesGroupID
	prospect.ExpectedRevenue = input.ExpectedRevenue
	prospect.Priority = input.Priority
	prospect.ExpectedClose = input.ExpectedClose
	if input.Probability != nil {
		if *input.Probability < 0 || *input.Probability > 100 {
			return nil, ErrInvalidProbability
		}
		prospect.Probability = *input.Probability
	} else {
		prospect.Probability = stage.Probability
	}

	return s.leads.Update(ctx, prospect)
}

func (s ProspectService) AdvanceStage(ctx context.Context, prospectID, stageID uint64) (*Prospect, error) {
	prospect, err := s.openOpportunity(ctx, prospectID)
	if err != nil {
		return nil, err
	}
	stage, err := s.resolveStage(ctx, prospect.OrganizationID, &stageID)
	if err != nil {
		return nil, err
	}
	if stage == nil {
		return nil, ErrStageNotFound
	}

	prospect.StageID = &stage.ID
	prospect.Probability = stage.Probability
	if stage.IsWon {
		now := time.Now()
		prospect.ClosedAt = &now
	}
	return s.leads.Update(ctx, prospect)
}

func (s ProspectService) Win(ctx context.Context, prospectID uint64) (*Prospect, error) {
	prospect, err := s.openOpportunity(ctx, prospectID)
	if err != nil {
		return nil, err
	}
	won, err := s.wonStage(ctx, prospect.OrganizationID)
	if err != nil {
		return nil, err
	}
	if won == nil {
		return nil, ErrNoWonStage
	}

	now := time.Now()
	prospect.StageID = &won.ID
	prospect.Probability = 100
	prospect.LostReason = nil
	prospect.ClosedAt = &now
	return s.leads.Update(ctx, prospect)
}

func (s ProspectService) Lose(ctx context.Context, prospectID uint64, lostReason string) (*Prospect, error) {
	prospect, err := s.openOpportunity(ctx, prospectID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(lostReason) == "" {
		return nil, ErrLostReasonRequired
	}

	now := time.Now()
	prospect.LostReason = &lostReason
	prospect.Probability = 0
	prospect.ClosedAt = &now
	return s.leads.Update(ctx, prospect)
}

func (s ProspectService) openOpportunity(ctx context.Context, prospectID uint64) (*Prospect, error) {
	prospect, err := s.leads.Find(ctx, prospectID)
	if err != nil {
		return nil, err
	}
	if prospect == nil {
		return nil, ErrLeadNotFound
	}
	if prospect.Type != ProspectKindOpportunity {
		return nil, ErrNotOpportunity
	}
	if prospect.ClosedAt != nil {
		return nil, ErrLeadClosed
	}
	return prospect, nil
}

func (s ProspectService) validateLead(ctx context.Context, prospect *Prospect) error {
	if strings.TrimSpace(prospect.Name) == "" {
		return ErrLeadNameRequired
	}
	if prospect.Type == "" {
		prospect.Type = ProspectKindLead
	}
	if !validLeadType(prospect.Type) {
		return ErrInvalidLeadType
	}
	if prospect.ExpectedRevenue < 0 {
		return ErrInvalidRevenue
	}
	if prospect.Probability < 0 || prospect.Probability > 100 {
		return ErrInvalidProbability
	}
	if prospect.Priority < 0 {
		return ErrInvalidPriority
	}
	if err := s.validateLeadRefs(ctx, prospect); err != nil {
		return err
	}
	return nil
}

func (s ProspectService) validateLeadRefs(ctx context.Context, prospect *Prospect) error {
	if prospect.ContactID != nil {
		contact, err := s.contacts.Find(ctx, *prospect.ContactID)
		if err != nil {
			return err
		}
		if contact == nil {
			return ErrContactNotFound
		}
	}
	if err := s.validateSalesperson(ctx, prospect.SalespersonID); err != nil {
		return err
	}
	if err := s.validateSalesGroup(ctx, prospect.SalesGroupID); err != nil {
		return err
	}
	return nil
}

func (s ProspectService) validatePromotionRefs(ctx context.Context, input PromotionInput) error {
	if err := s.validateSalesperson(ctx, input.SalespersonID); err != nil {
		return err
	}
	return s.validateSalesGroup(ctx, input.SalesGroupID)
}

func (s ProspectService) validateSalesperson(ctx context.Context, salespersonID *uint64) error {
	if salespersonID == nil {
		return nil
	}
	salesperson, err := s.contacts.Find(ctx, *salespersonID)
	if err != nil {
		return err
	}
	if salesperson == nil {
		return ErrSalespersonNotFound
	}
	return nil
}

func (s ProspectService) validateSalesGroup(ctx context.Context, salesTeamID *uint64) error {
	if salesTeamID == nil {
		return nil
	}
	team, err := s.teams.Find(ctx, *salesTeamID)
	if err != nil {
		return err
	}
	if team == nil {
		return ErrSalesGroupNotFound
	}
	return nil
}

func (s ProspectService) resolveStage(ctx context.Context, organizationID, stageID *uint64) (*reference.PipelineStage, error) {
	if stageID == nil {
		return nil, nil
	}
	stage, err := s.stages.Find(ctx, *stageID)
	if err != nil {
		return nil, err
	}
	if stage == nil {
		return nil, ErrStageNotFound
	}
	if organizationID != nil && stage.OrganizationID != nil && *stage.OrganizationID != *organizationID {
		return nil, ErrStageOrganization
	}
	return stage, nil
}

func (s ProspectService) wonStage(ctx context.Context, organizationID *uint64) (*reference.PipelineStage, error) {
	parsedQuery := &query.Query{}
	if organizationID != nil {
		parsedQuery.Filters = []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: *organizationID}}
	}
	page, err := s.stages.List(ctx, parsedQuery)
	if err != nil {
		return nil, err
	}
	for _, stage := range page.Items {
		if stage.IsWon {
			return stage, nil
		}
	}
	return nil, nil
}

func validLeadType(value string) bool {
	return value == ProspectKindLead || value == ProspectKindOpportunity
}

type PipelineStageService struct {
	stages dao.CRUD[reference.PipelineStage]
}

func NewPipelineStageService(stages dao.CRUD[reference.PipelineStage]) PipelineStageService {
	return PipelineStageService{stages: stages}
}

func (s PipelineStageService) List(ctx context.Context, q *query.Query) (*query.Page[reference.PipelineStage], error) {
	return s.stages.List(ctx, q)
}

func (s PipelineStageService) Find(ctx context.Context, id uint64) (*reference.PipelineStage, error) {
	return s.stages.Find(ctx, id)
}

func (s PipelineStageService) Create(ctx context.Context, stage *reference.PipelineStage) (*reference.PipelineStage, error) {
	return s.stages.Create(ctx, stage)
}

func (s PipelineStageService) Update(ctx context.Context, stage *reference.PipelineStage) (*reference.PipelineStage, error) {
	return s.stages.Update(ctx, stage)
}

func (s PipelineStageService) Delete(ctx context.Context, id uint64) error {
	return s.stages.Delete(ctx, id)
}

type SalesGroupService struct {
	teams dao.CRUD[reference.SalesGroup]
}

func NewSalesGroupService(teams dao.CRUD[reference.SalesGroup]) SalesGroupService {
	return SalesGroupService{teams: teams}
}

func (s SalesGroupService) List(ctx context.Context, q *query.Query) (*query.Page[reference.SalesGroup], error) {
	return s.teams.List(ctx, q)
}

func (s SalesGroupService) Find(ctx context.Context, id uint64) (*reference.SalesGroup, error) {
	return s.teams.Find(ctx, id)
}

func (s SalesGroupService) Create(ctx context.Context, team *reference.SalesGroup) (*reference.SalesGroup, error) {
	return s.teams.Create(ctx, team)
}

func (s SalesGroupService) Update(ctx context.Context, team *reference.SalesGroup) (*reference.SalesGroup, error) {
	return s.teams.Update(ctx, team)
}

func (s SalesGroupService) Delete(ctx context.Context, id uint64) error {
	return s.teams.Delete(ctx, id)
}
