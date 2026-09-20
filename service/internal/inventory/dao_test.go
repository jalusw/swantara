package inventory

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestDAOs_Constructors(t *testing.T) {
	db, _ := query.NewMockDB(t)

	tests := []struct {
		name string
		fn   func() interface{}
	}{
		{"NewWarehouseDAO", func() interface{} { return NewWarehouseDAO(db) }},
		{"NewStockLocationDAO", func() interface{} { return NewStockLocationDAO(db) }},
		{"NewShipmentDAO", func() interface{} { return NewShipmentDAO(db) }},
		{"NewStockMovementDAO", func() interface{} { return NewStockMovementDAO(db) }},
		{"NewStockBalanceDAO", func() interface{} { return NewStockBalanceDAO(db) }},
		{"NewBatchDAO", func() interface{} { return NewBatchDAO(db) }},
		{"NewStockHoldDAO", func() interface{} { return NewStockHoldDAO(db) }},
		{"NewCostLayerDAO", func() interface{} { return NewCostLayerDAO(db) }},
		{"NewReorderRuleDAO", func() interface{} { return NewReorderRuleDAO(db) }},
		{"NewStockCountDAO", func() interface{} { return NewStockCountDAO(db) }},
		{"NewStockCountLineDAO", func() interface{} { return NewStockCountLineDAO(db) }},
		{"NewWarehouseTransferDAO", func() interface{} { return NewWarehouseTransferDAO(db) }},
		{"NewInboundCostDAO", func() interface{} { return NewInboundCostDAO(db) }},
		{"NewInboundCostLineDAO", func() interface{} { return NewInboundCostLineDAO(db) }},
		{"NewInboundCostAdjustmentDAO", func() interface{} { return NewInboundCostAdjustmentDAO(db) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got == nil {
				t.Errorf("%s() = nil", tt.name)
			}
		})
	}
}

