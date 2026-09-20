package asset

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestFixedAssetDAO_UpdateTx_UpdatesAsset(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fixed_assets" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	dao := NewFixedAssetDAO(db)
	asset := &FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Laptop", State: AssetStateRunning}

	updated, err := dao.UpdateTx(ctx, db, asset)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != asset {
		t.Fatalf("expected same asset back")
	}

	query.AssertDBMockDone(t, mock)
}

func TestFixedAssetDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fixed_assets" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	dao := NewFixedAssetDAO(db)
	asset := &FixedAsset{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Laptop", State: AssetStateRunning}

	_, err := dao.UpdateTx(ctx, db, asset)
	if err == nil || err.Error() != "db down" {
		t.Fatalf("error = %v, want db down", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAssetDepreciationLineDAO_ListByAsset_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "asset_depreciation_lines" WHERE asset_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "asset_depreciation_lines" WHERE asset_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "asset_id", "sequence", "amount"}).AddRow(1, 1, 1, 100))

	dao := NewAssetDepreciationLineDAO(db)
	lines, err := dao.ListByAsset(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 1 || lines[0].Sequence != 1 {
		t.Fatalf("lines = %+v, want one line with sequence 1", lines)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAssetDepreciationLineDAO_ListByAsset_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "asset_depreciation_lines" WHERE asset_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))

	dao := NewAssetDepreciationLineDAO(db)
	_, err := dao.ListByAsset(ctx, 1)
	if err == nil || err.Error() != "db down" {
		t.Fatalf("error = %v, want db down", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAssetDepreciationLineDAO_CreateTx_CreatesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "asset_depreciation_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	dao := NewAssetDepreciationLineDAO(db)
	line := &AssetDepreciationLine{AssetID: 1, Sequence: 1, Amount: 100}

	created, err := dao.CreateTx(ctx, db, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("created id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAssetDepreciationLineDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "asset_depreciation_lines"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	dao := NewAssetDepreciationLineDAO(db)
	_, err := dao.CreateTx(ctx, db, &AssetDepreciationLine{AssetID: 1})
	if err == nil || err.Error() != "db down" {
		t.Fatalf("error = %v, want db down", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAssetDepreciationLineDAO_UpdateTx_UpdatesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "asset_depreciation_lines" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	dao := NewAssetDepreciationLineDAO(db)
	line := &AssetDepreciationLine{Base: model.Base{ID: 1}, AssetID: 1, Sequence: 1, Amount: 100, Posted: true}

	updated, err := dao.UpdateTx(ctx, db, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != line {
		t.Fatalf("expected same line back")
	}

	query.AssertDBMockDone(t, mock)
}

func TestAssetDepreciationLineDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "asset_depreciation_lines" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	dao := NewAssetDepreciationLineDAO(db)
	_, err := dao.UpdateTx(ctx, db, &AssetDepreciationLine{Base: model.Base{ID: 1}, AssetID: 1})
	if err == nil || err.Error() != "db down" {
		t.Fatalf("error = %v, want db down", err)
	}

	query.AssertDBMockDone(t, mock)
}
