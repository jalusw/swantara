package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type BudgetLineResponse struct {
	ID              uint64  `json:"id"`
	BudgetID        uint64  `json:"budget_id"`
	AccountID       uint64  `json:"account_id"`
	DimensionID     *uint64 `json:"dimension_id"`
	PlannedAmount   float64 `json:"planned_amount"`
	PracticalAmount float64 `json:"practical_amount"`
}

func newBudgetLineResponse(line *accounting.BudgetLine) BudgetLineResponse {
	return BudgetLineResponse{
		ID:              line.ID,
		BudgetID:        line.BudgetID,
		AccountID:       line.AccountID,
		DimensionID:     line.DimensionID,
		PlannedAmount:   line.PlannedAmount,
		PracticalAmount: line.PracticalAmount,
	}
}

type BudgetResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID *uint64   `json:"organization_id"`
	Name           *string   `json:"name"`
	DateStart      *string   `json:"date_start"`
	DateEnd        *string   `json:"date_end"`
	State          string    `json:"state"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newBudgetResponse(budget *accounting.Budget) BudgetResponse {
	return BudgetResponse{
		ID:             budget.ID,
		OrganizationID: budget.OrganizationID,
		Name:           budget.Name,
		DateStart:      helper.FormatDatePtr(budget.DateStart),
		DateEnd:        helper.FormatDatePtr(budget.DateEnd),
		State:          budget.State,
		CreatedAt:      budget.CreatedAt,
		UpdatedAt:      budget.UpdatedAt,
	}
}

type ListBudgetsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListBudgetsResponse `json:"data"`
}

type ListBudgetsResponse struct {
	Budgets []BudgetResponse `json:"budgets"`
}

type GetBudgetResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetBudgetResponse `json:"data"`
}

type GetBudgetResponse struct {
	Budget BudgetResponse       `json:"budget"`
	Lines  []BudgetLineResponse `json:"lines"`
}

type CreateBudgetLineRequest struct {
	AccountID     uint64  `json:"account_id" validate:"required,gt=0"`
	DimensionID   *uint64 `json:"dimension_id"`
	PlannedAmount float64 `json:"planned_amount" validate:"gte=0"`
}

type CreateBudgetRequest struct {
	OrganizationID *uint64                   `json:"organization_id"`
	Name           string                    `json:"name" validate:"required"`
	DateStart      string                    `json:"date_start" validate:"required"`
	DateEnd        string                    `json:"date_end" validate:"required"`
	Lines          []CreateBudgetLineRequest `json:"lines" validate:"required,min=1,dive"`
}

type CreateBudgetResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateBudgetResponse `json:"data"`
}

type CreateBudgetResponse struct {
	Budget BudgetResponse `json:"budget"`
}

type BudgetVarianceResponseEnvelope struct {
	httpx.EnvelopeBase
	Data BudgetVarianceResponse `json:"data"`
}

type BudgetVarianceResponse struct {
	Variance []accounting.BudgetVarianceLine `json:"variance"`
}