func TestShipmentDAO_CreateWithMovements(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		shipment  *Shipment
		movements []*StockMovement
		wantID    uint64
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shipments"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_movements"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
				mock.ExpectCommit()
			},
			shipment:  &Shipment{Name: helper.Ptr("P-1")},
			movements: []*StockMovement{{ItemID: 100, Qty: 5}},
			wantID:    1,
		},
		{
			name: "insert error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shipments"`)).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
			},
			shipment:  &Shipment{},
			movements: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewShipmentDAO(db)
			created, err := dao.CreateWithMovements(context.Background(), tt.shipment, tt.movements)
			if tt.wantErr {
				if err == nil {
					t.Fatal("CreateWithMovements() expected error")
				}
			} else {
				if err != nil {
					t.Fatalf("CreateWithMovements() error = %v", err)
				}
				if created != tt.shipment || tt.shipment.ID != tt.wantID {
					t.Errorf("CreateWithMovements() = %+v, want shipment id %d", created, tt.wantID)
				}
				if created.State != "" {
					t.Errorf("shipment state = %q, want empty", created.State)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockMovementDAO_CreateTx(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantID    uint64
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_movements"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
				mock.ExpectCommit()
			},
			wantID: 3,
		},
		{
			name: "insert error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_movements"`)).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)
			tx := db.Begin()

			dao := NewStockMovementDAO(db)
			movement := &StockMovement{ItemID: 100, Qty: 5}
			created, err := dao.CreateTx(context.Background(), tx, movement)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error when insert fails")
				}
				tx.Rollback()
			} else {
				if err != nil {
					t.Fatalf("CreateTx() error = %v", err)
				}
				if created.ID != tt.wantID {
					t.Errorf("CreateTx() id = %d, want %d", created.ID, tt.wantID)
				}
				tx.Commit()
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockMovementDAO_ListByShipment(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(mock sqlmock.Sqlmock)
		shipmentID uint64
		wantCount  int
		wantErr    bool
	}{
		{
			name: "returns movements for shipment",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "stock_movements" WHERE shipment_id = $1`)).
					WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_movements" WHERE shipment_id = $1`)).
					WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 100))
			},
			shipmentID: 5,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewStockMovementDAO(db)
			movements, err := dao.ListByShipment(context.Background(), tt.shipmentID)
			if err != nil {
				t.Fatalf("ListByShipment() error = %v", err)
			}
			if len(movements) != tt.wantCount || movements[0].ItemID != 100 {
				t.Errorf("ListByShipment() = %+v, want item 100", movements)
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockMovementDAO_ListByOrigin(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantMoves int
		wantErr   bool
	}{
		{
			name: "returns movements",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_movements" WHERE origin_type = $1 AND origin_id = $2 AND deleted_at IS NULL`)).
					WithArgs("purchase_order", 9).
					WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 100))
			},
			wantMoves: 1,
		},
		{
			name: "db error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_movements"`)).
					WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewStockMovementDAO(db)
			movements, err := dao.ListByOrigin(context.Background(), "purchase_order", 9)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ListByOrigin() expected error")
				}
			} else {
				if err != nil {
					t.Fatalf("ListByOrigin() error = %v", err)
				}
				if len(movements) != tt.wantMoves || movements[0].ItemID != 100 {
					t.Errorf("ListByOrigin() = %+v, want item 100", movements)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockMovementDAO_LedgerTotals(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(mock sqlmock.Sqlmock)
		organizationID *uint64
		wantLen        int
		wantItemID     uint64
		wantTotal      float64
	}{
		{
			name: "without organization",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT item_id, location_id, batch_id, SUM(qty) AS total`)).
					WithArgs(MovementStateDone, MovementStateDone).
					WillReturnRows(sqlmock.NewRows([]string{"item_id", "location_id", "batch_id", "total"}).
						AddRow(100, 10, nil, 50).AddRow(100, 20, 7, 12))
			},
			organizationID: nil,
			wantLen:        2,
			wantItemID:     100,
			wantTotal:      50,
		},
		{
			name: "with organization",
			setupMock: func(mock sqlmock.Sqlmock) {
				organizationID := uint64(3)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT item_id, location_id, batch_id, SUM(qty) AS total`)).
					WithArgs(MovementStateDone, organizationID, MovementStateDone, organizationID).
					WillReturnRows(sqlmock.NewRows([]string{"item_id", "location_id", "batch_id", "total"}).
						AddRow(100, 10, nil, 50))
			},
			organizationID: func() *uint64 { v := uint64(3); return &v }(),
			wantLen:        1,
			wantItemID:     100,
			wantTotal:      50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewStockMovementDAO(db)
			totals, err := dao.LedgerTotals(context.Background(), tt.organizationID)
			if err != nil {
				t.Fatalf("LedgerTotals() error = %v", err)
			}
			if len(totals) != tt.wantLen || totals[0].ItemID != tt.wantItemID || totals[0].Total != tt.wantTotal {
				t.Errorf("LedgerTotals() = %+v, want item %d total %v", totals, tt.wantItemID, tt.wantTotal)
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockMovementDAO_FindForUpdateTx(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(mock sqlmock.Sqlmock)
		movementID uint64
		wantState  string
		wantNil    bool
	}{
		{
			name: "locks and returns movement",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_movements" WHERE "stock_movements"."id" = $1 ORDER BY "stock_movements"."id" LIMIT $2 FOR UPDATE`)).
					WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "state"}).AddRow(1, "draft"))
			},
			movementID: 1,
			wantState:  MovementStateDraft,
		},
		{
			name: "not found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_movements" WHERE "stock_movements"."id" = $1 ORDER BY "stock_movements"."id" LIMIT $2 FOR UPDATE`)).
					WithArgs(9, 1).WillReturnError(gorm.ErrRecordNotFound)
			},
			movementID: 9,
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewStockMovementDAO(db)
			movement, err := dao.FindForUpdateTx(context.Background(), db, tt.movementID)
			if err != nil {
				t.Fatalf("FindForUpdateTx() error = %v", err)
			}
			if tt.wantNil {
				if movement != nil {
					t.Errorf("FindForUpdateTx() = %+v, want nil", movement)
				}
			} else {
				if movement == nil || movement.ID != tt.movementID || movement.State != tt.wantState {
					t.Errorf("FindForUpdateTx() = %+v, want draft movement %d", movement, tt.movementID)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockMovementDAO_ApplyTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	doneAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_movements" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances" WHERE (item_id = $1 AND location_id = $2) AND batch_id IS NULL ORDER BY "stock_balances"."id" LIMIT $3 FOR UPDATE`)).
		WithArgs(100, 10, 1).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_balances"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances" WHERE (item_id = $1 AND location_id = $2) AND batch_id IS NULL ORDER BY "stock_balances"."id" LIMIT $3 FOR UPDATE`)).
		WithArgs(100, 20, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "quantity"}).AddRow(2, 30))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_balances" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	dao := NewStockMovementDAO(db)
	var applied *StockMovement
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		movement := &StockMovement{Base: model.Base{ID: 1}, ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 20}
		movement.State = MovementStateDone
		movement.DateDone = &doneAt
		applied, err = dao.ApplyTx(ctx, tx, movement)
		return err
	})
	if err != nil {
		t.Fatalf("ApplyTx() error = %v", err)
	}
	if applied.State != MovementStateDone || applied.DateDone == nil {
		t.Errorf("ApplyTx() = %+v, want state done and date set", applied)
	}
	query.AssertDBMockDone(t, mock)
}

func TestStockMovementDAO_ApplyAllTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	doneAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_movements" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances"`)).
		WithArgs(100, 10, 1).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_balances"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances"`)).
		WithArgs(100, 20, 1).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_balances"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	dao := NewStockMovementDAO(db)
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		movement := &StockMovement{Base: model.Base{ID: 1}, ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 20}
		movement.State = MovementStateDone
		movement.DateDone = &doneAt
		return dao.ApplyAllTx(ctx, tx, []*StockMovement{movement})
	})
	if err != nil {
		t.Fatalf("ApplyAllTx() error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestStockBalanceDAO_FindByKey(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		batchID   *uint64
		wantNil   bool
		wantQty   float64
	}{
		{
			name: "not found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances" WHERE (item_id = $1 AND location_id = $2) AND batch_id IS NULL ORDER BY "stock_balances"."id" LIMIT $3`)).
					WithArgs(100, 10, 1).WillReturnError(gorm.ErrRecordNotFound)
			},
			batchID: nil,
			wantNil: true,
		},
		{
			name: "finds by batch",
			setupMock: func(mock sqlmock.Sqlmock) {
				batchID := uint64(7)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances" WHERE (item_id = $1 AND location_id = $2) AND batch_id = $3 ORDER BY "stock_balances"."id" LIMIT $4`)).
					WithArgs(100, 10, batchID, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "quantity"}).AddRow(5, 25))
			},
			batchID: func() *uint64 { v := uint64(7); return &v }(),
			wantQty: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewStockBalanceDAO(db)
			quant, err := dao.FindByKey(context.Background(), 100, 10, tt.batchID)
			if err != nil {
				t.Fatalf("FindByKey() error = %v", err)
			}
			if tt.wantNil {
				if quant != nil {
					t.Errorf("FindByKey() = %+v, want nil", quant)
				}
			} else {
				if quant == nil || quant.Quantity != tt.wantQty {
					t.Errorf("FindByKey() = %+v, want quantity %v", quant, tt.wantQty)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockBalanceDAO_Upsert(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		qty       float64
		wantQty   float64
	}{
		{
			name: "creates missing quant",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances"`)).
					WithArgs(100, 10, 1).WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_balances"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			qty:     5,
			wantQty: 5,
		},
		{
			name: "updates existing quant",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances"`)).
					WithArgs(100, 10, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "quantity"}).AddRow(1, 10))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_balances" SET`)).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			qty:     5,
			wantQty: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewStockBalanceDAO(db)
			quant, err := dao.Upsert(context.Background(), nil, 100, 10, nil, tt.qty)
			if err != nil {
				t.Fatalf("Upsert() error = %v", err)
			}
			if quant.Quantity != tt.wantQty {
				t.Errorf("Upsert() quantity = %v, want %v", quant.Quantity, tt.wantQty)
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockBalanceDAO_UpsertTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	tx := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances"`)).
		WithArgs(100, 20, 1).WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_balances"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	dao := NewStockBalanceDAO(db)
	quant, err := dao.UpsertTx(ctx, tx, nil, 100, 20, nil, -3)
	if err != nil {
		t.Fatalf("UpsertTx() error = %v", err)
	}
	if quant.Quantity != -3 {
		t.Errorf("UpsertTx() quantity = %v, want -3", quant.Quantity)
	}
	tx.Commit()
	query.AssertDBMockDone(t, mock)
}

func TestStockBalanceDAO_ListByItem(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "stock_balances" WHERE item_id = $1`)).
		WithArgs(100).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances" WHERE item_id = $1`)).
		WithArgs(100).WillReturnRows(sqlmock.NewRows([]string{"id", "quantity"}).AddRow(1, 5))

	dao := NewStockBalanceDAO(db)
	quants, err := dao.ListByItem(ctx, 100)
	if err != nil {
		t.Fatalf("ListByItem() error = %v", err)
	}
	if len(quants) != 1 || quants[0].Quantity != 5 {
		t.Errorf("ListByItem() = %+v, want quantity 5", quants)
	}
	query.AssertDBMockDone(t, mock)
}

