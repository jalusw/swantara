package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type ReorderRuleResponse struct {
	ID           uint64    `json:"id"`
	ItemID       uint64    `json:"item_id"`
	WarehouseID  *uint64   `json:"warehouse_id"`
	LocationID   *uint64   `json:"location_id"`
	MinQty       float64   `json:"min_qty"`
	MaxQty       float64   `json:"max_qty"`
	QtyMultiple  float64   `json:"qty_multiple"`
	LeadTimeDays *int      `json:"lead_time_days"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func newReorderRuleResponse(rule *inventory.ReorderRule) ReorderRuleResponse {
	return ReorderRuleResponse{
		ID:           rule.ID,
		ItemID:       rule.ItemID,
		WarehouseID:  rule.WarehouseID,
		LocationID:   rule.LocationID,
		MinQty:       rule.MinQty,
		MaxQty:       rule.MaxQty,
		QtyMultiple:  rule.QtyMultiple,
		LeadTimeDays: rule.LeadTimeDays,
		Active:       rule.Active,
		CreatedAt:    rule.CreatedAt,
		UpdatedAt:    rule.UpdatedAt,
	}
}

var reorderRuleQueryAllowlist = map[string]struct{}{
	"item_id":      {},
	"warehouse_id": {},
	"location_id":  {},
	"active":       {},
	"created_at":   {},
	"updated_at":   {},
}

type ListReorderRulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListReorderRulesResponse `json:"data"`
}
type ListReorderRulesResponse struct {
	Rules []ReorderRuleResponse `json:"rules"`
}

type GetReorderRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetReorderRuleResponse `json:"data"`
}
type GetReorderRuleResponse struct {
	Rule ReorderRuleResponse `json:"rule"`
}

type ReorderRuleRequest struct {
	ItemID       uint64  `json:"item_id" validate:"required,gt=0"`
	WarehouseID  *uint64 `json:"warehouse_id"`
	LocationID   *uint64 `json:"location_id"`
	MinQty       float64 `json:"min_qty"`
	MaxQty       float64 `json:"max_qty"`
	QtyMultiple  float64 `json:"qty_multiple"`
	LeadTimeDays *int    `json:"lead_time_days"`
	Active       *bool   `json:"active"`
}

type ReorderRuleResponseBodyEnvelope struct {
	httpx.EnvelopeBase
	Data ReorderRuleResponseBody `json:"data"`
}
type ReorderRuleResponseBody struct {
	Rule ReorderRuleResponse `json:"rule"`
}
