package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
)

type ReconcileRuleHandler struct {
	engine accounting.ReconcileRuleEngine
}

func NewReconcileRuleHandler(engine accounting.ReconcileRuleEngine) ReconcileRuleHandler {
	return ReconcileRuleHandler{engine: engine}
}

type ReconcileRuleResponse struct {
	ID              uint64    `json:"id"`
	OrganizationID  uint64    `json:"organization_id"`
	Name            *string   `json:"name"`
	AccountID       *uint64   `json:"account_id"`
	MatchContact    bool      `json:"match_contact"`
	MatchAmount     bool      `json:"match_amount"`
	MatchRef        bool      `json:"match_ref"`
	AmountTolerance float64   `json:"amount_tolerance"`
	Active          bool      `json:"active"`
	Sequence        int       `json:"sequence"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func newReconcileRuleResponse(r *accounting.ReconcileRule) ReconcileRuleResponse {
	return ReconcileRuleResponse{
		ID:              r.ID,
		OrganizationID:  r.OrganizationID,
		Name:            r.Name,
		AccountID:       r.AccountID,
		MatchContact:    r.MatchContact,
		MatchAmount:     r.MatchAmount,
		MatchRef:        r.MatchRef,
		AmountTolerance: r.AmountTolerance,
		Active:          r.Active,
		Sequence:        r.Sequence,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

var reconcileRuleQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}