func TestStockBalanceDAO_ListAll(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(mock sqlmock.Sqlmock)
		organizationID *uint64
		wantLen        int
	}{
		{
			name: "without organization",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "quantity"}).AddRow(1, 5).AddRow(2, 3))
			},
			organizationID: nil,
			wantLen:        2,
		},
		{
			name: "with organization",
			setupMock: func(mock sqlmock.Sqlmock) {
				organizationID := uint64(3)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_balances" WHERE organization_id = $1`)).
					WithArgs(organizationID).WillReturnRows(sqlmock.NewRows([]string{"id", "quantity"}).AddRow(1, 5))
			},
			organizationID: func() *uint64 { v := uint64(3); return &v }(),
			wantLen:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewStockBalanceDAO(db)
			quants, err := dao.ListAll(context.Background(), tt.organizationID)
			if err != nil {
				t.Fatalf("ListAll() error = %v", err)
			}
			if len(quants) != tt.wantLen {
				t.Errorf("ListAll() = %d quants, want %d", len(quants), tt.wantLen)
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestBatchDAO_ListByItem(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "batches" WHERE item_id = $1`)).
		WithArgs(100).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "batches" WHERE item_id = $1`)).
		WithArgs(100).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "LOT-1"))

	dao := NewBatchDAO(db)
	lots, err := dao.ListByItem(ctx, 100)
	if err != nil {
		t.Fatalf("ListByItem() error = %v", err)
	}
	if len(lots) != 1 || lots[0].Name != "LOT-1" {
		t.Errorf("ListByItem() = %+v, want LOT-1", lots)
	}
	query.AssertDBMockDone(t, mock)
}

