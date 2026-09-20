package inventory

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func errMockDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("db down"))
	mock.ExpectExec(".*").WillReturnError(errors.New("db down"))
	return db
}

func TestInventoryDAO_Errors(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		call func(db *gorm.DB) error
	}{
		{name: "movement list by shipment", call: func(db *gorm.DB) error { _, err := NewStockMovementDAO(db).ListByShipment(ctx, 1); return err }},
		{name: "movement list by origin", call: func(db *gorm.DB) error {
			_, err := NewStockMovementDAO(db).ListByOrigin(ctx, "shipment", 1)
			return err
		}},
		{name: "movement ledger totals", call: func(db *gorm.DB) error { _, err := NewStockMovementDAO(db).LedgerTotals(ctx, nil); return err }},
		{name: "movement find for update", call: func(db *gorm.DB) error { _, err := NewStockMovementDAO(db).FindForUpdateTx(ctx, db, 1); return err }},
		{name: "movement apply all", call: func(db *gorm.DB) error { return NewStockMovementDAO(db).ApplyAllTx(ctx, db, []*StockMovement{{}}) }},
		{name: "quant find by key", call: func(db *gorm.DB) error { _, err := NewStockBalanceDAO(db).FindByKey(ctx, 1, 2, nil); return err }},
		{name: "quant list by item", call: func(db *gorm.DB) error { _, err := NewStockBalanceDAO(db).ListByItem(ctx, 1); return err }},
		{name: "quant list all", call: func(db *gorm.DB) error { _, err := NewStockBalanceDAO(db).ListAll(ctx, nil); return err }},
		{name: "quant upsert", call: func(db *gorm.DB) error { _, err := NewStockBalanceDAO(db).Upsert(ctx, nil, 1, 2, nil, 5); return err }},
		{name: "batch list by item", call: func(db *gorm.DB) error { _, err := NewBatchDAO(db).ListByItem(ctx, 1); return err }},
		{name: "batch list in org", call: func(db *gorm.DB) error { _, err := NewBatchDAO(db).ListInOrg(ctx, nil, 10); return err }},
		{name: "batch find in org", call: func(db *gorm.DB) error { _, err := NewBatchDAO(db).FindInOrg(ctx, 1, 10); return err }},
		{name: "reservation list by movement", call: func(db *gorm.DB) error { _, err := NewStockHoldDAO(db).ListByMovement(ctx, 1); return err }},
		{name: "reservation delete by movement", call: func(db *gorm.DB) error { return NewStockHoldDAO(db).DeleteByMove(ctx, 1) }},
		{name: "reservation reserve", call: func(db *gorm.DB) error { _, err := NewStockHoldDAO(db).Reserve(ctx, 1, nil, 5); return err }},
		{name: "reservation release", call: func(db *gorm.DB) error { return NewStockHoldDAO(db).Release(ctx, 1) }},
		{name: "reservation release by movement", call: func(db *gorm.DB) error { return NewStockHoldDAO(db).ReleaseByMovement(ctx, 1) }},
		{name: "reservation list in org", call: func(db *gorm.DB) error { _, err := NewStockHoldDAO(db).ListInOrg(ctx, nil, 10); return err }},
		{name: "reservation find in org", call: func(db *gorm.DB) error { _, err := NewStockHoldDAO(db).FindInOrg(ctx, 1, 10); return err }},
		{name: "reservation release by movement in org", call: func(db *gorm.DB) error { return NewStockHoldDAO(db).ReleaseByMovementInOrg(ctx, 10, 1) }},
		{name: "layer list open", call: func(db *gorm.DB) error { _, err := NewCostLayerDAO(db).ListOpenByItem(ctx, 1); return err }},
		{name: "layer value", call: func(db *gorm.DB) error { _, err := NewCostLayerDAO(db).ValueForItem(ctx, 1); return err }},
		{name: "layer list by movement", call: func(db *gorm.DB) error { _, err := NewCostLayerDAO(db).ListByMovement(ctx, 1); return err }},
		{name: "reorder list active", call: func(db *gorm.DB) error { _, err := NewReorderRuleDAO(db).ListActive(ctx); return err }},
		{name: "reorder list in org", call: func(db *gorm.DB) error { _, err := NewReorderRuleDAO(db).ListInOrg(ctx, nil, 10); return err }},
		{name: "reorder list active in org", call: func(db *gorm.DB) error { _, err := NewReorderRuleDAO(db).ListActiveInOrg(ctx, 10); return err }},
		{name: "reorder find in org", call: func(db *gorm.DB) error { _, err := NewReorderRuleDAO(db).FindInOrg(ctx, 1, 10); return err }},
		{name: "count lines", call: func(db *gorm.DB) error { _, err := NewStockCountLineDAO(db).ListByCount(ctx, 1); return err }},
		{name: "landed lines", call: func(db *gorm.DB) error { _, err := NewInboundCostLineDAO(db).ListByInboundCost(ctx, 1); return err }},
		{name: "landed adjustments", call: func(db *gorm.DB) error {
			_, err := NewInboundCostAdjustmentDAO(db).ListByInboundCost(ctx, 1)
			return err
		}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			helper.AssertError(t, tt.call(errMockDB(t)), true, nil)
		})
	}
}
