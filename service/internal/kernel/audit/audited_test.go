package audit

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestAudited_CreateRecordsInsertDiff(t *testing.T) {
	db, mock := query.NewMockDB(t)
	recorder := &recorderMock{}
	audited := NewAudited(dao.NewBase[saleOrder](db), recorder, "sale_orders")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sale_orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectCommit()

	created, err := audited.Create(context.Background(), &saleOrder{Code: "SO/00042", State: "draft", AmountTotal: "1000.00"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 7 {
		t.Fatalf("expected created id 7, got %d", created.ID)
	}

	if len(recorder.entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(recorder.entries))
	}
	entry := recorder.entries[0]
	if entry.EntityTable != "sale_orders" || entry.RecordID != 7 || entry.Action != ActionInsert {
		t.Errorf("unexpected entry: %#v", entry)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAudited_UpdateRecordsChangedDiff(t *testing.T) {
	db, mock := query.NewMockDB(t)
	recorder := &recorderMock{}
	audited := NewAudited(dao.NewBase[saleOrder](db), recorder, "sale_orders")

	before := []string{"id", "created_at", "updated_at", "deleted_at", "code", "state", "amount_total"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sale_orders" WHERE id = $1`)).
		WillReturnRows(sqlmock.NewRows(before).AddRow(9, nil, nil, nil, "SO/00009", "draft", "100.00"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "sale_orders" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	_, err := audited.Update(context.Background(), &saleOrder{Base: model.Base{ID: 9}, Code: "SO/00009", State: "confirmed", AmountTotal: "100.00"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recorder.entries) != 1 || recorder.entries[0].Action != ActionUpdate {
		t.Fatalf("expected update audit entry, got %#v", recorder.entries)
	}
	if recorder.entries[0].RecordID != 9 {
		t.Errorf("expected record id 9, got %d", recorder.entries[0].RecordID)
	}
	var diff map[string]any
	if err := json.Unmarshal(recorder.entries[0].Diff, &diff); err != nil {
		t.Fatalf("invalid diff: %v", err)
	}
	if diff["State"] != "confirmed" {
		t.Errorf("expected state change captured, got %#v", diff)
	}
	if _, ok := diff["Code"]; ok {
		t.Errorf("expected code not in diff, got %#v", diff)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAudited_DeleteRecordsDeleteDiff(t *testing.T) {
	db, mock := query.NewMockDB(t)
	recorder := &recorderMock{}
	audited := NewAudited(dao.NewBase[saleOrder](db), recorder, "sale_orders")

	before := []string{"id", "created_at", "updated_at", "deleted_at", "code", "state", "amount_total"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sale_orders" WHERE id = $1`)).
		WillReturnRows(sqlmock.NewRows(before).AddRow(3, nil, nil, nil, "SO/00003", "cancelled", "50.00"))

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "sale_orders" WHERE "sale_orders"."id" = $1`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := audited.Delete(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recorder.entries) != 1 || recorder.entries[0].Action != ActionDelete {
		t.Fatalf("expected delete audit entry, got %#v", recorder.entries)
	}
	entry := recorder.entries[0]
	if entry.RecordID != 3 {
		t.Errorf("expected record id 3, got %d", entry.RecordID)
	}
	var diff map[string]any
	if err := json.Unmarshal(entry.Diff, &diff); err != nil {
		t.Fatalf("invalid diff: %v", err)
	}
	if diff["Code"] != "SO/00003" {
		t.Errorf("expected full record diff, got %#v", diff)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAudited_UpdateRejectsPostedRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	recorder := &recorderMock{}
	audited := NewAudited(dao.NewBase[postedMove](db), recorder, "stock_movements")

	before := []string{"id", "created_at", "updated_at", "deleted_at", "state"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "posted_moves" WHERE id = $1`)).
		WillReturnRows(sqlmock.NewRows(before).AddRow(5, nil, nil, nil, "posted"))

	_, err := audited.Update(context.Background(), &postedMove{Base: model.Base{ID: 5}, State: "done"})

	if helper.AssertError(t, err, true, ErrImmutable) {
		return
	}
	if len(recorder.entries) != 0 {
		t.Error("expected no audit entry for rejected update")
	}
	query.AssertDBMockDone(t, mock)
}

func TestAudited_DeleteRejectsPostedRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	recorder := &recorderMock{}
	audited := NewAudited(dao.NewBase[postedMove](db), recorder, "stock_movements")

	before := []string{"id", "created_at", "updated_at", "deleted_at", "state"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "posted_moves" WHERE id = $1`)).
		WillReturnRows(sqlmock.NewRows(before).AddRow(5, nil, nil, nil, "posted"))

	err := audited.Delete(context.Background(), 5)

	if helper.AssertError(t, err, true, ErrImmutable) {
		return
	}
	if len(recorder.entries) != 0 {
		t.Error("expected no audit entry for rejected delete")
	}
	query.AssertDBMockDone(t, mock)
}

func TestAudited_UpdateMissingRecord(t *testing.T) {
	db, mock := query.NewMockDB(t)
	audited := NewAudited(dao.NewBase[saleOrder](db), &recorderMock{}, "sale_orders")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sale_orders" WHERE id = $1`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := audited.Update(context.Background(), &saleOrder{Base: model.Base{ID: 9}, Code: "SO/00009"})

	if helper.AssertError(t, err, true, ErrRecordNotFound) {
		return
	}
	query.AssertDBMockDone(t, mock)
}

func TestAudited_UpdateNonAuditable(t *testing.T) {
	db, _ := query.NewMockDB(t)
	audited := NewAudited(dao.NewBase[plainRecord](db), &recorderMock{}, "records")

	_, err := audited.Update(context.Background(), &plainRecord{Name: "x"})

	if helper.AssertError(t, err, true, ErrNotAuditable) {
		return
	}
}