func TestBatchDAO_ListInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "batches"`)).
		WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "batches" JOIN`)).
		WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "LOT-1"))

	dao := NewBatchDAO(db)
	page, err := dao.ListInOrg(ctx, nil, 3)
	if err != nil {
		t.Fatalf("ListInOrg() error = %v", err)
	}
	if page.Count != 1 || len(page.Items) != 1 {
		t.Errorf("ListInOrg() = %+v, want 1 item", page)
	}
	query.AssertDBMockDone(t, mock)
}

func TestBatchDAO_FindInOrg(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantNil   bool
		wantName  string
	}{
		{
			name: "found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`FROM "batches" JOIN`)).
					WithArgs(3, 1, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "LOT-1"))
			},
			wantName: "LOT-1",
		},
		{
			name: "not found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`FROM "batches" JOIN`)).
					WithArgs(3, 1, 1).WillReturnError(gorm.ErrRecordNotFound)
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewBatchDAO(db)
			batch, err := dao.FindInOrg(context.Background(), 1, 3)
			if err != nil {
				t.Fatalf("FindInOrg() error = %v", err)
			}
			if tt.wantNil {
				if batch != nil {
					t.Errorf("FindInOrg() = %+v, want nil", batch)
				}
			} else {
				if batch == nil || batch.Name != tt.wantName {
					t.Errorf("FindInOrg() = %+v, want %s", batch, tt.wantName)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockHoldDAO_ListByMovement(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "stock_holds" WHERE movement_id = $1`)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_holds" WHERE movement_id = $1`)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "balance_id"}).AddRow(1, 2))

	dao := NewStockHoldDAO(db)
	reservations, err := dao.ListByMovement(ctx, 5)
	if err != nil {
		t.Fatalf("ListByMovement() error = %v", err)
	}
	if len(reservations) != 1 || reservations[0].BalanceID != 2 {
		t.Errorf("ListByMovement() = %+v, want quant 2", reservations)
	}
	query.AssertDBMockDone(t, mock)
}

