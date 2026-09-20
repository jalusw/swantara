package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
)

type ConsolidationHandler struct {
	svc interorganization.ConsolidationService
}

func NewConsolidationHandler(svc interorganization.ConsolidationService) ConsolidationHandler {
	return ConsolidationHandler{svc: svc}
}

type ConsolidationRunResponse struct {
	ID                  uint64  `json:"id"`
	GroupOrganizationID *uint64 `json:"group_organization_id"`
	PeriodID            *uint64 `json:"period_id"`
	ReportingCurrency   string  `json:"reporting_currency"`
	State               string  `json:"state"`
}

func newConsolidationRunResponse(run *interorganization.ConsolidationRun) ConsolidationRunResponse {
	return ConsolidationRunResponse{
		ID: run.ID, GroupOrganizationID: run.GroupOrganizationID, PeriodID: run.PeriodID,
		ReportingCurrency: run.ReportingCurrency, State: run.State,
	}
}

type ConsolidationEliminationResponse struct {
	ID                         uint64  `json:"id"`
	AccountID                  uint64  `json:"account_id"`
	CounterpartyOrganizationID *uint64 `json:"counterparty_organization_id"`
	Amount                     float64 `json:"amount"`
	Description                string  `json:"description"`
}

func newConsolidationEliminationResponse(e *interorganization.ConsolidationElimination) ConsolidationEliminationResponse {
	return ConsolidationEliminationResponse{
		ID: e.ID, AccountID: e.AccountID, CounterpartyOrganizationID: e.CounterpartyOrganizationID,
		Amount: e.Amount, Description: e.Description,
	}
}

type ConsolidatedBalanceResponse struct {
	OrganizationID uint64  `json:"organization_id"`
	AccountID      uint64  `json:"account_id"`
	Amount         float64 `json:"amount"`
}
