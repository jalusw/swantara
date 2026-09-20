package dao

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type sampleEntity struct {
	ID   uint64
	Name string
}

func (sampleEntity) TableName() string {
	return "sample_entities"
}

func TestBase_List(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	columns := []string{"id", "name"}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "sample_entities"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_entities"`)).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, "alpha").AddRow(2, "beta"))

	repo := NewBase[sampleEntity](db)
	page, err := repo.List(ctx, nil)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if page.Count != 2 {
		t.Errorf("Count = %d, want 2", page.Count)
	}
	if len(page.Items) != 2 || page.Items[0].Name != "alpha" || page.Items[1].Name != "beta" {
		t.Errorf("Items = %+v, want [alpha beta]", page.Items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_ListAppliesQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	q := &query.Query{
		Pagination: &query.Pagination{Page: 2, Size: 10},
		Sorts:      []query.Sort{{Field: "name", Direction: query.Descending}},
		Filters:    []query.Filter{{Field: "name", Operator: query.Equal, Value: "alpha"}},
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "sample_entities" WHERE name = $1`)).
		WithArgs("alpha").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_entities" WHERE name = $1 ORDER BY name DESC LIMIT $2 OFFSET $3`)).
		WithArgs("alpha", 10, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "alpha"))

	repo := NewBase[sampleEntity](db)
	page, err := repo.List(ctx, q)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(page.Items) != 1 {
		t.Errorf("len(Items) = %d, want 1", len(page.Items))
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_ListCountError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "sample_entities"`)).
		WillReturnError(errors.New("count failed"))

	repo := NewBase[sampleEntity](db)
	if _, err := repo.List(ctx, nil); err == nil {
		t.Fatal("List() expected error")
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_ListQueryError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "sample_entities"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_entities"`)).
		WillReturnError(errors.New("find failed"))

	repo := NewBase[sampleEntity](db)
	if _, err := repo.List(ctx, nil); err == nil {
		t.Fatal("List() expected error")
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_Search(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_entities" WHERE name = $1 LIMIT $2`)).
		WithArgs("alpha", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "alpha"))

	repo := NewBase[sampleEntity](db)
	entity, err := repo.Search(ctx, "name", "alpha")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if entity == nil || entity.Name != "alpha" {
		t.Errorf("Search() = %+v, want alpha", entity)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_SearchNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_entities" WHERE name = $1 LIMIT $2`)).
		WithArgs("missing", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewBase[sampleEntity](db)
	entity, err := repo.Search(ctx, "name", "missing")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if entity != nil {
		t.Errorf("Search() = %+v, want nil", entity)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_SearchError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_entities" WHERE name = $1 LIMIT $2`)).
		WithArgs("alpha", 1).
		WillReturnError(errors.New("db down"))

	repo := NewBase[sampleEntity](db)
	if _, err := repo.Search(ctx, "name", "alpha"); err == nil {
		t.Fatal("Search() expected error")
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_Find(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_entities" WHERE id = $1 ORDER BY "sample_entities"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "alpha"))

	repo := NewBase[sampleEntity](db)
	entity, err := repo.Find(ctx, 1)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if entity == nil || entity.ID != 1 {
		t.Errorf("Find() = %+v, want id 1", entity)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_FindNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_entities" WHERE id = $1 ORDER BY "sample_entities"."id" LIMIT $2`)).
		WithArgs(99, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewBase[sampleEntity](db)
	entity, err := repo.Find(ctx, 99)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if entity != nil {
		t.Errorf("Find() = %+v, want nil", entity)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_FindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sample_entities" WHERE id = $1 ORDER BY "sample_entities"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))

	repo := NewBase[sampleEntity](db)
	if _, err := repo.Find(ctx, 1); err == nil {
		t.Fatal("Find() expected error")
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_Create(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sample_entities" ("name") VALUES ($1) RETURNING "id"`)).
		WithArgs("new").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	repo := NewBase[sampleEntity](db)
	entity := &sampleEntity{Name: "new"}
	created, err := repo.Create(ctx, entity)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created != entity {
		t.Errorf("Create() returned %+v, want %+v", created, entity)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_CreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sample_entities" ("name") VALUES ($1) RETURNING "id"`)).
		WithArgs("new").
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	repo := NewBase[sampleEntity](db)
	if _, err := repo.Create(ctx, &sampleEntity{Name: "new"}); err == nil {
		t.Fatal("Create() expected error")
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_Update(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "sample_entities" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewBase[sampleEntity](db)
	entity := &sampleEntity{ID: 1, Name: "updated"}
	updated, err := repo.Update(ctx, entity)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated != entity {
		t.Errorf("Update() returned %+v, want %+v", updated, entity)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_UpdateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "sample_entities" SET`)).
		WillReturnError(errors.New("update failed"))
	mock.ExpectRollback()

	repo := NewBase[sampleEntity](db)
	if _, err := repo.Update(ctx, &sampleEntity{ID: 1}); err == nil {
		t.Fatal("Update() expected error")
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_Delete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "sample_entities" WHERE "sample_entities"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewBase[sampleEntity](db)
	if err := repo.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_DeleteError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "sample_entities" WHERE "sample_entities"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	repo := NewBase[sampleEntity](db)
	if err := repo.Delete(ctx, 1); err == nil {
		t.Fatal("Delete() expected error")
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_HardDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "sample_entities" WHERE "sample_entities"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewBase[sampleEntity](db)
	if err := repo.HardDelete(ctx, 1); err != nil {
		t.Fatalf("HardDelete() error = %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBase_HardDeleteError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "sample_entities" WHERE "sample_entities"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	repo := NewBase[sampleEntity](db)
	if err := repo.HardDelete(ctx, 1); err == nil {
		t.Fatal("HardDelete() expected error")
	}

	query.AssertDBMockDone(t, mock)
}

func TestCRUDMockDefaults(t *testing.T) {
	ctx := context.Background()
	mock := CRUDMock[sampleEntity]{}

	page, err := mock.List(ctx, nil)
	if err != nil || page.Count != 0 || len(page.Items) != 0 {
		t.Errorf("List() = (%v, %v), want empty page", page, err)
	}
	if got, err := mock.Search(ctx, "name", "x"); err != nil || got != nil {
		t.Errorf("Search() = (%v, %v), want (nil, nil)", got, err)
	}
	if got, err := mock.Find(ctx, 1); err != nil || got != nil {
		t.Errorf("Find() = (%v, %v), want (nil, nil)", got, err)
	}
	entity := &sampleEntity{}
	if got, err := mock.Create(ctx, entity); err != nil || got != entity {
		t.Errorf("Create() = (%v, %v), want original entity", got, err)
	}
	if got, err := mock.Update(ctx, entity); err != nil || got != entity {
		t.Errorf("Update() = (%v, %v), want original entity", got, err)
	}
	if err := mock.Delete(ctx, 1); err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}
	if err := mock.HardDelete(ctx, 1); err != nil {
		t.Errorf("HardDelete() error = %v, want nil", err)
	}
}