func TestStockHoldDAO_DeleteByMove(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "stock_holds" WHERE movement_id = $1`)).
		WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	dao := NewStockHoldDAO(db)
	if err := dao.DeleteByMove(ctx, 5); err != nil {
		t.Fatalf("DeleteByMove() error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestStockHoldDAO_Reserve(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
		wantQty   float64
		wantQuant uint64
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_balances" SET "reserved_qty"=reserved_qty + $1,"updated_at"=$2 WHERE id = $3 AND quantity - reserved_qty >= $4`)).
					WithArgs(float64(3), sqlmock.AnyArg(), 2, float64(3)).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_holds"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			wantQty:   3,
			wantQuant: 2,
		},
		{
			name: "overflow conflict",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_balances" SET "reserved_qty"=reserved_qty + $1,"updated_at"=$2`)).
					WithArgs(float64(3), sqlmock.AnyArg(), 2, float64(3)).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
			wantErr: ErrHoldConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)
			ctx := context.Background()

			dao := NewStockHoldDAO(db)
			reservation, err := dao.Reserve(ctx, 2, nil, 3)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Reserve() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("Reserve() error = %v", err)
				}
				if reservation.Qty != tt.wantQty || reservation.BalanceID != tt.wantQuant {
					t.Errorf("Reserve() = %+v, want qty %v on quant %d", reservation, tt.wantQty, tt.wantQuant)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockHoldDAO_Release(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_holds"`)).
					WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "balance_id", "qty"}).AddRow(1, 2, 3))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_balances" SET "reserved_qty"=reserved_qty - $1,"updated_at"=$2 WHERE id = $3`)).
					WithArgs(float64(3), sqlmock.AnyArg(), 2).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "stock_holds" WHERE "stock_holds"."id" = $1`)).
					WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
		},
		{
			name: "not found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_holds"`)).
					WithArgs(1, 1).WillReturnError(gorm.ErrRecordNotFound)
				mock.ExpectRollback()
			},
			wantErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewStockHoldDAO(db)
			err := dao.Release(context.Background(), 1)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Release() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("Release() error = %v", err)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockHoldDAO_ReleaseByMovement(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_holds" WHERE movement_id = $1`)).
					WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "balance_id", "qty"}).
					AddRow(1, 2, 3).AddRow(2, 4, 1))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_balances" SET "reserved_qty"=reserved_qty - $1,"updated_at"=$2 WHERE id = $3`)).
					WithArgs(float64(3), sqlmock.AnyArg(), 2).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_balances" SET "reserved_qty"=reserved_qty - $1,"updated_at"=$2 WHERE id = $3`)).
					WithArgs(float64(1), sqlmock.AnyArg(), 4).WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "stock_holds" WHERE movement_id = $1`)).
					WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 2))
				mock.ExpectCommit()
			},
		},
		{
			name: "find error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_holds" WHERE movement_id = $1`)).
					WithArgs(5).WillReturnError(errors.New("db down"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewStockHoldDAO(db)
			err := dao.ReleaseByMovement(context.Background(), 5)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ReleaseByMovement() expected error")
				}
			} else {
				if err != nil {
					t.Fatalf("ReleaseByMovement() error = %v", err)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestStockHoldDAO_ReleaseByMovementInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "stock_holds" JOIN`)).
		WithArgs(3, 5).WillReturnRows(sqlmock.NewRows([]string{"id", "balance_id", "qty"}).
		AddRow(1, 2, 3))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_balances" SET "reserved_qty"=reserved_qty - $1,"updated_at"=$2 WHERE id = $3`)).
		WithArgs(float64(3), sqlmock.AnyArg(), 2).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "stock_holds" WHERE movement_id = $1`)).
		WithArgs(5).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	dao := NewStockHoldDAO(db)
	if err := dao.ReleaseByMovementInOrg(ctx, 3, 5); err != nil {
		t.Fatalf("ReleaseByMovementInOrg() error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

func TestStockHoldDAO_ListInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "stock_holds"`)).
		WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "stock_holds" JOIN`)).
		WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"id", "balance_id"}).AddRow(1, 2))

	dao := NewStockHoldDAO(db)
	page, err := dao.ListInOrg(ctx, nil, 3)
	if err != nil {
		t.Fatalf("ListInOrg() error = %v", err)
	}
	if page.Count != 1 || len(page.Items) != 1 {
		t.Errorf("ListInOrg() = %+v, want 1 item", page)
	}
	query.AssertDBMockDone(t, mock)
}

func TestStockHoldDAO_FindInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "stock_holds" JOIN`)).
		WithArgs(3, 1, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "balance_id"}).AddRow(1, 2))

	dao := NewStockHoldDAO(db)
	reservation, err := dao.FindInOrg(ctx, 1, 3)
	if err != nil {
		t.Fatalf("FindInOrg() error = %v", err)
	}
	if reservation == nil || reservation.BalanceID != 2 {
		t.Errorf("FindInOrg() = %+v, want quant 2", reservation)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCostLayerDAO_ListOpenByItem(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cost_layers" WHERE item_id = $1 AND remaining_qty <> 0 AND deleted_at IS NULL ORDER BY id ASC`)).
		WithArgs(100).WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_qty"}).AddRow(1, 10))

	dao := NewCostLayerDAO(db)
	layers, err := dao.ListOpenByItem(ctx, 100)
	if err != nil {
		t.Fatalf("ListOpenByItem() error = %v", err)
	}
	if len(layers) != 1 || layers[0].RemainingQty != 10 {
		t.Errorf("ListOpenByItem() = %+v, want remaining 10", layers)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCostLayerDAO_ListOpenByItemInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "cost_layers" JOIN`)).
		WithArgs(100, 3).WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_qty"}).AddRow(1, 10))

	dao := NewCostLayerDAO(db)
	layers, err := dao.ListOpenByItemInOrg(ctx, 100, 3)
	if err != nil {
		t.Fatalf("ListOpenByItemInOrg() error = %v", err)
	}
	if len(layers) != 1 {
		t.Errorf("ListOpenByItemInOrg() = %d layers, want 1", len(layers))
	}
	query.AssertDBMockDone(t, mock)
}

func TestCostLayerDAO_ValueForItem(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		wantVal   float64
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(remaining_value), 0)`)).
					WithArgs(100).WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(250))
			},
			wantVal: 250,
		},
		{
			name: "db error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(remaining_value), 0)`)).
					WithArgs(100).WillReturnError(errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := query.NewMockDB(t)
			tt.setupMock(mock)

			dao := NewCostLayerDAO(db)
			total, err := dao.ValueForItem(context.Background(), 100)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ValueForItem() expected error")
				}
			} else {
				if err != nil {
					t.Fatalf("ValueForItem() error = %v", err)
				}
				if total != tt.wantVal {
					t.Errorf("ValueForItem() = %v, want %v", total, tt.wantVal)
				}
			}
			query.AssertDBMockDone(t, mock)
		})
	}
}

func TestCostLayerDAO_ListByMovement(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cost_layers" WHERE movement_id = $1 AND deleted_at IS NULL ORDER BY id ASC`)).
		WithArgs(5).WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_qty"}).AddRow(1, 10))

	dao := NewCostLayerDAO(db)
	layers, err := dao.ListByMovement(ctx, 5)
	if err != nil {
		t.Fatalf("ListByMovement() error = %v", err)
	}
	if len(layers) != 1 {
		t.Errorf("ListByMovement() = %d layers, want 1", len(layers))
	}
	query.AssertDBMockDone(t, mock)
}

func TestCostLayerDAO_CreateTxAndUpdateTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	tx := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "cost_layers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "cost_layers" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	dao := NewCostLayerDAO(db)
	layer := &CostLayer{ItemID: 100, RemainingQty: 10}
	created, err := dao.CreateTx(ctx, tx, layer)
	if err != nil {
		t.Fatalf("CreateTx() error = %v", err)
	}
	if _, err := dao.UpdateTx(ctx, tx, created); err != nil {
		t.Fatalf("UpdateTx() error = %v", err)
	}
	tx.Commit()
	query.AssertDBMockDone(t, mock)
}

func TestCostLayerDAO_ListOpenByItemForUpdateTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "cost_layers" WHERE item_id = $1 AND remaining_qty <> 0 AND deleted_at IS NULL ORDER BY id ASC FOR UPDATE`)).
		WithArgs(100).WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_qty"}).AddRow(3, 10))

	dao := NewCostLayerDAO(db)
	layers, err := dao.ListOpenByItemForUpdateTx(ctx, db, 100)
	if err != nil {
		t.Fatalf("ListOpenByItemForUpdateTx() error = %v", err)
	}
	if len(layers) != 1 || layers[0].ID != 3 || layers[0].RemainingQty != 10 {
		t.Errorf("ListOpenByItemForUpdateTx() = %+v, want single layer 3", layers)
	}
	query.AssertDBMockDone(t, mock)
}

func TestReorderRuleDAO_ListActive(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reorder_rules" WHERE active = $1 AND deleted_at IS NULL`)).
		WithArgs(true).WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 100))

	dao := NewReorderRuleDAO(db)
	rules, err := dao.ListActive(ctx)
	if err != nil {
		t.Fatalf("ListActive() error = %v", err)
	}
	if len(rules) != 1 || rules[0].ItemID != 100 {
		t.Errorf("ListActive() = %+v, want item 100", rules)
	}
	query.AssertDBMockDone(t, mock)
}

func TestReorderRuleDAO_ListInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "reorder_rules"`)).
		WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "reorder_rules" JOIN`)).
		WithArgs(3).WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 100))

	dao := NewReorderRuleDAO(db)
	page, err := dao.ListInOrg(ctx, nil, 3)
	if err != nil {
		t.Fatalf("ListInOrg() error = %v", err)
	}
	if page.Count != 1 || len(page.Items) != 1 {
		t.Errorf("ListInOrg() = %+v, want 1 item", page)
	}
	query.AssertDBMockDone(t, mock)
}

func TestReorderRuleDAO_ListActiveInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "reorder_rules" JOIN`)).
		WithArgs(3, true).WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 100))

	dao := NewReorderRuleDAO(db)
	rules, err := dao.ListActiveInOrg(ctx, 3)
	if err != nil {
		t.Fatalf("ListActiveInOrg() error = %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("ListActiveInOrg() = %d rules, want 1", len(rules))
	}
	query.AssertDBMockDone(t, mock)
}

func TestReorderRuleDAO_FindInOrg(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "reorder_rules" JOIN`)).
		WithArgs(3, 1, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 100))

	dao := NewReorderRuleDAO(db)
	rule, err := dao.FindInOrg(ctx, 1, 3)
	if err != nil {
		t.Fatalf("FindInOrg() error = %v", err)
	}
	if rule == nil || rule.ItemID != 100 {
		t.Errorf("FindInOrg() = %+v, want item 100", rule)
	}
	query.AssertDBMockDone(t, mock)
}

func TestStockCountDAO_CreateWithLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_counts"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_count_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	dao := NewStockCountDAO(db)
	count := &StockCount{Name: helper.Ptr("C-1")}
	created, err := dao.CreateWithLines(ctx, count, []*StockCountLine{{ItemID: 100}})
	if err != nil {
		t.Fatalf("CreateWithLines() error = %v", err)
	}
	if created != count || count.ID != 1 {
		t.Errorf("CreateWithLines() = %+v, want id 1", created)
	}
	query.AssertDBMockDone(t, mock)
}

func TestStockCountDAO_UpdateTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	tx := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_counts" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	dao := NewStockCountDAO(db)
	count := &StockCount{Base: model.Base{ID: 1}, State: CountStatePosted}
	updated, err := dao.UpdateTx(ctx, tx, count)
	if err != nil {
		t.Fatalf("UpdateTx() error = %v", err)
	}
	if updated.State != CountStatePosted {
		t.Errorf("UpdateTx() state = %q, want posted", updated.State)
	}
	tx.Commit()
	query.AssertDBMockDone(t, mock)
}

func TestStockCountLineDAO_ListByCount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "stock_count_lines" WHERE stock_count_id = $1`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_count_lines" WHERE stock_count_id = $1`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 100))

	dao := NewStockCountLineDAO(db)
	lines, err := dao.ListByCount(ctx, 1)
	if err != nil {
		t.Fatalf("ListByCount() error = %v", err)
	}
	if len(lines) != 1 || lines[0].ItemID != 100 {
		t.Errorf("ListByCount() = %+v, want item 100", lines)
	}
	query.AssertDBMockDone(t, mock)
}

func TestWarehouseTransferDAO_CreateWithShipments(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shipments"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_movements"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "shipments"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_movements"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "warehouse_transfers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectCommit()

	dao := NewWarehouseTransferDAO(db)
	transfer := &WarehouseTransfer{SrcWarehouseID: 1, DstWarehouseID: 2}
	created, err := dao.CreateWithShipments(ctx, transfer,
		&Shipment{Type: ShipmentTypeInternal}, []*StockMovement{{ItemID: 100}},
		&Shipment{Type: ShipmentTypeInternal}, []*StockMovement{{ItemID: 200}})
	if err != nil {
		t.Fatalf("CreateWithShipments() error = %v", err)
	}
	if created != transfer || transfer.ID != 5 {
		t.Errorf("CreateWithShipments() = %+v, want id 5", created)
	}
	if transfer.OutShipmentID == nil || *transfer.OutShipmentID != 1 || transfer.InShipmentID == nil || *transfer.InShipmentID != 3 {
		t.Errorf("shipment ids = out %v in %v, want 1/3", transfer.OutShipmentID, transfer.InShipmentID)
	}
	query.AssertDBMockDone(t, mock)
}

func TestWarehouseTransferDAO_UpdateTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	tx := db.Begin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "warehouse_transfers" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	transfers := NewWarehouseTransferDAO(db)
	transfer := &WarehouseTransfer{Base: model.Base{ID: 5}, State: TransferStateInTransit}
	updated, err := transfers.UpdateTx(ctx, tx, transfer)
	if err != nil {
		t.Fatalf("UpdateTx() error = %v", err)
	}
	if updated != transfer {
		t.Error("UpdateTx() returned a different transfer")
	}
	tx.Commit()
	query.AssertDBMockDone(t, mock)
}

func TestInboundCostDAO_CreateTxAndUpdateTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	tx := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "inbound_costs"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "inbound_costs" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	costs := NewInboundCostDAO(db)
	cost := &InboundCost{Name: "LC-1"}
	created, err := costs.CreateTx(ctx, tx, cost)
	if err != nil {
		t.Fatalf("CreateTx() error = %v", err)
	}
	if _, err := costs.UpdateTx(ctx, tx, created); err != nil {
		t.Fatalf("UpdateTx() error = %v", err)
	}
	tx.Commit()
	query.AssertDBMockDone(t, mock)
}

func TestInboundCostLineDAO_ListAndCreateTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "inbound_cost_lines" WHERE inbound_cost_id = $1`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "inbound_cost_lines" WHERE inbound_cost_id = $1`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 100))

	dao := NewInboundCostLineDAO(db)
	lines, err := dao.ListByInboundCost(ctx, 1)
	if err != nil {
		t.Fatalf("ListByInboundCost() error = %v", err)
	}
	if len(lines) != 1 || lines[0].ItemID != 100 {
		t.Errorf("ListByInboundCost() = %+v, want item 100", lines)
	}

	mock.ExpectBegin()
	tx := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "inbound_cost_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()
	line := &InboundCostLine{InboundCostID: 1, ItemID: 100}
	if _, err := dao.CreateTx(ctx, tx, line); err != nil {
		t.Fatalf("CreateTx() error = %v", err)
	}
	tx.Commit()
	query.AssertDBMockDone(t, mock)
}

func TestInboundCostAdjustmentDAO_ListAndCreateTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "inbound_cost_adjustments" WHERE inbound_cost_id = $1`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "inbound_cost_adjustments" WHERE inbound_cost_id = $1`)).
		WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(1, 100))

	dao := NewInboundCostAdjustmentDAO(db)
	adjustments, err := dao.ListByInboundCost(ctx, 1)
	if err != nil {
		t.Fatalf("ListByInboundCost() error = %v", err)
	}
	if len(adjustments) != 1 || adjustments[0].ItemID != 100 {
		t.Errorf("ListByInboundCost() = %+v, want item 100", adjustments)
	}

	mock.ExpectBegin()
	tx := db.Begin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "inbound_cost_adjustments"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()
	adjustment := &InboundCostAdjustment{InboundCostID: 1, ItemID: 100}
	if _, err := dao.CreateTx(ctx, tx, adjustment); err != nil {
		t.Fatalf("CreateTx() error = %v", err)
	}
	tx.Commit()
	query.AssertDBMockDone(t, mock)
}

func TestWarehouseDAO_And_StockLocationDAO_CRUD(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "warehouses" WHERE id = $1 ORDER BY "warehouses"."id" LIMIT $2`)).
		WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "WH"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "stock_locations" WHERE id = $1 ORDER BY "stock_locations"."id" LIMIT $2`)).
		WithArgs(1, 1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Loc"))

	warehouses := NewWarehouseDAO(db)
	warehouse, err := warehouses.Find(ctx, 1)
	if err != nil {
		t.Fatalf("WarehouseDAO.Find() error = %v", err)
	}
	if warehouse == nil || warehouse.Name != "WH" {
		t.Errorf("WarehouseDAO.Find() = %+v, want WH", warehouse)
	}

	locations := NewStockLocationDAO(db)
	location, err := locations.Find(ctx, 1)
	if err != nil {
		t.Fatalf("StockLocationDAO.Find() error = %v", err)
	}
	if location == nil || location.Name != "Loc" {
		t.Errorf("StockLocationDAO.Find() = %+v, want Loc", location)
	}
	query.AssertDBMockDone(t, mock)
}

func TestWarehouseDAO_And_StockLocationDAO_CreateAndUpdate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "warehouses"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "warehouses" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "stock_locations"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "stock_locations" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "warehouses" WHERE "warehouses"."id" = $1`)).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "stock_locations" WHERE "stock_locations"."id" = $1`)).
		WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	warehouses := NewWarehouseDAO(db)
	created, err := warehouses.Create(ctx, &reference.Warehouse{Name: "WH"})
	if err != nil {
		t.Fatalf("WarehouseDAO.Create() error = %v", err)
	}
	created.ID = 1
	if _, err := warehouses.Update(ctx, created); err != nil {
		t.Fatalf("WarehouseDAO.Update() error = %v", err)
	}

	locations := NewStockLocationDAO(db)
	locCreated, err := locations.Create(ctx, &reference.StockLocation{Name: "Loc"})
	if err != nil {
		t.Fatalf("StockLocationDAO.Create() error = %v", err)
	}
	locCreated.ID = 2
	if _, err := locations.Update(ctx, locCreated); err != nil {
		t.Fatalf("StockLocationDAO.Update() error = %v", err)
	}
	if err := warehouses.Delete(ctx, 1); err != nil {
		t.Fatalf("WarehouseDAO.Delete() error = %v", err)
	}
	if err := locations.Delete(ctx, 1); err != nil {
		t.Fatalf("StockLocationDAO.Delete() error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}
